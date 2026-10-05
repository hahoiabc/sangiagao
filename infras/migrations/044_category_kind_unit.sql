-- 044: Thêm "kiểu" (kind) + đơn vị (unit) cho danh mục — hỗ trợ nhóm phi-gạo
-- (máy móc, xe, ghe...). Luồng đăng tin + bảng giá đọc theo kind:
--   - 'nong_san' (gạo/nếp/tấm): giá đ/kg (5.000–99.000), số lượng kg, có mùa vụ/chứng nhận
--   - 'mat_hang'  (máy móc/xe...): giá đ/<unit>, khoảng giá rộng, ẩn số lượng kg/mùa vụ/chứng nhận
-- Mặc định 'nong_san'/'kg' → các danh mục gạo cũ GIỮ NGUYÊN hành vi.

ALTER TABLE rice_categories
    ADD COLUMN IF NOT EXISTS kind VARCHAR(20) NOT NULL DEFAULT 'nong_san',
    ADD COLUMN IF NOT EXISTS unit VARCHAR(20) NOT NULL DEFAULT 'kg';

-- Backfill 2 danh mục phi-gạo đã tạo sẵn (chủ thêm để mở rộng) → kiểu "mặt hàng".
-- Đơn vị gợi ý; chủ chỉnh lại trong quản lý danh mục (admin).
UPDATE rice_categories SET kind = 'mat_hang', unit = 'cái'   WHERE key = 'MAYMOC';
UPDATE rice_categories SET kind = 'mat_hang', unit = 'chiếc' WHERE key = 'XeGe';
