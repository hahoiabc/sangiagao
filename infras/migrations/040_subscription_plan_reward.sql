-- Migration 040 — Cho phép plan='reward' (thưởng thời gian sử dụng).
-- AdminReward tạo subscription plan='reward' nhưng constraint cũ chỉ cho
-- free_trial/paid → INSERT vi phạm → tính năng "Thưởng" báo lỗi.
-- Idempotent.
ALTER TABLE subscriptions DROP CONSTRAINT IF EXISTS subscriptions_plan_check;
ALTER TABLE subscriptions
    ADD CONSTRAINT subscriptions_plan_check
    CHECK (plan IN ('free_trial', 'paid', 'reward'));
