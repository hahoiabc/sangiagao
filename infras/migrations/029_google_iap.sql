-- Migration 029: Google Play In-App Billing support
-- Mirrors mig 023 (Apple IAP) structure but for Google Play subscriptions.
-- RTDN events arrive via Google Cloud Pub/Sub HTTP push.

-- 1. Extend subscriptions table with Google-specific columns
ALTER TABLE subscriptions
    ADD COLUMN IF NOT EXISTS google_purchase_token TEXT,
    ADD COLUMN IF NOT EXISTS google_order_id TEXT,
    ADD COLUMN IF NOT EXISTS google_product_id TEXT,
    ADD COLUMN IF NOT EXISTS google_subscription_id TEXT;

-- Allow 'google' as a payment source
ALTER TABLE subscriptions
    DROP CONSTRAINT IF EXISTS subscriptions_source_check;
ALTER TABLE subscriptions
    ADD CONSTRAINT subscriptions_source_check
    CHECK (source IN ('web', 'apple', 'google', 'admin'));

-- Index for fast lookup by purchase_token (renewals come in with same token)
CREATE UNIQUE INDEX IF NOT EXISTS idx_subscriptions_google_token
    ON subscriptions(google_purchase_token)
    WHERE google_purchase_token IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_subscriptions_google_subid
    ON subscriptions(google_subscription_id)
    WHERE google_subscription_id IS NOT NULL;

-- 2. Audit table for Google Real-time Developer Notifications
-- Idempotency: dedupe by (purchase_token, event_time_millis)
CREATE TABLE IF NOT EXISTS google_iap_notifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    notification_type INT NOT NULL,
    subscription_id TEXT,         -- google product/sub ID (e.g. com.sangiagao.premium.1m)
    purchase_token TEXT NOT NULL,
    package_name TEXT NOT NULL,
    event_time_millis BIGINT NOT NULL,
    raw_payload JSONB NOT NULL,
    processed BOOLEAN NOT NULL DEFAULT false,
    error TEXT,
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMPTZ,
    UNIQUE (purchase_token, event_time_millis)
);

CREATE INDEX IF NOT EXISTS idx_google_notif_token ON google_iap_notifications(purchase_token);
CREATE INDEX IF NOT EXISTS idx_google_notif_received ON google_iap_notifications(received_at DESC);
CREATE INDEX IF NOT EXISTS idx_google_notif_unprocessed ON google_iap_notifications(processed) WHERE processed = false;

-- 3. Map between Google product_id and our plan months + commission pricing
CREATE TABLE IF NOT EXISTS google_product_map (
    product_id       TEXT PRIMARY KEY,
    months           INTEGER NOT NULL CHECK (months > 0),
    gross_amount     BIGINT NOT NULL DEFAULT 0,
    platform_fee_pct NUMERIC(5,4) NOT NULL DEFAULT 0.15,  -- Google 15% first $1M, then 30%
    is_active        BOOLEAN NOT NULL DEFAULT true,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Seed Google products with same IDs as Apple for consistency + same pricing as iOS (+30%)
INSERT INTO google_product_map (product_id, months, gross_amount, platform_fee_pct) VALUES
    ('com.sangiagao.premium.1m',  1,  49000, 0.15),
    ('com.sangiagao.premium.3m',  3,  125000, 0.15),
    ('com.sangiagao.premium.6m',  6,  249000, 0.15),
    ('com.sangiagao.premium.12m', 12, 425000, 0.15)
ON CONFLICT (product_id) DO NOTHING;
