-- Migration 041 — Giới hạn thời gian hoa hồng lần 3 (stage 3).
-- Chính sách mới: hoa hồng lần 3 trở đi KHÔNG còn vĩnh viễn; chỉ trả cho các
-- lần thanh toán trong vòng N tháng kể từ ngày TV được giới thiệu đăng ký
-- (referred_at). 0 = không giới hạn (vĩnh viễn như cũ). Mặc định 24 tháng.
-- Cấu hình được theo từng thời kỳ + theo từng đối tác (per referral_code_id).
-- Idempotent.
ALTER TABLE commission_rules
    ADD COLUMN IF NOT EXISTS stage3_cap_months INT NOT NULL DEFAULT 24
    CHECK (stage3_cap_months >= 0);
