-- System Inbox: 1 record = 1 thông báo từ admin, nhiều user cùng đọc
CREATE TABLE IF NOT EXISTS system_inbox (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(200) NOT NULL,
    body TEXT NOT NULL,
    image_url TEXT,
    target VARCHAR(30) NOT NULL DEFAULT 'all_users',
    is_pinned BOOLEAN NOT NULL DEFAULT FALSE,
    expires_at TIMESTAMPTZ,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_system_inbox_target ON system_inbox(target);
CREATE INDEX IF NOT EXISTS idx_system_inbox_created_at ON system_inbox(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_system_inbox_pinned ON system_inbox(is_pinned, created_at DESC);

-- Read status: chỉ tạo khi user đọc (lazy tracking)
CREATE TABLE IF NOT EXISTS inbox_read_status (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    inbox_id UUID NOT NULL REFERENCES system_inbox(id) ON DELETE CASCADE,
    read_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, inbox_id)
);
