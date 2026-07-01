# 07 — Mobile (App Flutter)

> Cập nhật: 2026-06-30. App Flutter ở `mobile/`. Gồm: (A) cuộn vô hạn / scaling list,
> (B) build & upload store, (C) quy ước version.

---

## A. Phân trang / Cuộn vô hạn — mục tiêu scale 1 triệu user

---

## 1. Widget dùng chung — `PaginatedListView`

File: `mobile/lib/widgets/paginated_list_view.dart`

```dart
PaginatedListView<T>({
  required PageFetcher<T> fetcher,      // Future<List<T>> Function(int page) — page bắt đầu từ 1
  required PagedItemBuilder<T> itemBuilder,  // Widget Function(ctx, T item, int index)
  Widget? header,        // widget cố định TRÊN list, cuộn cùng (vd thẻ thống kê / hồ sơ)
  Widget? emptyState,    // hiện khi không có item
  EdgeInsetsGeometry padding,
  int pageSize = 20,     // dùng để suy ra "còn dữ liệu" (lô trả về đầy = còn nữa)
})
```

Widget tự lo: `ScrollController`, tải-thêm-khi-gần-cuối (maxScrollExtent − 320), kéo-để-làm-mới
(`RefreshIndicator`), spinner ở đáy, chống tải-trùng, trạng thái loading/rỗng/lỗi.

**"Còn dữ liệu" suy ra từ heuristic:** lô trả về có `length >= pageSize` ⇒ còn trang sau.
Ưu điểm: KHÔNG cần `total` từ mọi endpoint. Nhược điểm nhỏ: nếu trang cuối đúng `pageSize` item
sẽ fetch thừa 1 lần (trả rỗng → dừng).

### Cách dùng cơ bản (danh sách đơn giản)
```dart
PaginatedListView<AppNotification>(
  fetcher: (page) async =>
      (await ref.read(apiServiceProvider).getNotifications(page: page, limit: 20)).data,
  emptyState: const EmptyState(...),
  itemBuilder: (context, n, i) => Column(children: [ ListTile(...), const Divider(height: 1) ]),
)
```
> Widget dùng `ListView.builder` (KHÔNG có separatorBuilder) → đưa Divider/spacing vào trong `itemBuilder`.

### Có header (màn dashboard: phần trên cố định + danh sách bên dưới)
Đặt toàn bộ phần đầu màn vào `header:`, phần danh sách là items. Vd `referral_screen` (thẻ mã +
thống kê = header, lịch sử hoa hồng = items); `subscription_screen` (gói/trạng thái = header, lịch
sử gia hạn = items); `seller_profile_screen` (hồ sơ người bán = header, đánh giá = items).

### Tải lại từ ngoài (sau thao tác làm đổi dữ liệu)
```dart
final _listKey = GlobalKey<PaginatedListViewState>();
// ... PaginatedListView(key: _listKey, ...)
// sau khi xóa/bump/đổi filter:
_listKey.currentState?.refresh();   // tải lại từ trang 1
```
Dùng ở: `my_listings` (sau bump/xóa/sửa), `marketplace` (đổi filter/lọc người-bị-chặn),
`subscription`/`referral` (sau khi nạp lại phần header).

### Đổi bộ lọc (search/filter) — vd marketplace
`fetcher` đọc state filter HIỆN TẠI (instance field) tại thời điểm gọi. Khi filter đổi:
`setState(() { ...filter... })` rồi `_listKey.currentState?.refresh()` → fetch lại trang 1 với filter mới.

---

## 2. Danh sách màn (trạng thái)

**Đã cuộn vô hạn (11 màn):**
notifications, system_inbox, inbox (ds hội thoại), seller_profile (đánh giá), referral (lịch sử
hoa hồng), marketplace (Sàn gạo), my_listings (Tin của tôi), my_referees, my_payouts,
feedback_history (Lịch sử góp ý), subscription (lịch sử gia hạn).

