-- =============================================================
-- Seed subscription revenue test data
-- Uses existing seller accounts (0903000001-0903000060)
-- Creates subscriptions across the last 12 months with various plans
-- Idempotent: deletes existing test subs first
-- =============================================================

BEGIN;

-- Clear existing subscriptions to start fresh
DELETE FROM subscriptions WHERE user_id IN (SELECT id FROM users WHERE phone LIKE '0903%');

-- ========================
-- Month 1: 10 months ago — 3 subscriptions (early adopters)
-- ========================
INSERT INTO subscriptions (user_id, plan, duration_months, amount, started_at, expires_at, status, created_at)
SELECT id, 'paid', 1, 35000, NOW() - interval '10 months', NOW() - interval '9 months', 'expired', NOW() - interval '10 months'
FROM users WHERE phone = '0903000001';

INSERT INTO subscriptions (user_id, plan, duration_months, amount, started_at, expires_at, status, created_at)
SELECT id, 'paid', 1, 35000, NOW() - interval '10 months', NOW() - interval '9 months', 'expired', NOW() - interval '10 months'
FROM users WHERE phone = '0903000002';

INSERT INTO subscriptions (user_id, plan, duration_months, amount, started_at, expires_at, status, created_at)
SELECT id, 'free_trial', 0, 0, NOW() - interval '10 months', NOW() - interval '9 months', 'expired', NOW() - interval '10 months'
FROM users WHERE phone = '0903000003';

-- ========================
-- Month 2: 9 months ago — 5 subscriptions
-- ========================
INSERT INTO subscriptions (user_id, plan, duration_months, amount, started_at, expires_at, status, created_at)
SELECT id, 'paid', 3, 96000, NOW() - interval '9 months', NOW() - interval '6 months', 'expired', NOW() - interval '9 months'
FROM users WHERE phone = '0903000001';

INSERT INTO subscriptions (user_id, plan, duration_months, amount, started_at, expires_at, status, created_at)
SELECT id, 'paid', 1, 35000, NOW() - interval '9 months', NOW() - interval '8 months', 'expired', NOW() - interval '9 months'
FROM users WHERE phone = '0903000004';

INSERT INTO subscriptions (user_id, plan, duration_months, amount, started_at, expires_at, status, created_at)
SELECT id, 'paid', 1, 35000, NOW() - interval '9 months', NOW() - interval '8 months', 'expired', NOW() - interval '9 months'
FROM users WHERE phone = '0903000005';

INSERT INTO subscriptions (user_id, plan, duration_months, amount, started_at, expires_at, status, created_at)
SELECT id, 'free_trial', 0, 0, NOW() - interval '9 months', NOW() - interval '8 months', 'expired', NOW() - interval '9 months'
FROM users WHERE phone = '0903000006';

INSERT INTO subscriptions (user_id, plan, duration_months, amount, started_at, expires_at, status, created_at)
SELECT id, 'free_trial', 0, 0, NOW() - interval '9 months', NOW() - interval '8 months', 'expired', NOW() - interval '9 months'
FROM users WHERE phone = '0903000007';

-- ========================
-- Month 3: 8 months ago — 6 subscriptions
-- ========================
INSERT INTO subscriptions (user_id, plan, duration_months, amount, started_at, expires_at, status, created_at)
SELECT id, 'paid', 6, 180000, NOW() - interval '8 months', NOW() - interval '2 months', 'expired', NOW() - interval '8 months'
FROM users WHERE phone = '0903000002';

INSERT INTO subscriptions (user_id, plan, duration_months, amount, started_at, expires_at, status, created_at)
SELECT id, 'paid', 1, 35000, NOW() - interval '8 months', NOW() - interval '7 months', 'expired', NOW() - interval '8 months'
FROM users WHERE phone = '0903000008';

INSERT INTO subscriptions (user_id, plan, duration_months, amount, started_at, expires_at, status, created_at)
SELECT id, 'paid', 3, 96000, NOW() - interval '8 months', NOW() - interval '5 months', 'expired', NOW() - interval '8 months'
FROM users WHERE phone = '0903000009';

INSERT INTO subscriptions (user_id, plan, duration_months, amount, started_at, expires_at, status, created_at)
SELECT id, 'paid', 1, 35000, NOW() - interval '8 months', NOW() - interval '7 months', 'expired', NOW() - interval '8 months'
FROM users WHERE phone = '0903000010';

INSERT INTO subscriptions (user_id, plan, duration_months, amount, started_at, expires_at, status, created_at)
SELECT id, 'free_trial', 0, 0, NOW() - interval '8 months', NOW() - interval '7 months', 'expired', NOW() - interval '8 months'
FROM users WHERE phone = '0903000011';

