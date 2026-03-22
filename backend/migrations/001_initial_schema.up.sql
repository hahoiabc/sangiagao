-- Rice Marketplace - Initial Schema
-- Version: 001

-- Enable extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";  -- For full-text search

-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    phone VARCHAR(15) NOT NULL UNIQUE,
    role VARCHAR(10) NOT NULL CHECK (role IN ('member', 'seller', 'admin')),
    name VARCHAR(100),
    avatar_url TEXT,
    address TEXT,
    province VARCHAR(50),
    description TEXT,
    org_name VARCHAR(200),
    is_blocked BOOLEAN NOT NULL DEFAULT FALSE,
    block_reason TEXT,
    accepted_tos_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_phone ON users(phone);
CREATE INDEX idx_users_role ON users(role);

-- Subscriptions table
CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    plan VARCHAR(20) NOT NULL CHECK (plan IN ('free_trial', 'paid')),
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    status VARCHAR(10) NOT NULL CHECK (status IN ('active', 'expired')) DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX idx_subscriptions_status ON subscriptions(status);
CREATE INDEX idx_subscriptions_expires_at ON subscriptions(expires_at);

-- Listings table
CREATE TABLE listings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(200) NOT NULL,
    rice_type VARCHAR(100) NOT NULL,
    province VARCHAR(50) NOT NULL,
    district VARCHAR(100),
    quantity_kg NUMERIC(12, 2) NOT NULL CHECK (quantity_kg > 0),
    price_per_kg NUMERIC(12, 0) NOT NULL CHECK (price_per_kg > 0),
    description TEXT,
    certifications TEXT,
    images JSONB NOT NULL DEFAULT '[]'::jsonb,
    status VARCHAR(30) NOT NULL CHECK (status IN ('active', 'hidden_subscription', 'deleted')) DEFAULT 'active',
    view_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_listings_user_id ON listings(user_id);
CREATE INDEX idx_listings_status ON listings(status);
CREATE INDEX idx_listings_rice_type ON listings(rice_type);
CREATE INDEX idx_listings_province ON listings(province);
CREATE INDEX idx_listings_price ON listings(price_per_kg);
CREATE INDEX idx_listings_created_at ON listings(created_at DESC);

-- Full-text search index
ALTER TABLE listings ADD COLUMN search_vector tsvector;
CREATE INDEX idx_listings_search ON listings USING gin(search_vector);

-- Auto-update search_vector on insert/update
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

CREATE TRIGGER listings_search_trigger
    BEFORE INSERT OR UPDATE OF title, rice_type, description, province
    ON listings
    FOR EACH ROW
    EXECUTE FUNCTION listings_search_update();

-- Ratings table
CREATE TABLE ratings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    reviewer_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    seller_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    stars SMALLINT NOT NULL CHECK (stars >= 1 AND stars <= 5),
    comment TEXT NOT NULL CHECK (char_length(comment) >= 10),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(reviewer_id, seller_id)
);

CREATE INDEX idx_ratings_seller_id ON ratings(seller_id);

-- Reports table
CREATE TABLE reports (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    reporter_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    target_type VARCHAR(10) NOT NULL CHECK (target_type IN ('listing', 'user', 'rating')),
    target_id UUID NOT NULL,
    reason VARCHAR(50) NOT NULL,
    description TEXT,
    status VARCHAR(15) NOT NULL CHECK (status IN ('pending', 'resolved', 'dismissed')) DEFAULT 'pending',
    admin_action VARCHAR(50),
    resolved_by UUID REFERENCES users(id),
    resolved_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_reports_status ON reports(status);
CREATE INDEX idx_reports_created_at ON reports(created_at DESC);

-- Notifications table
CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(30) NOT NULL,
    title VARCHAR(200) NOT NULL,
    body TEXT NOT NULL,
    data JSONB,
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notifications_user_id ON notifications(user_id);
CREATE INDEX idx_notifications_is_read ON notifications(user_id, is_read);
CREATE INDEX idx_notifications_created_at ON notifications(created_at DESC);

-- Device tokens for push notifications
CREATE TABLE device_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token TEXT NOT NULL,
    platform VARCHAR(10) NOT NULL CHECK (platform IN ('ios', 'android')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, token)
);

-- OTP table for rate limiting
CREATE TABLE otp_requests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    phone VARCHAR(15) NOT NULL,
    code VARCHAR(6) NOT NULL,
    attempts INTEGER NOT NULL DEFAULT 0,
    expires_at TIMESTAMPTZ NOT NULL,
    verified BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_otp_phone ON otp_requests(phone, created_at DESC);

-- Updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at() RETURNS trigger AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER listings_updated_at BEFORE UPDATE ON listings
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
