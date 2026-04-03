-- Payment orders for SePay integration
CREATE TABLE IF NOT EXISTS payment_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    plan_months INTEGER NOT NULL,
    amount BIGINT NOT NULL,
    order_code VARCHAR(20) NOT NULL UNIQUE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'paid', 'expired', 'cancelled')),
    sepay_transaction_id BIGINT,
    paid_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_payment_orders_user ON payment_orders(user_id, created_at DESC);
CREATE INDEX idx_payment_orders_code ON payment_orders(order_code) WHERE status = 'pending';
CREATE INDEX idx_payment_orders_status ON payment_orders(status) WHERE status = 'pending';
