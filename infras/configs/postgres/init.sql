-- PostgreSQL initialization script
-- Runs automatically when container is created for the first time.

-- Enable extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    phone VARCHAR(15) NOT NULL UNIQUE,
    role VARCHAR(10) NOT NULL CHECK (role IN ('member', 'admin', 'editor', 'owner')),
    name VARCHAR(100),
    avatar_url TEXT,
    address TEXT,
    province VARCHAR(50),
    district VARCHAR(100),
    ward VARCHAR(100),
    description TEXT,
    org_name VARCHAR(200),
    is_blocked BOOLEAN NOT NULL DEFAULT FALSE,
    block_reason TEXT,
    password_hash TEXT,
    accepted_tos_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_phone ON users(phone);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
CREATE INDEX IF NOT EXISTS idx_users_phone_hash ON users(phone_hash);
CREATE INDEX IF NOT EXISTS idx_users_name_trgm ON users USING gin(name gin_trgm_ops);

-- Subscriptions table
CREATE TABLE IF NOT EXISTS subscriptions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    plan VARCHAR(20) NOT NULL CHECK (plan IN ('free_trial', 'paid')),
    duration_months INTEGER NOT NULL DEFAULT 1,
    amount BIGINT NOT NULL DEFAULT 0,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    status VARCHAR(10) NOT NULL CHECK (status IN ('active', 'expired')) DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_status ON subscriptions(status);
CREATE INDEX IF NOT EXISTS idx_subscriptions_expires_at ON subscriptions(expires_at);
CREATE INDEX IF NOT EXISTS idx_subscriptions_active_expiry ON subscriptions(user_id, expires_at) WHERE status = 'active';

-- Listings table
CREATE TABLE IF NOT EXISTS listings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(200) NOT NULL,
    rice_type VARCHAR(100) NOT NULL,
    province VARCHAR(50),
    district VARCHAR(100),
    quantity_kg NUMERIC(12, 2) NOT NULL CHECK (quantity_kg > 0),
    price_per_kg NUMERIC(12, 0) NOT NULL CHECK (price_per_kg > 0),
    harvest_season VARCHAR(50),
    description TEXT,
    certifications TEXT,
    images JSONB NOT NULL DEFAULT '[]'::jsonb,
    status VARCHAR(30) NOT NULL CHECK (status IN ('active', 'hidden_subscription', 'deleted')) DEFAULT 'active',
    view_count INTEGER NOT NULL DEFAULT 0,
    category VARCHAR(50),
    search_vector tsvector,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_listings_user_id ON listings(user_id);
CREATE INDEX IF NOT EXISTS idx_listings_status ON listings(status);
CREATE INDEX IF NOT EXISTS idx_listings_rice_type ON listings(rice_type);
CREATE INDEX IF NOT EXISTS idx_listings_province ON listings(province);
CREATE INDEX IF NOT EXISTS idx_listings_price ON listings(price_per_kg);
CREATE INDEX IF NOT EXISTS idx_listings_created_at ON listings(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_listings_search ON listings USING gin(search_vector);

-- Full-text search trigger
CREATE OR REPLACE FUNCTION listings_search_update() RETURNS trigger AS $$
BEGIN
    NEW.search_vector :=
        to_tsvector('simple', COALESCE(NEW.title, '')) ||
        to_tsvector('simple', COALESCE(NEW.rice_type, '')) ||
        to_tsvector('simple', COALESCE(NEW.description, '')) ||
        to_tsvector('simple', COALESCE(NEW.province, ''));
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS listings_search_trigger ON listings;
CREATE TRIGGER listings_search_trigger
    BEFORE INSERT OR UPDATE OF title, rice_type, description, province
    ON listings FOR EACH ROW EXECUTE FUNCTION listings_search_update();

-- Ratings table
CREATE TABLE IF NOT EXISTS ratings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    reviewer_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    seller_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    stars SMALLINT NOT NULL CHECK (stars >= 1 AND stars <= 5),
    comment TEXT DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(reviewer_id, seller_id)
);

