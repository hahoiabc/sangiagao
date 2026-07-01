# 01 — Kiến trúc tổng thể

## Monorepo 4 service + mobile + hạ tầng

| Service | Công nghệ | Port nội bộ | Thư mục | Ghi chú |
|---|---|---|---|---|
| Backend API | Go 1.25 / Gin | 8080 | `backend/` | Layered: handler→service→repository |
| Web công khai | Next.js / React | 3001 | `web/` | SEO (SSG/ISR), marketplace, member area |
| Admin | Next.js / React | 3000 | `admin/` | Quản trị (shadcn/ui) |
| Chat | Elixir / Phoenix | 4000 | `chat/` | WebSocket realtime |
| Mobile | Flutter | — | `mobile/` | iOS + Android |

**Hạ tầng:** PostgreSQL 16, Redis 7, MongoDB 7, MinIO (S3), Nginx, Cloudflare CDN.

---

## Backend layered (quan trọng)

- `internal/handler/` — HTTP: parse request, map lỗi → status.
- `internal/service/` — business logic. Interface repo định nghĩa ở `service/deps.go` (consumer-defined).
- `internal/repository/` — query Postgres (pgx/v5).
- `internal/handler/deps.go` — interface service mà handler dùng (để test + tránh import cycle).
- `cmd/api/main.go` — DI wiring + đăng ký route + cron goroutine.
- **Test:** `testify/mock`, mock struct trong `*_test.go`. CHƯA có integration test (cần testcontainers).

> Khi thêm method vào 1 service dùng qua interface: nhớ cập nhật interface ở **deps.go** (cả `service/` lẫn `handler/`) VÀ các mock struct trong `*_test.go` — nếu không sẽ vỡ build test.

---

## Luồng dữ liệu chính

- **Đăng tin / marketplace:** app/web → Backend REST → Postgres `listings`. Ảnh lên MinIO.
- **Chat:** app gửi REST `/conversations/:id/messages` → Postgres `messages` (NGUỒN THẬT), rồi relay qua Phoenix CHỈ để realtime. Chi tiết ở [03-mo-hinh-du-lieu](03-mo-hinh-du-lieu.md).
- **Thanh toán:** SePay webhook → Backend → cập nhật subscription + commission engine. Chi tiết ở [06-thanh-toan-goi](06-thanh-toan-goi.md).
- **Realtime/notify:** Redis (cache + token revoke) + FCM (push) + Phoenix WS (chat).
