package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sangiagao/rice-marketplace/internal/model"
)

var ErrReferralCodeNotFound = errors.New("referral code not found")
var ErrCommissionRuleNotFound = errors.New("commission rule not found")
var ErrCommissionRecordExists = errors.New("commission record already exists for this payment event")

type AffiliateRepo struct {
	pool *pgxpool.Pool
}

func NewAffiliateRepo(pool *pgxpool.Pool) *AffiliateRepo {
	return &AffiliateRepo{pool: pool}
}

// Pool exposes the underlying pgxpool for handlers that need ad-hoc SQL
// (admin leaderboard, payable preview). Use sparingly.
func (r *AffiliateRepo) Pool() *pgxpool.Pool { return r.pool }

// --- referral_codes ---

func (r *AffiliateRepo) GetCodeByUser(ctx context.Context, userID string) (*model.ReferralCode, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, user_id, code, active, created_at
		   FROM referral_codes WHERE user_id = $1`, userID)
	var c model.ReferralCode
	if err := row.Scan(&c.ID, &c.UserID, &c.Code, &c.Active, &c.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrReferralCodeNotFound
		}
		return nil, err
	}
	return &c, nil
}

func (r *AffiliateRepo) GetCodeByCode(ctx context.Context, code string) (*model.ReferralCode, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, user_id, code, active, created_at
		   FROM referral_codes WHERE code = $1 AND active = TRUE`, code)
	var c model.ReferralCode
	if err := row.Scan(&c.ID, &c.UserID, &c.Code, &c.Active, &c.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrReferralCodeNotFound
		}
		return nil, err
	}
	return &c, nil
}

func (r *AffiliateRepo) CreateCode(ctx context.Context, userID, code string) (*model.ReferralCode, error) {
	row := r.pool.QueryRow(ctx,
		`INSERT INTO referral_codes (user_id, code) VALUES ($1, $2)
		 RETURNING id, user_id, code, active, created_at`, userID, code)
	var c model.ReferralCode
	if err := row.Scan(&c.ID, &c.UserID, &c.Code, &c.Active, &c.CreatedAt); err != nil {
		return nil, err
	}
	return &c, nil
}

// --- commission_rules ---

// GetActiveRule returns the rule for a given referral_code_id, falling back to
// the default rule (referral_code_id IS NULL) if no override exists.
func (r *AffiliateRepo) GetActiveRule(ctx context.Context, referralCodeID *string) (*model.CommissionRule, error) {
	// Try per-partner override first
	if referralCodeID != nil {
		row := r.pool.QueryRow(ctx,
			`SELECT id, referral_code_id, stage1_days, stage1_pct, stage2_days, stage2_pct,
			        stage3_pct, base_type, minimum_payout, active_from, active_to, created_at, updated_at
			   FROM commission_rules
			  WHERE referral_code_id = $1 AND active_to IS NULL
			  LIMIT 1`, *referralCodeID)
		rule, err := scanRule(row)
		if err == nil {
			return rule, nil
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}
	}
	// Fallback to default
	row := r.pool.QueryRow(ctx,
		`SELECT id, referral_code_id, stage1_days, stage1_pct, stage2_days, stage2_pct,
		        stage3_pct, base_type, minimum_payout, active_from, active_to, created_at, updated_at
		   FROM commission_rules
		  WHERE referral_code_id IS NULL AND active_to IS NULL
		  LIMIT 1`)
	rule, err := scanRule(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCommissionRuleNotFound
		}
		return nil, err
	}
	return rule, nil
}

func (r *AffiliateRepo) ListRules(ctx context.Context) ([]*model.CommissionRule, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, referral_code_id, stage1_days, stage1_pct, stage2_days, stage2_pct,
		        stage3_pct, base_type, minimum_payout, active_from, active_to, created_at, updated_at
		   FROM commission_rules
		  ORDER BY referral_code_id NULLS FIRST, created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.CommissionRule
	for rows.Next() {
		rule, err := scanRule(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, rule)
	}
	return out, rows.Err()
}

