-- Add notifications.broadcast permission for owner and admin
INSERT INTO role_permissions (role, permission_key, allowed) VALUES
    ('owner', 'notifications.broadcast', true),
    ('admin', 'notifications.broadcast', true),
    ('editor', 'notifications.broadcast', false)
ON CONFLICT (role, permission_key) DO NOTHING;
