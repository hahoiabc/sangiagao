package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sangiagao/rice-marketplace/internal/model"
)

var ErrPaymentNotFound = errors.New("payment order not found")

type PaymentRepo struct {
	pool *pgxpool.Pool
}

func NewPaymentRepo(pool *pgxpool.Pool) *PaymentRepo {
	return &PaymentRepo{pool: pool}
}

func (r *PaymentRepo) Create(ctx context.Context, userID string, planMonths int, amount int64, orderCode string, expiresAt time.Time) (*model.PaymentOrder, error) {
	var order model.PaymentOrder
	err := r.pool.QueryRow(ctx,
		`INSERT INTO payment_orders (user_id, plan_months, amount, order_code, expires_at)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, user_id, plan_months, amount, order_code, status, sepay_transaction_id, paid_at, expires_at, created_at`,
		userID, planMonths, amount, orderCode, expiresAt,
	).Scan(&order.ID, &order.UserID, &order.PlanMonths, &order.Amount, &order.OrderCode,
		&order.Status, &order.SepayTransactionID, &order.PaidAt, &order.ExpiresAt, &order.CreatedAt)
	return &order, err
}

func (r *PaymentRepo) GetByOrderCode(ctx context.Context, orderCode string) (*model.PaymentOrder, error) {
	var order model.PaymentOrder
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, plan_months, amount, order_code, status, sepay_transaction_id, paid_at, expires_at, created_at
		 FROM payment_orders WHERE order_code = $1`, orderCode,
	).Scan(&order.ID, &order.UserID, &order.PlanMonths, &order.Amount, &order.OrderCode,
		&order.Status, &order.SepayTransactionID, &order.PaidAt, &order.ExpiresAt, &order.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrPaymentNotFound
	}
	return &order, err
}

func (r *PaymentRepo) GetByID(ctx context.Context, id string) (*model.PaymentOrder, error) {
	var order model.PaymentOrder
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, plan_months, amount, order_code, status, sepay_transaction_id, paid_at, expires_at, created_at
		 FROM payment_orders WHERE id = $1`, id,
	).Scan(&order.ID, &order.UserID, &order.PlanMonths, &order.Amount, &order.OrderCode,
		&order.Status, &order.SepayTransactionID, &order.PaidAt, &order.ExpiresAt, &order.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrPaymentNotFound
	}
	return &order, err
}

func (r *PaymentRepo) MarkPaid(ctx context.Context, orderCode string, sepayTxID int64) error {
	result, err := r.pool.Exec(ctx,
		`UPDATE payment_orders SET status = 'paid', sepay_transaction_id = $1, paid_at = NOW()
		 WHERE order_code = $2 AND status = 'pending'`,
		sepayTxID, orderCode,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrPaymentNotFound
	}
	return nil
}

func (r *PaymentRepo) ExpireOverdue(ctx context.Context) (int, error) {
	result, err := r.pool.Exec(ctx,
		`UPDATE payment_orders SET status = 'expired'
		 WHERE status = 'pending' AND expires_at <= NOW()`,
	)
	if err != nil {
		return 0, err
	}
	return int(result.RowsAffected()), nil
}

func (r *PaymentRepo) HasPendingByUser(ctx context.Context, userID string) (bool, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM payment_orders WHERE user_id = $1 AND status = 'pending' AND expires_at > NOW()`,
		userID,
	).Scan(&count)
	return count > 0, err
}

func (r *PaymentRepo) ListAll(ctx context.Context, page, limit int) ([]*model.PaymentOrder, int, error) {
	offset := (page - 1) * limit
	var total int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM payment_orders`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.pool.Query(ctx,
		`SELECT p.id, p.user_id, p.plan_months, p.amount, p.order_code, p.status,
		        p.sepay_transaction_id, p.paid_at, p.expires_at, p.created_at,
		        u.name, u.phone
		 FROM payment_orders p
		 JOIN users u ON u.id = p.user_id
		 ORDER BY p.created_at DESC LIMIT $1 OFFSET $2`, limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var orders []*model.PaymentOrder
	for rows.Next() {
		var o model.PaymentOrder
		var userName, userPhone *string
		if err := rows.Scan(&o.ID, &o.UserID, &o.PlanMonths, &o.Amount, &o.OrderCode,
			&o.Status, &o.SepayTransactionID, &o.PaidAt, &o.ExpiresAt, &o.CreatedAt,
			&userName, &userPhone); err != nil {
			return nil, 0, err
		}
		o.UserName = userName
		o.UserPhone = userPhone
		orders = append(orders, &o)
	}
	if orders == nil {
		orders = []*model.PaymentOrder{}
	}
	return orders, total, rows.Err()
}

func (r *PaymentRepo) HasSepayTxID(ctx context.Context, txID int64) (bool, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM payment_orders WHERE sepay_transaction_id = $1`,
		txID,
	).Scan(&count)
	return count > 0, err
}