CREATE INDEX IF NOT EXISTS idx_ratings_seller_id ON ratings(seller_id);

-- Reports table
CREATE TABLE IF NOT EXISTS reports (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    reporter_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    target_type VARCHAR(10) NOT NULL CHECK (target_type IN ('listing', 'user', 'rating')),
    target_id UUID NOT NULL,
    reason VARCHAR(50) NOT NULL,
    description TEXT,
    status VARCHAR(15) NOT NULL CHECK (status IN ('pending', 'resolved', 'dismissed')) DEFAULT 'pending',
    admin_action VARCHAR(50),
    admin_note TEXT,
    resolved_by UUID REFERENCES users(id),
    resolved_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_reports_status ON reports(status);
CREATE INDEX IF NOT EXISTS idx_reports_created_at ON reports(created_at DESC);

-- Notifications table
CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(30) NOT NULL,
    title VARCHAR(200) NOT NULL,
    body TEXT NOT NULL,
    data JSONB,
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_is_read ON notifications(user_id, is_read);
CREATE INDEX IF NOT EXISTS idx_notifications_created_at ON notifications(created_at DESC);

-- Device tokens
CREATE TABLE IF NOT EXISTS device_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token TEXT NOT NULL,
    platform VARCHAR(10) NOT NULL CHECK (platform IN ('ios', 'android')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, token)
);

CREATE INDEX IF NOT EXISTS idx_device_tokens_user_id ON device_tokens(user_id);

-- System Inbox (1 record = 1 thông báo, nhiều user cùng đọc)
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

CREATE TABLE IF NOT EXISTS inbox_read_status (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    inbox_id UUID NOT NULL REFERENCES system_inbox(id) ON DELETE CASCADE,
    read_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, inbox_id)
);

-- OTP requests
CREATE TABLE IF NOT EXISTS otp_requests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    phone VARCHAR(15) NOT NULL,
    code VARCHAR(6) NOT NULL,
    attempts INTEGER NOT NULL DEFAULT 0,
    expires_at TIMESTAMPTZ NOT NULL,
    verified BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_otp_phone ON otp_requests(phone, created_at DESC);

-- Auto update updated_at
CREATE OR REPLACE FUNCTION update_updated_at() RETURNS trigger AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS users_updated_at ON users;
CREATE TRIGGER users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

DROP TRIGGER IF EXISTS listings_updated_at ON listings;
CREATE TRIGGER listings_updated_at BEFORE UPDATE ON listings
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- Conversations table
CREATE TABLE IF NOT EXISTS conversations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    member_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    seller_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    listing_id UUID REFERENCES listings(id) ON DELETE SET NULL,
    last_message_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(member_id, seller_id, listing_id)
);

CREATE INDEX IF NOT EXISTS idx_conversations_member_id ON conversations(member_id);
CREATE INDEX IF NOT EXISTS idx_conversations_seller_id ON conversations(seller_id);
CREATE INDEX IF NOT EXISTS idx_conversations_last_message ON conversations(last_message_at DESC);
CREATE INDEX IF NOT EXISTS idx_conversations_participants ON conversations(LEAST(member_id, seller_id), GREATEST(member_id, seller_id));