INSERT INTO subscriptions (user_id, plan, duration_months, amount, started_at, expires_at, status, created_at)
SELECT id, 'free_trial', 0, 0, NOW() - interval '8 months', NOW() - interval '7 months', 'expired', NOW() - interval '8 months'
FROM users WHERE phone = '0903000012';

-- ========================
-- Month 4: 7 months ago — 8 subscriptions (growing)
-- ========================
INSERT INTO subscriptions (user_id, plan, duration_months, amount, started_at, expires_at, status, created_at)
SELECT id, 'paid', 1, 35000, NOW() - interval '7 months', NOW() - interval '6 months', 'expired', NOW() - interval '7 months'
FROM users WHERE phone IN ('0903000004', '0903000005', '0903000013', '0903000014');

INSERT INTO subscriptions (user_id, plan, duration_months, amount, started_at, expires_at, status, created_at)
SELECT id, 'paid', 3, 96000, NOW() - interval '7 months', NOW() - interval '4 months', 'expired', NOW() - interval '7 months'
FROM users WHERE phone IN ('0903000015', '0903000016');

INSERT INTO subscriptions (user_id, plan, duration_months, amount, started_at, expires_at, status, created_at)
SELECT id, 'free_trial', 0, 0, NOW() - interval '7 months', NOW() - interval '6 months', 'expired', NOW() - interval '7 months'
FROM users WHERE phone IN ('0903000017', '0903000018');

-- ========================
-- Month 5: 6 months ago — 10 subscriptions
-- ========================
INSERT INTO subscriptions (user_id, plan, duration_months, amount, started_at, expires_at, status, created_at)
SELECT id, 'paid', 6, 180000, NOW() - interval '6 months', NOW(), 'expired', NOW() - interval '6 months'
FROM users WHERE phone IN ('0903000004', '0903000005');

INSERT INTO subscriptions (user_id, plan, duration_months, amount, started_at, expires_at, status, created_at)
SELECT id, 'paid', 1, 35000, NOW() - interval '6 months', NOW() - interval '5 months', 'expired', NOW() - interval '6 months'
FROM users WHERE phone IN ('0903000008', '0903000010', '0903000019', '0903000020');

INSERT INTO subscriptions (user_id, plan, duration_months, amount, started_at, expires_at, status, created_at)
SELECT id, 'paid', 12, 300000, NOW() - interval '6 months', NOW() + interval '6 months', 'active', NOW() - interval '6 months'
FROM users WHERE phone = '0903000001';

INSERT INTO subscriptions (user_id, plan, duration_months, amount, started_at, expires_at, status, created_at)
SELECT id, 'free_trial', 0, 0, NOW() - interval '6 months', NOW() - interval '5 months', 'expired', NOW() - interval '6 months'
FROM users WHERE phone IN ('0903000021', '0903000022', '0903000023');

-- ========================
-- Month 6: 5 months ago — 12 subscriptions
-- ========================
INSERT INTO subscriptions (user_id, plan, duration_months, amount, started_at, expires_at, status, created_at)
SELECT id, 'paid', 3, 96000, NOW() - interval '5 months', NOW() - interval '2 months', 'expired', NOW() - interval '5 months'
FROM users WHERE phone IN ('0903000008', '0903000010', '0903000013', '0903000014');

INSERT INTO subscriptions (user_id, plan, duration_months, amount, started_at, expires_at, status, created_at)
SELECT id, 'paid', 1, 35000, NOW() - interval '5 months', NOW() - interval '4 months', 'expired', NOW() - interval '5 months'
FROM users WHERE phone IN ('0903000019', '0903000020', '0903000024', '0903000025');

INSERT INTO subscriptions (user_id, plan, duration_months, amount, started_at, expires_at, status, created_at)
SELECT id, 'paid', 6, 180000, NOW() - interval '5 months', NOW() + interval '1 month', 'active', NOW() - interval '5 months'
FROM users WHERE phone = '0903000002';

INSERT INTO subscriptions (user_id, plan, duration_months, amount, started_at, expires_at, status, created_at)
SELECT id, 'free_trial', 0, 0, NOW() - interval '5 months', NOW() - interval '4 months', 'expired', NOW() - interval '5 months'
FROM users WHERE phone IN ('0903000026', '0903000027', '0903000028');

-- ========================
-- Month 7: 4 months ago — 14 subscriptions
-- ========================
INSERT INTO subscriptions (user_id, plan, duration_months, amount, started_at, expires_at, status, created_at)
SELECT id, 'paid', 6, 180000, NOW() - interval '4 months', NOW() + interval '2 months', 'active', NOW() - interval '4 months'
FROM users WHERE phone IN ('0903000009', '0903000015');

