# Checklist chuyển đổi codebase sang dự án khác

## 1. Thay đổi bắt buộc trước deploy

### A. Branding & Identity
- [ ] Đổi tên app: `sangiagao` / `Sàn Giao Gạo` → tên mới
- [ ] Đổi domain: `sangiagao.vn` → domain mới
- [ ] Đổi logo, icon, splash screen (mobile + admin)
- [ ] Đổi `applicationId` / `bundleId` trong Flutter (Android + iOS)
- [ ] Đổi tên package Go: `github.com/sangiagao/rice-marketplace` → package mới

### B. Secrets & Credentials (QUAN TRỌNG NHẤT)
- [ ] Tạo **JWT secret mới** (không dùng lại!)
- [ ] Tạo **database passwords mới** (PG, MongoDB, Redis)
- [ ] Tạo **MinIO access/secret key mới**
- [ ] Tạo **SMS/Zalo OTP credentials mới**
- [ ] Tạo **Firebase project mới** (push notification) + thay `google-services.json` / `GoogleService-Info.plist`
- [ ] Cập nhật toàn bộ `.env` file

### C. Database
- [ ] Chạy migration từ đầu trên DB trống (không copy data cũ)
- [ ] Xóa seed data / test data nếu có
- [ ] Tạo admin account mới

### D. Business Logic cần sửa
- [ ] **Catalog**: đổi danh mục sản phẩm (gạo → sản phẩm mới)
- [ ] **Validation rules**: price range, quantity range phù hợp ngành mới
- [ ] **Vietnamese text**: rà soát error messages, labels phù hợp sản phẩm mới
- [ ] **Subscription plans**: tạo gói phù hợp

### E. Infrastructure
- [ ] VPS mới hoặc tách biệt hoàn toàn
- [ ] Cloudflare account/zone mới cho domain mới
- [ ] MinIO bucket mới
- [ ] UFW + backup config lại từ đầu
- [ ] SSL certificate (Cloudflare tự xử lý nếu dùng proxy)

## 2. Files cần grep & replace

```
Tìm và thay toàn bộ:
  "sangiagao"          → tên dự án mới
  "Sàn Giao Gạo"      → tên hiển thị mới
  "sangiagao.vn"       → domain mới
  "rice-marketplace"   → package name mới
  "rice-images"        → bucket name mới
  "rice_chat"          → chat app name mới (Elixir)
  "rice_internal"      → docker network name mới
```

## 3. Thứ tự thực hiện

| Bước | Việc | Lý do |
|------|------|-------|
| 1 | Fork/copy repo, xóa `.git`, `git init` mới | Lịch sử sạch |
| 2 | Đổi tên package + branding | Nền tảng |
| 3 | Tạo `.env` mới với credentials mới | Bảo mật |
| 4 | Sửa business logic (catalog, validation, text) | Phù hợp ngành |
| 5 | `go build`, `flutter analyze`, `npx tsc --noEmit` | Verify compile |
| 6 | Chạy local bằng Docker Compose | Test end-to-end |
| 7 | Setup VPS + deploy | Go live |

## 4. Những gì KHÔNG cần sửa (dùng lại được ngay)

- Auth flow (OTP + JWT + refresh token)
- Chat system (Elixir Phoenix WebSocket)
- Upload presigned (MinIO)
- Admin panel structure
- Rate limiting, anti-spam
- Subscription/payment flow
- Notification system (push + inbox)
- Monitoring page
- Backup cronjob script
