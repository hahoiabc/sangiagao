-- Allow editor to broadcast notifications
UPDATE role_permissions SET allowed = true WHERE role = 'editor' AND permission_key = 'notifications.broadcast';

-- Add notifications.send_individual permission
INSERT INTO role_permissions (role, permission_key, allowed) VALUES
    ('owner', 'notifications.send_individual', true),
    ('admin', 'notifications.send_individual', true),
    ('editor', 'notifications.send_individual', false)
ON CONFLICT (role, permission_key) DO NOTHING;