func (r *AffiliateRepo) UpsertRule(ctx context.Context, rule *model.CommissionRule) error {
	// Close existing active rule for same referral_code_id, then insert new
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	now := time.Now().UTC()
	if rule.ReferralCodeID != nil {
		_, err = tx.Exec(ctx,
			`UPDATE commission_rules SET active_to = $1, updated_at = $1
			   WHERE referral_code_id = $2 AND active_to IS NULL`, now, *rule.ReferralCodeID)
	} else {
		_, err = tx.Exec(ctx,
			`UPDATE commission_rules SET active_to = $1, updated_at = $1
			   WHERE referral_code_id IS NULL AND active_to IS NULL`, now)
	}
	if err != nil {
		return err
	}

	row := tx.QueryRow(ctx,
		`INSERT INTO commission_rules
		    (referral_code_id, stage1_days, stage1_pct, stage2_days, stage2_pct,
		     stage3_pct, base_type, minimum_payout, active_from)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 RETURNING id, created_at, updated_at`,
		rule.ReferralCodeID, rule.Stage1Days, rule.Stage1Pct, rule.Stage2Days, rule.Stage2Pct,
		rule.Stage3Pct, rule.BaseType, rule.MinimumPayout, now)
	if err := row.Scan(&rule.ID, &rule.CreatedAt, &rule.UpdatedAt); err != nil {
		return err
	}
	rule.ActiveFrom = now
	return tx.Commit(ctx)
}

// --- commission_records ---

// InsertRecord inserts a commission_record idempotently. Returns
// ErrCommissionRecordExists if (payment_source, payment_event_id) already
// exists.
func (r *AffiliateRepo) InsertRecord(ctx context.Context, rec *model.CommissionRecord) error {
	row := r.pool.QueryRow(ctx,
		`INSERT INTO commission_records
		    (referrer_user_id, referee_user_id, subscription_id, payment_source, payment_event_id,
		     gross_amount, platform_fee, net_amount, base_amount, stage, rate, commission_amount,
		     referee_age_days, rule_id, payable_after)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		 ON CONFLICT (payment_source, payment_event_id) DO NOTHING
		 RETURNING id, status, created_at`,
		rec.ReferrerUserID, rec.RefereeUserID, rec.SubscriptionID, rec.PaymentSource, rec.PaymentEventID,
		rec.GrossAmount, rec.PlatformFee, rec.NetAmount, rec.BaseAmount, rec.Stage, rec.Rate,
		rec.CommissionAmount, rec.RefereeAgeDays, rec.RuleID, rec.PayableAfter)
	if err := row.Scan(&rec.ID, &rec.Status, &rec.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrCommissionRecordExists
		}
		return err
	}
	return nil
}

// StatsForReferrer aggregates commission totals for a single referrer.
func (r *AffiliateRepo) StatsForReferrer(ctx context.Context, userID string) (*model.ReferralStats, error) {
	stats := &model.ReferralStats{}
	row := r.pool.QueryRow(ctx,
		`SELECT
		    COUNT(DISTINCT referee_user_id) AS total_referrals,
		    COALESCE(SUM(commission_amount), 0) AS total_earned,
		    COALESCE(SUM(CASE WHEN status='payable' THEN commission_amount END), 0) AS payable,
		    COALESCE(SUM(CASE WHEN status='pending' THEN commission_amount END), 0) AS pending,
		    COALESCE(SUM(CASE WHEN status='paid' THEN commission_amount END), 0) AS paid
		   FROM commission_records WHERE referrer_user_id = $1`, userID)
	if err := row.Scan(&stats.TotalReferrals, &stats.TotalEarned,
		&stats.PayableAmount, &stats.PendingAmount, &stats.PaidAmount); err != nil {
		return nil, err
	}
	// Active referees: count distinct referees with active subscription
	row = r.pool.QueryRow(ctx,
		`SELECT COUNT(DISTINCT u.id) FROM users u
		   JOIN subscriptions s ON s.user_id = u.id AND s.status = 'active'
		  WHERE u.referrer_user_id = $1`, userID)
	if err := row.Scan(&stats.ActiveReferees); err != nil {
		return nil, err
	}
	return stats, nil
}

