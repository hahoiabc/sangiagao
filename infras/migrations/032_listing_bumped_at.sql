-- "Làm mới tin đăng" + Unaccent search + ranking content-based.
-- Cooldown 5h54m (5.9h) giữa 2 lần bump, lifetime cap 240 lần / tin.
-- bumped_at NULL cho tin cũ → ranking dùng GREATEST(created_at, updated_at, COALESCE(bumped_at, '1970-01-01'))
-- bump_count tracking cho lifetime quota — 60 ngày × 4 bump/ngày = 240 lần.

ALTER TABLE listings
  ADD COLUMN IF NOT EXISTS bumped_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS bump_count INTEGER NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_listings_bumped_at
  ON listings(bumped_at DESC)
  WHERE bumped_at IS NOT NULL;

-- Rebuild full-text search trigger với unaccent — user gõ không dấu ("gao st thom")
-- vẫn match được tin có dấu ("Gạo ST25 thơm"). Extension unaccent đã có từ mig 025.
CREATE OR REPLACE FUNCTION listings_search_update() RETURNS trigger AS $$
BEGIN
    NEW.search_vector :=
        to_tsvector('simple', unaccent(COALESCE(NEW.title, ''))) ||
        to_tsvector('simple', unaccent(COALESCE(NEW.rice_type, ''))) ||
        to_tsvector('simple', unaccent(COALESCE(NEW.description, ''))) ||
        to_tsvector('simple', unaccent(COALESCE(NEW.province, '')));
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Backfill: trigger UPDATE để regenerate search_vector cho tin cũ.
-- WHERE clause vô hại — chỉ trigger BEFORE UPDATE OF (title, rice_type, description, province)
-- nên cần UPDATE 1 trong 4 cột đó. Set title = title là no-op nhưng kích hoạt trigger.
UPDATE listings SET title = title WHERE status != 'deleted';
