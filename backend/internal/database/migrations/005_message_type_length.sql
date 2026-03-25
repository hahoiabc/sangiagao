-- Increase type column length to accommodate 'listing_link' (12 chars)
ALTER TABLE messages ALTER COLUMN type TYPE VARCHAR(20);
