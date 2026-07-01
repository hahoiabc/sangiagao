# Tài liệu Sàn Giao Gạo

> Cập nhật: 2026-06-30. Gom kiến thức "ngầm" (không hiển nhiên từ code) cho người
> vào fix/nâng cấp về sau. Mỗi khi phát hiện thêm kiến thức ngầm hoặc đổi hạ tầng
> → cập nhật tài liệu tương ứng.

## Mục lục

| # | Tài liệu | Đọc khi |
|---|---|---|
| 01 | [Kiến trúc](01-kien-truc.md) | Muốn hiểu tổng thể hệ thống |
| 02 | [Triển khai](02-trien-khai.md) | Deploy, áp migration, truy cập server, backup khóa |
| 03 | [Mô hình dữ liệu](03-mo-hinh-du-lieu.md) | Đụng schema/bảng/role/status |
| 04 | [Auth & RBAC](04-auth-rbac.md) | Đụng đăng nhập, token, phân quyền, tài khoản |
| 05 | [Affiliate](05-affiliate.md) | Đụng hoa hồng/đối tác |
| 06 | [Thanh toán & Gói](06-thanh-toan-goi.md) | Đụng thanh toán/gói thành viên/IAP |
| 07 | [Mobile](07-mobile.md) | Đụng app Flutter (list, build, upload store) |
| 08 | [Bảo mật](08-bao-mat.md) | Rà soát bảo mật |

Ảnh chụp lịch sử: [`archive/`](archive/) — audit 06/2026 (đã fix), store-readiness 04/2026 (đã live), PRD, reuse-checklist.

---

## Trạng thái dự án (30/06/2026)

- **Đã LIVE:** web + admin + backend (prod `14.225.213.73`, `/opt/sangiagao`). App trên cả 2 store.
- **Phiên bản app:** 1.6.6(52) đã duyệt · 1.6.7 BỎ QUA · **1.6.8(53)** (cuộn vô hạn) chờ duyệt · **1.6.9(54)** đã commit (2 màn lịch sử) chờ build. *(Số kế tiếp khi build: 1.6.9+54.)*
- **Migration mới nhất:** `infras/migrations/041`. Số kế tiếp = **042**.
- **Thanh toán:** chỉ SePay (QR chuyển khoản) hoạt động. Apple/Google IAP TẮT trên prod.
- **Affiliate:** đã mở cho mọi vai trò; `commission_records` vẫn rỗng (chưa phát sinh thật).

---

## Nợ kỹ thuật & Roadmap

| Mục | Mô tả | Ưu tiên |
|---|---|---|
| Test integration | Dựng testcontainers (Postgres+Redis) phủ: refund→clawback, expire→hide→renew, OTP, block, xóa TK. Hiện chỉ unit test. | Cao |
| APP_ENV=production | Đang `development` → guard chặn-OTP-mock + Redis-bắt-buộc đang ngủ. Bật sau khi verify CORS/DB/MinIO/Redis. | TB |
| Mobile single-flight refresh | Thêm khoá/Completer vào `mobile/.../api_service.dart` (web đã có). Hiện dựa ân hạn `rot:` 60s. | TB |
| Web chat polling | Web REST polling 15s thay vì nối Phoenix WS (hạ tầng WS đã có). | TB |
| Observability | Chưa có metrics/tracing/Crashlytics. | TB |
| Stats tổng cho list phân trang | my_referees/my_payouts: tổng chỉ tính theo trang đã tải → cần endpoint stats riêng nếu muốn chính xác. | Thấp |
| Giới hạn thiết bị đăng nhập | CỐ Ý không giới hạn (chủ chọn, siết sau) — ĐỪNG nhầm là bug. | Hoãn |
| Apple/Google IAP | Đang tắt; bật cần env `APP_STORE_*` / Google service account. | Khi cần |

---

## Quy ước khi fix (đọc trước khi sửa)

1. **Phân tích phụ thuộc, fix nền tảng trước** (schema/migration trước code).
2. **Fix chạm người dùng / ý đồ thiết kế / tiền / auth-OTP → TRÌNH DUYỆT chủ trước.** Bảo mật thuần (RBAC/DoS/IDOR nội bộ/escape XSS) → tự làm.
3. **Migration:** áp THỦ CÔNG (psql) TRƯỚC khi deploy binary; idempotent + tương thích ngược.
4. **Commit:** chỉ `git add` file của mình (repo hay có file dở dang pre-existing ở mobile/web). `git push origin HEAD:main`.
5. **Build/test trước commit.** Backend: `cd backend && go build ./... && go test ./internal/... && gofmt -l`. Web/admin: `npx tsc --noEmit`. App: `flutter analyze`.
6. **2 pattern lỗi hay gặp:** (a) *"định danh lệch pha"* — bảng/cột/role/status/constraint đổi mà không đồng bộ mọi nơi; (b) *lỗi nhánh phụ bị nuốt* — log Warn rồi return nil → tính năng tưởng chạy mà ghi 0. Khi fix money/security nhớ kiểm cả nhánh phụ + constraint.
