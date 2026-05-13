package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sangiagao/rice-marketplace/internal/apple"
)

// AppleIAPService verifies Apple StoreKit transactions and syncs them with our
// subscriptions table. Isolated from SubscriptionService to keep existing tests
// untouched.
type AppleIAPService struct {
	pool        *pgxpool.Pool
	client      *apple.Client
	expectedBID string // bundleId we expect transactions to come from
	engine      *CommissionEngine
}

func NewAppleIAPService(pool *pgxpool.Pool, client *apple.Client, bundleID string) *AppleIAPService {
	return &AppleIAPService{
		pool:        pool,
		client:      client,
		expectedBID: bundleID,
	}
}

// AttachCommissionEngine wires the affiliate commission engine. Optional —
// if nil, no commission is recorded. Set in main.go after engine construction.
func (s *AppleIAPService) AttachCommissionEngine(e *CommissionEngine) {
	s.engine = e
}

// productInfo is the pricing snapshot used to compute net_amount + commission.
type productInfo struct {
	Months         int
	GrossAmount    int64
	PlatformFeePct float64
}

// AppleVerifyResult is returned to mobile after successful verification.
type AppleVerifyResult struct {
	SubscriptionID  string    `json:"subscription_id"`
	ProductID       string    `json:"product_id"`
	ExpiresAt       time.Time `json:"expires_at"`
	Environment     string    `json:"environment"`
	Months          int       `json:"months"`
	IsNewActivation bool      `json:"is_new_activation"`
}

var (
	ErrAppleBundleMismatch     = errors.New("transaction bundle does not match expected app")
	ErrAppleProductUnknown     = errors.New("product is not registered for this app")
	ErrAppleTransactionRevoked = errors.New("transaction has been revoked")
)

// VerifyTransaction looks up an Apple transaction by ID and upserts a subscription
// record for the given user. Handles both first purchase and renewal transactions
// (Apple uses originalTransactionId as the stable subscription identifier).
func (s *AppleIAPService) VerifyTransaction(ctx context.Context, userID, transactionID string) (*AppleVerifyResult, error) {
	if userID == "" || transactionID == "" {
		return nil, errors.New("apple_iap: userID and transactionID required")
	}

	info, err := s.client.GetTransactionInfo(ctx, transactionID)
	if err != nil {
		return nil, fmt.Errorf("apple_iap: fetch transaction: %w", err)
	}

	if !strings.EqualFold(info.BundleID, s.expectedBID) {
		return nil, fmt.Errorf("%w: got %s want %s", ErrAppleBundleMismatch, info.BundleID, s.expectedBID)
	}
	if info.RevocationDate > 0 {
		return nil, ErrAppleTransactionRevoked
	}

	prod, err := s.lookupProduct(ctx, info.ProductID)
	if err != nil {
		return nil, err
	}
	if info.ExpiresDate <= 0 {
		return nil, errors.New("apple_iap: missing expiresDate")
	}

	expiresAt := info.ExpiresTime()
	res, err := s.upsert(ctx, userID, info, prod, expiresAt)
	if err != nil {
		return nil, err
	}

	if _, err := s.restoreListings(ctx, userID); err != nil {
		log.Printf("apple_iap: restore listings for %s: %v", userID, err)
	}

	// Commission engine: record affiliate commission if referee has a referrer.
	// Best-effort — failures must not block the verify response.
	s.recordCommission(ctx, userID, res.SubscriptionID, info, prod)

	return res, nil
}

