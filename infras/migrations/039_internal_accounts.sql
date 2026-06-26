-- Migration 039 — Tài khoản nội bộ (admin tạo thủ công).
-- Cờ is_internal: tài khoản do admin tạo cho người không có Zalo / nội bộ.
--   - MIỄN gói thành viên (như staff) — không bị ẩn tin, đăng tin không cần gói.
--   - LOẠI khỏi thống kê/báo cáo (số demo không làm sai số liệu).
--   - Hiện nhãn "Nội bộ" trong danh sách user admin.
-- Idempotent.
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_internal BOOLEAN NOT NULL DEFAULT false;

-- Permission key cho tạo tài khoản thủ công (route POST /admin/users).
-- Mặc định: owner + admin được; editor/aff/member không.
INSERT INTO role_permissions (role, permission_key, allowed) VALUES
    ('owner',  'users.create', true),
    ('admin',  'users.create', true),
    ('editor', 'users.create', false),
    ('aff',    'users.create', false),
    ('member', 'users.create', false)
ON CONFLICT (role, permission_key) DO NOTHING;
