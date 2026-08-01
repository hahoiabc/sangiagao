-- 042: Số ngày tin đăng được hiển thị trên sàn.
-- '0' = KHÔNG giới hạn (mặc định, giữ nguyên hành vi cũ). Khi > 0, marketplace/
-- tìm kiếm/bảng giá chỉ hiện tin có hoạt động (tạo/sửa/bump) trong N ngày; chủ tin
-- bấm "Làm mới" để tin tươi lại và hiện lại. Chỉnh trong Admin (Cài đặt tin đăng).
INSERT INTO site_settings (key, value) VALUES ('listing_display_days', '0')
ON CONFLICT (key) DO NOTHING;
