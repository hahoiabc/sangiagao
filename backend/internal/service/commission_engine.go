package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sangiagao/rice-marketplace/internal/model"
	"github.com/sangiagao/rice-marketplace/internal/repository"
)

// PayableDelayDays is the T+N delay before pending → payable. Hedge against
// platform refunds (Apple, ZaloPay) since we explicitly do NOT clawback.
const PayableDelayDays = 45

// CommissionEngine attaches commission records to payment events. Idempotent
// by (payment_source, payment_event_id). Pure calculation is split out for
// unit testing.
type CommissionEngine struct {
	pool    *pgxpool.Pool
	affRepo *repository.AffiliateRepo
}

func NewCommissionEngine(pool *pgxpool.Pool, affRepo *repository.AffiliateRepo) *CommissionEngine {
	return &CommissionEngine{pool: pool, affRepo: affRepo}
}

// PaymentEvent captures the inputs the engine needs from any source (Apple
// webhook, SePay webhook, admin grant). All amounts in VND minor unit (int64
// VND — VND has no fractional unit so this is just VND).
type PaymentEvent struct {
	RefereeUserID  string
	SubscriptionID string
	Source         string // apple | sepay | admin
	EventID        string // apple_transaction_id or sepay_order_id (unique per source)
	GrossAmount    int64
	PlatformFeePct float64 // 0.30 Apple, 0 SePay
	OccurredAt     time.Time
}

// CommissionCalc is the pure-function result of stage + rate + amount given
// rule + payment sequence. Exported for unit tests.
type CommissionCalc struct {
	Stage            int
	Rate             float64
	BaseAmount       int64
	NetAmount        int64
	PlatformFee      int64
	CommissionAmount int64
	PaymentSequence  int // 1, 2, 3...
}

// Calculate is pure: given rule + payment sequence + gross, return what should
// be recorded. Stage is the payment sequence capped at 3:
//   - payment #1 → stage 1 (45% default)
//   - payment #2 → stage 2 (30% default)
//   - payment #3+ → stage 3 (15% default, perpetual)
//
// Đổi từ time-based sang payment-count-based 2026-05-19 — đơn giản hơn, đẩy
// gói dài, công bằng giữa monthly và yearly subscribers.
func Calculate(rule *model.CommissionRule, gross int64, platformFeePct float64, paymentSequence int) CommissionCalc {
	if paymentSequence < 1 {
		paymentSequence = 1
	}
	var stage int
	var rate float64
	switch paymentSequence {
	case 1:
		stage, rate = 1, rule.Stage1Pct
	case 2:
		stage, rate = 2, rule.Stage2Pct
	default:
		stage, rate = 3, rule.Stage3Pct
	}

	// Fee + net
	platformFee := int64(math.Round(float64(gross) * platformFeePct))
	net := gross - platformFee

	// Base
	var base int64
	if rule.BaseType == "gross" {
		base = gross
	} else {
		base = net
	}
	commission := int64(math.Round(float64(base) * rate))

	return CommissionCalc{
		Stage:            stage,
		Rate:             rate,
		BaseAmount:       base,
		NetAmount:        net,
		PlatformFee:      platformFee,
		CommissionAmount: commission,
		PaymentSequence:  paymentSequence,
	}
}

