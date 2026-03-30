package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sangiagao/rice-marketplace/internal/model"
)

var ErrInboxNotFound = errors.New("inbox message not found")

type InboxRepo struct {
	pool *pgxpool.Pool
}

func NewInboxRepo(pool *pgxpool.Pool) *InboxRepo {
	return &InboxRepo{pool: pool}
}

func (r *InboxRepo) Create(ctx context.Context, adminID string, req *model.CreateInboxRequest) (*model.InboxMessage, error) {
	target := req.Target
	if target == "" {
		target = "all_users"
	}

	var expiresAt *time.Time
	if req.ExpiresAt != nil && *req.ExpiresAt != "" {
		t, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err != nil {
			return nil, fmt.Errorf("invalid expires_at format: %w", err)
		}
		expiresAt = &t
	}

	var msg model.InboxMessage
	err := r.pool.QueryRow(ctx,
		`INSERT INTO system_inbox (title, body, image_url, target, is_pinned, expires_at, created_by)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id, title, body, image_url, target, is_pinned, expires_at, created_by, created_at`,
		req.Title, req.Body, req.ImageURL, target, req.IsPinned, expiresAt, adminID,
	).Scan(&msg.ID, &msg.Title, &msg.Body, &msg.ImageURL, &msg.Target, &msg.IsPinned, &msg.ExpiresAt, &msg.CreatedBy, &msg.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

func (r *InboxRepo) Update(ctx context.Context, id string, req *model.UpdateInboxRequest) (*model.InboxMessage, error) {
	var expiresAt *time.Time
	if req.ExpiresAt != nil && *req.ExpiresAt != "" {
		t, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err != nil {
			return nil, fmt.Errorf("invalid expires_at format: %w", err)
		}
		expiresAt = &t
	}

	var msg model.InboxMessage
	err := r.pool.QueryRow(ctx,
		`UPDATE system_inbox SET
		    title = COALESCE($2, title),
		    body = COALESCE($3, body),
		    image_url = COALESCE($4, image_url),
		    is_pinned = COALESCE($5, is_pinned),
		    expires_at = CASE WHEN $6 THEN $7 ELSE expires_at END
		 WHERE id = $1
		 RETURNING id, title, body, image_url, target, is_pinned, expires_at, created_by, created_at`,
		id, req.Title, req.Body, req.ImageURL, req.IsPinned, req.ExpiresAt != nil, expiresAt,
	).Scan(&msg.ID, &msg.Title, &msg.Body, &msg.ImageURL, &msg.Target, &msg.IsPinned, &msg.ExpiresAt, &msg.CreatedBy, &msg.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrInboxNotFound
	}
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

func (r *InboxRepo) Delete(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM system_inbox WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrInboxNotFound
	}
	return nil
}

// targetFilter builds WHERE clause for target matching user role.
// target can be: "all_users", "role:member", "role:seller", etc.
func targetFilter(userRole string) string {
	return `(si.target = 'all_users' OR si.target = 'role:' || $2)`
}

func (r *InboxRepo) ListForUser(ctx context.Context, userID, userRole string, page, limit int) ([]*model.InboxMessage, int, error) {
	offset := (page - 1) * limit

	// Count
	var total int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM system_inbox si
		 WHERE (`+targetFilter(userRole)+`)
		   AND (si.expires_at IS NULL OR si.expires_at > NOW())`,
		userID, userRole,
	).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// List with read status
	rows, err := r.pool.Query(ctx,
		`SELECT si.id, si.title, si.body, si.image_url, si.target, si.is_pinned, si.expires_at, si.created_by, si.created_at,
		        (irs.user_id IS NOT NULL) AS is_read
		 FROM system_inbox si
		 LEFT JOIN inbox_read_status irs ON irs.inbox_id = si.id AND irs.user_id = $1
		 WHERE (`+targetFilter(userRole)+`)
		   AND (si.expires_at IS NULL OR si.expires_at > NOW())
		 ORDER BY si.is_pinned DESC, si.created_at DESC
		 LIMIT $3 OFFSET $4`,
		userID, userRole, limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var msgs []*model.InboxMessage
	for rows.Next() {
		var m model.InboxMessage
		if err := rows.Scan(&m.ID, &m.Title, &m.Body, &m.ImageURL, &m.Target, &m.IsPinned, &m.ExpiresAt, &m.CreatedBy, &m.CreatedAt, &m.IsRead); err != nil {
			return nil, 0, err
		}
		msgs = append(msgs, &m)
	}
	if msgs == nil {
		msgs = []*model.InboxMessage{}
	}
	return msgs, total, rows.Err()
}

func (r *InboxRepo) GetByID(ctx context.Context, id, userID string) (*model.InboxMessage, error) {
	var m model.InboxMessage
	err := r.pool.QueryRow(ctx,
		`SELECT si.id, si.title, si.body, si.image_url, si.target, si.is_pinned, si.expires_at, si.created_by, si.created_at,
		        (irs.user_id IS NOT NULL) AS is_read
		 FROM system_inbox si
		 LEFT JOIN inbox_read_status irs ON irs.inbox_id = si.id AND irs.user_id = $2
		 WHERE si.id = $1`,
		id, userID,
	).Scan(&m.ID, &m.Title, &m.Body, &m.ImageURL, &m.Target, &m.IsPinned, &m.ExpiresAt, &m.CreatedBy, &m.CreatedAt, &m.IsRead)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrInboxNotFound
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *InboxRepo) MarkRead(ctx context.Context, userID, inboxID string) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO inbox_read_status (user_id, inbox_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		userID, inboxID,
	)
	return err
}

func (r *InboxRepo) UnreadCount(ctx context.Context, userID, userRole string) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM system_inbox si
		 WHERE (`+targetFilter(userRole)+`)
		   AND (si.expires_at IS NULL OR si.expires_at > NOW())
		   AND si.id NOT IN (SELECT inbox_id FROM inbox_read_status WHERE user_id = $1)`,
		userID, userRole,
	).Scan(&count)
	return count, err
}

func (r *InboxRepo) ListAll(ctx context.Context, page, limit int) ([]*model.InboxMessage, int, error) {
	offset := (page - 1) * limit

	var total int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM system_inbox`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.Query(ctx,
		`SELECT id, title, body, image_url, target, is_pinned, expires_at, created_by, created_at
		 FROM system_inbox
		 ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var msgs []*model.InboxMessage
	for rows.Next() {
		var m model.InboxMessage
		if err := rows.Scan(&m.ID, &m.Title, &m.Body, &m.ImageURL, &m.Target, &m.IsPinned, &m.ExpiresAt, &m.CreatedBy, &m.CreatedAt); err != nil {
			return nil, 0, err
		}
		msgs = append(msgs, &m)
	}
	if msgs == nil {
		msgs = []*model.InboxMessage{}
	}
	return msgs, total, rows.Err()
}

// GetTargetUserIDs returns user IDs matching a target string for push notifications.
func (r *InboxRepo) GetTargetUserIDs(ctx context.Context, target string) ([]string, error) {
	var query string
	var args []interface{}

	if target == "all_users" {
		query = `SELECT id FROM users WHERE is_blocked = false AND role NOT IN ('admin', 'owner')`
	} else if len(target) > 5 && target[:5] == "role:" {
		role := target[5:]
		query = `SELECT id FROM users WHERE is_blocked = false AND role = $1`
		args = append(args, role)
	} else {
		return nil, nil
	}

	rows, err := r.pool.Query(ctx, query, args...)
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