// recordCommission is a best-effort hook to write a commission_record + update
// subscription net_amount. Errors are logged, not returned.
func (s *AppleIAPService) recordCommission(ctx context.Context, refereeID, subID string, info *apple.TransactionInfo, prod productInfo) {
	if s.engine == nil || prod.GrossAmount == 0 {
		return
	}
	// Update subscriptions.net_amount + platform_fee_pct for reporting.
	netAmount := prod.GrossAmount - int64(float64(prod.GrossAmount)*prod.PlatformFeePct)
	if err := s.engine.UpdateSubscriptionNet(ctx, subID, netAmount, prod.PlatformFeePct); err != nil {
		log.Printf("apple_iap: update net_amount: %v", err)
	}

	_, err := s.engine.RecordForPayment(ctx, PaymentEvent{
		RefereeUserID:  refereeID,
		SubscriptionID: subID,
		Source:         "apple",
		EventID:        info.TransactionID,
		GrossAmount:    prod.GrossAmount,
		PlatformFeePct: prod.PlatformFeePct,
		OccurredAt:     info.PurchaseTime(),
	})
	if err != nil {
		log.Printf("apple_iap: commission record: %v", err)
	}
}

func (s *AppleIAPService) lookupProduct(ctx context.Context, productID string) (productInfo, error) {
	var p productInfo
	err := s.pool.QueryRow(ctx,
		`SELECT months, gross_amount, platform_fee_pct
		   FROM apple_product_map WHERE product_id = $1 AND is_active = true`,
		productID,
	).Scan(&p.Months, &p.GrossAmount, &p.PlatformFeePct)
	if err != nil {
		return productInfo{}, fmt.Errorf("%w: %s", ErrAppleProductUnknown, productID)
	}
	return p, nil
}

func (s *AppleIAPService) upsert(ctx context.Context, userID string, info *apple.TransactionInfo, prod productInfo, expiresAt time.Time) (*AppleVerifyResult, error) {
	months := prod.Months
	grossAmount := prod.GrossAmount
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var (
		existingID string
		isNew      bool
	)
	err = tx.QueryRow(ctx,
		`SELECT id FROM subscriptions WHERE apple_original_transaction_id = $1`,
		info.OriginalTransactionID,
	).Scan(&existingID)

	if err == nil {
		// Renewal — extend existing subscription.
		_, err := tx.Exec(ctx,
			`UPDATE subscriptions
			   SET apple_transaction_id = $1,
			       apple_product_id = $2,
			       apple_environment = $3,
			       auto_renew_status = true,
			       expires_at = $4,
			       duration_months = $5,
			       status = 'active',
			       updated_at = NOW()
			 WHERE id = $6`,
			info.TransactionID, info.ProductID, info.Environment,
			expiresAt, months, existingID,
		)
		if err != nil {
			return nil, fmt.Errorf("apple_iap: update sub: %w", err)
		}
	} else {
		// New subscription record.
		isNew = true
		err = tx.QueryRow(ctx,
			`INSERT INTO subscriptions (
				user_id, plan, started_at, expires_at, status,
				duration_months, amount, source,
				apple_transaction_id, apple_original_transaction_id,
				apple_product_id, apple_environment, auto_renew_status
			 ) VALUES (
				$1, 'paid', $2, $3, 'active',
				$4, $5, 'apple',
				$6, $7, $8, $9, true
			 ) RETURNING id`,
			userID, info.PurchaseTime(), expiresAt,
			months, grossAmount,
			info.TransactionID, info.OriginalTransactionID,
			info.ProductID, info.Environment,
		).Scan(&existingID)
		if err != nil {
			return nil, fmt.Errorf("apple_iap: insert sub: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &AppleVerifyResult{
		SubscriptionID:  existingID,
		ProductID:       info.ProductID,
		ExpiresAt:       expiresAt,
		Environment:     info.Environment,
		Months:          months,
		IsNewActivation: isNew,
	}, nil
}

func (s *AppleIAPService) restoreListings(ctx context.Context, userID string) (int, error) {
	tag, err := s.pool.Exec(ctx,
		`UPDATE listings SET status = 'active', updated_at = NOW()
		  WHERE user_id = $1 AND status = 'hidden'`,
		userID,
	)
	if err != nil {
		return 0, err
	}
	return int(tag.RowsAffected()), nil
}
