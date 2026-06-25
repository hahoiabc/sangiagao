-- Migration 036 — Nền tảng cho đợt fix audit (25/06/2026)
-- Gom các thay đổi schema mà nhiều code-fix phụ thuộc. Áp THỦ CÔNG qua psql.
-- Idempotent (IF EXISTS / IF NOT EXISTS) để chạy lại an toàn.
--
-- Liên quan: AUDIT_FINDINGS.md các BUG #7, #8, #9, #3.

-- ── BUG #7: Role constraint hợp nhất đủ 6 role ──────────────────────────────
-- mig 026 đã thu hẹp về (member,seller,admin,aff), làm MẤT owner/editor mà code
-- vẫn dùng (bypass sub, guard /admin, permission fallback). Mở lại đủ 6.
-- Tập mới là SUPERSET nên mọi hàng hiện có đều hợp lệ.
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_check;
ALTER TABLE users
    ADD CONSTRAINT users_role_check
    CHECK (role IN ('member', 'seller', 'admin', 'editor', 'owner', 'aff'));

-- ── BUG #8: cột reports.admin_note (code đã dùng nhưng bảng thiếu) ───────────
-- report_repo.go SELECT/UPDATE admin_note → 42703 → kiểm duyệt report 500.
ALTER TABLE reports ADD COLUMN IF NOT EXISTS admin_note TEXT;

-- ── BUG #9: cột users.deleted_at cho soft-delete tài khoản ───────────────────
-- Xóa tài khoản chuyển sang soft-delete + ẩn danh (giữ commission/payment để
-- đối soát). Code sẽ filter deleted_at IS NULL ở các lookup.
ALTER TABLE users ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ NULL;
CREATE INDEX IF NOT EXISTS idx_users_active
    ON users (id) WHERE deleted_at IS NULL;

-- ── BUG #3: chuẩn hóa giá trị status ẩn-do-hết-hạn ──────────────────────────
-- Thống nhất 'hidden_subscription'. 'hidden' hiện CHỈ do IAP refund/revoke đặt
-- (không phải user tự ẩn) nên chuyển an toàn. Code Apple/Google cũng sửa kèm.
UPDATE listings SET status = 'hidden_subscription' WHERE status = 'hidden';
