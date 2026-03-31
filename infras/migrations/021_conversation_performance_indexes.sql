-- Composite indexes for fast ListByUser query (filter soft-deleted + ORDER BY last_message_at)
CREATE INDEX IF NOT EXISTS idx_conversations_member_inbox
    ON conversations (member_id, last_message_at DESC)
    WHERE deleted_by_member = FALSE;

CREATE INDEX IF NOT EXISTS idx_conversations_seller_inbox
    ON conversations (seller_id, last_message_at DESC)
    WHERE deleted_by_seller = FALSE;

-- Index for unread count subquery (messages not read by recipient)
CREATE INDEX IF NOT EXISTS idx_messages_unread
    ON messages (conversation_id, sender_id)
    WHERE read_at IS NULL;
