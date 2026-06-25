-- Migration 037 — Seed permission keys cho các route /admin trước đây mồ côi
-- (BUG #5). Áp THỦ CÔNG qua psql TRƯỚC khi deploy binary mới — nếu không,
-- owner/admin sẽ bị chặn các route này (matrix có nhưng key vắng → false).
--
-- Quy tắc: permissions.manage CHỈ owner/admin (editor KHÔNG được ghi đè ma trận
-- quyền — đây là lỗ leo thang đã phát hiện). Các key khác cho staff xem/sửa.

INSERT INTO role_permissions (role, permission_key, allowed) VALUES
    -- permissions.manage — xem/ghi đè ma trận phân quyền (NHẠY CẢM: chỉ owner/admin)
    ('owner',  'permissions.manage', true),
    ('admin',  'permissions.manage', true),
    ('editor', 'permissions.manage', false),
    ('aff',    'permissions.manage', false),
    ('member', 'permissions.manage', false),
    -- payments.view — xem đơn thanh toán
    ('owner',  'payments.view', true),
    ('admin',  'payments.view', true),
    ('editor', 'payments.view', true),
    ('aff',    'payments.view', false),
    ('member', 'payments.view', false),
    -- site_settings.manage — sửa slogan/màu/video/about
    ('owner',  'site_settings.manage', true),
    ('admin',  'site_settings.manage', true),
    ('editor', 'site_settings.manage', true),
    ('aff',    'site_settings.manage', false),
    ('member', 'site_settings.manage', false),
    -- inbox.manage — quản lý system inbox
    ('owner',  'inbox.manage', true),
    ('admin',  'inbox.manage', true),
    ('editor', 'inbox.manage', true),
    ('aff',    'inbox.manage', false),
    ('member', 'inbox.manage', false),
    -- zns.view — xem trạng thái Zalo ZNS
    ('owner',  'zns.view', true),
    ('admin',  'zns.view', true),
    ('editor', 'zns.view', false),
    ('aff',    'zns.view', false),
    ('member', 'zns.view', false)
ON CONFLICT (role, permission_key) DO NOTHING;