INSERT INTO subscriptions (user_id, plan, duration_months, amount, started_at, expires_at, status, created_at)
SELECT id, 'paid', 3, 96000, NOW() - interval '4 months', NOW() - interval '1 month', 'expired', NOW() - interval '4 months'
FROM users WHERE phone IN ('0903000019', '0903000020', '0903000024', '0903000025');

INSERT INTO subscriptions (user_id, plan, duration_months, amount, started_at, expires_at, status, created_at)
SELECT id, 'paid', 1, 35000, NOW() - interval '4 months', NOW() - interval '3 months', 'expired', NOW() - interval '4 months'
FROM users WHERE phone IN ('0903000029', '0903000030', '0903000031', '0903000032');

INSERT INTO subscriptions (user_id, plan, duration_months, amount, started_at, expires_at, status, created_at)
SELECT id, 'free_trial', 0, 0, NOW() - interval '4 months', NOW() - interval '3 months', 'expired', NOW() - interval '4 months'
FROM users WHERE phone IN ('0903000033', '0903000034', '0903000035', '0903000036');

-- ========================
-- Month 8: 3 months ago — 16 subscriptions (peak growth)
-- ========================
INSERT INTO subscriptions (user_id, plan, duration_months, amount, started_at, expires_at, status, created_at)
SELECT id, 'paid', 12, 300000, NOW() - interval '3 months', NOW() + interval '9 months', 'active', NOW() - interval '3 months'
FROM users WHERE phone IN ('0903000003', '0903000016');

INSERT INTO subscriptions (user_id, plan, duration_months, amount, started_at, expires_at, status, created_at)
SELECT id, 'paid', 6, 180000, NOW() - interval '3 months', NOW() + interval '3 months', 'active', NOW() - interval '3 months'
FROM users WHERE phone IN ('0903000008', '0903000010', '0903000013');

INSERT INTO subscriptions (user_id, plan, duration_months, amount, started_at, expires_at, status, created_at)
SELECT id, 'paid', 3, 96000, NOW() - interval '3 months', NOW(), 'expired', NOW() - interval '3 months'
FROM users WHERE phone IN ('0903000029', '0903000030', '0903000031', '0903000032');

INSERT INTO subscriptions (user_id, plan, duration_months, amount, started_at, expires_at, status, created_at)
SELECT id, 'paid', 1, 35000, NOW() - interval '3 months', NOW() - interval '2 months', 'expired', NOW() - interval '3 months'
FROM users WHERE phone IN ('0903000037', '0903000038');

INSERT INTO subscriptions (user_id, plan, duration_months, amount, started_at, expires_at, status, created_at)
SELECT id, 'free_trial', 0, 0, NOW() - interval '3 months', NOW() - interval '2 months', 'expired', NOW() - interval '3 months'
FROM users WHERE phone IN ('0903000039', '0903000040', '0903000041', '0903000042');

-- ========================
-- Month 9: 2 months ago — 15 subscriptions
-- ========================
INSERT INTO subscriptions (user_id, plan, duration_months, amount, started_at, expires_at, status, created_at)
SELECT id, 'paid', 6, 180000, NOW() - interval '2 months', NOW() + interval '4 months', 'active', NOW() - interval '2 months'
FROM users WHERE phone IN ('0903000019', '0903000020', '0903000024');

INSERT INTO subscriptions (user_id, plan, duration_months, amount, started_at, expires_at, status, created_at)
SELECT id, 'paid', 3, 96000, NOW() - interval '2 months', NOW() + interval '1 month', 'active', NOW() - interval '2 months'
FROM users WHERE phone IN ('0903000037', '0903000038', '0903000025');

INSERT INTO subscriptions (user_id, plan, duration_months, amount, started_at, expires_at, status, created_at)
SELECT id, 'paid', 1, 35000, NOW() - interval '2 months', NOW() - interval '1 month', 'expired', NOW() - interval '2 months'
FROM users WHERE phone IN ('0903000043', '0903000044', '0903000045', '0903000046');

INSERT INTO subscriptions (user_id, plan, duration_months, amount, started_at, expires_at, status, created_at)
SELECT id, 'free_trial', 0, 0, NOW() - interval '2 months', NOW() - interval '1 month', 'expired', NOW() - interval '2 months'
FROM users WHERE phone IN ('0903000047', '0903000048', '0903000049', '0903000050', '0903000051');

-- ========================
-- Month 10: 1 month ago — 18 subscriptions
-- ========================
INSERT INTO subscriptions (user_id, plan, duration_months, amount, started_at, expires_at, status, created_at)
SELECT id, 'paid', 12, 300000, NOW() - interval '1 month', NOW() + interval '11 months', 'active', NOW() - interval '1 month'
FROM users WHERE phone IN ('0903000004', '0903000005');

