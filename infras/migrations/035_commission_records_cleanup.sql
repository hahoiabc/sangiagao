-- Cleanup commission_records sau khi đổi sang payment-count model (mig 034).
-- Dự án còn ở demo, chưa có aff thật nên xoá sạch records cũ + drop field
-- deprecated + thêm unique index chống race condition concurrent webhook.
--
-- quick-deploy.sh apply mọi .sql mỗi lần deploy — wrap TRUNCATE trong guard
-- để chỉ chạy 1 lần (khi column referee_age_days vẫn còn).

-- 1. Clear data demo (one-shot — guard bằng column tồn tại)
DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_name = 'commission_records'
      AND column_name = 'referee_age_days'
  ) THEN
    TRUNCATE TABLE commission_records;
  END IF;
END $$;

-- 2. Drop deprecated column. Đã chuyển sang đếm payment_sequence,
-- referee_age_days không còn ý nghĩa.
ALTER TABLE commission_records
  DROP COLUMN IF EXISTS referee_age_days;

-- 3. Unique index trên (referrer, referee, payment_sequence) — chống 2 webhook
-- concurrent (Apple retry + SePay parallel) cùng compute paymentSequence=N+1
-- → cùng INSERT → 1 thằng vi phạm constraint, engine fail gracefully (đã có
-- SELECT FOR UPDATE serialize ở layer ứng dụng, đây là defense layer 2).
-- Partial WHERE để cancelled records không chiếm slot.
CREATE UNIQUE INDEX IF NOT EXISTS commission_records_seq_unique
  ON commission_records (referrer_user_id, referee_user_id, payment_sequence)
  WHERE status != 'cancelled';
