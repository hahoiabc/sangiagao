# Hạ Tầng Rice Marketplace (Infrastructure)

Thư mục này chứa toàn bộ cấu hình hạ tầng cho môi trường development local.

---

## Tổng Quan

| Service | Image | Port (local) | Mục đích |
|---------|-------|-------------|----------|
| **PostgreSQL** | postgres:16-alpine | `5435` | Database chính (users, listings, subscriptions...) |
| **Redis** | redis:7-alpine | `6381` | Cache, sessions, rate limiting |
| **MongoDB** | mongo:7 | `27018` | Lưu trữ tin nhắn chat |

> **Lưu ý**: Ports được đặt khác default để tránh xung đột với các dự án khác trên máy.

---

## Yêu Cầu

- **Docker Desktop** (đã cài và đang chạy)
- Không cần cài PostgreSQL, Redis, MongoDB trực tiếp — tất cả chạy qua Docker.

---

## Cấu Trúc Thư Mục

```
infras/
├── .env                          # Biến môi trường (KHÔNG commit bản production)
├── .env.example                  # Template .env (commit lên git)
├── README.md                     # File này
├── docker/
│   └── docker-compose.yml        # Docker Compose cho 3 services
├── configs/
│   ├── postgres/
│   │   └── init.sql              # Script khởi tạo DB (8 tables + indexes + admin user)
│   └── mongodb/
│       └── init.js               # Script khởi tạo collections + indexes
└── scripts/
    ├── start.sh                  # Khởi động hạ tầng
    ├── stop.sh                   # Dừng hạ tầng (giữ data)
    ├── status.sh                 # Xem trạng thái
    ├── test.sh                   # Test kết nối tất cả services
    └── reset.sh                  # ⚠️ Xóa sạch data và tạo lại từ đầu
```

---

## Bắt Đầu Nhanh

### 1. Cấu hình môi trường

```bash
# Nếu chưa có file .env, tạo từ template:
cp infras/.env.example infras/.env

# Chỉnh sửa nếu cần (ports, passwords...):
# vim infras/.env
```

File `.env` hiện tại dùng các port sau (đã tránh xung đột):
- PostgreSQL: **5435** (thay vì 5432 mặc định)
- Redis: **6381** (thay vì 6379 mặc định)
- MongoDB: **27018** (thay vì 27017 mặc định)

### 2. Khởi động hạ tầng

```bash
# Cách 1: Dùng script (khuyến nghị)
./infras/scripts/start.sh

# Cách 2: Dùng docker compose trực tiếp
cd infras/docker
docker compose --env-file ../.env up -d
```

Script `start.sh` sẽ:
- Kiểm tra Docker daemon đang chạy
- Kiểm tra file .env tồn tại
- Pull images và khởi động containers
- Đợi tất cả services healthy
- In thông tin kết nối

### 3. Kiểm tra kết nối

```bash
./infras/scripts/test.sh
```

Kết quả mong đợi:
```
🧪 Test kết nối Rice Marketplace Infrastructure
================================================
1️⃣  PostgreSQL (localhost:5435)
   ✅ Kết nối OK
   ✅ Số tables: 8
   ✅ Admin user: Admin
   ✅ Extensions: uuid-ossp, pg_trgm
2️⃣  Redis (localhost:6381)
   ✅ Kết nối OK (PONG)
   ✅ SET/GET hoạt động
3️⃣  MongoDB (localhost:27018)
   ✅ Kết nối OK
   ✅ Collections: conversations, messages
   ✅ Messages indexes: 4
================================================
🎉 KẾT QUẢ: 3/3 services PASSED
================================================
```

---

## Lệnh Thường Dùng

### Scripts

| Lệnh | Mô tả |
|-------|--------|
| `./infras/scripts/start.sh` | Khởi động toàn bộ hạ tầng |
| `./infras/scripts/stop.sh` | Dừng hạ tầng (data vẫn còn) |
| `./infras/scripts/status.sh` | Xem trạng thái + resource usage |
| `./infras/scripts/test.sh` | Test kết nối tất cả services |
| `./infras/scripts/reset.sh` | **Xóa sạch** data và tạo lại từ đầu |

### Kết nối trực tiếp vào database

