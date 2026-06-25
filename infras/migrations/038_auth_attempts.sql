-- Migration 038 — Bảng auth_attempts (chống spam đăng nhập/OTP/đăng ký).
--
-- LÝ DO: schema gốc nằm ở backend/migrations/005_spam_protection.up.sql NHƯNG
-- thư mục đó KHÔNG thuộc nhóm auto-run (auto-run = backend/internal/database/
-- migrations). Vì vậy bảng này CHƯA BAO GIỜ được tạo trên prod → spam_service
-- query lỗi 42P01. Trước đây spam fail-OPEN nên login vẫn chạy (âm thầm bỏ qua);
-- sau khi đổi spam sang fail-CLOSED, lỗi này chặn TOÀN BỘ đăng nhập
-- ("Hệ thống đang bận"). Migration này tạo bảng để spam-check chạy đúng.
--
-- Đã áp tay trên prod 25/06; file này để lần deploy mới (fresh) không thiếu bảng.
-- Idempotent.

CREATE TABLE IF NOT EXISTS auth_attempts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    ip_address VARCHAR(45) NOT NULL,
    device_id VARCHAR(128),
    phone VARCHAR(15),
    action VARCHAR(20) NOT NULL,
    success BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_auth_attempts_ip_action ON auth_attempts(ip_address, action, created_at);
CREATE INDEX IF NOT EXISTS idx_auth_attempts_device ON auth_attempts(device_id, action, created_at);
CREATE INDEX IF NOT EXISTS idx_auth_attempts_cleanup ON auth_attempts(created_at);
