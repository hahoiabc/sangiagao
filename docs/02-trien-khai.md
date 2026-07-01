# 02 — Triển khai (Deploy)

## Truy cập server

- **VPS prod:** `root@14.225.213.73`, repo tại **`/opt/sangiagao`**.
- **Đăng nhập:** `ssh root@14.225.213.73` (bằng SSH key `~/.ssh/id_ed25519`, đã cài). Mật khẩu dự phòng còn (chưa tắt password auth).
- ⚠️ VPS này CHẠY CHUNG với **Sell 365** (app khác) — container `sell365_*`. **TUYỆT ĐỐI không đụng container sell365.**

**Container sàn giao gạo:** `backend`, `web`, `admin`, `rice_chat`, `rice_nginx`, `rice_postgres` (DB), `rice_redis`, `rice_mongodb`, `minio`, `rice_certbot`, `rice_db_backup`.

DB: `DB_NAME=rice_marketplace`, `DB_USER=rice_user`, host `postgres` (alias mạng `rice_internal`).

---

## Deploy — KHÔNG dùng docker-compose

Dùng script `infras/scripts/quick-deploy.sh`:
```bash
ssh root@14.225.213.73
cd /opt/sangiagao
git pull origin main
bash infras/scripts/quick-deploy.sh backend    # hoặc: web | admin | chat | all | migrate
```
Script tự: `git pull` → `docker build --no-cache` → stop/rm/run container (`docker run --env-file infras/.env.backend ...`) → restart nginx → health check.

**Bẫy:**
- `docker-compose.production.yml` ở repo gốc **KHÔNG được deploy dùng** (chỉ tham khảo). Env THẬT ở **`/opt/sangiagao/infras/.env.backend`** (và `.env.chat`).
- Health check báo **"chat: NOT running" là BÁO ĐỘNG GIẢ** (tìm tên `chat`, container thật `rice_chat`). "Deploy completed with 1 error(s)" thường do cái này — xem các dòng `[OK]` là chính.

**Quy trình chuẩn từ máy dev:** sửa code → build/test cục bộ → `git commit` + `git push origin HEAD:main` → SSH server `git pull` + `quick-deploy.sh <phần>`.

---

## Cấu hình môi trường quan trọng

- **`APP_ENV` KHÔNG set trên prod → chạy `env=development`.** Hệ quả: `config.Validate()` production KHÔNG chạy; 2 guard bảo mật NGỦ (chặn OTP mock + Redis bắt buộc). Không sao vì SMS=`zalo+mock` thật + Redis chạy. Muốn bật: set `APP_ENV=production` NHƯNG verify CORS_ORIGINS/DB_PASSWORD/MinIO/Redis qua Validate trước (kẻo app không khởi động).
- **`SMS_PROVIDER`** chỉ chấp nhận **`zalo`** hoặc **`zalo+mock`** (code chỉ hiện thực Zalo ZNS). Giá trị khác (`esms`/`mock`/rỗng) → `MockSender` → OTP không gửi. Cần đủ `ZALO_APP_ID/SECRET/TEMPLATE_ID/REFRESH_TOKEN`.
- **IAP:** Apple + Google IAP đang **TẮT** (thiếu env). Chỉ SePay hoạt động.

---

## Migrations (CÓ 2 HỆ THỐNG — dễ nhầm)

1. **Tự động** khi backend khởi động: `backend/internal/database/migrations/` (embedded, ~001–014). Backend tự `RunMigrations`.
2. **Thủ công** qua psql: `infras/migrations/` (015 → hiện tại). KHÔNG tự chạy:
   ```bash
   docker exec -i rice_postgres psql -U rice_user -d rice_marketplace < infras/migrations/NNN_xxx.sql
   ```
   (`quick-deploy.sh migrate` chạy LẠI tất cả file, dựa idempotency.)

**Quy tắc:** idempotent (`IF NOT EXISTS`, `DROP ... IF EXISTS`, `ON CONFLICT DO NOTHING`) + tương thích ngược (binary cũ chạy được với schema mới). **Áp migration TRƯỚC khi deploy binary.**

**Migration mới nhất: `041`. Số kế tiếp = 042.** (Cập nhật số này mỗi khi thêm migration.)

---

## Truy cập cơ sở dữ liệu

```bash
ssh root@14.225.213.73
docker exec -it rice_postgres psql -U rice_user -d rice_marketplace
# Backup nhanh:
docker exec rice_postgres pg_dump -U rice_user rice_marketplace | gzip > backup.sql.gz
```

---

## Backup khóa & bí mật

Bộ khóa quan trọng (keystore Android, API key Apple, SSH key, .env backend) được sao lưu
ở máy chủ tại `~/Desktop/SANGIAGAO-KEYS-BACKUP/` (kèm README + hướng dẫn vận hành). **File
bí mật — không commit lên git.** Quan trọng nhất: `upload-keystore.jks` (mất = không cập
nhật được app Android) và `PHONE_ENCRYPT_KEY` trong `.env.backend` (mất = không giải mã
được SĐT đã lưu). Xem thêm build/upload app ở [07-mobile](07-mobile.md).