```bash
# PostgreSQL - Mở psql shell
docker exec -it rice_postgres psql -U rice_user -d rice_marketplace

# Ví dụ queries:
#   \dt                           -- Liệt kê tables
#   SELECT * FROM users;          -- Xem users
#   \d listings                   -- Xem cấu trúc table listings

# Redis - Mở redis-cli
docker exec -it rice_redis redis-cli

# Ví dụ commands:
#   PING                          -- Test kết nối
#   KEYS *                        -- Liệt kê tất cả keys
#   INFO memory                   -- Xem memory usage

# MongoDB - Mở mongosh
docker exec -it rice_mongodb mongosh -u rice_user -p rice_secret_dev --authenticationDatabase admin rice_chat

# Ví dụ commands:
#   db.getCollectionNames()       -- Liệt kê collections
#   db.messages.find()            -- Xem messages
#   db.messages.getIndexes()      -- Xem indexes
```

### Docker Compose

```bash
# Xem logs (tất cả services)
cd infras/docker && docker compose --env-file ../.env logs -f

# Xem logs 1 service cụ thể
cd infras/docker && docker compose --env-file ../.env logs -f postgres

# Restart 1 service
cd infras/docker && docker compose --env-file ../.env restart redis
```

---

## Thông Tin Kết Nối (Cho Backend)

Sử dụng các giá trị này khi cấu hình backend Golang:

### PostgreSQL
```
Host:     localhost
Port:     5435
User:     rice_user
Password: rice_secret_dev
Database: rice_marketplace
DSN:      postgres://rice_user:rice_secret_dev@localhost:5435/rice_marketplace?sslmode=disable
```

### Redis
```
Host:     localhost
Port:     6381
URL:      redis://localhost:6381
```

### MongoDB
```
Host:     localhost
Port:     27018
User:     rice_user
Password: rice_secret_dev
Database: rice_chat
URL:      mongodb://rice_user:rice_secret_dev@localhost:27018/rice_chat?authSource=admin
```

---

## Database Schema

### PostgreSQL (8 tables)

| Table | Mô tả | Cột chính |
|-------|--------|-----------|
| `users` | Người dùng | id, phone, role, name, province, is_blocked |
| `subscriptions` | Gói subscription | user_id, plan, expires_at, status |
| `listings` | Tin đăng sản phẩm | user_id, title, rice_type, province, price_per_kg, images, status |
| `ratings` | Đánh giá người bán | reviewer_id, seller_id, stars, comment |
| `reports` | Báo cáo vi phạm | reporter_id, target_type, target_id, reason, status |
| `notifications` | Thông báo | user_id, type, title, body, is_read |
| `device_tokens` | FCM tokens | user_id, token, platform |
| `otp_requests` | OTP xác minh | phone, code, attempts, expires_at |

**Đặc biệt:**
- Full-text search trên `listings` (tsvector) — tự động cập nhật khi title/description thay đổi
- Extensions: `uuid-ossp` (UUID generation), `pg_trgm` (trigram search)
- Trigger `update_updated_at` tự động cập nhật `updated_at`
- Admin user mặc định: SĐT `0900000000`

### MongoDB (2 collections)

| Collection | Mô tả | Schema |
|------------|--------|--------|
| `messages` | Tin nhắn chat | conversation_id, sender_id, content, type, timestamp, read_at |
| `conversations` | Cuộc hội thoại | participants[], listing_id, last_message_at |

---

## Xử Lý Sự Cố

### Docker daemon chưa chạy
```
❌ Docker daemon chưa chạy!
→ Mở Docker Desktop, đợi icon Docker trên menu bar chuyển sang xanh, rồi chạy lại.
```

### Port bị chiếm
```
Error: port is already allocated
→ Sửa port trong infras/.env (POSTGRES_PORT, REDIS_PORT, MONGO_PORT)
→ Rồi chạy lại: ./infras/scripts/start.sh
```

### Init script không chạy
```
Tables = 0 (chưa có tables)
→ Docker chỉ chạy init script khi tạo volume lần đầu.
→ Chạy: ./infras/scripts/reset.sh để xóa volume và tạo lại.
```

### Muốn xem logs chi tiết
```bash
# Xem logs realtime
cd infras/docker && docker compose --env-file ../.env logs -f

# Xem logs của 1 service
docker logs rice_postgres --tail 50
docker logs rice_redis --tail 50
docker logs rice_mongodb --tail 50
```