// RecordForPayment is the main entry point called from Apple/SePay/admin
// payment handlers. Idempotent: returns nil without error if event already
// recorded. Returns nil error + no record if referee has no referrer.
func (e *CommissionEngine) RecordForPayment(ctx context.Context, ev PaymentEvent) (*model.CommissionRecord, error) {
	if ev.RefereeUserID == "" || ev.SubscriptionID == "" || ev.EventID == "" {
		return nil, errors.New("commission: missing required fields")
	}
	if ev.GrossAmount <= 0 {
		slog.Info("commission: skip zero-amount event", "event_id", ev.EventID)
		return nil, nil
	}

	// Look up referrer
	var referrerUserID *string
	err := e.pool.QueryRow(ctx,
		`SELECT referrer_user_id FROM users WHERE id = $1`, ev.RefereeUserID).
		Scan(&referrerUserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("commission: referee %s not found", ev.RefereeUserID)
		}
		return nil, err
	}
	if referrerUserID == nil {
		// Referee was not referred — no commission, not an error
		return nil, nil
	}

	// Self-referral guard (defense-in-depth, attribution layer should also check)
	if *referrerUserID == ev.RefereeUserID {
		slog.Warn("commission: self-referral blocked", "user_id", ev.RefereeUserID)
		return nil, nil
	}

	// Đếm payment sequence: số commission records hợp lệ (not cancelled) đã
	// có sẵn cho cặp (referrer, referee). +1 cho lần này.
	// Stage 1 = payment #1, Stage 2 = payment #2, Stage 3 = payment #3+
	var existingCount int
	err = e.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM commission_records
		  WHERE referrer_user_id = $1 AND referee_user_id = $2
		    AND status != 'cancelled'`,
		*referrerUserID, ev.RefereeUserID).Scan(&existingCount)
	if err != nil {
		return nil, fmt.Errorf("commission: count existing: %w", err)
	}
	paymentSequence := existingCount + 1

	// Phương án C (2026-05-18): bỏ guard "aff-only". Mọi role có code đều earn
	// (owner/admin/editor/aff). Vector gian lận tự bơm tiền KHÔNG có lời vì
	// commission < amount paid; vector duy nhất khả thi là Apple/Google refund
	// abuse — xử lý bằng CancelCommissionsForSubscription khi nhận REFUND
	// webhook (clawback pending/payable records).

	// Get referrer's referral_code_id (for per-partner rule lookup)
	var referralCodeID *string
	err = e.pool.QueryRow(ctx,
		`SELECT id FROM referral_codes WHERE user_id = $1`, *referrerUserID).Scan(&referralCodeID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	// If referrer has no code (edge case — attributed before code generation),
	// referralCodeID stays nil → default rule applies.

	rule, err := e.affRepo.GetActiveRule(ctx, referralCodeID)
	if err != nil {
		return nil, fmt.Errorf("commission: load rule: %w", err)
	}

	calc := Calculate(rule, ev.GrossAmount, ev.PlatformFeePct, paymentSequence)

	rec := &model.CommissionRecord{
		ReferrerUserID:   *referrerUserID,
		RefereeUserID:    ev.RefereeUserID,
		SubscriptionID:   ev.SubscriptionID,
		PaymentSource:    ev.Source,
		PaymentEventID:   ev.EventID,
		GrossAmount:      ev.GrossAmount,
		PlatformFee:      calc.PlatformFee,
		NetAmount:        calc.NetAmount,
		BaseAmount:       calc.BaseAmount,
		Stage:            calc.Stage,
		Rate:             calc.Rate,
		CommissionAmount: calc.CommissionAmount,
		PaymentSequence:  calc.PaymentSequence,
		RuleID:           rule.ID,
		PayableAfter:     ev.OccurredAt.Add(PayableDelayDays * 24 * time.Hour),
	}

	if err := e.affRepo.InsertRecord(ctx, rec); err != nil {
		if errors.Is(err, repository.ErrCommissionRecordExists) {
			slog.Info("commission: duplicate event, ignored", "event_id", ev.EventID)
			return nil, nil
		}
		return nil, fmt.Errorf("commission: insert: %w", err)
	}

	slog.Info("commission: recorded",
		"referrer", *referrerUserID,
		"referee", ev.RefereeUserID,
		"stage", calc.Stage,
		"rate", calc.Rate,
		"commission", calc.CommissionAmount,
		"source", ev.Source)

	return rec, nil
}

// UpdateSubscriptionNet writes net_amount + platform_fee_pct onto the
// subscriptions row. Called alongside RecordForPayment so admin reports can
// show net revenue.
func (e *CommissionEngine) UpdateSubscriptionNet(ctx context.Context, subscriptionID string, netAmount int64, platformFeePct float64) error {
	_, err := e.pool.Exec(ctx,
		`UPDATE subscriptions
		    SET net_amount = $1, platform_fee_pct = $2, updated_at = NOW()
		  WHERE id = $3`, netAmount, platformFeePct, subscriptionID)
	return err
}

// CancelCommissionsForSubscription — clawback hoa hồng cho 1 subscription bị
// refund / revoke. Chỉ cancel records ở trạng thái pending/payable (chưa rút).
// Records đã paid giữ nguyên (không thể clawback tiền đã chuyển ngân hàng).
//
// Gọi từ Apple/Google REFUND webhook handler. Trả về số records cancel để
// log + admin báo cáo.
func (e *CommissionEngine) CancelCommissionsForSubscription(ctx context.Context, subscriptionID string) (int64, error) {
	tag, err := e.pool.Exec(ctx,
		`UPDATE commissions
		    SET status = 'cancelled', updated_at = NOW()
		  WHERE subscription_id = $1
		    AND status IN ('pending', 'payable')`,
		subscriptionID,
	)
	if err != nil {
		return 0, err
	}
	count := tag.RowsAffected()
	if count > 0 {
		slog.Info("commission: cancelled records for refunded subscription",
			"subscription_id", subscriptionID, "cancelled_count", count)
	}
	return count, nil
}
