# 04 — Auth, Session & RBAC & Tài khoản

## JWT & Session

- **Token:** access **15 phút** + refresh **30 ngày**. Ký HS256 (ép cứng, chống alg confusion). `JWT_SECRET` + `PHONE_ENCRYPT_KEY` (64 hex) bắt buộc.
- **Đăng nhập bằng mật khẩu** (`LoginPassword`) KHÔNG cần OTP. OTP chỉ dùng khi **đăng ký** + **đặt lại mật khẩu**.
- Mật khẩu (luồng tự đăng ký / đặt lại): yêu cầu ≥6 ký tự + có **chữ hoa + chữ thường + ký tự đặc biệt** (`validatePassword`). Riêng **admin tạo tài khoản thủ công** chỉ cần ≥6 ký tự (nới lỏng).

### Thu hồi token (3 cơ chế, đều qua Redis)
1. `blacklist:{hash(token)}` — logout, rotation huỷ token cũ.
2. `tvf:{userID}` (tokens_valid_from) — set khi **đổi/reset mật khẩu** hoặc **khóa/xóa tài khoản**; JWTAuth từ chối token có `IssuedAt < tvf`. Helper `middleware.RevokeUserTokens`.
3. `rot:{hash(refresh cũ)}` — ân hạn xoay 60s: refresh song song trong 60s nhận LẠI cùng cặp token mới (idempotent).

- **Refresh single-flight:** Web/Admin CÓ khoá tuần tự (`isRefreshing`). **Mobile CHƯA có** → dựa ân hạn `rot:` 60s. *(Nợ kỹ thuật.)*
- Handler thu hồi token cần `SetCache(appCache)` trong `main.go`.

> **Bài học:** đổi thành fail-closed trên LUỒNG AUTH (vd spam guard) là user-facing → phải VERIFY prerequisite (bảng tồn tại trên PROD) TRƯỚC khi deploy. Từng gây "Hệ thống đang bận" toàn bộ login vì bảng `auth_attempts` chưa có trên prod.

---

## RBAC / Phân quyền

- Bảng `role_permissions(role, permission_key, allowed)`. `permission_service` load thành matrix (cache). **Key vắng = `false`** (fail-closed).
- ➜ Thêm route mới gắn `RequirePermission(key)` thì **PHẢI seed key** cho role tương ứng (migration), nếu không owner/admin cũng bị chặn.
- **Bẫy "route mồ côi":** route `/admin` chỉ gác `RequireRole` mà quên `RequirePermission` → bỏ qua matrix. *(Đã gắn key cho payments/site-settings/inbox/permissions/zns.)*
- `permissions.manage` (sửa ma trận quyền) chỉ owner/admin (KHÔNG editor).
- Ví dụ key mới: `users.create` (mig 039) cho owner/admin → tạo tài khoản thủ công.

---

## Tài khoản: Tạo thủ công & Nội bộ

Dành cho người **không có Zalo** (không nhận được OTP) hoặc tài khoản nội bộ/demo.

- **Tạo thủ công (admin):** `POST /admin/users` (perm `users.create`, owner/admin). UI: tab **"Tạo thành viên"** ở `admin/.../users/page.tsx`. Nhập SĐT + tên + mật khẩu + vai trò → tài khoản **đăng nhập ngay bằng SĐT + mật khẩu**, KHÔNG qua OTP.
- **Cờ `is_internal`** (mig 039): tài khoản nội bộ → **miễn gói** (như staff, không bị ẩn tin) + **loại khỏi thống kê** total_users + hiện nhãn "Nội bộ" trong danh sách admin.
- Vai trò tạo được: member / seller / editor / admin (không tạo owner).

---

## Xóa tài khoản (semantics quan trọng)

`userRepo.DeleteUser` — dùng cho **CẢ user tự xóa LẪN owner xóa từ admin**:

- **Ẩn danh** dòng `users` (giữ để neo FK tiền): xóa PII, `phone`→mã `'D'+14hex`, `phone_hash`→`'deleted:'+id`, set `deleted_at` + `is_blocked=true`. ➜ **giải phóng SĐT gốc** để đăng ký lại như mới.
- **Xóa sạch** dữ liệu cá nhân: listings, conversations+messages, ratings, reports (của họ), notifications, device_tokens, feedbacks, inbox_read_status, user_blocks, message_reactions.
- **GIỮ** (đối soát + log): commission/payment/subscription/referral + `admin_audit_logs`.
- Owner xóa được tài khoản **admin-trở-xuống** (chặn xóa owner/chính mình) — `adminService.DeleteUser`. UI: nút owner-only + dialog cảnh báo.
- KHÔNG cần migration (giữ dòng user → không vỡ FK).
- **Xóa ≠ Khóa:** khóa (`is_blocked`, block_reason) chỉ chặn đăng nhập, giữ nguyên dữ liệu + SĐT; xóa thì ẩn danh + giải phóng SĐT + xóa dữ liệu cá nhân.

**Admin "Đã xóa":** danh sách chính ẩn tài khoản đã xóa (deleted_at IS NULL); tab riêng "🗑 Đã xóa" liệt kê chúng (ListUsers có param `deleted`).