// ListRecordsForReferrer returns paginated commission records.
func (r *AffiliateRepo) ListRecordsForReferrer(ctx context.Context, userID string, limit, offset int) ([]*model.CommissionRecord, error) {
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	rows, err := r.pool.Query(ctx,
		`SELECT id, referrer_user_id, referee_user_id, subscription_id,
		        payment_source, payment_event_id, gross_amount, platform_fee, net_amount,
		        base_amount, stage, rate, commission_amount, referee_age_days, rule_id,
		        status, payable_after, paid_at, payout_id, created_at
		   FROM commission_records
		  WHERE referrer_user_id = $1
		  ORDER BY created_at DESC
		  LIMIT $2 OFFSET $3`, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.CommissionRecord
	for rows.Next() {
		var rec model.CommissionRecord
		if err := rows.Scan(
			&rec.ID, &rec.ReferrerUserID, &rec.RefereeUserID, &rec.SubscriptionID,
			&rec.PaymentSource, &rec.PaymentEventID, &rec.GrossAmount, &rec.PlatformFee, &rec.NetAmount,
			&rec.BaseAmount, &rec.Stage, &rec.Rate, &rec.CommissionAmount, &rec.RefereeAgeDays, &rec.RuleID,
			&rec.Status, &rec.PayableAfter, &rec.PaidAt, &rec.PayoutID, &rec.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, &rec)
	}
	return out, rows.Err()
}

// PromotePayableRecords moves status pending → payable for records past payable_after.
// Returns count of rows updated. Called by daily cron.
func (r *AffiliateRepo) PromotePayableRecords(ctx context.Context) (int64, error) {
	ct, err := r.pool.Exec(ctx,
		`UPDATE commission_records
		    SET status = 'payable'
		  WHERE status = 'pending' AND payable_after <= NOW()`)
	if err != nil {
		return 0, err
	}
	return ct.RowsAffected(), nil
}

// --- payouts ---

// CreatePayout in a transaction: insert payout row, mark records paid, return.
func (r *AffiliateRepo) CreatePayout(ctx context.Context, p *model.Payout, recordIDs []string) error {
	if len(recordIDs) == 0 {
		return errors.New("no records to pay out")
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	now := time.Now().UTC()
	row := tx.QueryRow(ctx,
		`INSERT INTO payouts
		    (referrer_user_id, total_amount, record_count, method, bank_info, note, status, created_by)
		 VALUES ($1, $2, $3, $4, $5, $6, 'pending', $7)
		 RETURNING id, created_at`,
		p.ReferrerUserID, p.TotalAmount, p.RecordCount, p.Method, p.BankInfo, p.Note, p.CreatedBy)
	if err := row.Scan(&p.ID, &p.CreatedAt); err != nil {
		return err
	}

	// Mark each record as paid + link to this payout. Defensive: only update if
	// still payable + belongs to referrer (prevents accidental double-payout).
	ct, err := tx.Exec(ctx,
		`UPDATE commission_records
		    SET status = 'paid', paid_at = $1, payout_id = $2
		  WHERE id = ANY($3::uuid[]) AND status = 'payable' AND referrer_user_id = $4`,
		now, p.ID, recordIDs, p.ReferrerUserID)
	if err != nil {
		return err
	}
	if int(ct.RowsAffected()) != len(recordIDs) {
		return errors.New("some records were not in payable state or did not belong to referrer")
	}

	p.Status = "pending"
	return tx.Commit(ctx)
}

func (r *AffiliateRepo) ListPayouts(ctx context.Context, referrerUserID *string, limit, offset int) ([]*model.Payout, error) {
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	args := []any{limit, offset}
	where := ""
	if referrerUserID != nil {
		where = "WHERE referrer_user_id = $3"
		args = append(args, *referrerUserID)
	}
	q := `SELECT id, referrer_user_id, total_amount, record_count, method, bank_info, note,
	             status, created_by, sent_at, created_at
	        FROM payouts ` + where + `
	       ORDER BY created_at DESC
	       LIMIT $1 OFFSET $2`
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.Payout
	for rows.Next() {
		var p model.Payout
		if err := rows.Scan(
			&p.ID, &p.ReferrerUserID, &p.TotalAmount, &p.RecordCount, &p.Method, &p.BankInfo, &p.Note,
			&p.Status, &p.CreatedBy, &p.SentAt, &p.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, &p)
	}
	return out, rows.Err()
}

func (r *AffiliateRepo) MarkPayoutSent(ctx context.Context, payoutID string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE payouts SET status = 'sent', sent_at = NOW() WHERE id = $1 AND status = 'pending'`, payoutID)
	return err
}

// scanRule reads a CommissionRule from a pgx row/rows scanner.
func scanRule(scanner interface{ Scan(...any) error }) (*model.CommissionRule, error) {
	var rule model.CommissionRule
	if err := scanner.Scan(
		&rule.ID, &rule.ReferralCodeID, &rule.Stage1Days, &rule.Stage1Pct, &rule.Stage2Days, &rule.Stage2Pct,
		&rule.Stage3Pct, &rule.BaseType, &rule.MinimumPayout, &rule.ActiveFrom, &rule.ActiveTo,
		&rule.CreatedAt, &rule.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &rule, nil
}
