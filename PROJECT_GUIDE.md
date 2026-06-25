# PROJECT GUIDE — Sàn Giao Gạo (sangiagao.vn)

> Cẩm nang định hướng dự án cho người vào fix/nâng cấp về sau.
> Gom các kiến thức "ngầm" (không hiển nhiên từ code) — cập nhật 25/06/2026.
> Bổ trợ cho: `AUDIT_FINDINGS.md` (danh sách lỗ + cách fix), `CLAUDE.md` (kiến trúc tóm tắt).

---

## 1. Kiến trúc tổng thể

Monorepo 4 service + hạ tầng:

| Service | Công nghệ | Port nội bộ | Thư mục | Ghi chú |
|---------|-----------|-------------|---------|---------|
| Backend API | Go 1.25 / Gin | 8080 | `backend/` | Layered: handler→service→repository |
| Web public | Next.js / React | 3001 | `web/` | SEO (SSG/ISR), marketplace, member area |
| Admin | Next.js / React | 3000 | `admin/` | Quản trị (shadcn/ui) |
| Chat | Elixir / Phoenix | 4000 | `chat/` | WebSocket realtime |

Hạ tầng: PostgreSQL 16, Redis 7, MongoDB 7, MinIO (S3), Nginx, Cloudflare CDN.

**Backend layered (quan trọng):**
- `internal/handler/` — HTTP, parse request, map lỗi → status.
- `internal/service/` — business logic. Interface repo định nghĩa ở `service/deps.go` (consumer-defined).
- `internal/repository/` — query Postgres (pgx/v5).
- `internal/handler/deps.go` — interface service mà handler dùng (để test + tránh import cycle).
- `cmd/api/main.go` — DI wiring + đăng ký route + cron goroutine.
- Test: `testify/mock`, mock struct trong `*_test.go`. **CHƯA có integration test** (cần testcontainers).

---

## 2. Deploy THỰC TẾ (đọc kỹ — nhiều chỗ KHÔNG hiển nhiên)

**Server prod:** `root@14.225.213.73`, repo tại **`/opt/sangiagao`**. VPS này CHẠY CHUNG với **Sell 365** (app khác) — container `sell365_*`. **TUYỆT ĐỐI không đụng container sell365.**

**Container sàn giao gạo:** `backend`, `web`, `admin`, `rice_chat`, `rice_nginx`, `rice_postgres` (DB), `rice_redis`, `rice_mongodb`, `minio`, `rice_certbot`, `rice_db_backup`.

**⚠️ Deploy KHÔNG dùng docker-compose.** Dùng script `infras/scripts/quick-deploy.sh`:
```bash
cd /opt/sangiagao
bash infras/scripts/quick-deploy.sh backend   # hoặc web | admin | chat | all | migrate
```
Script tự: `git pull origin main` → `docker build --no-cache` → stop/rm/run container (`docker run --env-file infras/.env.backend ...`) → restart nginx → health check.

- ➜ **`docker-compose.production.yml` ở repo gốc KHÔNG được deploy dùng** (gây hiểu nhầm — nó chỉ là tham khảo). Env thật nằm ở **`/opt/sangiagao/infras/.env.backend`** (và `.env.chat`).
- ➜ Health check trong script báo **"chat: NOT running" là BÁO ĐỘNG GIẢ** — nó tìm tên `chat` nhưng container thật tên `rice_chat`. Bỏ qua lỗi đó.
- DB connection: `DB_NAME=rice_marketplace`, `DB_USER=rice_user`, host `postgres` (alias mạng `rice_internal`).

**⚠️ `APP_ENV` KHÔNG được set trên prod → app chạy `env=development`.** Hệ quả:
- Các kiểm tra production trong `config.Validate()` KHÔNG chạy (CORS phải explicit, DB pass đổi, MinIO/Redis...).
- 2 guard bảo mật NGỦ: **chặn OTP mock** (whitelist SMS_PROVIDER) và **Redis bắt buộc**. Không sao vì hiện SMS=`zalo+mock` thật + Redis chạy.
- Muốn bật: set `APP_ENV=production` trong `infras/.env.backend` NHƯNG phải verify CORS_ORIGINS/DB_PASSWORD/MinIO/Redis qua Validate trước, kẻo app không khởi động.

**SMS/OTP:** `SMS_PROVIDER` chỉ chấp nhận **`zalo`** hoặc **`zalo+mock`** (code chỉ hiện thực Zalo ZNS). Giá trị khác (`esms`, `mock`, rỗng) → rơi vào `MockSender` → OTP không gửi. Hiện prod dùng `zalo+mock`, có đủ `ZALO_APP_ID/SECRET/TEMPLATE_ID/REFRESH_TOKEN`.

**IAP:** **Apple IAP và Google IAP đang TẮT** trên prod (thiếu env `APP_STORE_*` / Google). Chỉ thanh toán qua **SePay** (chuyển khoản QR, source `web`/`sepay`) hoạt động.

---

