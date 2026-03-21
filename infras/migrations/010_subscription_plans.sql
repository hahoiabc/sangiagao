-- Add subscription plan pricing columns
ALTER TABLE subscriptions ADD COLUMN IF NOT EXISTS duration_months INTEGER NOT NULL DEFAULT 1;
ALTER TABLE subscriptions ADD COLUMN IF NOT EXISTS amount BIGINT NOT NULL DEFAULT 0;

-- Backfill existing paid subscriptions with 1-month / 35000 VND
UPDATE subscriptions SET duration_months = 1, amount = 35000 WHERE plan = 'paid' AND amount = 0;
