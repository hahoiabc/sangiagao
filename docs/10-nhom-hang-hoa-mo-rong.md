# 10 — Mở rộng nhóm hàng hoá (nông sản / mặt hàng)

> Sàn Giá Gạo không chỉ bán gạo. Tài liệu này mô tả cơ chế cho phép đăng tin & hiển thị
> bảng giá cho **mọi nhóm** (gạo, máy móc, dịch vụ vận chuyển...). Triển khai 05/10/2026.

## 1. Khái niệm

Mỗi **danh mục** (`rice_categories`) có 3 thuộc tính điều khiển hành vi:

| Cột | Giá trị | Ý nghĩa |
|---|---|---|
| `kind` | `nong_san` \| `mat_hang` | Kiểu danh mục → quyết định **luồng đăng tin** |
| `unit` | `kg`, `cái`, `km`, `chiếc`... | Đơn vị giá (hiển thị "đ/\<unit\>") |
| `aggregate_price` | `true` \| `false` | Cách hiển thị trên **bảng giá** |

### Ngữ nghĩa 2 kiểu

- **`nong_san`** (gạo, nếp, tấm): hàng chuẩn hoá. Giá **đ/kg** (chặn 5.000–99.000), số lượng
  **kg bắt buộc**, có **mùa vụ** + **chứng nhận**.
- **`mat_hang`** (máy móc, xe, ghe...): giá **đ/\<unit\>**, chỉ cần **> 0** (trần "vệ sinh" 100 tỷ,
  chặn gõ nhầm), **ẩn số lượng kg** (ngầm = 1), **ẩn mùa vụ + chứng nhận**. Người bán tự ghi
  chi tiết trong Mô tả.

### Cờ `aggregate_price` (chỉ ảnh hưởng bảng giá, không ảnh hưởng đăng tin)

- `true`  → bảng giá hiện **giá thấp nhất** "Từ X đ/\<unit\>" (so sánh được: gạo đ/kg; vận chuyển đ/km).
- `false` → bảng giá chỉ **đếm tin** ("N tin → Xem"), KHÔNG gộp giá (khi đơn vị trong nhóm đa dạng,
  vd máy móc cái/bộ — gộp min sẽ sai lệch).

## 2. Các nhóm hiện có (05/10/2026)

| key | label | kind | unit | aggregate | Ý nghĩa |
|---|---|---|---|---|---|
| gao_deo_thom, gao_kho, tam_deo_thom, tam_kho, nep | (gạo...) | nong_san | kg | true | Gạo — min đ/kg |
| `MAYMOC` | Máy - Thiết Bị | mat_hang | cái | **false** | **Bán máy móc** → đếm tin |
| `XeGe` | Xe Tải - Ghe | mat_hang | km | **true** | **Dịch vụ vận chuyển** → min đ/km |

## 3. Migrations

- `044_category_kind_unit.sql` — thêm `kind` (default `nong_san`) + `unit` (default `kg`); backfill
  MAYMOC/XeGe = `mat_hang`.
- `045_category_aggregate_price.sql` — thêm `aggregate_price` (default `true`); mat_hang → `false`.
- ⚠ Migrations 015+ **áp THỦ CÔNG** bằng psql (embed trong backend chỉ tới 014):
  `docker exec -i rice_postgres psql -U rice_user -d rice_marketplace < infras/migrations/0XX.sql`

## 4. Backend

- **Model**: `CatalogCategory`, `RiceCategory`, `PriceBoardCategory` đều có `Kind/Unit/AggregatePrice`.
  Hằng `model.CategoryKindCommodity` / `CategoryKindItem`.
- **Validate đăng tin** (`listing_service.go` → `Create`/`Update`): phân nhánh theo `kind`
  (helper `categoryKindUnit`). `validItemPrice` (>0, <100 tỷ) cho mặt hàng; `validPricePerKG` cho
  nông sản. Mặt hàng: `QuantityKG <= 0` → đặt = 1.
- **`CreateListingRequest.QuantityKG`** nới thành `omitempty,gte=0` (service enforce theo kind).
- **Bảng giá** (`GetPriceBoard`): bỏ qua danh mục 0 sản phẩm; `products` luôn non-nil (`[]`); trả
  kèm `kind/unit/aggregate_price` để client quyết hiển thị.
- **Admin API**: `CreateCategoryRequest`/`UpdateCategoryRequest` nhận `kind/unit/aggregate_price`.

## 5. Admin (quản lý danh mục)

`/catalog` → dialog danh mục có: **Kiểu** (select), **Đơn vị giá** (khoá "kg" khi nông sản),
checkbox **"Hiện giá thấp nhất trên bảng giá"** (hiện khi mặt hàng). Đổi Kiểu tự set mặc định
(nong_san→unit kg + aggregate true; mat_hang→aggregate false).

## 6. Web

- **Đăng/sửa tin** (`tin-dang/tao-moi`, `tin-dang/sua/[id]`): đọc `kind` của danh mục → nhãn giá
  "Giá (đ/\<unit\>)", **ẩn** Số lượng/Mùa vụ cho mặt hàng, validate theo kind.
- **Bảng giá** (`bang-gia`): caption theo nhóm; `aggregate_price=false` → hiện "N tin" thay vì giá.

## 7. Mobile (1.6.14)

- `product_catalog.dart`, `price_board.dart`: thêm `kind/unit/aggregatePrice` (parse null-safe).
- `create_listing_screen` / `edit_listing_screen`: nhãn giá đ/\<unit\>, ẩn số lượng/mùa vụ cho mặt hàng.
- `price_board_screen`: caption theo nhóm + đếm tin khi `aggregate_price=false`.

## 8. Thêm một nhóm mới (quy trình chủ tự làm)

1. **Admin → Danh mục → Thêm**: đặt Mã (key, không dấu, vd `phan_bon`), Tên, chọn **Kiểu** + **Đơn vị**
   + **Hiện giá thấp nhất** (nếu muốn).
2. **Admin → Sản phẩm → Thêm**: thêm vài **loại** dưới danh mục đó (vd "Phân Urê", "Phân NPK").
   ⚠ BẮT BUỘC có ≥1 loại thì người dùng mới **đăng tin được** (form đăng tin bắt chọn loại) và bảng giá mới hiện.
3. Xong — người dùng đăng tin được ngay (web). App cần bản ≥1.6.14.

## 9. Bẫy / lưu ý

- **Cache bảng giá (Redis)**: key `priceboard:v1:d<N>`, TTL **5 phút**. Đổi cấu hình danh mục bằng
  SQL/admin → bảng giá trễ ≤5'. Xả ngay: lấy mật khẩu từ `REDIS_URL` (`redis://:PASS@host`) rồi
  `redis-cli -a $PASS --scan --pattern 'priceboard:*' | xargs redis-cli -a $PASS del`.
  **KHÔNG restart `rice_redis`** (giữ refresh-token Zalo).
- **Catalog cache (in-memory backend)**: TTL **1 giờ**. Đổi `kind` ở admin → validation đăng tin
  trễ ≤1h. (Bảng giá đọc DB trực tiếp nên chỉ trễ do cache Redis 5'.)
- **Tương thích ngược**: client cũ không gửi `kind` → danh mục mặc định `nong_san` → gạo y nguyên.
- Cột DB `price_per_kg` / `quantity_kg` **giữ nguyên tên** (chỉ là số) — đổi nhãn ở UI theo kind,
  không migrate dữ liệu.
