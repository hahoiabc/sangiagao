-- Marketplace browse: status + created_at for sorted pagination
CREATE INDEX IF NOT EXISTS idx_listings_status_created
  ON listings (status, created_at DESC) WHERE deleted_at IS NULL;

-- My listings: user_id + status for seller dashboard
CREATE INDEX IF NOT EXISTS idx_listings_user_status
  ON listings (user_id, status) WHERE deleted_at IS NULL;

-- Unread notifications: fast count + list for user
CREATE INDEX IF NOT EXISTS idx_notifications_unread
  ON notifications (user_id) WHERE is_read = false;

-- Pending reports: admin queue sorted by date
CREATE INDEX IF NOT EXISTS idx_reports_pending
  ON reports (created_at DESC) WHERE status = 'pending';
