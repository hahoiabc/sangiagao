-- Unread message count per conversation: speeds up inbox unread badge query
-- Used by: conversation_repo.go ListByUser subquery (sender_id <> $1 AND read_at IS NULL)
CREATE INDEX IF NOT EXISTS idx_messages_unread
  ON messages (conversation_id, sender_id) WHERE read_at IS NULL;

-- Subscription expiry cron: fast lookup of active subscriptions near/past expiry
-- Used by: subscription_repo.go ExpireOverdue, HideListingsForExpired
CREATE INDEX IF NOT EXISTS idx_subscriptions_active_expires
  ON subscriptions (status, expires_at) WHERE status = 'active';

-- Subscription lookup by user: GetActiveByUserID, GetByUserID
CREATE INDEX IF NOT EXISTS idx_subscriptions_user_active
  ON subscriptions (user_id, status) WHERE status = 'active';

-- Admin audit log table
CREATE TABLE IF NOT EXISTS admin_audit_logs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  admin_id TEXT NOT NULL REFERENCES users(id),
  action TEXT NOT NULL,
  target_type TEXT NOT NULL,
  target_id TEXT NOT NULL,
  details JSONB,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_admin
  ON admin_audit_logs (admin_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_audit_logs_target
  ON admin_audit_logs (target_type, target_id, created_at DESC);
