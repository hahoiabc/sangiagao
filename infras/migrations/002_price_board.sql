-- Migration: Add category column to listings and create product_sponsors table

-- 1. Add category column to listings
ALTER TABLE listings ADD COLUMN IF NOT EXISTS category VARCHAR(50);

-- 2. Indexes for price board
CREATE INDEX IF NOT EXISTS idx_listings_category ON listings(category);
CREATE INDEX IF NOT EXISTS idx_listings_category_rice_type_price
    ON listings(category, rice_type, price_per_kg)
    WHERE status = 'active';

-- 3. Product sponsors table
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
