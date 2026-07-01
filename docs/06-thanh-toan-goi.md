# 06 — Thanh toán & Gói thành viên

> Hoa hồng đối tác (affiliate) tách riêng ở [05-affiliate](05-affiliate.md).

## Nguồn thanh toán

| Nguồn | Trạng thái | Ghi chú |
|---|---|---|
| **SePay** (QR chuyển khoản) | ✅ ACTIVE | source `web`/`sepay`; webhook xác nhận |
| Apple IAP | ❌ TẮT trên prod | thiếu env `APP_STORE_*` |
| Google IAP | ❌ TẮT trên prod | thiếu Google service account |

App có màn IAP (StoreKit / Play Billing) sẵn trong code (`iap_service.dart`, `subscription_screen.dart`) + nút "Khôi phục giao dịch" (Apple bắt buộc) + disclosure tự-gia-hạn. Bật IAP khi có env.

## Gói thành viên (subscription)

- Bảng `subscriptions`: `plan` ∈ `free_trial` / `paid` / `reward` (mig 041 thêm `reward`). `status` ∈ `active` / `expired`. `expires_at`.
- **Hết hạn → ẩn tin:** cron `HideListingsForExpired` đổi listing `active`→`hidden_subscription` cho user hết hạn. **Loại trừ** staff (owner/admin/editor) + `is_internal` (miễn gói).
- **Gia hạn/kích hoạt lại** → restore listing về `active`.
- **Thưởng thời gian** (admin): `AdminReward` tạo subscription `plan='reward'` (miễn phí, không tính doanh thu). *(Bẫy: constraint `subscriptions_plan_check` phải chứa `'reward'` — mig 041.)*
- `PayableDelayDays = 45`: dùng cho vòng đời hoa hồng (T+45), xem [05-affiliate](05-affiliate.md).

## Miễn gói (không bị ẩn tin)

- **Staff:** `owner` / `admin` / `editor` — hàm `isPlatformStaff`.
- **Tài khoản nội bộ:** cờ `is_internal=true` (tạo thủ công). Xem [04-auth-rbac](04-auth-rbac.md).
- Cả 2 được loại trong `listing_service.Create` (không chặn đăng) + `HideListingsForExpired` (không ẩn).

## Hiển thị "gói vĩnh viễn" / trạng thái

- Màn Gói dịch vụ (`subscription_screen.dart`) hiện: trạng thái (còn N ngày / hết hạn), bảng giá gia hạn (IAP hoặc grid + liên hệ Zalo trên web/dev), lịch sử gia hạn (đã cuộn vô hạn — xem [07-mobile](07-mobile.md)).
