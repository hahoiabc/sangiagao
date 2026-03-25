-- Migration 011: Create subscription_plans table for dynamic plan management
-- Owner can edit plans from admin panel

CREATE TABLE IF NOT EXISTS subscription_plans (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    months INTEGER NOT NULL UNIQUE CHECK (months > 0),
    amount BIGINT NOT NULL CHECK (amount >= 0),
    label VARCHAR(100) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Seed default plans
INSERT INTO subscription_plans (months, amount, label, sort_order) VALUES
    (1, 35000, '1 tháng', 1),
    (3, 96000, '3 tháng', 2),
    (6, 180000, '6 tháng', 3),
    (12, 300000, '12 tháng', 4)
ON CONFLICT (months) DO NOTHING;
