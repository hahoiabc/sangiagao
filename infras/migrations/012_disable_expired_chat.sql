-- Disable chat for expired subscriptions
UPDATE role_permissions SET allowed = false WHERE role = 'expired' AND permission_key IN ('chat.send', 'chat.send_image');
