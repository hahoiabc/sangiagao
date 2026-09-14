package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SpamRepo struct {
	pool *pgxpool.Pool
}

func NewSpamRepo(pool *pgxpool.Pool) *SpamRepo {
	return &SpamRepo{pool: pool}
}

func (r *SpamRepo) LogAttempt(ctx context.Context, ip, deviceID, phone, action string, success bool) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO auth_attempts (ip_address, device_id, phone, action, success)
		 VALUES ($1, $2, $3, $4, $5)`,
		ip, nilIfEmpty(deviceID), nilIfEmpty(phone), action, success,
	)
	return err
}

func (r *SpamRepo) CountByIP(ctx context.Context, ip, action string, since time.Time) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM auth_attempts
		 WHERE ip_address = $1 AND action = $2 AND created_at > $3`,
		ip, action, since,
	).Scan(&count)
	return count, err
}

func (r *SpamRepo) CountByDevice(ctx context.Context, deviceID, action string, since time.Time) (int, error) {
	if deviceID == "" {
		return 0, nil
	}
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM auth_attempts
		 WHERE device_id = $1 AND action = $2 AND created_at > $3`,
		deviceID, action, since,
	).Scan(&count)
	return count, err
}

// NthRecentByIP trả thời điểm của bản ghi mới-nhất-thứ (offset+1) khớp action
// trong khoảng since→now (offset 0-based). Dùng tính thời gian mở lại chặn login
// theo cửa sổ trượt. ok=false nếu chưa đủ số bản ghi.
func (r *SpamRepo) NthRecentByIP(ctx context.Context, ip, action string, offset int, since time.Time) (time.Time, bool, error) {
	var t time.Time
	err := r.pool.QueryRow(ctx,
		`SELECT created_at FROM auth_attempts
		 WHERE ip_address = $1 AND action = $2 AND created_at > $3
		 ORDER BY created_at DESC OFFSET $4 LIMIT 1`,
		ip, action, since, offset,
	).Scan(&t)
	if errors.Is(err, pgx.ErrNoRows) {
		return time.Time{}, false, nil
	}
	if err != nil {
		return time.Time{}, false, err
	}
	return t, true, nil
}

// NthRecentByDevice — như NthRecentByIP nhưng theo device_id.
func (r *SpamRepo) NthRecentByDevice(ctx context.Context, deviceID, action string, offset int, since time.Time) (time.Time, bool, error) {
	if deviceID == "" {
		return time.Time{}, false, nil
	}
	var t time.Time
	err := r.pool.QueryRow(ctx,
		`SELECT created_at FROM auth_attempts
		 WHERE device_id = $1 AND action = $2 AND created_at > $3
		 ORDER BY created_at DESC OFFSET $4 LIMIT 1`,
		deviceID, action, since, offset,
	).Scan(&t)
	if errors.Is(err, pgx.ErrNoRows) {
		return time.Time{}, false, nil
	}
	if err != nil {
		return time.Time{}, false, err
	}
	return t, true, nil
}

func (r *SpamRepo) CountByDeviceAllTime(ctx context.Context, deviceID, action string) (int, error) {
	if deviceID == "" {
		return 0, nil
	}
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM auth_attempts
		 WHERE device_id = $1 AND action = $2 AND success = true`,
		deviceID, action,
	).Scan(&count)
	return count, err
}

func (r *SpamRepo) Cleanup(ctx context.Context, olderThan time.Time) (int, error) {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM auth_attempts WHERE created_at < $1`, olderThan,
	)
	if err != nil {
		return 0, err
	}
	return int(tag.RowsAffected()), nil
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
