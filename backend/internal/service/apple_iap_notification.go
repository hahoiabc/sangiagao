package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/sangiagao/rice-marketplace/internal/apple"
)

// HandleNotification ingests a verified V2 notification payload and updates
// our subscriptions table. Idempotent by notificationUUID.
func (s *AppleIAPService) HandleNotification(ctx context.Context, payload *apple.NotificationPayload, tx *apple.TransactionInfo, rawJSON []byte) error {
	if payload == nil {
		return errors.New("apple_iap: nil payload")
	}
	if !strings.EqualFold(payload.Data.BundleID, s.expectedBID) {
		return fmt.Errorf("%w: got %s want %s", ErrAppleBundleMismatch, payload.Data.BundleID, s.expectedBID)
	}

	// 1. Idempotency check + audit insert.
	already, err := s.recordNotification(ctx, payload, tx, rawJSON)
	if err != nil {
		return err
	}
	if already {
		slog.Info("apple_iap: notification already processed", "uuid", payload.NotificationUUID)
		return nil
	}

	// 2. Apply state change based on type.
	switch payload.NotificationType {
	case apple.NotifSubscribed, apple.NotifDidRenew:
		if tx != nil {
			if err := s.applyActiveTransaction(ctx, tx); err != nil {
				return s.markNotificationError(ctx, payload.NotificationUUID, err)
			}
		}
	case apple.NotifExpired, apple.NotifGracePeriod:
		if tx != nil {
			if err := s.applyExpired(ctx, tx); err != nil {
				return s.markNotificationError(ctx, payload.NotificationUUID, err)
			}
		}
	case apple.NotifRefund, apple.NotifRevoke:
		if tx != nil {
			if err := s.applyRevoked(ctx, tx); err != nil {
				return s.markNotificationError(ctx, payload.NotificationUUID, err)
			}
		}
	case apple.NotifBillingRetry:
		// Apple is retrying; don't expire yet. Just log.
		slog.Info("apple_iap: billing retry", "tx", txID(tx))
	case apple.NotifPriceIncrease, apple.NotifDidChangeStatus, apple.NotifConsumption:
		// Informational; nothing to apply server-side for now.
	case apple.NotifTest:
		slog.Info("apple_iap: TEST notification received", "uuid", payload.NotificationUUID)
	default:
		slog.Info("apple_iap: unhandled notification type", "type", payload.NotificationType)
	}

	return s.markNotificationProcessed(ctx, payload.NotificationUUID)
}

func (s *AppleIAPService) recordNotification(ctx context.Context, p *apple.NotificationPayload, t *apple.TransactionInfo, rawJSON []byte) (alreadyExists bool, err error) {
	var (
		txID, originalTxID, productID *string
		expiresAt                     *time.Time
	)
	if t != nil {
		txID = strPtr(t.TransactionID)
		originalTxID = strPtr(t.OriginalTransactionID)
		productID = strPtr(t.ProductID)
		if t.ExpiresDate > 0 {
			ts := t.ExpiresTime()
			expiresAt = &ts
		}
	}

	// Use ON CONFLICT to make the insert idempotent.
	cmdTag, err := s.pool.Exec(ctx,
		`INSERT INTO apple_notifications (
			notification_uuid, notification_type, subtype,
			bundle_id, environment,
			transaction_id, original_transaction_id, product_id, expires_date,
			raw_payload
		 ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		 ON CONFLICT (notification_uuid) DO NOTHING`,
		p.NotificationUUID, p.NotificationType, p.Subtype,
		p.Data.BundleID, p.Data.Environment,
		txID, originalTxID, productID, expiresAt,
		json.RawMessage(rawJSON),
	)
	if err != nil {
		return false, fmt.Errorf("apple_iap: insert audit: %w", err)
	}
	return cmdTag.RowsAffected() == 0, nil
}

