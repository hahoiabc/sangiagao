-- Migration 025: Enable unaccent extension for SEO price-board queries
-- Used in /api/v1/seo/listings to match province/rice_type slugs (ASCII)
-- against Vietnamese-accented names in listings table.

CREATE EXTENSION IF NOT EXISTS unaccent;
