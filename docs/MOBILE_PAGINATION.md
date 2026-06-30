# Phân trang / Cuộn vô hạn (App Mobile)

> Cập nhật: 2026-06-30. Áp dụng cho app Flutter `mobile/`.
> Mục tiêu: mọi danh sách phình theo số thành viên đều cuộn vô hạn (scale 1 triệu user).

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
