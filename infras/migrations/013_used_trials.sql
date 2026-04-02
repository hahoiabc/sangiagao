-- Track phone_hash that already used free trial (prevent abuse via delete+re-register)
CREATE TABLE IF NOT EXISTS used_trials (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    phone_hash TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Backfill from existing free_trial subscriptions
INSERT INTO used_trials (phone_hash)
SELECT DISTINCT u.phone_hash
FROM subscriptions s
JOIN users u ON u.id = s.user_id
WHERE s.plan = 'free_trial' AND u.phone_hash IS NOT NULL
ON CONFLICT (phone_hash) DO NOTHING;