func (s *AppleIAPService) markNotificationProcessed(ctx context.Context, uuid string) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE apple_notifications SET processed = true, processed_at = NOW(), error = NULL
		  WHERE notification_uuid = $1`,
		uuid,
	)
	return err
}

func (s *AppleIAPService) markNotificationError(ctx context.Context, uuid string, processErr error) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE apple_notifications SET processed = false, error = $2, processed_at = NOW()
		  WHERE notification_uuid = $1`,
		uuid, processErr.Error(),
	)
	if err != nil {
		slog.Error("apple_iap: mark error failed", "err", err)
	}
	return processErr
}

// applyActiveTransaction extends/upserts subscription with the latest expiresDate.
// We do NOT know userID from notification — must look up via existing record by
// originalTransactionId. If no match (e.g. renewal arrives before /verify call),
// we just record audit and let next /verify reconcile.
func (s *AppleIAPService) applyActiveTransaction(ctx context.Context, t *apple.TransactionInfo) error {
	expiresAt := t.ExpiresTime()

	cmdTag, err := s.pool.Exec(ctx,
		`UPDATE subscriptions
		    SET apple_transaction_id = $1,
		        apple_product_id = $2,
		        apple_environment = $3,
		        auto_renew_status = true,
		        expires_at = $4,
		        status = 'active',
		        updated_at = NOW()
		  WHERE apple_original_transaction_id = $5`,
		t.TransactionID, t.ProductID, t.Environment, expiresAt, t.OriginalTransactionID,
	)
	if err != nil {
		return fmt.Errorf("apple_iap: update active: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		// No matching subscription — likely renewal arrived before mobile /verify call.
		// Mobile will call /verify on next launch and reconcile.
		slog.Info("apple_iap: no matching subscription for original_tx", "original_tx", t.OriginalTransactionID)
	}

	// Look up user + subscription_id for downstream operations.
	var userID, subID string
	row := s.pool.QueryRow(ctx,
		`SELECT user_id, id FROM subscriptions WHERE apple_original_transaction_id = $1`,
		t.OriginalTransactionID,
	)
	if err := row.Scan(&userID, &subID); err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			slog.Warn("apple_iap: lookup user for renewal", "err", err)
		}
		return nil
	}

	_, _ = s.restoreListings(ctx, userID)

	// Record affiliate commission for renewal. Best-effort, errors logged.
	if s.engine != nil {
		prod, err := s.lookupProduct(ctx, t.ProductID)
		if err == nil && prod.GrossAmount > 0 {
			s.recordCommission(ctx, userID, subID, t, prod)
		} else if err != nil {
			slog.Info("apple_iap: skip commission on renewal (product unknown)", "product_id", t.ProductID)
		}
	}
	return nil
}

func (s *AppleIAPService) applyExpired(ctx context.Context, t *apple.TransactionInfo) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE subscriptions
		    SET status = 'expired', auto_renew_status = false, updated_at = NOW()
		  WHERE apple_original_transaction_id = $1`,
		t.OriginalTransactionID,
	)
	return err
}

func (s *AppleIAPService) applyRevoked(ctx context.Context, t *apple.TransactionInfo) error {
	// Mark expired AND hide listings (tougher than expire — refund/revoke means
	// retroactively no service).
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var userID, subID string
	err = tx.QueryRow(ctx,
		`UPDATE subscriptions
		    SET status = 'expired', auto_renew_status = false, updated_at = NOW()
		  WHERE apple_original_transaction_id = $1
		  RETURNING user_id, id`,
		t.OriginalTransactionID,
	).Scan(&userID, &subID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return err
	}

	_, err = tx.Exec(ctx,
		`UPDATE listings SET status = 'hidden', updated_at = NOW()
		  WHERE user_id = $1 AND status = 'active'`,
		userID,
	)
	if err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}

	// Refund/revoke = clawback hoa hồng pending/payable. Records đã paid không
	// rollback được (tiền đã chuyển ngân hàng). Chạy ngoài tx vì cross-table
	// và nếu fail cũng không nên rollback việc mark expired.
	if s.engine != nil {
		if _, err := s.engine.CancelCommissionsForSubscription(ctx, subID); err != nil {
			slog.Warn("commission clawback failed on Apple revoke", "sub_id", subID, "err", err)
		}
	}
	return nil
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func txID(t *apple.TransactionInfo) string {
	if t == nil {
		return ""
	}
	return t.TransactionID
}
