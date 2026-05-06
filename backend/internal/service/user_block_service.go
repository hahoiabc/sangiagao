package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// UserBlockService implements user-level block per Apple Guideline 1.2.
// Blocked users' listings are hidden from feed and chat for the blocker.
type UserBlockService struct {
	pool *pgxpool.Pool
}

func NewUserBlockService(pool *pgxpool.Pool) *UserBlockService {
	return &UserBlockService{pool: pool}
}

type BlockedUser struct {
	ID         string    `json:"id"`
	BlockedID  string    `json:"blocked_id"`
	BlockedName string   `json:"blocked_name,omitempty"`
	Reason     string    `json:"reason,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

var (
	ErrCannotBlockSelf = errors.New("không thể tự chặn chính mình")
	ErrAlreadyBlocked  = errors.New("đã chặn người dùng này")
	ErrNotBlocked      = errors.New("chưa chặn người dùng này")
)

func (s *UserBlockService) Block(ctx context.Context, blockerID, blockedID, reason string) error {
	if blockerID == blockedID {
		return ErrCannotBlockSelf
	}
	_, err := s.pool.Exec(ctx,
		`INSERT INTO user_blocks (blocker_id, blocked_id, reason) VALUES ($1, $2, $3)
		 ON CONFLICT (blocker_id, blocked_id) DO NOTHING`,
		blockerID, blockedID, reason,
	)
	if err != nil {
		return fmt.Errorf("user_blocks: insert: %w", err)
	}
	return nil
}

func (s *UserBlockService) Unblock(ctx context.Context, blockerID, blockedID string) error {
	tag, err := s.pool.Exec(ctx,
		`DELETE FROM user_blocks WHERE blocker_id = $1 AND blocked_id = $2`,
		blockerID, blockedID,
	)
	if err != nil {
		return fmt.Errorf("user_blocks: delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotBlocked
	}
	return nil
}

func (s *UserBlockService) List(ctx context.Context, blockerID string) ([]BlockedUser, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT b.id, b.blocked_id, COALESCE(u.name, ''), COALESCE(b.reason, ''), b.created_at
		   FROM user_blocks b
		   LEFT JOIN users u ON u.id = b.blocked_id
		  WHERE b.blocker_id = $1
		  ORDER BY b.created_at DESC`,
		blockerID,
	)
	if err != nil {
		return nil, fmt.Errorf("user_blocks: list: %w", err)
	}
	defer rows.Close()

	out := make([]BlockedUser, 0)
	for rows.Next() {
		var b BlockedUser
		if err := rows.Scan(&b.ID, &b.BlockedID, &b.BlockedName, &b.Reason, &b.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

// IsBlocked returns true if blocker has blocked blocked.
func (s *UserBlockService) IsBlocked(ctx context.Context, blockerID, blockedID string) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM user_blocks WHERE blocker_id = $1 AND blocked_id = $2)`,
		blockerID, blockedID,
	).Scan(&exists)
	return exists, err
}