INSERT INTO subscriptions (user_id, plan, duration_months, amount, started_at, expires_at, status, created_at)
SELECT id, 'paid', 6, 180000, NOW() - interval '1 month', NOW() + interval '5 months', 'active', NOW() - interval '1 month'
FROM users WHERE phone IN ('0903000029', '0903000030', '0903000031');

INSERT INTO subscriptions (user_id, plan, duration_months, amount, started_at, expires_at, status, created_at)
SELECT id, 'paid', 3, 96000, NOW() - interval '1 month', NOW() + interval '2 months', 'active', NOW() - interval '1 month'
FROM users WHERE phone IN ('0903000043', '0903000044', '0903000045', '0903000046');

INSERT INTO subscriptions (user_id, plan, duration_months, amount, started_at, expires_at, status, created_at)
SELECT id, 'paid', 1, 35000, NOW() - interval '1 month', NOW(), 'expired', NOW() - interval '1 month'
FROM users WHERE phone IN ('0903000052', '0903000053', '0903000054');

INSERT INTO subscriptions (user_id, plan, duration_months, amount, started_at, expires_at, status, created_at)
SELECT id, 'free_trial', 0, 0, NOW() - interval '1 month', NOW(), 'expired', NOW() - interval '1 month'
FROM users WHERE phone IN ('0903000055', '0903000056', '0903000057', '0903000058', '0903000059', '0903000060');

-- ========================
-- Month 11: Current month — 20 subscriptions (highest)
-- ========================
INSERT INTO subscriptions (user_id, plan, duration_months, amount, started_at, expires_at, status, created_at)
SELECT id, 'paid', 12, 300000, NOW() - interval '5 days', NOW() + interval '355 days', 'active', NOW() - interval '5 days'
FROM users WHERE phone IN ('0903000006', '0903000007', '0903000011');

INSERT INTO subscriptions (user_id, plan, duration_months, amount, started_at, expires_at, status, created_at)
SELECT id, 'paid', 6, 180000, NOW() - interval '3 days', NOW() + interval '177 days', 'active', NOW() - interval '3 days'
FROM users WHERE phone IN ('0903000032', '0903000052', '0903000053');

INSERT INTO subscriptions (user_id, plan, duration_months, amount, started_at, expires_at, status, created_at)
SELECT id, 'paid', 3, 96000, NOW() - interval '2 days', NOW() + interval '88 days', 'active', NOW() - interval '2 days'
FROM users WHERE phone IN ('0903000054', '0903000055', '0903000033', '0903000034');

INSERT INTO subscriptions (user_id, plan, duration_months, amount, started_at, expires_at, status, created_at)
SELECT id, 'paid', 1, 35000, NOW() - interval '1 day', NOW() + interval '29 days', 'active', NOW() - interval '1 day'
FROM users WHERE phone IN ('0903000035', '0903000036', '0903000039', '0903000040');

INSERT INTO subscriptions (user_id, plan, duration_months, amount, started_at, expires_at, status, created_at)
SELECT id, 'free_trial', 0, 0, NOW(), NOW() + interval '30 days', 'active', NOW()
FROM users WHERE phone IN ('0903000041', '0903000042', '0903000047', '0903000048', '0903000049', '0903000050');

COMMIT;

-- Summary
DO $$
DECLARE
  v_total int;
  v_active int;
  v_expired int;
  v_paid int;
  v_trial int;
  v_revenue bigint;
BEGIN
  SELECT COUNT(*) INTO v_total FROM subscriptions;
  SELECT COUNT(*) INTO v_active FROM subscriptions WHERE status = 'active' AND expires_at > NOW();
  SELECT COUNT(*) INTO v_expired FROM subscriptions WHERE status = 'expired' OR expires_at <= NOW();
  SELECT COUNT(*) INTO v_paid FROM subscriptions WHERE plan = 'paid';
  SELECT COUNT(*) INTO v_trial FROM subscriptions WHERE plan = 'free_trial';
  SELECT COALESCE(SUM(amount), 0) INTO v_revenue FROM subscriptions WHERE plan = 'paid';
  RAISE NOTICE '========== Subscription Seed Summary ==========';
  RAISE NOTICE 'Total subscriptions: %', v_total;
  RAISE NOTICE 'Active: %', v_active;
  RAISE NOTICE 'Expired: %', v_expired;
  RAISE NOTICE 'Paid: %', v_paid;
  RAISE NOTICE 'Free trial: %', v_trial;
  RAISE NOTICE 'Total revenue: % VND', v_revenue;
  RAISE NOTICE '================================================';
END $$;
