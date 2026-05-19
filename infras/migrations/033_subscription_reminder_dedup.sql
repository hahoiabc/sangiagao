-- Dedup cho thông báo nhắc gia hạn — chống spam.
-- Trước fix: cron mỗi giờ + KHÔNG track "đã gửi" → user nhận 24+ thông báo trùng
-- trong 1 ngày khi sub sắp hết hạn (1 lần/giờ × 24h).
--
-- Fix: thêm cột reminder_sent_at. Cron chỉ gửi nếu:
--   reminder_sent_at IS NULL OR reminder_sent_at < NOW() - INTERVAL '24 hours'
-- → tối đa 1 reminder / sub / 24 giờ.

ALTER TABLE subscriptions
  ADD COLUMN IF NOT EXISTS reminder_sent_at TIMESTAMPTZ;

-- Backfill: set cho subs đang active = NOW() để dừng spam ngay lập tức cho
-- user đang bị ảnh hưởng. Sau đó cron sẽ chờ 24h mới gửi reminder mới.
UPDATE subscriptions
   SET reminder_sent_at = NOW()
 WHERE status = 'active'
   AND expires_at > NOW()
   AND expires_at <= NOW() + INTERVAL '72 hours';

CREATE INDEX IF NOT EXISTS idx_subscriptions_reminder_pending
  ON subscriptions(expires_at)
  WHERE status = 'active' AND reminder_sent_at IS NULL;
