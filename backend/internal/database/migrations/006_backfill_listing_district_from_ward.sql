-- Backfill listing district from user ward (xã/phường)
-- listing.district stores the ward name from user profile
UPDATE listings SET district = u.ward
FROM users u
WHERE listings.user_id = u.id
  AND (listings.district IS NULL OR listings.district = '')
  AND u.ward IS NOT NULL AND u.ward <> '';
