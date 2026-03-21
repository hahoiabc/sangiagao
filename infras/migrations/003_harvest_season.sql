-- Migration: Add harvest_season column, make province optional

ALTER TABLE listings ADD COLUMN IF NOT EXISTS harvest_season VARCHAR(50);
ALTER TABLE listings ALTER COLUMN province DROP NOT NULL;
