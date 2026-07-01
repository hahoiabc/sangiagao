# 03 — Mô hình dữ liệu & quy ước (các BẪY đã gặp)

## Vai trò (roles)

6 vai trò: `member`, `seller`, `admin`, `editor`, `owner`, `aff`. Constraint `users_role_check` (mig 036 gộp đủ 6).

- Đăng ký gán cứng **`member`**. **KHÔNG ai được gán `seller` tự động** → "người bán" thực chất là `member` (hoặc `aff` sau khi opt-in affiliate). *(Bug cũ từng lọc nhầm `role='seller'`.)*
- `owner` / `admin` / `editor` = **staff**: bypass gate gói + được nhiều đặc quyền. Xem [04-auth-rbac](04-auth-rbac.md).
- `aff` = đối tác. Cách lên aff: [05-affiliate](05-affiliate.md).
- **Miễn gói:** staff + tài khoản có cờ `is_internal=true` (tạo thủ công) đều bỏ qua kiểm tra hết hạn gói (không bị ẩn tin).

## Listing (tin đăng)

- Status: `active` / `hidden_subscription` (ẩn do hết hạn gói) / `deleted`. Marketplace chỉ hiện `active`.
- *(Trước đây IAP dùng nhầm `'hidden'` — đã thống nhất `hidden_subscription` mọi nơi.)*

## Bảng hoa hồng

- Tên bảng là **`commission_records`** (KHÔNG phải `commissions`). *(Bug cũ từng UPDATE nhầm `commissions`.)* Chi tiết [05-affiliate](05-affiliate.md).

## Chat / tin nhắn (lưu 2 nơi)

- Postgres `messages` = **NGUỒN THẬT** (app gửi REST `/conversations/:id/messages` → ghi Postgres).
- MongoDB (Elixir) — chỉ realtime relay qua Phoenix (relay KHÔNG ghi Mongo).
- ➜ Muốn đếm/đọc tin nhắn chuẩn → dùng **Postgres**.

## Số điện thoại (mã hóa)

- `phone_hash` (SHA-256, để lookup) + `phone_encrypt` (AES, để hiển thị). Cột `phone` plaintext UNIQUE vẫn còn (ràng buộc 1 SĐT/1 account).
- ⚠️ `PHONE_ENCRYPT_KEY` (64 hex, trong `.env.backend`) — mất = KHÔNG giải mã được SĐT đã lưu. Backup kỹ.

## Thời gian & trường computed

- `created_at` (timestamp). Subscription có `expires_at`.
- `users.subscription_expires_at` là **computed** (subquery MAX active sub), KHÔNG phải cột thật.

---

## BẪY chung của dự án

- **"Định danh lệch pha":** bảng/cột/role/status/**constraint** đổi mà không cập nhật đồng bộ mọi nơi → tính năng ghi 0 hoặc rollback âm thầm. Vd: constraint `subscriptions_plan_check` từng thiếu `'reward'` khiến tính năng "thưởng" lỗi; `commission_rules` thiếu cột dùng ở code.
- **Cột UNIQUE toàn cục:** khi thêm bảng/cột có UNIQUE, cân nhắc nó có nên là per-scope không (bài học từ Sell365 — với sàn gạo, 1 SĐT/1 account là cố ý toàn cục).
- Khi thêm `source_type`/`plan`/`status`/`role` code mới → nhớ cập nhật **CHECK constraint** tương ứng.
