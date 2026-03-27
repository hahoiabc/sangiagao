-- Prevent duplicate reports: same reporter can only report same target once while pending
CREATE UNIQUE INDEX IF NOT EXISTS idx_reports_unique_pending
ON reports (reporter_id, target_type, target_id)
WHERE status = 'pending';
