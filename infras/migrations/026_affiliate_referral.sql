-- Migration 026: Affiliate / Referral system
-- Adds new role 'aff', referral attribution on users, net_amount tracking on
-- subscriptions, and 5 new tables: referral_codes, commission_rules,
-- commission_records, payouts. Single-tier, configurable per-partner rules,
-- 3-stage age-based commission (stage1_days/stage2_days/stage3_perpetual),
-- T+45 payable delay (no clawback strategy).

-- 1. Extend users: add 'aff' role + referral attribution
ALTER TABLE users
    DROP CONSTRAINT IF EXISTS users_role_check;

ALTER TABLE users
    ADD CONSTRAINT users_role_check
    CHECK (role IN ('member', 'seller', 'admin', 'aff'));

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS referrer_user_id UUID NULL REFERENCES users(id),
    ADD COLUMN IF NOT EXISTS referred_at TIMESTAMPTZ NULL;

CREATE INDEX IF NOT EXISTS idx_users_referrer ON users(referrer_user_id) WHERE referrer_user_id IS NOT NULL;

-- 2. Extend subscriptions: track NET amount for commission base
ALTER TABLE subscriptions
    ADD COLUMN IF NOT EXISTS net_amount BIGINT NULL,
    ADD COLUMN IF NOT EXISTS platform_fee_pct NUMERIC(5,4) NOT NULL DEFAULT 0;

COMMENT ON COLUMN subscriptions.net_amount IS 'Revenue after platform fee (Apple 30%/15%, SePay 0%). Base for commission when rule.base_type=net.';
COMMENT ON COLUMN subscriptions.platform_fee_pct IS '0.30 Apple standard, 0.15 Apple Small Business, 0 SePay/admin.';

-- 3. Referral codes (1-1 per user, lazy-generated)
CREATE TABLE IF NOT EXISTS referral_codes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    code VARCHAR(8) NOT NULL UNIQUE,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_referral_codes_code ON referral_codes(code) WHERE active = TRUE;

-- 4. Commission rules: default (referral_code_id = NULL) + per-partner override
CREATE TABLE IF NOT EXISTS commission_rules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    referral_code_id UUID NULL REFERENCES referral_codes(id) ON DELETE CASCADE,
    stage1_days INT NOT NULL CHECK (stage1_days > 0),
    stage1_pct NUMERIC(5,4) NOT NULL CHECK (stage1_pct >= 0 AND stage1_pct <= 1),
    stage2_days INT NOT NULL CHECK (stage2_days > 0),
    stage2_pct NUMERIC(5,4) NOT NULL CHECK (stage2_pct >= 0 AND stage2_pct <= 1),
    stage3_pct NUMERIC(5,4) NOT NULL CHECK (stage3_pct >= 0 AND stage3_pct <= 1),
    base_type VARCHAR(8) NOT NULL DEFAULT 'net' CHECK (base_type IN ('gross', 'net')),
    minimum_payout BIGINT NOT NULL DEFAULT 100000 CHECK (minimum_payout >= 0),
    active_from TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    active_to TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Only one default rule active at a time (referral_code_id IS NULL)
CREATE UNIQUE INDEX IF NOT EXISTS idx_commission_rules_default_active
    ON commission_rules ((referral_code_id IS NULL))
    WHERE referral_code_id IS NULL AND active_to IS NULL;

-- One active rule per partner at a time
CREATE UNIQUE INDEX IF NOT EXISTS idx_commission_rules_partner_active
    ON commission_rules (referral_code_id)
    WHERE referral_code_id IS NOT NULL AND active_to IS NULL;

-- 5. Commission ledger: 1 row per payment event
CREATE TABLE IF NOT EXISTS commission_records (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    referrer_user_id UUID NOT NULL REFERENCES users(id),
    referee_user_id UUID NOT NULL REFERENCES users(id),
    subscription_id UUID NOT NULL REFERENCES subscriptions(id),
    payment_source VARCHAR(16) NOT NULL CHECK (payment_source IN ('apple', 'sepay', 'admin')),
    payment_event_id VARCHAR(64) NOT NULL,
    gross_amount BIGINT NOT NULL,
    platform_fee BIGINT NOT NULL DEFAULT 0,
    net_amount BIGINT NOT NULL,
    base_amount BIGINT NOT NULL,        -- gross or net depending on rule
    stage INT NOT NULL CHECK (stage IN (1, 2, 3)),
    rate NUMERIC(5,4) NOT NULL,
    commission_amount BIGINT NOT NULL,
    referee_age_days INT NOT NULL,
    rule_id UUID NOT NULL REFERENCES commission_rules(id),
    status VARCHAR(16) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'payable', 'paid', 'cancelled')),
    payable_after TIMESTAMPTZ NOT NULL,
    paid_at TIMESTAMPTZ NULL,
    payout_id UUID NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (payment_source, payment_event_id)
);

CREATE INDEX IF NOT EXISTS idx_commission_records_referrer_status
    ON commission_records(referrer_user_id, status);
CREATE INDEX IF NOT EXISTS idx_commission_records_payable
    ON commission_records(payable_after) WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS idx_commission_records_payout
    ON commission_records(payout_id) WHERE payout_id IS NOT NULL;

-- 6. Payout batches
CREATE TABLE IF NOT EXISTS payouts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    referrer_user_id UUID NOT NULL REFERENCES users(id),
    total_amount BIGINT NOT NULL CHECK (total_amount > 0),
    record_count INT NOT NULL CHECK (record_count > 0),
    method VARCHAR(16) NOT NULL CHECK (method IN ('bank', 'momo', 'cash', 'other')),
    bank_info JSONB NULL,
    note TEXT NULL,
    status VARCHAR(16) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'sent', 'failed')),
    created_by UUID NOT NULL REFERENCES users(id),
    sent_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_payouts_referrer ON payouts(referrer_user_id, created_at DESC);

-- 7. FK from commission_records.payout_id back to payouts
ALTER TABLE commission_records
    DROP CONSTRAINT IF EXISTS fk_commission_payout;
ALTER TABLE commission_records
    ADD CONSTRAINT fk_commission_payout FOREIGN KEY (payout_id) REFERENCES payouts(id);

-- 7b. Extend apple_product_map with pricing for net calculation
ALTER TABLE apple_product_map
    ADD COLUMN IF NOT EXISTS gross_amount BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS platform_fee_pct NUMERIC(5,4) NOT NULL DEFAULT 0.30;

UPDATE apple_product_map SET gross_amount = 49000  WHERE product_id = 'com.sangiagao.premium.1m'  AND gross_amount = 0;
UPDATE apple_product_map SET gross_amount = 125000 WHERE product_id = 'com.sangiagao.premium.3m'  AND gross_amount = 0;
UPDATE apple_product_map SET gross_amount = 249000 WHERE product_id = 'com.sangiagao.premium.6m'  AND gross_amount = 0;
UPDATE apple_product_map SET gross_amount = 425000 WHERE product_id = 'com.sangiagao.premium.12m' AND gross_amount = 0;

-- 8. Seed default rule: 90d 50% / 180d 30% / perpetual 20%, NET base, 100k threshold
INSERT INTO commission_rules
    (referral_code_id, stage1_days, stage1_pct, stage2_days, stage2_pct, stage3_pct, base_type, minimum_payout)
SELECT NULL, 90, 0.5000, 180, 0.3000, 0.2000, 'net', 100000
WHERE NOT EXISTS (
    SELECT 1 FROM commission_rules
    WHERE referral_code_id IS NULL AND active_to IS NULL
);
