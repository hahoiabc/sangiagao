package repository

import (
	"context"
	"time"

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
