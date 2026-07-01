# Sàn Giao Gạo — sangiagao.vn

Sàn giao dịch & kết nối mua bán gạo trực tuyến Việt Nam. Người bán đăng tin gạo,
người mua tìm kiếm theo loại/giá/vùng, nhắn tin trực tiếp; có bảng giá, gói thành
viên, và chương trình đối tác (affiliate).

- **Web/API:** https://sangiagao.vn
- **Admin:** https://admin.sangiagao.vn
- **App:** iOS (App Store, id 6761744869) + Android (Google Play)

---

## Kiến trúc (monorepo 4 service + mobile)

| Service | Công nghệ | Thư mục |
|---|---|---|
| Backend API | Go 1.25 / Gin | `backend/` |
| Web công khai | Next.js / React | `web/` |
| Admin | Next.js / React | `admin/` |
| Chat realtime | Elixir / Phoenix | `chat/` |
| Mobile | Flutter | `mobile/` |
| Hạ tầng (Docker/nginx/migrations) | — | `infras/` |

Hạ tầng: PostgreSQL 16, Redis 7, MongoDB 7, MinIO (S3), Nginx, Cloudflare.

---

## Bắt đầu nhanh (dev)

```bash
# Backend
cd backend && go build ./... && go test ./internal/...

# Web / Admin
cd web && npm install && npm run dev      # cổng 3001
cd admin && npm install && npm run dev    # cổng 3000

# Mobile
cd mobile && flutter pub get && flutter run
```

Deploy production: xem [`docs/02-trien-khai.md`](docs/02-trien-khai.md).

---

## 📚 Tài liệu

Toàn bộ tài liệu ở **[`docs/`](docs/)** — bắt đầu từ **[`docs/README.md`](docs/README.md)** (mục lục + trạng thái dự án + roadmap).

| # | Tài liệu | Nội dung |
|---|---|---|
| 01 | [Kiến trúc](docs/01-kien-truc.md) | 4 service, layered backend, hạ tầng |
| 02 | [Triển khai](docs/02-trien-khai.md) | Deploy, server, migration, backup khóa |
| 03 | [Mô hình dữ liệu](docs/03-mo-hinh-du-lieu.md) | Bảng/quy ước + các "bẫy" |
| 04 | [Auth & RBAC](docs/04-auth-rbac.md) | JWT/session, phân quyền, tài khoản (tạo/nội bộ/xóa) |
| 05 | [Affiliate](docs/05-affiliate.md) | Hoa hồng, công thức, thỏa thuận |
| 06 | [Thanh toán & Gói](docs/06-thanh-toan-goi.md) | SePay, IAP, gói thành viên |
| 07 | [Mobile](docs/07-mobile.md) | Cuộn vô hạn, scaling list, build/upload store |
| 08 | [Bảo mật](docs/08-bao-mat.md) | Tóm tắt bảo mật (audit đã xử lý) |

Ảnh chụp lịch sử (audit cũ, PRD…): [`docs/archive/`](docs/archive/). Hướng dẫn cho AI/Claude: [`CLAUDE.md`](CLAUDE.md).
