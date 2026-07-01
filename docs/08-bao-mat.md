# 08 — Bảo mật

> Chi tiết audit đầy đủ (25/06/2026, ảnh chụp lịch sử): [`archive/2026-06-audit-findings.md`](archive/2026-06-audit-findings.md).
> Đợt audit đó đã fix **21 bug** — mục dưới là các LỚP bảo mật hiện hành + bài học.

## Các lớp bảo mật hiện hành

- **Auth/token:** JWT HS256 ép cứng, 3 cơ chế thu hồi qua Redis (blacklist / tvf / rot). Xem [04-auth-rbac](04-auth-rbac.md).
- **RBAC fail-closed:** matrix `role_permissions`, key vắng = false. Route admin phải có `RequirePermission` (bẫy "route mồ côi").
- **PII:** SĐT mã hóa (`phone_hash` + `phone_encrypt`). `PHONE_ENCRYPT_KEY` bí mật, backup kỹ.
- **OTP:** chỉ Zalo ZNS; guard chặn mock ở production (đang ngủ vì `APP_ENV=development`). Rate-limit đăng ký/OTP.
- **Rate-limit / DoS:** `UserRateLimit` cho broadcast/notif/OTP; `auth_attempts` chặn brute-force login.
- **IDOR / cách ly:** inbox, conversation, listing kiểm ownership. Report chống self-report + 23505.
- **XSS:** escape JSON-LD (web SEO) + validate domain URL ảnh (chống chèn URL lạ).
- **Refund abuse (affiliate):** clawback `pending`/`payable` khi nhận webhook REFUND; đã `paid` thì không thu hồi. Xem [05-affiliate](05-affiliate.md).

## Đã xử lý trong audit 25/06 (tham chiếu)

Backdoor OTP mock prod, route `/admin` mồ côi permission, role constraint (mig 026 thiếu owner/editor → mig 036 gộp 6), clawback UPDATE sai bảng, freemium filter role sai, refresh rotation, revoke token khi đổi-MK/khóa, spam fail-closed, escape XSS, soft-delete + ẩn danh… (21 mục).

## Bài học bảo mật (đọc khi fix)

1. **Fail-closed trên luồng auth = user-facing** → verify prerequisite trên PROD trước khi deploy (đừng tin migration có trong repo = đã áp).
2. **Đổi mật khẩu/khóa/xóa tài khoản** phải thu hồi token cũ (`RevokeUserTokens`) — nếu không, phiên cũ vẫn dùng được.
3. **Money/security fix:** kiểm cả nhánh phụ bị nuốt (log Warn rồi return nil) + constraint DB (định danh lệch pha).
4. **Fix bảo mật thuần** (RBAC/DoS/IDOR nội bộ/escape) → tự làm; fix chạm auth-OTP/người dùng → xin duyệt chủ trước.
