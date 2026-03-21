package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sangiagao/rice-marketplace/internal/config"
	phonecrypto "github.com/sangiagao/rice-marketplace/pkg/crypto"
)

// migrate-phones backfills phone_hash and phone_encrypt for existing users and otp_requests.
// Run this AFTER applying migration 008_phone_encryption.sql.
//
// Usage: go run ./cmd/migrate-phones
func main() {
	cfg := config.Load()

	pc, err := phonecrypto.New(cfg.PhoneEncryptKey)
	if err != nil {
		log.Fatal("Invalid PHONE_ENCRYPT_KEY:", err)
	}

	pool, err := pgxpool.New(context.Background(), cfg.DBDSN())
	if err != nil {
		log.Fatal("DB connection failed:", err)
	}
	defer pool.Close()

	ctx := context.Background()

	// --- Migrate users table ---
	fmt.Println("Migrating users table...")
	rows, err := pool.Query(ctx, `SELECT id, phone FROM users WHERE phone_hash IS NULL AND phone IS NOT NULL`)
	if err != nil {
		log.Fatal("Query users failed:", err)
	}

	userCount := 0
	type userRow struct {
		ID    string
		Phone string
	}
	var users []userRow
	for rows.Next() {
		var u userRow
		if err := rows.Scan(&u.ID, &u.Phone); err != nil {
			log.Fatal("Scan user failed:", err)
		}
		users = append(users, u)
	}
	rows.Close()

	for _, u := range users {
		phoneHash := pc.Hash(u.Phone)
		phoneEnc, err := pc.Encrypt(u.Phone)
		if err != nil {
			log.Printf("Encrypt failed for user %s: %v", u.ID, err)
			continue
		}
		_, err = pool.Exec(ctx,
			`UPDATE users SET phone_hash = $1, phone_encrypt = $2 WHERE id = $3`,
			phoneHash, phoneEnc, u.ID,
		)
		if err != nil {
			log.Printf("Update user %s failed: %v", u.ID, err)
			continue
		}
		userCount++
	}
	fmt.Printf("  Migrated %d users\n", userCount)

	// --- Migrate otp_requests table ---
	fmt.Println("Migrating otp_requests table...")
	otpRows, err := pool.Query(ctx, `SELECT id, phone FROM otp_requests WHERE phone_hash IS NULL AND phone IS NOT NULL`)
	if err != nil {
		log.Fatal("Query otp_requests failed:", err)
	}

	otpCount := 0
	type otpRow struct {
		ID    string
		Phone string
	}
	var otps []otpRow
	for otpRows.Next() {
		var o otpRow
		if err := otpRows.Scan(&o.ID, &o.Phone); err != nil {
			log.Fatal("Scan otp failed:", err)
		}
		otps = append(otps, o)
	}
	otpRows.Close()

	for _, o := range otps {
		phoneHash := pc.Hash(o.Phone)
		_, err = pool.Exec(ctx,
			`UPDATE otp_requests SET phone_hash = $1 WHERE id = $2`,
			phoneHash, o.ID,
		)
		if err != nil {
			log.Printf("Update otp %s failed: %v", o.ID, err)
			continue
		}
		otpCount++
	}
	fmt.Printf("  Migrated %d otp_requests\n", otpCount)

	fmt.Println("Done!")
	fmt.Println("")
	fmt.Println("Next steps:")
	fmt.Println("  1. Verify the data: SELECT id, phone, phone_hash, phone_encrypt FROM users LIMIT 5;")
	fmt.Println("  2. After confirming, you can drop the old phone column (see migration 008 comments)")

	os.Exit(0)
}
