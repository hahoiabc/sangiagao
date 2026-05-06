-- Migration 023: Apple In-App Purchase support
-- Adds columns to subscriptions table + audit log table for App Store Server Notifications V2

-- Extend subscriptions table with Apple-specific columns
ALTER TABLE subscriptions
    ADD COLUMN IF NOT EXISTS apple_transaction_id TEXT,
    ADD COLUMN IF NOT EXISTS apple_original_transaction_id TEXT,
    ADD COLUMN IF NOT EXISTS apple_product_id TEXT,
    ADD COLUMN IF NOT EXISTS apple_environment TEXT CHECK (apple_environment IN ('Sandbox', 'Production')),
    ADD COLUMN IF NOT EXISTS auto_renew_status BOOLEAN,
    ADD COLUMN IF NOT EXISTS source TEXT NOT NULL DEFAULT 'web' CHECK (source IN ('web', 'apple', 'admin'));

-- Apple uses original_transaction_id as the stable identifier across renewals
CREATE UNIQUE INDEX IF NOT EXISTS idx_subscriptions_apple_original_tx
    ON subscriptions(apple_original_transaction_id)
    WHERE apple_original_transaction_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_subscriptions_source ON subscriptions(source);

-- Audit table for App Store Server Notifications V2
-- Idempotency: dedupe by notification_uuid
CREATE TABLE IF NOT EXISTS apple_notifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    notification_uuid TEXT NOT NULL UNIQUE,
    notification_type TEXT NOT NULL,
    subtype TEXT,
    bundle_id TEXT NOT NULL,
    environment TEXT NOT NULL CHECK (environment IN ('Sandbox', 'Production')),
    transaction_id TEXT,
    original_transaction_id TEXT,
    product_id TEXT,
    expires_date TIMESTAMPTZ,
    raw_payload JSONB NOT NULL,
    processed BOOLEAN NOT NULL DEFAULT false,
    error TEXT,
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_apple_notif_original_tx ON apple_notifications(original_transaction_id);
CREATE INDEX IF NOT EXISTS idx_apple_notif_received ON apple_notifications(received_at DESC);
CREATE INDEX IF NOT EXISTS idx_apple_notif_unprocessed ON apple_notifications(processed) WHERE processed = false;

-- Map between Apple product_id and our subscription_plans.months
-- Allows us to look up plan duration when receiving an Apple transaction
CREATE TABLE IF NOT EXISTS apple_product_map (
    product_id TEXT PRIMARY KEY,
    months INTEGER NOT NULL CHECK (months > 0),
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO apple_product_map (product_id, months) VALUES
    ('com.sangiagao.premium.1m', 1),
    ('com.sangiagao.premium.3m', 3),
    ('com.sangiagao.premium.6m', 6),
    ('com.sangiagao.premium.12m', 12)
ON CONFLICT (product_id) DO NOTHING;