-- Messages table
CREATE TABLE IF NOT EXISTS messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    sender_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    type VARCHAR(20) NOT NULL DEFAULT 'text' CHECK (type IN ('text', 'image', 'audio', 'recalled', 'listing_link')),
    read_at TIMESTAMPTZ,
    deleted_by_sender BOOLEAN NOT NULL DEFAULT false,
    deleted_by_receiver BOOLEAN NOT NULL DEFAULT false,
    reply_to_id UUID REFERENCES messages(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_messages_conversation_id ON messages(conversation_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_messages_sender_id ON messages(sender_id);
CREATE INDEX IF NOT EXISTS idx_messages_unread ON messages(conversation_id, read_at) WHERE read_at IS NULL;

-- Message reactions table
CREATE TABLE IF NOT EXISTS message_reactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    message_id UUID NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    emoji VARCHAR(10) NOT NULL DEFAULT '👍',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(message_id, user_id, emoji)
);

CREATE INDEX IF NOT EXISTS idx_message_reactions_message ON message_reactions(message_id);
CREATE INDEX IF NOT EXISTS idx_message_reactions_user ON message_reactions(user_id);

-- Rice categories table
CREATE TABLE IF NOT EXISTS rice_categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    key VARCHAR(50) NOT NULL UNIQUE,
    label VARCHAR(100) NOT NULL,
    icon VARCHAR(50) DEFAULT 'category',
    sort_order INTEGER NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_rice_categories_active ON rice_categories(is_active, sort_order);

-- Rice products table
CREATE TABLE IF NOT EXISTS rice_products (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    key VARCHAR(50) NOT NULL UNIQUE,
    label VARCHAR(100) NOT NULL,
    category_id UUID NOT NULL REFERENCES rice_categories(id) ON DELETE CASCADE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_rice_products_category ON rice_products(category_id);
CREATE INDEX IF NOT EXISTS idx_rice_products_active ON rice_products(is_active, sort_order);

-- Auto update triggers for catalog
DROP TRIGGER IF EXISTS rice_categories_updated_at ON rice_categories;
CREATE TRIGGER rice_categories_updated_at BEFORE UPDATE ON rice_categories
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

DROP TRIGGER IF EXISTS rice_products_updated_at ON rice_products;
CREATE TRIGGER rice_products_updated_at BEFORE UPDATE ON rice_products
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- Product sponsors table
CREATE TABLE IF NOT EXISTS product_sponsors (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    product_key VARCHAR(50) NOT NULL UNIQUE,
    logo_url TEXT NOT NULL,
    sponsor_name VARCHAR(200),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_product_sponsors_active ON product_sponsors(is_active);

-- Feedbacks table
CREATE TABLE IF NOT EXISTS feedbacks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id),
    content TEXT NOT NULL,
    reply TEXT,
    replied_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_feedbacks_user_id ON feedbacks(user_id);
CREATE INDEX IF NOT EXISTS idx_feedbacks_created_at ON feedbacks(created_at DESC);

-- Seed admin user
INSERT INTO users (phone, role, name, accepted_tos_at)
VALUES ('0900000000', 'admin', 'Admin', NOW())
ON CONFLICT (phone) DO NOTHING;

-- Seed catalog data
INSERT INTO rice_categories (key, label, icon, sort_order) VALUES
('gao_deo_thom', 'Gạo dẻo thơm', 'rice_bowl', 1),
('gao_kho', 'Gạo khô', 'grass', 2),
('tam_deo_thom', 'Tấm dẻo thơm', 'grain', 3),
('tam_kho', 'Tấm khô', 'scatter_plot', 4),
('nep', 'Nếp', 'spa', 5)
ON CONFLICT (key) DO NOTHING;

INSERT INTO rice_products (key, label, category_id, sort_order)
SELECT p.key, p.label, c.id, p.sort_order FROM (VALUES
    ('st_25', 'ST 25', 1), ('st_24', 'ST 24', 2), ('st_21', 'ST 21', 3),
    ('om_18', 'OM 18', 4), ('om_49', 'OM 49', 5), ('om_5451', 'OM 5451', 6),
    ('dai_thom_8', 'Đài Thơm 8', 7), ('om_6976', 'OM 6976', 8),
    ('nhat', 'Nhật', 9), ('lien_huong', 'Liên Hương', 10),
    ('mien', 'Miên', 11), ('dai_loan', 'Đài Loan', 12)
) AS p(key, label, sort_order)
CROSS JOIN rice_categories c WHERE c.key = 'gao_deo_thom'
ON CONFLICT (key) DO NOTHING;

INSERT INTO rice_products (key, label, category_id, sort_order)
SELECT p.key, p.label, c.id, p.sort_order FROM (VALUES
    ('tai_nguyen', 'Tài Nguyên', 1), ('soc', 'Sóc', 2), ('so_ri', 'Sơ Ri', 3),
    ('mong_chim', 'Móng Chim', 4), ('ham_chau_sieu', 'Hàm Châu siêu', 5),
    ('ir_504', 'IR 504', 6), ('q5', 'Q5', 7), ('an_no', 'Ấn nở', 8),
    ('myanmar', 'Myanmar', 9)
) AS p(key, label, sort_order)
CROSS JOIN rice_categories c WHERE c.key = 'gao_kho'
ON CONFLICT (key) DO NOTHING;

INSERT INTO rice_products (key, label, category_id, sort_order)
SELECT p.key, p.label, c.id, p.sort_order FROM (VALUES
    ('tam_st_25', 'Tấm ST 25', 1), ('tam_st_24', 'Tấm ST 24', 2), ('tam_st_21', 'Tấm ST 21', 3),
    ('tam_om_18', 'Tấm OM 18', 4), ('tam_om_49', 'Tấm OM 49', 5), ('tam_om_5451', 'Tấm OM 5451', 6),
    ('tam_dai_thom_8', 'Tấm Đài Thơm 8', 7), ('tam_om_6976', 'Tấm OM 6976', 8),
    ('tam_nhat', 'Tấm Nhật', 9), ('tam_lien_huong', 'Tấm Liên Hương', 10),
    ('tam_mien', 'Tấm Miên', 11), ('tam_dai_loan', 'Tấm Đài Loan', 12)
) AS p(key, label, sort_order)
CROSS JOIN rice_categories c WHERE c.key = 'tam_deo_thom'
ON CONFLICT (key) DO NOTHING;

INSERT INTO rice_products (key, label, category_id, sort_order)
SELECT p.key, p.label, c.id, p.sort_order FROM (VALUES
    ('tam_tai_nguyen', 'Tấm Tài Nguyên', 1), ('tam_soc', 'Tấm Sóc', 2), ('tam_so_ri', 'Tấm Sơ Ri', 3),
    ('tam_mong_chim', 'Tấm Móng Chim', 4), ('tam_ham_chau_sieu', 'Tấm Hàm Châu siêu', 5),
    ('tam_ir_504', 'Tấm IR 504', 6), ('tam_q5', 'Tấm Q5', 7), ('tam_an_no', 'Tấm Ấn nở', 8),
    ('tam_myanmar', 'Tấm Myanmar', 9)
) AS p(key, label, sort_order)
CROSS JOIN rice_categories c WHERE c.key = 'tam_kho'
ON CONFLICT (key) DO NOTHING;

INSERT INTO rice_products (key, label, category_id, sort_order)
SELECT p.key, p.label, c.id, p.sort_order FROM (VALUES
    ('sap_moi', 'Sáp Mới', 1), ('sap_cu', 'Sáp cũ', 2),
    ('nep_la_moi', 'Nếp Lá mới', 3), ('nep_la_cu', 'Nếp Lá cũ', 4),
    ('bac_hat_lon', 'Bắc Hạt Lớn', 5), ('bac_hat_nho', 'Bắc Hạt Nhỏ', 6),
    ('nep_thai', 'Nếp Thái', 7), ('nep_than', 'Nếp Than', 8),
    ('huyet_rong', 'Huyết Rồng', 9)
) AS p(key, label, sort_order)
CROSS JOIN rice_categories c WHERE c.key = 'nep'
ON CONFLICT (key) DO NOTHING;

-- Subscription plans table
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

DROP TRIGGER IF EXISTS subscription_plans_updated_at ON subscription_plans;
CREATE TRIGGER subscription_plans_updated_at BEFORE UPDATE ON subscription_plans
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- Seed default plans
INSERT INTO subscription_plans (months, amount, label, sort_order) VALUES
    (1, 35000, '1 tháng', 1),
    (3, 96000, '3 tháng', 2),
    (6, 180000, '6 tháng', 3),
    (12, 300000, '12 tháng', 4)
ON CONFLICT (months) DO NOTHING;

-- Role permissions table
CREATE TABLE IF NOT EXISTS role_permissions (
    role VARCHAR(20) NOT NULL,
    permission_key VARCHAR(50) NOT NULL,
    allowed BOOLEAN NOT NULL DEFAULT false,
    PRIMARY KEY (role, permission_key)
);

CREATE INDEX IF NOT EXISTS idx_role_permissions_role ON role_permissions(role);

-- Seed default permissions
INSERT INTO role_permissions (role, permission_key, allowed) VALUES
    -- Owner: all permissions
    ('owner', 'dashboard.view', true), ('owner', 'dashboard.charts', true), ('owner', 'system.monitor', true),
    ('owner', 'users.list', true), ('owner', 'users.detail', true), ('owner', 'users.block', true),
    ('owner', 'users.role', true), ('owner', 'users.batch_block', true),
    ('owner', 'listings.create', true), ('owner', 'listings.edit_own', true),
    ('owner', 'listings.delete_any', true), ('owner', 'listings.batch_delete', true),
    ('owner', 'marketplace.browse', true), ('owner', 'marketplace.search', true),
    ('owner', 'marketplace.detail', true), ('owner', 'marketplace.priceboard', true),
    ('owner', 'marketplace.seller_profile', true),
    ('owner', 'chat.send', true), ('owner', 'chat.send_image', true), ('owner', 'ratings.create', true),
    ('owner', 'reports.create', true), ('owner', 'reports.manage', true),
    ('owner', 'sub.activate', true), ('owner', 'sub.revenue', true), ('owner', 'sub.plans', true),
    ('owner', 'catalog.manage', true), ('owner', 'sponsors.manage', true),
    ('owner', 'feedback.create', true), ('owner', 'feedback.reply', true),
    ('owner', 'notifications.broadcast', true), ('owner', 'notifications.send_individual', true),
    -- Admin
    ('admin', 'dashboard.view', true), ('admin', 'dashboard.charts', true), ('admin', 'system.monitor', true),
    ('admin', 'users.list', true), ('admin', 'users.detail', true), ('admin', 'users.block', true),
    ('admin', 'users.role', true), ('admin', 'users.batch_block', true),
    ('admin', 'listings.create', true), ('admin', 'listings.edit_own', true),
    ('admin', 'listings.delete_any', true), ('admin', 'listings.batch_delete', true),
    ('admin', 'marketplace.browse', true), ('admin', 'marketplace.search', true),
    ('admin', 'marketplace.detail', true), ('admin', 'marketplace.priceboard', true),
    ('admin', 'marketplace.seller_profile', true),
    ('admin', 'chat.send', true), ('admin', 'chat.send_image', true), ('admin', 'ratings.create', true),
    ('admin', 'reports.create', true), ('admin', 'reports.manage', true),
    ('admin', 'sub.activate', true), ('admin', 'sub.revenue', true), ('admin', 'sub.plans', false),
    ('admin', 'catalog.manage', true), ('admin', 'sponsors.manage', true),
    ('admin', 'feedback.create', true), ('admin', 'feedback.reply', true),
    ('admin', 'notifications.broadcast', true), ('admin', 'notifications.send_individual', true),
    -- Editor
    ('editor', 'dashboard.view', true), ('editor', 'dashboard.charts', true), ('editor', 'system.monitor', true),
    ('editor', 'users.list', false), ('editor', 'users.detail', false), ('editor', 'users.block', false),
    ('editor', 'users.role', false), ('editor', 'users.batch_block', false),
    ('editor', 'listings.create', true), ('editor', 'listings.edit_own', true),
    ('editor', 'listings.delete_any', true), ('editor', 'listings.batch_delete', true),
    ('editor', 'marketplace.browse', true), ('editor', 'marketplace.search', true),
    ('editor', 'marketplace.detail', true), ('editor', 'marketplace.priceboard', true),
    ('editor', 'marketplace.seller_profile', true),
    ('editor', 'chat.send', true), ('editor', 'chat.send_image', true), ('editor', 'ratings.create', true),
    ('editor', 'reports.create', true), ('editor', 'reports.manage', true),
    ('editor', 'sub.activate', true), ('editor', 'sub.revenue', true), ('editor', 'sub.plans', false),
    ('editor', 'catalog.manage', true), ('editor', 'sponsors.manage', true),
    ('editor', 'feedback.create', true), ('editor', 'feedback.reply', true),
    ('editor', 'notifications.broadcast', true), ('editor', 'notifications.send_individual', false),
    -- Member
    ('member', 'dashboard.view', false), ('member', 'dashboard.charts', false), ('member', 'system.monitor', false),
    ('member', 'users.list', false), ('member', 'users.detail', false), ('member', 'users.block', false),
    ('member', 'users.role', false), ('member', 'users.batch_block', false),
    ('member', 'listings.create', true), ('member', 'listings.edit_own', true),
    ('member', 'listings.delete_any', false), ('member', 'listings.batch_delete', false),
    ('member', 'marketplace.browse', true), ('member', 'marketplace.search', true),
    ('member', 'marketplace.detail', true), ('member', 'marketplace.priceboard', true),
    ('member', 'marketplace.seller_profile', true),
    ('member', 'chat.send', true), ('member', 'chat.send_image', true), ('member', 'ratings.create', true),
    ('member', 'reports.create', true), ('member', 'reports.manage', false),
    ('member', 'sub.activate', false), ('member', 'sub.revenue', false), ('member', 'sub.plans', false),
    ('member', 'catalog.manage', false), ('member', 'sponsors.manage', false),
    ('member', 'feedback.create', true), ('member', 'feedback.reply', false),
    -- Expired member
    ('expired', 'dashboard.view', false), ('expired', 'dashboard.charts', false), ('expired', 'system.monitor', false),
    ('expired', 'users.list', false), ('expired', 'users.detail', false), ('expired', 'users.block', false),
    ('expired', 'users.role', false), ('expired', 'users.batch_block', false),
    ('expired', 'listings.create', false), ('expired', 'listings.edit_own', false),
    ('expired', 'listings.delete_any', false), ('expired', 'listings.batch_delete', false),
    ('expired', 'marketplace.browse', true), ('expired', 'marketplace.search', true),
    ('expired', 'marketplace.detail', true), ('expired', 'marketplace.priceboard', true),
    ('expired', 'marketplace.seller_profile', true),
    ('expired', 'chat.send', false), ('expired', 'chat.send_image', false), ('expired', 'ratings.create', true),
    ('expired', 'reports.create', true), ('expired', 'reports.manage', false),
    ('expired', 'sub.activate', false), ('expired', 'sub.revenue', false), ('expired', 'sub.plans', false),
    ('expired', 'catalog.manage', false), ('expired', 'sponsors.manage', false),
    ('expired', 'feedback.create', true), ('expired', 'feedback.reply', false),
    -- Guest (not logged in)
    ('guest', 'dashboard.view', false), ('guest', 'dashboard.charts', false), ('guest', 'system.monitor', false),
    ('guest', 'users.list', false), ('guest', 'users.detail', false), ('guest', 'users.block', false),
    ('guest', 'users.role', false), ('guest', 'users.batch_block', false),
    ('guest', 'listings.create', false), ('guest', 'listings.edit_own', false),
    ('guest', 'listings.delete_any', false), ('guest', 'listings.batch_delete', false),
    ('guest', 'marketplace.browse', true), ('guest', 'marketplace.search', true),
    ('guest', 'marketplace.detail', true), ('guest', 'marketplace.priceboard', true),
    ('guest', 'marketplace.seller_profile', true),
    ('guest', 'chat.send', false), ('guest', 'chat.send_image', false), ('guest', 'ratings.create', false),
    ('guest', 'reports.create', false), ('guest', 'reports.manage', false),
    ('guest', 'sub.activate', false), ('guest', 'sub.revenue', false), ('guest', 'sub.plans', false),
    ('guest', 'catalog.manage', false), ('guest', 'sponsors.manage', false),
    ('guest', 'feedback.create', false), ('guest', 'feedback.reply', false)
ON CONFLICT (role, permission_key) DO NOTHING;

-- Done
DO $$ BEGIN RAISE NOTICE '✅ Rice Marketplace database initialized successfully!'; END $$;
