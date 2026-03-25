-- Migration 008: Phone number encryption (Hash + AES)
-- Adds phone_hash (for lookups) and phone_encrypt (for display) columns.
-- After running the Go migration tool to backfill data, the old phone column can be dropped.

-- Step 1: Add new columns to users
ALTER TABLE users ADD COLUMN IF NOT EXISTS phone_hash VARCHAR(64);
ALTER TABLE users ADD COLUMN IF NOT EXISTS phone_encrypt TEXT;

-- Step 2: Add new columns to otp_requests
ALTER TABLE otp_requests ADD COLUMN IF NOT EXISTS phone_hash VARCHAR(64);

-- Step 3: Create index on phone_hash for fast lookups
CREATE INDEX IF NOT EXISTS idx_users_phone_hash ON users(phone_hash);
CREATE INDEX IF NOT EXISTS idx_otp_phone_hash ON otp_requests(phone_hash);

-- Step 4: After backfill migration tool runs successfully:
-- Run these manually once all data is migrated:
--
-- ALTER TABLE users DROP CONSTRAINT users_phone_key;
-- ALTER TABLE users DROP COLUMN phone;
-- ALTER TABLE users ADD CONSTRAINT users_phone_hash_unique UNIQUE(phone_hash);
-- ALTER TABLE otp_requests DROP COLUMN phone;
-- DROP INDEX IF EXISTS idx_users_phone;
-- DROP INDEX IF EXISTS idx_otp_phone;
