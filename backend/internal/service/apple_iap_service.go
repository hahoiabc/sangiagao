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
}

func NewAppleIAPService(pool *pgxpool.Pool, client *apple.Client, bundleID string) *AppleIAPService {
	return &AppleIAPService{
		pool:        pool,
		client:      client,
		expectedBID: bundleID,
	}
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

	months, err := s.lookupMonths(ctx, info.ProductID)
	if err != nil {
		return nil, err
	}
	if info.ExpiresDate <= 0 {
		return nil, errors.New("apple_iap: missing expiresDate")
	}

	expiresAt := info.ExpiresTime()
	res, err := s.upsert(ctx, userID, info, months, expiresAt)
	if err != nil {
		return nil, err
	}

	if _, err := s.restoreListings(ctx, userID); err != nil {
		log.Printf("apple_iap: restore listings for %s: %v", userID, err)
	}

	return res, nil
}

func (s *AppleIAPService) lookupMonths(ctx context.Context, productID string) (int, error) {
	var months int
	err := s.pool.QueryRow(ctx,
		`SELECT months FROM apple_product_map WHERE product_id = $1 AND is_active = true`,
		productID,
	).Scan(&months)
	if err != nil {
		return 0, fmt.Errorf("%w: %s", ErrAppleProductUnknown, productID)
	}
	return months, nil
}

func (s *AppleIAPService) upsert(ctx context.Context, userID string, info *apple.TransactionInfo, months int, expiresAt time.Time) (*AppleVerifyResult, error) {
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
				$4, 0, 'apple',
				$5, $6, $7, $8, true
			 ) RETURNING id`,
			userID, info.PurchaseTime(), expiresAt,
			months,
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