**Bounded — CỐ Ý KHÔNG phân trang** (không phình theo số user):
- Bảng giá (`price_board`): `GROUP BY category, rice_type` → tối đa vài chục dòng theo danh mục gạo.
- Danh mục sản phẩm (`getProductCatalog`): cố định theo catalog.
- Màn chi tiết (tin đăng / hộp thư): 1 item.
- Chat (tin nhắn trong 1 hội thoại): đã có cuộn-tải-tin-cũ riêng (`chat_screen`, kiểu reverse).

---

## 3. Quy ước backend phân trang

Đa số endpoint list ĐÃ nhận `?page=&limit=` và trả `{ data, total, page, limit }`. Khi thêm
endpoint list mới, theo mẫu này. Helper ví dụ (`referral_handler.go`):
```go
func parsePageLimit(c *gin.Context) (page, limit, offset int) // mặc định 1/20, trần 50
```
`getMyReferees` / `getMyPayouts` đã đổi từ LIMIT cứng (200/100) sang page/limit (mig không cần,
chỉ sửa handler).

App `api_service.dart`: hàm list nhận `{int page = 1, int limit = 20}`. Fetcher trích `.data`
(hoặc list trực tiếp) để truyền cho `PaginatedListView`.

---

## 4. Đánh đổi đã chấp nhận (đọc trước khi sửa)

- **my_referees**: bỏ thẻ "Tổng quan" (tổng count/active/hoa hồng) — vì tính từ list-đã-tải sẽ SAI
  khi phân trang. Muốn lại: cần endpoint stats tổng riêng → nhồi vào `header:`.
- **my_payouts**: tổng "Đã nhận/Chờ" cộng dồn theo trang đã tải (không phải toàn cục) — tương tự.
- **marketplace**: lọc người-bị-chặn nằm trong fetcher ⇒ 1 trang có người bị chặn trả < pageSize
  có thể dừng phân trang sớm. Hiếm, chấp nhận (đúng tinh thần "ẩn ngay" Apple GL 1.2).

---

## B. Build & upload store

Chạy trong `mobile/`. Nhớ set `version` ở `pubspec.yaml` trước (xem mục C).

**Android (AAB — tự tải lên Google Play):**
```bash
flutter build appbundle --release
# → build/app/outputs/bundle/release/app-release.aab
```
Cảnh báo "failed to strip debug symbols" là vô hại. Play Console → Production → Create new release → upload AAB.

**iOS (IPA — đẩy qua Transporter):**
```bash
flutter build ipa --release --export-options-plist=ios/ExportOptions.plist
# → build/ios/ipa/SanGiaGao.vn.ipa
```
- `ExportOptions.plist` để ký **TỰ ĐỘNG** (Xcode tự tạo profile). ĐỪNG đổi về manual (từng lỗi thiếu Associated Domains). Cần đăng nhập Apple ID trong Xcode.
- Mở app **Transporter** → kéo `.ipa` vào → **Deliver**. Sau ~5–30 phút build hiện ở App Store Connect (TestFlight/Distribution). Cài thử qua TestFlight trước khi duyệt.

**Định danh store:** App ID Apple `6761744869` · Team `4398LD7T8U` · Bundle `com.sangiagao.riceMarketplace`. Chữ ký Android: `mobile/android/app/upload-keystore.jks` (mật khẩu ở `key.properties`) — **mất là không cập nhật được app**, backup kỹ (xem [02-trien-khai](02-trien-khai.md)).

---

## C. Quy ước version

- `mobile/pubspec.yaml` dòng `version: X.Y.Z+build`. Build number PHẢI tăng mỗi lần upload.
- **Hiện tại:** 1.6.6(52) đã duyệt · **1.6.7 BỎ QUA** (lỡ dùng trên Apple) · 1.6.8(53) chờ duyệt · **1.6.9(54)** đã commit chờ build.
- Số kế tiếp khi build: **1.6.9+54**.

---

## D. Nạp danh sách kiểu POS/offline (nếu sau này cần)

Hiện các list đều tải server (offset/limit + cuộn vô hạn). Nếu về sau có màn cần cache offline
(như Sell365 POS), cân nhắc SQLite local + cột search không dấu. Chưa áp dụng ở sàn gạo.
