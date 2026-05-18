package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sangiagao/rice-marketplace/internal/google"
)

// GoogleIAPService mirrors AppleIAPService for Google Play subscriptions.
type GoogleIAPService struct {
	pool        *pgxpool.Pool
	client      *google.Client
	expectedPkg string
	engine      *CommissionEngine
}

func NewGoogleIAPService(pool *pgxpool.Pool, client *google.Client, packageName string) *GoogleIAPService {
	return &GoogleIAPService{pool: pool, client: client, expectedPkg: packageName}
}

func (s *GoogleIAPService) AttachCommissionEngine(e *CommissionEngine) { s.engine = e }

type GoogleVerifyResult struct {
	SubscriptionID string    `json:"subscription_id"`
	ProductID      string    `json:"product_id"`
	ExpiresAt      time.Time `json:"expires_at"`
	Months         int       `json:"months"`
	IsNewActivation bool     `json:"is_new_activation"`
}

type googleProductInfo struct {
	Months         int
	GrossAmount    int64
	PlatformFeePct float64
}

func (s *GoogleIAPService) lookupProduct(ctx context.Context, productID string) (googleProductInfo, error) {
	var p googleProductInfo
	err := s.pool.QueryRow(ctx,
		`SELECT months, gross_amount, platform_fee_pct
		   FROM google_product_map WHERE product_id = $1 AND is_active = true`,
		productID,
	).Scan(&p.Months, &p.GrossAmount, &p.PlatformFeePct)
	if err != nil {
		return googleProductInfo{}, fmt.Errorf("google_iap: unknown product %s", productID)
	}
	return p, nil
}

// VerifyPurchase is called by mobile after a successful Google Play purchase.
// Fetches authoritative state from Google + upserts subscription.
func (s *GoogleIAPService) VerifyPurchase(ctx context.Context, userID, productID, purchaseToken string) (*GoogleVerifyResult, error) {
	if userID == "" || productID == "" || purchaseToken == "" {
		return nil, errors.New("google_iap: missing required field")
	}
	prod, err := s.lookupProduct(ctx, productID)
	if err != nil {
		return nil, err
	}
	pur, err := s.client.GetSubscriptionPurchase(ctx, productID, purchaseToken)
	if err != nil {
		return nil, fmt.Errorf("google_iap: fetch purchase: %w", err)
	}
	if pur.ExpiresTime().IsZero() {
		return nil, errors.New("google_iap: missing expiryTimeMillis")
	}

	res, err := s.upsert(ctx, userID, productID, purchaseToken, pur, prod)
	if err != nil {
		return nil, err
	}

	// Acknowledge purchase (Google requires within 3 days, idempotent).
	if pur.AcknowledgementState == 0 {
		if ackErr := s.client.AcknowledgePurchase(ctx, productID, purchaseToken); ackErr != nil {
			slog.Warn("google_iap: acknowledge failed", "err", ackErr)
		}
	}

	// Restore listings + record commission
	_, _ = s.restoreListings(ctx, userID)
	s.recordCommission(ctx, userID, res.SubscriptionID, pur, productID, prod)
	return res, nil
}

