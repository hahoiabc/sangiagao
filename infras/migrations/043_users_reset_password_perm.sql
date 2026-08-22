-- Migration 043 — Quyền "Đặt lại mật khẩu khách hàng" (users.reset_password).
-- Chủ (owner) mặc định CÓ; chủ tự giao cho nhân viên (admin/editor...) qua trang
-- Phân quyền. Guard nghiệp vụ ở service: KHÔNG cho reset tài khoản nội bộ/owner/
-- admin trừ khi người thực hiện là owner (chống nhân viên chiếm tài khoản quản trị).
-- Idempotent.
INSERT INTO role_permissions (role, permission_key, allowed) VALUES
    ('owner',  'users.reset_password', true),
    ('admin',  'users.reset_password', false),
    ('editor', 'users.reset_password', false),
    ('aff',    'users.reset_password', false),
    ('member', 'users.reset_password', false)
ON CONFLICT (role, permission_key) DO NOTHING;
