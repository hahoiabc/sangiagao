-- Soft delete conversations per user (each side can hide independently)
ALTER TABLE conversations ADD COLUMN IF NOT EXISTS deleted_by_member BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE conversations ADD COLUMN IF NOT EXISTS deleted_by_seller BOOLEAN NOT NULL DEFAULT FALSE;