func (s *GoogleIAPService) upsert(ctx context.Context, userID, productID, purchaseToken string, pur *google.SubscriptionPurchase, prod googleProductInfo) (*GoogleVerifyResult, error) {
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
		`SELECT id FROM subscriptions WHERE google_purchase_token = $1`,
		purchaseToken,
	).Scan(&existingID)

	expiresAt := pur.ExpiresTime()
	startedAt := pur.StartTime()
	if startedAt.IsZero() {
		startedAt = time.Now().UTC()
	}

	if err == nil {
		// Renewal — extend existing subscription
		_, err = tx.Exec(ctx,
			`UPDATE subscriptions
			    SET google_order_id    = $1,
			        google_product_id  = $2,
			        google_subscription_id = $2,
			        expires_at         = $3,
			        duration_months    = $4,
			        status             = 'active',
			        updated_at         = NOW()
			  WHERE id = $5`,
			pur.OrderID, productID, expiresAt, prod.Months, existingID)
		if err != nil {
			return nil, fmt.Errorf("google_iap: update sub: %w", err)
		}
	} else {
		// New subscription
		isNew = true
		err = tx.QueryRow(ctx,
			`INSERT INTO subscriptions (
				user_id, plan, started_at, expires_at, status,
				duration_months, amount, source,
				google_purchase_token, google_order_id,
				google_product_id, google_subscription_id
			 ) VALUES (
				$1, 'paid', $2, $3, 'active',
				$4, $5, 'google',
				$6, $7, $8, $8
			 ) RETURNING id`,
			userID, startedAt, expiresAt,
			prod.Months, prod.GrossAmount,
			purchaseToken, pur.OrderID, productID,
		).Scan(&existingID)
		if err != nil {
			return nil, fmt.Errorf("google_iap: insert sub: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &GoogleVerifyResult{
		SubscriptionID:  existingID,
		ProductID:       productID,
		ExpiresAt:       expiresAt,
		Months:          prod.Months,
		IsNewActivation: isNew,
	}, nil
}

func (s *GoogleIAPService) recordCommission(ctx context.Context, refereeID, subID string, pur *google.SubscriptionPurchase, productID string, prod googleProductInfo) {
	if s.engine == nil || prod.GrossAmount == 0 {
		return
	}
	netAmount := prod.GrossAmount - int64(float64(prod.GrossAmount)*prod.PlatformFeePct)
	if err := s.engine.UpdateSubscriptionNet(ctx, subID, netAmount, prod.PlatformFeePct); err != nil {
		log.Printf("google_iap: update net_amount: %v", err)
	}
	_, err := s.engine.RecordForPayment(ctx, PaymentEvent{
		RefereeUserID:  refereeID,
		SubscriptionID: subID,
		Source:         "google",
		EventID:        pur.OrderID,
		GrossAmount:    prod.GrossAmount,
		PlatformFeePct: prod.PlatformFeePct,
		OccurredAt:     pur.StartTime(),
	})
	if err != nil {
		log.Printf("google_iap: commission record: %v", err)
	}
}

func (s *GoogleIAPService) restoreListings(ctx context.Context, userID string) (int, error) {
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

// HandleNotification processes one decoded RTDN payload.
// Idempotent via (purchase_token, event_time_millis) unique index.
func (s *GoogleIAPService) HandleNotification(ctx context.Context, payload *google.RTDNPayload, rawJSON []byte) error {
	if payload == nil {
		return errors.New("google_iap: nil payload")
	}
	if payload.PackageName != "" && payload.PackageName != s.expectedPkg {
		return fmt.Errorf("google_iap: wrong package %s, expected %s", payload.PackageName, s.expectedPkg)
	}
	if payload.TestNotification != nil {
		slog.Info("google_iap: TEST notification received")
		return nil
	}
	if payload.SubscriptionNotification == nil {
		return nil // ignore one-time products etc.
	}
	sn := payload.SubscriptionNotification

	// Audit
	already, err := s.recordNotification(ctx, payload, rawJSON)
	if err != nil {
		return err
	}
	if already {
		slog.Info("google_iap: notification already processed", "token", sn.PurchaseToken)
		return nil
	}

	// State transition based on type
	switch sn.NotificationType {
	case google.NotifSubscriptionPurchased,
		google.NotifSubscriptionRenewed,
		google.NotifSubscriptionRecovered,
		google.NotifSubscriptionRestarted:
		if err := s.applyActive(ctx, sn.SubscriptionID, sn.PurchaseToken); err != nil {
			return s.markError(ctx, sn.PurchaseToken, payload.EventTimeMillis, err)
		}
	case google.NotifSubscriptionExpired,
		google.NotifSubscriptionCanceled,
		google.NotifSubscriptionOnHold:
		if err := s.applyExpired(ctx, sn.PurchaseToken); err != nil {
			return s.markError(ctx, sn.PurchaseToken, payload.EventTimeMillis, err)
		}
	case google.NotifSubscriptionRevoked: // refund
		if err := s.applyRevoked(ctx, sn.PurchaseToken); err != nil {
			return s.markError(ctx, sn.PurchaseToken, payload.EventTimeMillis, err)
		}
	default:
		slog.Info("google_iap: unhandled notification type", "type", sn.NotificationType)
	}

	return s.markProcessed(ctx, sn.PurchaseToken, payload.EventTimeMillis)
}

func (s *GoogleIAPService) recordNotification(ctx context.Context, p *google.RTDNPayload, rawJSON []byte) (bool, error) {
	sn := p.SubscriptionNotification
	cmd, err := s.pool.Exec(ctx,
		`INSERT INTO google_iap_notifications
		    (notification_type, subscription_id, purchase_token, package_name,
		     event_time_millis, raw_payload)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT (purchase_token, event_time_millis) DO NOTHING`,
		sn.NotificationType, sn.SubscriptionID, sn.PurchaseToken, p.PackageName,
		parseMillisStr(p.EventTimeMillis), json.RawMessage(rawJSON),
	)
	if err != nil {
		return false, err
	}
	return cmd.RowsAffected() == 0, nil
}

func (s *GoogleIAPService) markProcessed(ctx context.Context, token, eventMs string) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE google_iap_notifications SET processed = true, processed_at = NOW(), error = NULL
		  WHERE purchase_token = $1 AND event_time_millis = $2`, token, parseMillisStr(eventMs))
	return err
}

func (s *GoogleIAPService) markError(ctx context.Context, token, eventMs string, processErr error) error {
	_, _ = s.pool.Exec(ctx,
		`UPDATE google_iap_notifications SET processed = false, error = $1, processed_at = NOW()
		  WHERE purchase_token = $2 AND event_time_millis = $3`,
		processErr.Error(), token, parseMillisStr(eventMs))
	return processErr
}

func (s *GoogleIAPService) applyActive(ctx context.Context, productID, purchaseToken string) error {
	pur, err := s.client.GetSubscriptionPurchase(ctx, productID, purchaseToken)
	if err != nil {
		return err
	}
	prod, err := s.lookupProduct(ctx, productID)
	if err != nil {
		return err
	}
	// Look up user
	var userID, subID string
	row := s.pool.QueryRow(ctx,
		`SELECT user_id, id FROM subscriptions WHERE google_purchase_token = $1`,
		purchaseToken)
	if err := row.Scan(&userID, &subID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			slog.Info("google_iap: renewal before /verify, will reconcile on next mobile open",
				"token", purchaseToken)
			return nil
		}
		return err
	}
	if _, err := s.upsert(ctx, userID, productID, purchaseToken, pur, prod); err != nil {
		return err
	}
	_, _ = s.restoreListings(ctx, userID)
	s.recordCommission(ctx, userID, subID, pur, productID, prod)
	return nil
}

func (s *GoogleIAPService) applyExpired(ctx context.Context, purchaseToken string) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE subscriptions
		    SET status = 'expired', updated_at = NOW()
		  WHERE google_purchase_token = $1`, purchaseToken)
	return err
}

func (s *GoogleIAPService) applyRevoked(ctx context.Context, purchaseToken string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var userID, subID string
	err = tx.QueryRow(ctx,
		`UPDATE subscriptions SET status = 'expired', updated_at = NOW()
		  WHERE google_purchase_token = $1 RETURNING user_id, id`, purchaseToken).
		Scan(&userID, &subID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return err
	}
	_, err = tx.Exec(ctx,
		`UPDATE listings SET status = 'hidden', updated_at = NOW()
		  WHERE user_id = $1 AND status = 'active'`, userID)
	if err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}

	// Refund/revoke = clawback hoa hồng pending/payable (paid không rollback được).
	if s.engine != nil {
		if _, err := s.engine.CancelCommissionsForSubscription(ctx, subID); err != nil {
			slog.Warn("commission clawback failed on Google revoke", "sub_id", subID, "err", err)
		}
	}
	return nil
}

func parseMillisStr(s string) int64 {
	var ms int64
	_, _ = fmt.Sscanf(s, "%d", &ms)
	return ms
}
