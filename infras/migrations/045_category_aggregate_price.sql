-- 045: Cờ "gộp giá" cho danh mục trên BẢNG GIÁ.
--   aggregate_price = true  → hiện GIÁ THẤP NHẤT (đ/unit)  — gạo, vận chuyển (đ/km)
--   aggregate_price = false → chỉ ĐẾM TIN ("N tin → Xem")  — máy móc/xe (đơn vị đa dạng, min dễ sai)
-- Mặc định true (gạo + danh mục cũ giữ hành vi gộp min). Mặt hàng phi chuẩn-hoá đặt false.

ALTER TABLE rice_categories
    ADD COLUMN IF NOT EXISTS aggregate_price BOOLEAN NOT NULL DEFAULT TRUE;

-- Máy móc/xe: đơn vị đa dạng (cái/bộ/chiếc) → đếm tin thay vì gộp min.
-- Vận chuyển (đ/km) sau này tạo thì để aggregate_price = true (chỉnh ở admin).
UPDATE rice_categories SET aggregate_price = FALSE WHERE kind = 'mat_hang';
