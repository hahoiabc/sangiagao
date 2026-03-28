package repository

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sangiagao/rice-marketplace/internal/model"
)

var ErrNotificationNotFound = errors.New("notification not found")

type NotificationRepo struct {
	pool *pgxpool.Pool
}

func NewNotificationRepo(pool *pgxpool.Pool) *NotificationRepo {
	return &NotificationRepo{pool: pool}
}

func (r *NotificationRepo) Create(ctx context.Context, userID, nType, title, body string, data json.RawMessage) (*model.Notification, error) {
	var n model.Notification
	err := r.pool.QueryRow(ctx,
		`INSERT INTO notifications (user_id, type, title, body, data)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, user_id, type, title, body, data, is_read, created_at`,
		userID, nType, title, body, data,
	).Scan(&n.ID, &n.UserID, &n.Type, &n.Title, &n.Body, &n.Data, &n.IsRead, &n.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func (r *NotificationRepo) ListByUser(ctx context.Context, userID string, page, limit int) ([]*model.Notification, int, error) {
	offset := (page - 1) * limit

	var total int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM notifications WHERE user_id = $1`, userID,
	).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, type, title, body, data, is_read, created_at
		 FROM notifications WHERE user_id = $1
		 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var notifications []*model.Notification
	for rows.Next() {
		var n model.Notification
		if err := rows.Scan(&n.ID, &n.UserID, &n.Type, &n.Title, &n.Body, &n.Data, &n.IsRead, &n.CreatedAt); err != nil {
			return nil, 0, err
		}
		notifications = append(notifications, &n)
	}
	if notifications == nil {
		notifications = []*model.Notification{}
	}
	return notifications, total, rows.Err()
}

func (r *NotificationRepo) MarkRead(ctx context.Context, id, userID string) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE notifications SET is_read = true WHERE id = $1 AND user_id = $2`, id, userID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotificationNotFound
	}
	return nil
}

func (r *NotificationRepo) RegisterDevice(ctx context.Context, userID, token, platform string) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO device_tokens (user_id, token, platform)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (user_id, token) DO NOTHING`,
		userID, token, platform,
	)
	return err
}

func (r *NotificationRepo) GetDeviceTokens(ctx context.Context, userID string) ([]string, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT token FROM device_tokens WHERE user_id = $1 LIMIT 50`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tokens := make([]string, 0, 10)
	for rows.Next() {
		var token string
		if err := rows.Scan(&token); err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
	}
	return tokens, rows.Err()
}

func (r *NotificationRepo) UnreadCount(ctx context.Context, userID string) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND is_read = false`, userID,
	).Scan(&count)
	return count, err
}

func (r *NotificationRepo) GetAllUserIDs(ctx context.Context) ([]string, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id FROM users WHERE is_blocked = false AND role NOT IN ('admin', 'owner')`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (r *NotificationRepo) GetAllDeviceTokens(ctx context.Context) ([]string, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT DISTINCT dt.token FROM device_tokens dt
		 JOIN users u ON u.id = dt.user_id
		 WHERE u.is_blocked = false`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []string
	for rows.Next() {
		var token string
		if err := rows.Scan(&token); err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
	}
	return tokens, rows.Err()
}

func (r *NotificationRepo) CreateBatch(ctx context.Context, userIDs []string, nType, title, body string, data json.RawMessage) error {
	if len(userIDs) == 0 {
		return nil
	}
	batch := &pgx.Batch{}
	for _, uid := range userIDs {
		batch.Queue(
			`INSERT INTO notifications (user_id, type, title, body, data) VALUES ($1, $2, $3, $4, $5)`,
			uid, nType, title, body, data,
		)
	}
	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()
	for range userIDs {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}
	return nil
}
