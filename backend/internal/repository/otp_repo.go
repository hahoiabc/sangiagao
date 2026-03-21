package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sangiagao/rice-marketplace/pkg/crypto"
)

type OTPRepo struct {
	pool   *pgxpool.Pool
	crypto *crypto.PhoneCrypto
}

func NewOTPRepo(pool *pgxpool.Pool, phoneCrypto *crypto.PhoneCrypto) *OTPRepo {
	return &OTPRepo{pool: pool, crypto: phoneCrypto}
}

type OTPRecord struct {
	ID        string
	Phone     string
	Code      string
	Attempts  int
	ExpiresAt time.Time
	Verified  bool
}

func (r *OTPRepo) Create(ctx context.Context, phone, code string, expiresAt time.Time) error {
	phoneHash := r.crypto.Hash(phone)
	_, err := r.pool.Exec(ctx,
		`INSERT INTO otp_requests (phone, code, expires_at, phone_hash) VALUES ($1, $2, $3, $4)`,
		phone, code, expiresAt, phoneHash,
	)
	return err
}

func (r *OTPRepo) GetLatest(ctx context.Context, phone string) (*OTPRecord, error) {
	phoneHash := r.crypto.Hash(phone)
	var otp OTPRecord
	err := r.pool.QueryRow(ctx,
		`SELECT id, phone, code, attempts, expires_at, verified
		 FROM otp_requests
		 WHERE phone_hash = $1 AND verified = FALSE
		 ORDER BY created_at DESC LIMIT 1`,
		phoneHash,
	).Scan(&otp.ID, &otp.Phone, &otp.Code, &otp.Attempts, &otp.ExpiresAt, &otp.Verified)
	if err != nil {
		// Fallback: try plaintext phone for unmigrated data
		err = r.pool.QueryRow(ctx,
			`SELECT id, phone, code, attempts, expires_at, verified
			 FROM otp_requests
			 WHERE phone = $1 AND phone_hash IS NULL AND verified = FALSE
			 ORDER BY created_at DESC LIMIT 1`,
			phone,
		).Scan(&otp.ID, &otp.Phone, &otp.Code, &otp.Attempts, &otp.ExpiresAt, &otp.Verified)
		if err != nil {
			return nil, err
		}
	}
	return &otp, nil
}

func (r *OTPRepo) IncrementAttempts(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE otp_requests SET attempts = attempts + 1 WHERE id = $1`, id,
	)
	return err
}

func (r *OTPRepo) MarkVerified(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE otp_requests SET verified = TRUE WHERE id = $1`, id,
	)
	return err
}

func (r *OTPRepo) CountRecent(ctx context.Context, phone string, since time.Time) (int, error) {
	phoneHash := r.crypto.Hash(phone)
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM otp_requests WHERE (phone_hash = $1 OR (phone = $2 AND phone_hash IS NULL)) AND created_at > $3`,
		phoneHash, phone, since,
	).Scan(&count)
	return count, err
}
