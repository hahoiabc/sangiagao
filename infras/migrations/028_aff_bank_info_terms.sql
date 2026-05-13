-- Migration 028: Bank info for affiliate payouts + T&C acceptance tracking +
-- transfer fee column on payouts. Implements Phase A (bank info) + Phase B (T&C).

-- 1. Per-aff bank info (1-1 with user, lazy-create when aff first opens form)
CREATE TABLE IF NOT EXISTS aff_bank_info (
    user_id      UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    account_no   VARCHAR(32)  NOT NULL,
    bank_name    VARCHAR(100) NOT NULL,
    holder_name  VARCHAR(120) NOT NULL,
    note         TEXT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 2. T&C acceptance — aff must accept once per terms version before payout works
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS aff_terms_accepted_at TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS aff_terms_version    VARCHAR(8)  NULL;

-- 3. Transfer fee deducted from payout (admin enters per payout, paid by aff)
ALTER TABLE payouts
    ADD COLUMN IF NOT EXISTS transfer_fee BIGINT NOT NULL DEFAULT 0 CHECK (transfer_fee >= 0);

COMMENT ON COLUMN payouts.transfer_fee IS 'VND. Bank transfer cost borne by aff partner — deducted from total_amount when displaying net received.';