## 3. Migrations (CÓ 2 HỆ THỐNG — dễ nhầm)

1. **Tự động** khi backend khởi động: `backend/internal/database/migrations/` (embedded, ~001–014). Backend tự `RunMigrations`.
2. **Thủ công** qua psql: `infras/migrations/` (015 → hiện tại). KHÔNG tự chạy. Áp bằng:
   ```bash
   docker exec -i rice_postgres psql -U rice_user -d rice_marketplace < infras/migrations/NNN_xxx.sql
   ```
   (Hoặc `quick-deploy.sh migrate` — nhưng nó chạy LẠI tất cả file, dựa vào idempotency.)

**Quy tắc:** migration phải **idempotent** (`IF NOT EXISTS`, `DROP ... IF EXISTS`, `ON CONFLICT DO NOTHING`) và **tương thích ngược** (binary cũ vẫn chạy được với schema mới → cho phép áp migration trước khi deploy binary, rollback an toàn). **Áp migration TRƯỚC khi deploy binary mới.**

Migration gần nhất: **037**. Số kế tiếp = **038**.

---

## 4. Mô hình dữ liệu & quy ước (các BẪY đã gặp)

- **Roles (6):** `member`, `seller`, `admin`, `editor`, `owner`, `aff`. Constraint `users_role_check` (mig 036 hợp nhất đủ 6).
  - Đăng ký gán cứng **`member`**. **KHÔNG ai được gán `seller` tự động** → "người bán" thực chất là `member` (hoặc `aff` sau khi opt-in affiliate). *(Bug #2 từng lọc nhầm `role='seller'`.)*
  - `owner`/`admin`/`editor` = staff, bypass subscription gate.
- **Listing status:** `active` / `hidden_subscription` (ẩn do hết hạn gói) / `deleted`. Marketplace chỉ hiện `active`. *(Trước đây IAP dùng nhầm `'hidden'` — đã thống nhất `hidden_subscription` ở mọi nơi.)*
- **Bảng hoa hồng tên `commission_records`** (KHÔNG phải `commissions`). *(Bug #1 từng UPDATE nhầm `commissions`.)*
- **Chat lưu 2 nơi:** Postgres `messages` (NGUỒN THẬT — app gửi qua REST `/conversations/:id/messages` → ghi Postgres) + MongoDB (Elixir). Mobile gửi REST trước, rồi `relay` qua Phoenix CHỈ để realtime (relay KHÔNG ghi Mongo). ➜ Muốn đếm/đọc tin nhắn chuẩn → dùng Postgres.
- **Phone:** lưu mã hóa. `phone_hash` (SHA-256, để lookup) + `phone_encrypt` (AES, để hiển thị). Cột `phone` plaintext UNIQUE vẫn còn (ràng buộc 1 SĐT/1 account).
- **Thời gian:** `created_at` (timestamp). Subscription `expires_at`. `subscription_expires_at` trên User là computed (subquery), không phải cột thật.

---

## 5. Auth & Session

- **JWT:** access token **15 phút** + refresh token **30 ngày**. Ký HS256 (ép cứng, chống alg confusion). Secret + `PHONE_ENCRYPT_KEY` (64 hex) bắt buộc.
- **Thu hồi token (3 cơ chế, đều qua Redis):**
  1. `blacklist:{hash(token)}` — logout, rotation huỷ token cũ.
  2. `tvf:{userID}` (tokens_valid_from) — set khi **đổi/reset mật khẩu** hoặc **khóa/xóa tài khoản**; JWTAuth từ chối token có `IssuedAt < tvf`. Helper `middleware.RevokeUserTokens`.
  3. `rot:{hash(refresh cũ)}` — ân hạn xoay 60s: refresh song song trong 60s nhận LẠI cùng cặp token mới (idempotent).
- **Refresh single-flight:** Web/Admin CÓ khoá tuần tự (`isRefreshing`). **Mobile CHƯA có** → dựa vào ân hạn `rot:` 60s ở backend để không văng đăng nhập. *(Nợ: thêm single-flight cho mobile — mục #11-B.)*
- Handler nào thu hồi token cần `SetCache(appCache)` trong `main.go` (auth/user/admin handler đều đã wire).

---

## 6. Thanh toán & Hoa hồng (Affiliate)

- **3 nguồn:** `sepay` (QR chuyển khoản — ACTIVE), `apple`/`google` IAP (TẮT trên prod).
- **Commission engine** (`service/commission_engine.go`): stage theo **số lần thanh toán** (lần 1→45%, lần 2→30%, lần 3+→15%, cấu hình ở `commission_rules`). Idempotent theo `(payment_source, payment_event_id)` + `FOR UPDATE` lock referee. `PayableDelayDays = 45` (T+45 mới `pending`→`payable`).
- **Clawback khi refund:** `CancelCommissionsForSubscription` — gọi từ webhook REFUND/REVOKE Apple/Google. *(Bug #1 đã fix: trước UPDATE sai bảng.)* SePay KHÔNG có refund tự động (xử lý tay).
- **Trạng thái prod (25/06):** `commission_records` RỖNG, 0 payout — hệ affiliate **chưa dùng thật**.

---

## 7. Xóa tài khoản (semantics quan trọng)

`userRepo.DeleteUser` — dùng cho **CẢ user tự xóa LẪN owner xóa từ admin**:
- **Ẩn danh** dòng `users` (giữ lại để neo FK tiền): xóa PII, đổi `phone`→mã `'D'+14hex`, `phone_hash`→`'deleted:'+id`, `deleted_at`, `is_blocked=true`. ➜ **giải phóng SĐT gốc** để đăng ký lại như mới.
- **Xóa sạch** dữ liệu cá nhân: listings, conversations+messages, ratings, reports (của họ), notifications, device_tokens, feedbacks, inbox_read_status, user_blocks, message_reactions.
- **GIỮ** (đối soát kế toán + log): commission/payment/subscription/referral + `admin_audit_logs`.
- Owner xóa được tài khoản **admin-trở-xuống** (chặn xóa owner/chính mình) — `adminService.DeleteUser`. UI: nút trong `admin/.../users/page.tsx` (owner-only) + dialog cảnh báo.
- KHÔNG cần migration (giữ dòng user → không vỡ FK).

---

## 8. RBAC / Phân quyền

- Bảng `role_permissions(role, permission_key, allowed)`. `permission_service` load thành matrix (cache). Key vắng = `false` (fail-closed). Thêm route mới gắn `RequirePermission(key)` thì **phải seed key** cho role tương ứng (migration), nếu không owner/admin cũng bị chặn.
- **Bẫy "route mồ côi":** route `/admin` chỉ gác `RequireRole` mà quên `RequirePermission` → bỏ qua matrix. *(Bug #5 đã gắn key cho payments/site-settings/inbox/permissions/zns.)*
- `permissions.manage` (sửa ma trận quyền) chỉ owner/admin (KHÔNG editor).

---

## 9. Nợ kỹ thuật & Roadmap (cho lần nâng cấp sau)

| Mục | Mô tả | Ưu tiên |
|-----|-------|---------|
| **APP_ENV=production** | Đang `development` → guard #4 (chặn OTP mock) + #14 (Redis bắt buộc) ngủ. Bật sau khi verify CORS/DB/MinIO/Redis. | TB |
| **Mobile single-flight refresh** (#11-B) | Thêm khoá/Completer vào `mobile/.../api_service.dart _refreshToken()` (web đã có). Hiện dựa ân hạn 60s. | TB |
| **Test integration** | Cần dựng testcontainers (Postgres+Redis) phủ: refund→clawback, expire→hide→renew→restore, OTP, block, xóa TK. Hiện chỉ unit test. | Cao |
| **#15 resolve report** | Hiện báo lỗi nếu action fail (không nuốt). Lý tưởng: chạy resolve+action trong 1 transaction. | Thấp |
| **Commission rules** | Hardcode/migration; chưa có admin UI sửa rule động. | Thấp |
| **Cron jobs** | 4 cron chạy bằng goroutine trong `main.go` — instance chết = mất job. Cân nhắc scheduler (Asynq). | Thấp |
| **Web chat polling** | Web dùng REST polling 15s thay vì nối Phoenix WS client (hạ tầng WS đã có). | TB |
| **Apple/Google IAP** | Đang tắt. Bật cần env `APP_STORE_*` / Google service account. | Khi cần |
| **Observability** | Chưa có metrics (Prometheus)/tracing/Crashlytics. Log slog lẫn log.Printf. | TB |

---

## 10. Quy ước khi fix (đọc trước khi sửa)

1. **Phân tích phụ thuộc, fix nền tảng trước** (vd schema/migration trước code).
2. **Fix chạm người dùng / ý đồ thiết kế / tiền / auth-OTP → TRÌNH DUYỆT chủ trước.** Fix bảo mật thuần (RBAC/DoS/IDOR nội bộ/escape XSS) → tự làm. (Chủ yêu cầu rõ.)
3. **Migration:** áp THỦ CÔNG trước khi deploy binary; idempotent + tương thích ngược.
4. **Commit:** chỉ `git add` file của mình (repo hay có file dở dang pre-existing ở mobile/web). `git push origin HEAD:main` (deploy pull main).
5. Build/test/gofmt trước commit. Backend: `cd backend && go build ./... && go test ./internal/... && gofmt -l`.
6. Pattern lỗi hay gặp ở dự án này: **"định danh lệch pha"** (bảng/cột/role/status đổi nhưng không cập nhật đồng bộ) + **lỗi nhánh phụ bị nuốt** (log Warn rồi return nil) → tính năng tưởng chạy mà ghi 0/không chạy. Khi fix money/security nhớ kiểm cả nhánh phụ.

---

*Tài liệu này nên cập nhật mỗi khi phát hiện thêm "kiến thức ngầm" hoặc thay đổi hạ tầng/quy ước.*
