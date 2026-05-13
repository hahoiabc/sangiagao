-- Migration 027: Seed permission keys for affiliate referral system.
-- Aligns with admin/src/app/(admin)/users/page.tsx "Hoa hồng giới thiệu" group.
-- Admin can re-toggle these per role via /users → Vai trò & Quyền hạn.

INSERT INTO role_permissions (role, permission_key, allowed) VALUES
    -- referrals.view_own — see own commission only
    ('owner',  'referrals.view_own', true),
    ('admin',  'referrals.view_own', true),
    ('editor', 'referrals.view_own', false),
    ('aff',    'referrals.view_own', true),
    ('member', 'referrals.view_own', true),
    -- referrals.view_all — see all partners (admin filter bypass)
    ('owner',  'referrals.view_all', true),
    ('admin',  'referrals.view_all', true),
    ('editor', 'referrals.view_all', false),
    ('aff',    'referrals.view_all', false),
    ('member', 'referrals.view_all', false),
    -- referrals.manage_rules — edit commission rule defaults / overrides
    ('owner',  'referrals.manage_rules', true),
    ('admin',  'referrals.manage_rules', true),
    ('editor', 'referrals.manage_rules', false),
    ('aff',    'referrals.manage_rules', false),
    ('member', 'referrals.manage_rules', false),
    -- referrals.create_payout — create + mark payout sent
    ('owner',  'referrals.create_payout', true),
    ('admin',  'referrals.create_payout', true),
    ('editor', 'referrals.create_payout', false),
    ('aff',    'referrals.create_payout', false),
    ('member', 'referrals.create_payout', false)
ON CONFLICT (role, permission_key) DO NOTHING;
