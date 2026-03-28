package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sangiagao/rice-marketplace/internal/model"
)

const subCols = `id, user_id, plan, duration_months, amount, started_at, expires_at, status, created_at`

type SubscriptionRepo struct {
	pool *pgxpool.Pool
}

func NewSubscriptionRepo(pool *pgxpool.Pool) *SubscriptionRepo {
	return &SubscriptionRepo{pool: pool}
}

func scanSub(row pgx.Row) (*model.Subscription, error) {
	var s model.Subscription
	err := row.Scan(&s.ID, &s.UserID, &s.Plan, &s.DurationMonths, &s.Amount, &s.StartedAt, &s.ExpiresAt, &s.Status, &s.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func scanSubRows(rows pgx.Rows) ([]*model.Subscription, error) {
	var subs []*model.Subscription
	for rows.Next() {
		var s model.Subscription
		if err := rows.Scan(&s.ID, &s.UserID, &s.Plan, &s.DurationMonths, &s.Amount, &s.StartedAt, &s.ExpiresAt, &s.Status, &s.CreatedAt); err != nil {
			return nil, err
		}
		subs = append(subs, &s)
	}
	return subs, nil
}

func (r *SubscriptionRepo) Create(ctx context.Context, userID, plan string, daysValid int) (*model.Subscription, error) {
	return scanSub(r.pool.QueryRow(ctx,
		`INSERT INTO subscriptions (user_id, plan, duration_months, amount, expires_at)
		 VALUES ($1, $2, 0, 0, NOW() + ($3 * interval '1 day'))
		 RETURNING `+subCols,
		userID, plan, daysValid,
	))
}

func (r *SubscriptionRepo) GetActiveByUserID(ctx context.Context, userID string) (*model.Subscription, error) {
	s, err := scanSub(r.pool.QueryRow(ctx,
		`SELECT `+subCols+`
		 FROM subscriptions
		 WHERE user_id = $1 AND status = 'active' AND expires_at > NOW()
		 ORDER BY created_at DESC LIMIT 1`,
		userID,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return s, err
}

func (r *SubscriptionRepo) GetByUserID(ctx context.Context, userID string) (*model.Subscription, error) {
	s, err := scanSub(r.pool.QueryRow(ctx,
		`SELECT `+subCols+`
		 FROM subscriptions
		 WHERE user_id = $1
		 ORDER BY created_at DESC LIMIT 1`,
		userID,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return s, err
}

func (r *SubscriptionRepo) ExpireOverdue(ctx context.Context) (int, error) {
	tag, err := r.pool.Exec(ctx,
		`UPDATE subscriptions SET status = 'expired'
		 WHERE status = 'active' AND expires_at <= NOW()`,
	)
	if err != nil {
		return 0, err
	}
	return int(tag.RowsAffected()), nil
}

func (r *SubscriptionRepo) HideListingsForExpired(ctx context.Context) (int, error) {
	tag, err := r.pool.Exec(ctx,
		`UPDATE listings SET status = 'hidden_subscription'
		 WHERE status = 'active'
		   AND user_id IN (
		     SELECT DISTINCT u.id FROM users u
		     WHERE u.role = 'seller'
		       AND NOT EXISTS (
		         SELECT 1 FROM subscriptions s
		         WHERE s.user_id = u.id AND s.status = 'active' AND s.expires_at > NOW()
		       )
		   )`,
	)
	if err != nil {
		return 0, err
	}
	return int(tag.RowsAffected()), nil
}

func (r *SubscriptionRepo) ActivateByUserID(ctx context.Context, userID string, daysValid int, durationMonths int, amount int64, plan string) (*model.Subscription, error) {
	return scanSub(r.pool.QueryRow(ctx,
		`INSERT INTO subscriptions (user_id, plan, duration_months, amount, expires_at)
		 VALUES ($1, $5, $2, $3, NOW() + ($4 * interval '1 day'))
		 RETURNING `+subCols,
		userID, durationMonths, amount, daysValid, plan,
	))
}

func (r *SubscriptionRepo) ExtendSubscription(ctx context.Context, subID string, extraDays int, durationMonths int, amount int64) (*model.Subscription, error) {
	return scanSub(r.pool.QueryRow(ctx,
		`UPDATE subscriptions
		 SET expires_at = expires_at + ($2 * interval '1 day'),
		     duration_months = duration_months + $3,
		     amount = amount + $4
		 WHERE id = $1
		 RETURNING `+subCols,
		subID, extraDays, durationMonths, amount,
	))
}

func (r *SubscriptionRepo) ListByUserID(ctx context.Context, userID string, page, limit int) ([]*model.Subscription, int, error) {
	offset := (page - 1) * limit

	var total int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM subscriptions WHERE user_id = $1`, userID,
	).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.Query(ctx,
		`SELECT `+subCols+`
		 FROM subscriptions
		 WHERE user_id = $1
		 ORDER BY created_at DESC
		 LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	subs, err := scanSubRows(rows)
	if err != nil {
		return nil, 0, err
	}
	return subs, total, nil
}

func (r *SubscriptionRepo) GetExpiringSoon(ctx context.Context, withinHours int) ([]*model.Subscription, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+subCols+`
		 FROM subscriptions
		 WHERE status = 'active'
		   AND expires_at > NOW()
		   AND expires_at <= NOW() + ($1 * interval '1 hour')`,
		withinHours,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanSubRows(rows)
}

type SubRevenueMonth struct {
	Month      string `json:"month"`
	PaidCount  int    `json:"paid_count"`
	TrialCount int    `json:"trial_count"`
	Revenue    int64  `json:"revenue"`
}

type SubRevenueStats struct {
	TotalSubscriptions int               `json:"total_subscriptions"`
	ActiveCount        int               `json:"active_count"`
	ExpiredCount       int               `json:"expired_count"`
	PaidCount          int               `json:"paid_count"`
	TrialCount         int               `json:"trial_count"`
	TotalRevenue       int64             `json:"total_revenue"`
	MonthlyRevenue     []SubRevenueMonth `json:"monthly_revenue"`
}

func (r *SubscriptionRepo) GetRevenueStats(ctx context.Context) (*SubRevenueStats, error) {
	stats := &SubRevenueStats{}

	// Single query for all aggregate counts
	err := r.pool.QueryRow(ctx, `
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE status = 'active' AND expires_at > NOW()),
			COUNT(*) FILTER (WHERE status = 'expired' OR expires_at <= NOW()),
			COUNT(*) FILTER (WHERE plan = 'paid'),
			COUNT(*) FILTER (WHERE plan = 'free_trial'),
			COALESCE(SUM(amount) FILTER (WHERE plan = 'paid'), 0)
		FROM subscriptions
	`).Scan(&stats.TotalSubscriptions, &stats.ActiveCount, &stats.ExpiredCount, &stats.PaidCount, &stats.TrialCount, &stats.TotalRevenue)
	if err != nil {
		return nil, err
	}

	// Monthly breakdown (last 12 months)
	rows, err := r.pool.Query(ctx,
		`SELECT
			TO_CHAR(created_at, 'YYYY-MM') AS month,
			COUNT(*) FILTER (WHERE plan = 'paid') AS paid_count,
			COUNT(*) FILTER (WHERE plan = 'free_trial') AS trial_count,
			COALESCE(SUM(amount) FILTER (WHERE plan = 'paid'), 0) AS revenue
		 FROM subscriptions
		 WHERE created_at >= NOW() - INTERVAL '12 months'
		 GROUP BY TO_CHAR(created_at, 'YYYY-MM')
		 ORDER BY month`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var m SubRevenueMonth
		if err := rows.Scan(&m.Month, &m.PaidCount, &m.TrialCount, &m.Revenue); err != nil {
			return nil, err
		}
		stats.MonthlyRevenue = append(stats.MonthlyRevenue, m)
	}

	return stats, nil
}

type SubRevenueDay struct {
	Date       string `json:"date"`
	PaidCount  int    `json:"paid_count"`
	TrialCount int    `json:"trial_count"`
	Revenue    int64  `json:"revenue"`
}

type SubDailyRevenueReport struct {
	From       string          `json:"from"`
	To         string          `json:"to"`
	TotalPaid  int             `json:"total_paid"`
	TotalTrial int             `json:"total_trial"`
	TotalRevenue int64         `json:"total_revenue"`
	Days       []SubRevenueDay `json:"days"`
}

func (r *SubscriptionRepo) GetDailyRevenue(ctx context.Context, from, to string) (*SubDailyRevenueReport, error) {
	report := &SubDailyRevenueReport{From: from, To: to}

	rows, err := r.pool.Query(ctx,
		`SELECT
			TO_CHAR(created_at, 'YYYY-MM-DD') AS day,
			COUNT(*) FILTER (WHERE plan = 'paid') AS paid_count,
			COUNT(*) FILTER (WHERE plan = 'free_trial') AS trial_count,
			COALESCE(SUM(amount) FILTER (WHERE plan = 'paid'), 0) AS revenue
		 FROM subscriptions
		 WHERE created_at >= $1::date AND created_at < ($2::date + interval '1 day')
		 GROUP BY TO_CHAR(created_at, 'YYYY-MM-DD')
		 ORDER BY day`,
		from, to,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var d SubRevenueDay
		if err := rows.Scan(&d.Date, &d.PaidCount, &d.TrialCount, &d.Revenue); err != nil {
			return nil, err
		}
		report.TotalPaid += d.PaidCount
		report.TotalTrial += d.TrialCount
		report.TotalRevenue += d.Revenue
		report.Days = append(report.Days, d)
	}

	return report, nil
}

func (r *SubscriptionRepo) RestoreListings(ctx context.Context, userID string) (int, error) {
	tag, err := r.pool.Exec(ctx,
		`UPDATE listings SET status = 'active'
		 WHERE user_id = $1 AND status = 'hidden_subscription'`,
		userID,
	)
	if err != nil {
		return 0, err
	}
	return int(tag.RowsAffected()), nil
}
