package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sangiagao/rice-marketplace/internal/model"
)

var ErrPlanNotFound = errors.New("plan not found")

type PlanRepo struct {
	pool *pgxpool.Pool
}

func NewPlanRepo(pool *pgxpool.Pool) *PlanRepo {
	return &PlanRepo{pool: pool}
}

const planCols = `id, months, amount, label, is_active, sort_order, created_at, updated_at`

func scanPlan(row pgx.Row) (*model.SubscriptionPlan, error) {
	var p model.SubscriptionPlan
	var createdAt, updatedAt time.Time
	err := row.Scan(&p.ID, &p.Months, &p.Amount, &p.Label, &p.IsActive, &p.SortOrder, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}
	p.CreatedAt = createdAt.Format(time.RFC3339)
	p.UpdatedAt = updatedAt.Format(time.RFC3339)
	return &p, nil
}

func scanPlans(rows pgx.Rows) ([]model.SubscriptionPlan, error) {
	var plans []model.SubscriptionPlan
	for rows.Next() {
		var p model.SubscriptionPlan
		var createdAt, updatedAt time.Time
		if err := rows.Scan(&p.ID, &p.Months, &p.Amount, &p.Label, &p.IsActive, &p.SortOrder, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		p.CreatedAt = createdAt.Format(time.RFC3339)
		p.UpdatedAt = updatedAt.Format(time.RFC3339)
		plans = append(plans, p)
	}
	if plans == nil {
		plans = []model.SubscriptionPlan{}
	}
	return plans, rows.Err()
}

func (r *PlanRepo) ListActive(ctx context.Context) ([]model.SubscriptionPlan, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+planCols+` FROM subscription_plans WHERE is_active = true ORDER BY sort_order, months`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPlans(rows)
}

func (r *PlanRepo) ListAll(ctx context.Context) ([]model.SubscriptionPlan, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+planCols+` FROM subscription_plans ORDER BY sort_order, months`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPlans(rows)
}

func (r *PlanRepo) GetByMonths(ctx context.Context, months int) (*model.SubscriptionPlan, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT `+planCols+` FROM subscription_plans WHERE months = $1 AND is_active = true`, months)
	p, err := scanPlan(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return p, err
}

func (r *PlanRepo) Create(ctx context.Context, req *model.CreatePlanRequest) (*model.SubscriptionPlan, error) {
	row := r.pool.QueryRow(ctx,
		`INSERT INTO subscription_plans (months, amount, label, sort_order)
		 VALUES ($1, $2, $3, (SELECT COALESCE(MAX(sort_order), 0) + 1 FROM subscription_plans))
		 RETURNING `+planCols,
		req.Months, req.Amount, req.Label)
	return scanPlan(row)
}

func (r *PlanRepo) Update(ctx context.Context, id string, req *model.UpdatePlanRequest) (*model.SubscriptionPlan, error) {
	row := r.pool.QueryRow(ctx,
		`UPDATE subscription_plans SET
			months = COALESCE($2, months),
			amount = COALESCE($3, amount),
			label = COALESCE($4, label),
			is_active = COALESCE($5, is_active)
		 WHERE id = $1
		 RETURNING `+planCols,
		id, req.Months, req.Amount, req.Label, req.IsActive)
	p, err := scanPlan(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrPlanNotFound
	}
	return p, err
}

func (r *PlanRepo) Delete(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM subscription_plans WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrPlanNotFound
	}
	return nil
}
