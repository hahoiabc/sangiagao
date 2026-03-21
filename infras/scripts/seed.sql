-- =============================================================
-- Rice Marketplace — Seed 100 users
-- 60 sellers (30 with active 30-day subscription) + 40 buyers
-- Working avatars via ui-avatars.com
-- Idempotent: ON CONFLICT DO NOTHING
-- =============================================================

BEGIN;

-- ========================
-- SELLERS (60 accounts) — phone 0903xxxxxx
-- ========================
INSERT INTO users (phone, role, name, avatar_url, address, province, description, org_name, accepted_tos_at, created_at, updated_at) VALUES
('0903000001', 'seller', 'Nguyễn Văn An',      'https://ui-avatars.com/api/?name=Nguyen+An&background=0D8ABC&color=fff&size=200',       '123 Lê Lợi, Q.1',             'TP. Hồ Chí Minh', 'Chuyên cung cấp gạo ST25 chính hãng Sóc Trăng',              'HTX Gạo Sóc Trăng',        NOW() - interval '50 days', NOW() - interval '50 days', NOW() - interval '50 days'),
('0903000002', 'seller', 'Trần Thị Bích',       'https://ui-avatars.com/api/?name=Tran+Bich&background=E91E63&color=fff&size=200',       '456 Trần Hưng Đạo',           'An Giang',         'Đại lý gạo lớn nhất An Giang, hơn 10 năm kinh nghiệm',      'Đại Lý Gạo Phú Tân',       NOW() - interval '45 days', NOW() - interval '45 days', NOW() - interval '45 days'),
('0903000003', 'seller', 'Lê Hoàng Cường',      'https://ui-avatars.com/api/?name=Le+Cuong&background=4CAF50&color=fff&size=200',        '789 Nguyễn Trãi',              'Đồng Tháp',        'Gạo hữu cơ, gạo lứt, gạo mầm cao cấp',                      'Gạo Hữu Cơ Đồng Tháp',    NOW() - interval '40 days', NOW() - interval '40 days', NOW() - interval '40 days'),
('0903000004', 'seller', 'Phạm Minh Đức',       'https://ui-avatars.com/api/?name=Pham+Duc&background=FF9800&color=fff&size=200',        '12 Phan Đình Phùng',           'Kiên Giang',       'Gạo Jasmine, gạo thơm đặc sản miền Tây',                     'Công Ty Gạo Rạch Giá',     NOW() - interval '38 days', NOW() - interval '38 days', NOW() - interval '38 days'),
('0903000005', 'seller', 'Hoàng Thị Ái',        'https://ui-avatars.com/api/?name=Hoang+Ai&background=9C27B0&color=fff&size=200',        '34 Hai Bà Trưng',              'Cần Thơ',          'Chuyên gạo xuất khẩu, giá sỉ cạnh tranh',                    'HTX Nông Nghiệp Cần Thơ',  NOW() - interval '35 days', NOW() - interval '35 days', NOW() - interval '35 days'),
('0903000006', 'seller', 'Võ Thanh Phong',       'https://ui-avatars.com/api/?name=Vo+Phong&background=00BCD4&color=fff&size=200',        '56 Lý Thường Kiệt',            'Long An',          'Gạo Nàng Hoa, gạo thơm lài Long An',                         'Gạo Tân An',               NOW() - interval '32 days', NOW() - interval '32 days', NOW() - interval '32 days'),
('0903000007', 'seller', 'Đặng Văn Giang',      'https://ui-avatars.com/api/?name=Dang+Giang&background=795548&color=fff&size=200',      '78 Điện Biên Phủ',             'Hậu Giang',        'Nông sản sạch Hậu Giang, gạo thơm đặc sản',                 'Nông Sản Hậu Giang',       NOW() - interval '30 days', NOW() - interval '30 days', NOW() - interval '30 days'),
('0903000008', 'seller', 'Bùi Thị Hạnh',        'https://ui-avatars.com/api/?name=Bui+Hanh&background=607D8B&color=fff&size=200',        '90 Võ Văn Tần',                'Vĩnh Long',        'Gạo lứt đỏ, gạo mầm tốt cho sức khỏe',                      'Gạo Sạch Vĩnh Long',       NOW() - interval '28 days', NOW() - interval '28 days', NOW() - interval '28 days'),
('0903000009', 'seller', 'Ngô Quốc Huy',        'https://ui-avatars.com/api/?name=Ngo+Huy&background=F44336&color=fff&size=200',         '102 Cách Mạng Tháng 8',        'Tiền Giang',       'Gạo tấm, gạo nếp, gạo đặc sản Mỹ Tho',                     'Gạo Mỹ Tho',               NOW() - interval '25 days', NOW() - interval '25 days', NOW() - interval '25 days'),
('0903000010', 'seller', 'Lý Thanh Tùng',       'https://ui-avatars.com/api/?name=Ly+Tung&background=3F51B5&color=fff&size=200',         '14 Nam Kỳ Khởi Nghĩa',         'Bạc Liêu',         'Gạo một bụi đỏ Bạc Liêu, đặc sản ĐBSCL',                    'HTX Gạo Bạc Liêu',         NOW() - interval '22 days', NOW() - interval '22 days', NOW() - interval '22 days'),
('0903000011', 'seller', 'Trương Thị Kim',      'https://ui-avatars.com/api/?name=Truong+Kim&background=8BC34A&color=fff&size=200',      '26 Pasteur',                    'Sóc Trăng',        'Gạo ST25 đạt giải gạo ngon nhất thế giới',                   'Gạo ST25 Sóc Trăng',       NOW() - interval '20 days', NOW() - interval '20 days', NOW() - interval '20 days'),
('0903000012', 'seller', 'Mai Xuân Lâm',        'https://ui-avatars.com/api/?name=Mai+Lam&background=CDDC39&color=333&size=200',         '38 Nguyễn Du',                  'Trà Vinh',         'Gạo Khao Dawk Mali, gạo thơm Trà Vinh',                     'Nông Trại Trà Vinh',       NOW() - interval '18 days', NOW() - interval '18 days', NOW() - interval '18 days'),
('0903000013', 'seller', 'Đinh Công Minh',      'https://ui-avatars.com/api/?name=Dinh+Minh&background=FF5722&color=fff&size=200',       '50 Lê Duẩn',                   'Cà Mau',           'Gạo hữu cơ vùng đất mũi Cà Mau',                            'Gạo Đất Mũi',              NOW() - interval '16 days', NOW() - interval '16 days', NOW() - interval '16 days'),
('0903000014', 'seller', 'Phan Thị Ngọc',       'https://ui-avatars.com/api/?name=Phan+Ngoc&background=673AB7&color=fff&size=200',       '62 Trường Chinh',               'Bến Tre',          'Gạo dẻo Bến Tre, gạo nàng thơm chợ Đào',                    'Gạo Bến Tre Xanh',         NOW() - interval '14 days', NOW() - interval '14 days', NOW() - interval '14 days'),
('0903000015', 'seller', 'Huỳnh Quốc Bảo',     'https://ui-avatars.com/api/?name=Huynh+Bao&background=009688&color=fff&size=200',       '74 Nguyễn Văn Cừ',             'TP. Hồ Chí Minh',  'Phân phối gạo sỉ và lẻ khu vực TP.HCM',                      'Gạo Sài Gòn',              NOW() - interval '12 days', NOW() - interval '12 days', NOW() - interval '12 days'),
('0903000016', 'seller', 'Lê Thị Phương',       'https://ui-avatars.com/api/?name=Le+Phuong&background=2196F3&color=fff&size=200',       '86 Bùi Viện',                   'Đồng Tháp',        'Gạo sen Đồng Tháp, gạo thơm cao cấp',                       'Gạo Sen Tháp Mười',        NOW() - interval '10 days', NOW() - interval '10 days', NOW() - interval '10 days'),
('0903000017', 'seller', 'Trần Đình Quân',      'https://ui-avatars.com/api/?name=Tran+Quan&background=FFC107&color=333&size=200',       '98 Phạm Ngũ Lão',              'An Giang',          'Gạo IR 50404, gạo xuất khẩu An Giang',                       'Công Ty XK Gạo An Giang',  NOW() - interval '8 days',  NOW() - interval '8 days',  NOW() - interval '8 days'),
('0903000018', 'seller', 'Nguyễn Hoàng Sơn',    'https://ui-avatars.com/api/?name=Nguyen+Son&background=E040FB&color=fff&size=200',      '110 Tôn Đức Thắng',            'Kiên Giang',        'Gạo nếp than, gạo nếp cái hoa vàng',                         'Gạo Nếp Phú Quốc',         NOW() - interval '6 days',  NOW() - interval '6 days',  NOW() - interval '6 days'),
('0903000019', 'seller', 'Đỗ Thị Thanh',        'https://ui-avatars.com/api/?name=Do+Thanh&background=76FF03&color=333&size=200',        '122 Hùng Vương',                'Cần Thơ',           'Chuyên gạo lứt gạo mầm cho sức khỏe',                       'Gạo Lứt Cần Thơ',          NOW() - interval '5 days',  NOW() - interval '5 days',  NOW() - interval '5 days'),
('0903000020', 'seller', 'Vũ Minh Triết',       'https://ui-avatars.com/api/?name=Vu+Triet&background=40C4FF&color=333&size=200',        '134 Lê Hồng Phong',            'Long An',           'Gạo tài nguyên Long An, giá rẻ chất lượng',                  'HTX Gạo Long An',          NOW() - interval '4 days',  NOW() - interval '4 days',  NOW() - interval '4 days'),
('0903000021', 'seller', 'Cao Văn Trung',       'https://ui-avatars.com/api/?name=Cao+Trung&background=FF6E40&color=fff&size=200',       '146 Ngô Quyền',                'Hậu Giang',         'Gạo sạch không thuốc trừ sâu',                               'Nông Sản Sạch HG',         NOW() - interval '48 days', NOW() - interval '48 days', NOW() - interval '48 days'),
('0903000022', 'seller', 'Tô Thị Uyên',         'https://ui-avatars.com/api/?name=To+Uyen&background=EA80FC&color=333&size=200',         '158 Bà Triệu',                 'Vĩnh Long',         'Gạo thơm Vĩnh Long, đóng gói đẹp làm quà',                  'Gạo Quà Vĩnh Long',        NOW() - interval '42 days', NOW() - interval '42 days', NOW() - interval '42 days'),
('0903000023', 'seller', 'Lâm Quốc Việt',       'https://ui-avatars.com/api/?name=Lam+Viet&background=B388FF&color=333&size=200',        '170 Lý Tự Trọng',              'Tiền Giang',        'Gạo tấm thơm, gạo dẻo giá sỉ',                              'Gạo Sỉ Tiền Giang',        NOW() - interval '36 days', NOW() - interval '36 days', NOW() - interval '36 days'),
('0903000024', 'seller', 'Hồ Thị Xuân',         'https://ui-avatars.com/api/?name=Ho+Xuan&background=82B1FF&color=333&size=200',         '182 Hai Bà Trưng',              'Bạc Liêu',          'Gạo đặc sản miền Tây, ship toàn quốc',                       'Gạo Miền Tây Online',      NOW() - interval '34 days', NOW() - interval '34 days', NOW() - interval '34 days'),
('0903000025', 'seller', 'Dương Văn Yên',       'https://ui-avatars.com/api/?name=Duong+Yen&background=80D8FF&color=333&size=200',       '194 Đinh Tiên Hoàng',           'Sóc Trăng',         'Nhà phân phối gạo ST24, ST25 chính gốc',                     'Gạo Chính Gốc ST',         NOW() - interval '30 days', NOW() - interval '30 days', NOW() - interval '30 days'),
('0903000026', 'seller', 'Nguyễn Thị Ánh',      'https://ui-avatars.com/api/?name=Nguyen+Anh&background=A7FFEB&color=333&size=200',      '206 Trần Phú',                  'Trà Vinh',          'Gạo thơm lài, gạo sạch từ đồng ruộng',                       'Gạo Nhà Nông Trà Vinh',    NOW() - interval '28 days', NOW() - interval '28 days', NOW() - interval '28 days'),
('0903000027', 'seller', 'Châu Minh Khải',      'https://ui-avatars.com/api/?name=Chau+Khai&background=B9F6CA&color=333&size=200',       '218 Lê Lai',                    'Cà Mau',            'Gạo organic Cà Mau, không hóa chất',                         'Organic Farm Cà Mau',      NOW() - interval '26 days', NOW() - interval '26 days', NOW() - interval '26 days'),
('0903000028', 'seller', 'Trịnh Thị Duyên',     'https://ui-avatars.com/api/?name=Trinh+Duyen&background=CCFF90&color=333&size=200',     '230 Phạm Văn Đồng',            'Bến Tre',           'Gạo dừa Bến Tre, gạo thơm xứ dừa',                          'Gạo Xứ Dừa',               NOW() - interval '24 days', NOW() - interval '24 days', NOW() - interval '24 days'),
('0903000029', 'seller', 'Lưu Hoàng Đạo',       'https://ui-avatars.com/api/?name=Luu+Dao&background=F4FF81&color=333&size=200',         '242 Võ Thị Sáu',               'TP. Hồ Chí Minh',   'Kho gạo lớn Q.8, bán sỉ lẻ đa dạng',                        'Kho Gạo Quận 8',           NOW() - interval '22 days', NOW() - interval '22 days', NOW() - interval '22 days'),
('0903000030', 'seller', 'Tạ Thị Hồng',         'https://ui-avatars.com/api/?name=Ta+Hong&background=FFD180&color=333&size=200',         '254 Nguyễn Thái Học',           'Đồng Tháp',         'HTX gạo sạch Sa Đéc, chất lượng cao',                        'HTX Gạo Sa Đéc',           NOW() - interval '20 days', NOW() - interval '20 days', NOW() - interval '20 days'),
-- Sellers 31-60 (without active subscription)
('0903000031', 'seller', 'Phùng Văn Hải',       'https://ui-avatars.com/api/?name=Phung+Hai&background=FF8A80&color=333&size=200',       '10 Trần Quốc Toản',            'An Giang',          'Gạo Châu Đốc, gạo thơm An Giang',                           'Gạo Châu Đốc',             NOW() - interval '60 days', NOW() - interval '60 days', NOW() - interval '60 days'),
('0903000032', 'seller', 'Kiều Thị Lan',        'https://ui-avatars.com/api/?name=Kieu+Lan&background=FF80AB&color=333&size=200',        '22 Lê Văn Sỹ',                 'Kiên Giang',        'Gạo Rạch Giá thơm ngon',                                     'Gạo Kiên Giang Xanh',      NOW() - interval '55 days', NOW() - interval '55 days', NOW() - interval '55 days'),
('0903000033', 'seller', 'Sơn Minh Long',       'https://ui-avatars.com/api/?name=Son+Long&background=8C9EFF&color=fff&size=200',        '34 Hoàng Văn Thụ',              'Cần Thơ',           'Gạo Cần Thơ chất lượng',                                     'Gạo Ninh Kiều',            NOW() - interval '52 days', NOW() - interval '52 days', NOW() - interval '52 days'),
('0903000034', 'seller', 'Quách Thị Mai',       'https://ui-avatars.com/api/?name=Quach+Mai&background=B388FF&color=333&size=200',       '46 Nguyễn Huệ',                'Long An',           'Gạo tấm Long An giá tốt',                                    'Gạo Tấm Tân An',           NOW() - interval '49 days', NOW() - interval '49 days', NOW() - interval '49 days'),
('0903000035', 'seller', 'Âu Dương Nam',        'https://ui-avatars.com/api/?name=Au+Nam&background=82B1FF&color=333&size=200',          '58 Phan Chu Trinh',             'Hậu Giang',         'Đại lý gạo Hậu Giang',                                       'Đại Lý Gạo HG',            NOW() - interval '46 days', NOW() - interval '46 days', NOW() - interval '46 days'),
('0903000036', 'seller', 'Thái Thị Oanh',       'https://ui-avatars.com/api/?name=Thai+Oanh&background=80CBC4&color=333&size=200',       '70 Bạch Đằng',                  'Vĩnh Long',         'Gạo ngon Vĩnh Long giá rẻ',                                  'Gạo Vĩnh Long 365',        NOW() - interval '43 days', NOW() - interval '43 days', NOW() - interval '43 days'),
('0903000037', 'seller', 'La Quang Phúc',       'https://ui-avatars.com/api/?name=La+Phuc&background=A5D6A7&color=333&size=200',         '82 Trương Định',                'Tiền Giang',        'Nông sản Tiền Giang chất lượng',                              'NS Tiền Giang',             NOW() - interval '40 days', NOW() - interval '40 days', NOW() - interval '40 days'),
('0903000038', 'seller', 'Mạc Thị Quyên',      'https://ui-avatars.com/api/?name=Mac+Quyen&background=C5E1A5&color=333&size=200',       '94 Đồng Khởi',                  'Bạc Liêu',          'Gạo Bạc Liêu tươi mới',                                      'Gạo Tươi BL',              NOW() - interval '37 days', NOW() - interval '37 days', NOW() - interval '37 days'),
('0903000039', 'seller', 'Trịnh Văn Sang',      'https://ui-avatars.com/api/?name=Trinh+Sang&background=E6EE9C&color=333&size=200',      '106 Lê Thánh Tôn',             'Sóc Trăng',         'Gạo ST Sóc Trăng chính hãng',                                'ST Sóc Trăng Store',        NOW() - interval '33 days', NOW() - interval '33 days', NOW() - interval '33 days'),
('0903000040', 'seller', 'Khổng Thị Tâm',       'https://ui-avatars.com/api/?name=Khong+Tam&background=FFF59D&color=333&size=200',       '118 Nguyễn Bỉnh Khiêm',         'Trà Vinh',          'Gạo sạch Trà Vinh giá hợp lý',                               'Gạo Sạch TV',              NOW() - interval '29 days', NOW() - interval '29 days', NOW() - interval '29 days'),
('0903000041', 'seller', 'Đoàn Minh Tuấn',      'https://ui-avatars.com/api/?name=Doan+Tuan&background=FFE082&color=333&size=200',       '130 Trần Quang Diệu',           'Cà Mau',            'Gạo miền đất mũi',                                           'Gạo Cà Mau Xanh',          NOW() - interval '27 days', NOW() - interval '27 days', NOW() - interval '27 days'),
('0903000042', 'seller', 'Từ Thị Út',           'https://ui-avatars.com/api/?name=Tu+Ut&background=FFCC80&color=333&size=200',           '142 Lý Chính Thắng',            'Bến Tre',           'Gạo nàng thơm Bến Tre',                                      'Gạo Nàng Thơm BT',         NOW() - interval '25 days', NOW() - interval '25 days', NOW() - interval '25 days'),
('0903000043', 'seller', 'Uông Văn Vinh',       'https://ui-avatars.com/api/?name=Uong+Vinh&background=FFAB91&color=333&size=200',       '154 Phạm Viết Chánh',           'TP. Hồ Chí Minh',   'Đại lý gạo Q.Bình Thạnh',                                    'Gạo Bình Thạnh',           NOW() - interval '23 days', NOW() - interval '23 days', NOW() - interval '23 days'),
('0903000044', 'seller', 'Ông Thị Xuyến',       'https://ui-avatars.com/api/?name=Ong+Xuyen&background=BCAAA4&color=333&size=200',       '166 Khánh Hội',                 'Đồng Tháp',         'Gạo Lai Vung chất lượng cao',                                 'Gạo Lai Vung',             NOW() - interval '21 days', NOW() - interval '21 days', NOW() - interval '21 days'),
('0903000045', 'seller', 'Nghiêm Đức Hoàn',     'https://ui-avatars.com/api/?name=Nghiem+Hoan&background=B0BEC5&color=333&size=200',     '178 An Dương Vương',            'An Giang',          'Gạo Long Xuyên uy tín',                                      'Gạo Long Xuyên',           NOW() - interval '19 days', NOW() - interval '19 days', NOW() - interval '19 days'),
('0903000046', 'seller', 'Vương Thị Yến',       'https://ui-avatars.com/api/?name=Vuong+Yen&background=EF9A9A&color=333&size=200',       '190 Nguyễn Công Trứ',           'Kiên Giang',        'Gạo Hà Tiên thơm ngon',                                      'Gạo Hà Tiên',              NOW() - interval '17 days', NOW() - interval '17 days', NOW() - interval '17 days'),
('0903000047', 'seller', 'Tăng Quốc Anh',       'https://ui-avatars.com/api/?name=Tang+Anh&background=F48FB1&color=333&size=200',        '202 Lê Đại Hành',              'Cần Thơ',            'Gạo Ô Môn Cần Thơ',                                          'Gạo Ô Môn',                NOW() - interval '15 days', NOW() - interval '15 days', NOW() - interval '15 days'),
('0903000048', 'seller', 'Giáp Thị Bảo',        'https://ui-avatars.com/api/?name=Giap+Bao&background=CE93D8&color=333&size=200',        '214 Trần Khánh Dư',             'Long An',           'Gạo Đức Huệ Long An',                                        'Gạo Đức Huệ',              NOW() - interval '13 days', NOW() - interval '13 days', NOW() - interval '13 days'),
('0903000049', 'seller', 'Phí Văn Cảnh',        'https://ui-avatars.com/api/?name=Phi+Canh&background=B39DDB&color=333&size=200',        '226 Phan Bội Châu',             'Hậu Giang',         'Gạo Vị Thanh Hậu Giang',                                     'Gạo Vị Thanh',             NOW() - interval '11 days', NOW() - interval '11 days', NOW() - interval '11 days'),
('0903000050', 'seller', 'Diệp Thị Diễm',      'https://ui-avatars.com/api/?name=Diep+Diem&background=9FA8DA&color=333&size=200',       '238 Nguyễn Thiện Thuật',        'Vĩnh Long',         'Gạo ngon mỗi ngày Vĩnh Long',                                'Gạo Mỗi Ngày',             NOW() - interval '9 days',  NOW() - interval '9 days',  NOW() - interval '9 days'),
('0903000051', 'seller', 'Hà Công Em',          'https://ui-avatars.com/api/?name=Ha+Em&background=90CAF9&color=333&size=200',           '250 Lê Hồng Phong',            'Tiền Giang',        'Gạo Gò Công Tiền Giang',                                     'Gạo Gò Công',              NOW() - interval '7 days',  NOW() - interval '7 days',  NOW() - interval '7 days'),
('0903000052', 'seller', 'Khuất Thị Gấm',      'https://ui-avatars.com/api/?name=Khuat+Gam&background=81D4FA&color=333&size=200',       '262 Đinh Công Tráng',           'Bạc Liêu',          'Gạo Giá Rai Bạc Liêu',                                       'Gạo Giá Rai',              NOW() - interval '5 days',  NOW() - interval '5 days',  NOW() - interval '5 days'),
('0903000053', 'seller', 'Đào Văn Hiếu',        'https://ui-avatars.com/api/?name=Dao+Hieu&background=80DEEA&color=333&size=200',        '274 Trương Công Định',          'Sóc Trăng',         'Gạo Mỹ Xuyên Sóc Trăng',                                     'Gạo Mỹ Xuyên',             NOW() - interval '3 days',  NOW() - interval '3 days',  NOW() - interval '3 days'),
('0903000054', 'seller', 'Lương Thị Khuê',      'https://ui-avatars.com/api/?name=Luong+Khue&background=80CBC4&color=333&size=200',      '286 Nguyễn Trung Trực',         'Trà Vinh',          'Gạo Cầu Ngang Trà Vinh',                                     'Gạo Cầu Ngang',            NOW() - interval '2 days',  NOW() - interval '2 days',  NOW() - interval '2 days'),
('0903000055', 'seller', 'Nông Đức Lợi',        'https://ui-avatars.com/api/?name=Nong+Loi&background=A5D6A7&color=333&size=200',        '298 Phan Văn Hân',              'Cà Mau',            'Gạo Năm Căn Cà Mau',                                         'Gạo Năm Căn',              NOW() - interval '1 day',   NOW() - interval '1 day',   NOW() - interval '1 day'),
('0903000056', 'seller', 'Ôn Thị Mộng',         'https://ui-avatars.com/api/?name=On+Mong&background=C8E6C9&color=333&size=200',         '310 Lê Quý Đôn',               'Bến Tre',           'Gạo Mỏ Cày Bến Tre',                                         'Gạo Mỏ Cày',               NOW() - interval '58 days', NOW() - interval '58 days', NOW() - interval '58 days'),
('0903000057', 'seller', 'Quản Văn Nhật',       'https://ui-avatars.com/api/?name=Quan+Nhat&background=DCEDC8&color=333&size=200',       '322 Trần Bình Trọng',           'TP. Hồ Chí Minh',   'Gạo Q.Tân Phú giá sỉ',                                       'Gạo Tân Phú',              NOW() - interval '53 days', NOW() - interval '53 days', NOW() - interval '53 days'),
('0903000058', 'seller', 'Triệu Thị Phượng',    'https://ui-avatars.com/api/?name=Trieu+Phuong&background=F0F4C3&color=333&size=200',    '334 Nguyễn Đình Chiểu',        'Đồng Tháp',         'Gạo Tháp Mười ngon nổi tiếng',                                'Gạo Tháp Mười',            NOW() - interval '47 days', NOW() - interval '47 days', NOW() - interval '47 days'),
('0903000059', 'seller', 'Vy Quốc Sơn',         'https://ui-avatars.com/api/?name=Vy+Son&background=FFF9C4&color=333&size=200',          '346 Lý Thái Tổ',               'An Giang',           'Gạo Thoại Sơn An Giang',                                     'Gạo Thoại Sơn',            NOW() - interval '44 days', NOW() - interval '44 days', NOW() - interval '44 days'),
('0903000060', 'seller', 'Xã Thị Thảo',         'https://ui-avatars.com/api/?name=Xa+Thao&background=FFECB3&color=333&size=200',         '358 Trần Cao Vân',              'Kiên Giang',        'Gạo Châu Thành Kiên Giang',                                  'Gạo Châu Thành KG',        NOW() - interval '41 days', NOW() - interval '41 days', NOW() - interval '41 days')
ON CONFLICT (phone) DO NOTHING;

-- ========================
-- BUYERS (40 accounts) — phone 0904xxxxxx
-- ========================
INSERT INTO users (phone, role, name, avatar_url, address, province, accepted_tos_at, created_at, updated_at) VALUES
('0904000001', 'buyer', 'Nguyễn Thị Hoa',     'https://ui-avatars.com/api/?name=Nguyen+Hoa&background=E91E63&color=fff&size=200',    '15 Lý Thái Tổ, Q.10',          'TP. Hồ Chí Minh', NOW() - interval '45 days', NOW() - interval '45 days', NOW() - interval '45 days'),
('0904000002', 'buyer', 'Trần Văn Bình',      'https://ui-avatars.com/api/?name=Tran+Binh&background=3F51B5&color=fff&size=200',     '27 Trần Phú',                   'Hà Nội',           NOW() - interval '42 days', NOW() - interval '42 days', NOW() - interval '42 days'),
('0904000003', 'buyer', 'Lê Thị Cúc',         'https://ui-avatars.com/api/?name=Le+Cuc&background=4CAF50&color=fff&size=200',        '39 Nguyễn Văn Linh',            'Đà Nẵng',          NOW() - interval '40 days', NOW() - interval '40 days', NOW() - interval '40 days'),
('0904000004', 'buyer', 'Phạm Quốc Dũng',     'https://ui-avatars.com/api/?name=Pham+Dung&background=FF9800&color=fff&size=200',     '51 Lê Lợi',                     'Cần Thơ',          NOW() - interval '38 days', NOW() - interval '38 days', NOW() - interval '38 days'),
('0904000005', 'buyer', 'Hoàng Thị Én',       'https://ui-avatars.com/api/?name=Hoang+En&background=9C27B0&color=fff&size=200',      '63 Hai Bà Trưng',               'Hải Phòng',        NOW() - interval '36 days', NOW() - interval '36 days', NOW() - interval '36 days'),
('0904000006', 'buyer', 'Võ Đức Phong',        'https://ui-avatars.com/api/?name=Vo+Phong&background=00BCD4&color=fff&size=200',      '75 Phan Đình Phùng',            'Huế',              NOW() - interval '34 days', NOW() - interval '34 days', NOW() - interval '34 days'),
('0904000007', 'buyer', 'Đặng Thị Giang',     'https://ui-avatars.com/api/?name=Dang+Giang&background=795548&color=fff&size=200',    '87 Nguyễn Trãi',                'Bình Dương',       NOW() - interval '32 days', NOW() - interval '32 days', NOW() - interval '32 days'),
('0904000008', 'buyer', 'Bùi Văn Hà',         'https://ui-avatars.com/api/?name=Bui+Ha&background=607D8B&color=fff&size=200',        '99 Võ Văn Kiệt',                'Đồng Nai',         NOW() - interval '30 days', NOW() - interval '30 days', NOW() - interval '30 days'),
('0904000009', 'buyer', 'Ngô Thị Lan Anh',    'https://ui-avatars.com/api/?name=Ngo+Lan&background=F44336&color=fff&size=200',       '111 Điện Biên Phủ',             'Nha Trang',        NOW() - interval '28 days', NOW() - interval '28 days', NOW() - interval '28 days'),
('0904000010', 'buyer', 'Lý Minh Khôi',       'https://ui-avatars.com/api/?name=Ly+Khoi&background=2196F3&color=fff&size=200',       '123 Trần Hưng Đạo',            'Vũng Tàu',         NOW() - interval '26 days', NOW() - interval '26 days', NOW() - interval '26 days'),
('0904000011', 'buyer', 'Trương Thị Linh',    'https://ui-avatars.com/api/?name=Truong+Linh&background=8BC34A&color=fff&size=200',   '135 Nguyễn Du',                 'Long Xuyên',       NOW() - interval '24 days', NOW() - interval '24 days', NOW() - interval '24 days'),
('0904000012', 'buyer', 'Mai Văn Minh',        'https://ui-avatars.com/api/?name=Mai+Minh&background=FFC107&color=333&size=200',      '147 Lê Duẩn',                   'Quy Nhơn',         NOW() - interval '22 days', NOW() - interval '22 days', NOW() - interval '22 days'),
('0904000013', 'buyer', 'Đinh Thị Ngà',       'https://ui-avatars.com/api/?name=Dinh+Nga&background=673AB7&color=fff&size=200',      '159 Phan Chu Trinh',            'Buôn Ma Thuột',    NOW() - interval '20 days', NOW() - interval '20 days', NOW() - interval '20 days'),
('0904000014', 'buyer', 'Phan Quốc Oai',      'https://ui-avatars.com/api/?name=Phan+Oai&background=009688&color=fff&size=200',      '171 Bạch Đằng',                 'Đà Lạt',           NOW() - interval '18 days', NOW() - interval '18 days', NOW() - interval '18 days'),
('0904000015', 'buyer', 'Huỳnh Thị Phúc',     'https://ui-avatars.com/api/?name=Huynh+Phuc&background=FF5722&color=fff&size=200',    '183 Trường Chinh',              'Thanh Hóa',        NOW() - interval '16 days', NOW() - interval '16 days', NOW() - interval '16 days'),
('0904000016', 'buyer', 'Lê Văn Quang',       'https://ui-avatars.com/api/?name=Le+Quang&background=E040FB&color=fff&size=200',      '195 Hùng Vương',                'Vinh',             NOW() - interval '14 days', NOW() - interval '14 days', NOW() - interval '14 days'),
('0904000017', 'buyer', 'Trần Thị Rạng',      'https://ui-avatars.com/api/?name=Tran+Rang&background=76FF03&color=333&size=200',     '207 Nguyễn Văn Cừ',            'Thái Nguyên',      NOW() - interval '12 days', NOW() - interval '12 days', NOW() - interval '12 days'),
('0904000018', 'buyer', 'Nguyễn Đức Sơn',     'https://ui-avatars.com/api/?name=Nguyen+Son&background=40C4FF&color=333&size=200',    '219 Lê Hồng Phong',            'Nam Định',         NOW() - interval '10 days', NOW() - interval '10 days', NOW() - interval '10 days'),
('0904000019', 'buyer', 'Phạm Thị Trang',     'https://ui-avatars.com/api/?name=Pham+Trang&background=FF6E40&color=fff&size=200',    '231 Ngô Quyền',                'Hải Dương',        NOW() - interval '8 days',  NOW() - interval '8 days',  NOW() - interval '8 days'),
('0904000020', 'buyer', 'Hoàng Văn Uy',       'https://ui-avatars.com/api/?name=Hoang+Uy&background=EA80FC&color=333&size=200',      '243 Trần Quang Diệu',          'Ninh Bình',        NOW() - interval '6 days',  NOW() - interval '6 days',  NOW() - interval '6 days'),
('0904000021', 'buyer', 'Võ Thị Vân',         'https://ui-avatars.com/api/?name=Vo+Van&background=0D8ABC&color=fff&size=200',        '255 Lý Tự Trọng',              'TP. Hồ Chí Minh',  NOW() - interval '44 days', NOW() - interval '44 days', NOW() - interval '44 days'),
('0904000022', 'buyer', 'Đặng Quốc Xuân',     'https://ui-avatars.com/api/?name=Dang+Xuan&background=4CAF50&color=fff&size=200',     '267 Pasteur',                   'Hà Nội',           NOW() - interval '41 days', NOW() - interval '41 days', NOW() - interval '41 days'),
('0904000023', 'buyer', 'Bùi Thị Yến',        'https://ui-avatars.com/api/?name=Bui+Yen&background=FF9800&color=fff&size=200',       '279 Cách Mạng Tháng 8',        'Đà Nẵng',          NOW() - interval '39 days', NOW() - interval '39 days', NOW() - interval '39 days'),
('0904000024', 'buyer', 'Ngô Văn An',         'https://ui-avatars.com/api/?name=Ngo+An&background=9C27B0&color=fff&size=200',        '291 Nam Kỳ Khởi Nghĩa',        'Cần Thơ',          NOW() - interval '37 days', NOW() - interval '37 days', NOW() - interval '37 days'),
('0904000025', 'buyer', 'Lý Thị Bé',          'https://ui-avatars.com/api/?name=Ly+Be&background=00BCD4&color=fff&size=200',         '303 Lê Lai',                    'Bình Dương',       NOW() - interval '35 days', NOW() - interval '35 days', NOW() - interval '35 days'),
('0904000026', 'buyer', 'Trương Văn Công',     'https://ui-avatars.com/api/?name=Truong+Cong&background=795548&color=fff&size=200',   '315 Đinh Tiên Hoàng',           'Đồng Nai',         NOW() - interval '33 days', NOW() - interval '33 days', NOW() - interval '33 days'),
('0904000027', 'buyer', 'Mai Thị Diệu',       'https://ui-avatars.com/api/?name=Mai+Dieu&background=607D8B&color=fff&size=200',      '327 Trần Phú',                  'Hải Phòng',        NOW() - interval '31 days', NOW() - interval '31 days', NOW() - interval '31 days'),
('0904000028', 'buyer', 'Đinh Hoàng Gia',     'https://ui-avatars.com/api/?name=Dinh+Gia&background=F44336&color=fff&size=200',      '339 Nguyễn Thái Học',           'Huế',              NOW() - interval '29 days', NOW() - interval '29 days', NOW() - interval '29 days'),
('0904000029', 'buyer', 'Phan Thị Hiền',      'https://ui-avatars.com/api/?name=Phan+Hien&background=3F51B5&color=fff&size=200',     '351 Võ Thị Sáu',               'Nha Trang',        NOW() - interval '27 days', NOW() - interval '27 days', NOW() - interval '27 days'),
('0904000030', 'buyer', 'Huỳnh Văn Kiên',     'https://ui-avatars.com/api/?name=Huynh+Kien&background=8BC34A&color=fff&size=200',    '363 Nguyễn Bỉnh Khiêm',        'Vũng Tàu',         NOW() - interval '25 days', NOW() - interval '25 days', NOW() - interval '25 days'),
('0904000031', 'buyer', 'Lê Thị Liễu',        'https://ui-avatars.com/api/?name=Le+Lieu&background=FFC107&color=333&size=200',       '375 Trần Khánh Dư',            'Long Xuyên',       NOW() - interval '23 days', NOW() - interval '23 days', NOW() - interval '23 days'),
('0904000032', 'buyer', 'Trần Minh Nhân',     'https://ui-avatars.com/api/?name=Tran+Nhan&background=673AB7&color=fff&size=200',     '387 Phan Bội Châu',             'Quy Nhơn',         NOW() - interval '21 days', NOW() - interval '21 days', NOW() - interval '21 days'),
('0904000033', 'buyer', 'Nguyễn Thị Oanh',    'https://ui-avatars.com/api/?name=Nguyen+Oanh&background=009688&color=fff&size=200',   '399 Lê Đại Hành',              'Buôn Ma Thuột',    NOW() - interval '19 days', NOW() - interval '19 days', NOW() - interval '19 days'),
('0904000034', 'buyer', 'Phạm Đình Phát',     'https://ui-avatars.com/api/?name=Pham+Phat&background=FF5722&color=fff&size=200',     '411 Trần Bình Trọng',           'Đà Lạt',           NOW() - interval '17 days', NOW() - interval '17 days', NOW() - interval '17 days'),
('0904000035', 'buyer', 'Hoàng Thị Quỳnh',    'https://ui-avatars.com/api/?name=Hoang+Quynh&background=E040FB&color=fff&size=200',   '423 Nguyễn Công Trứ',           'Thanh Hóa',        NOW() - interval '15 days', NOW() - interval '15 days', NOW() - interval '15 days'),
('0904000036', 'buyer', 'Võ Thanh Sang',       'https://ui-avatars.com/api/?name=Vo+Sang&background=76FF03&color=333&size=200',       '435 Lý Chính Thắng',            'Vinh',             NOW() - interval '13 days', NOW() - interval '13 days', NOW() - interval '13 days'),
('0904000037', 'buyer', 'Đặng Thị Thi',       'https://ui-avatars.com/api/?name=Dang+Thi&background=40C4FF&color=333&size=200',      '447 Phạm Văn Đồng',            'Thái Nguyên',      NOW() - interval '11 days', NOW() - interval '11 days', NOW() - interval '11 days'),
('0904000038', 'buyer', 'Bùi Đức Thắng',      'https://ui-avatars.com/api/?name=Bui+Thang&background=FF6E40&color=fff&size=200',     '459 Đinh Công Tráng',           'Nam Định',         NOW() - interval '9 days',  NOW() - interval '9 days',  NOW() - interval '9 days'),
('0904000039', 'buyer', 'Ngô Thị Uyển',       'https://ui-avatars.com/api/?name=Ngo+Uyen&background=EA80FC&color=333&size=200',      '471 Nguyễn Đình Chiểu',         'Hải Dương',        NOW() - interval '7 days',  NOW() - interval '7 days',  NOW() - interval '7 days'),
('0904000040', 'buyer', 'Lý Hoàng Vũ',        'https://ui-avatars.com/api/?name=Ly+Vu&background=B388FF&color=333&size=200',         '483 Trần Cao Vân',              'Ninh Bình',        NOW() - interval '5 days',  NOW() - interval '5 days',  NOW() - interval '5 days')
ON CONFLICT (phone) DO NOTHING;

-- ========================
-- SUBSCRIPTIONS — 30 sellers (0903000001 → 0903000030) get active 30-day paid subscription
-- ========================
INSERT INTO subscriptions (user_id, plan, started_at, expires_at, status)
SELECT id, 'paid', NOW(), NOW() + interval '30 days', 'active'
FROM users
WHERE phone IN (
  '0903000001','0903000002','0903000003','0903000004','0903000005',
  '0903000006','0903000007','0903000008','0903000009','0903000010',
  '0903000011','0903000012','0903000013','0903000014','0903000015',
  '0903000016','0903000017','0903000018','0903000019','0903000020',
  '0903000021','0903000022','0903000023','0903000024','0903000025',
  '0903000026','0903000027','0903000028','0903000029','0903000030'
)
AND NOT EXISTS (
  SELECT 1 FROM subscriptions s WHERE s.user_id = users.id AND s.status = 'active' AND s.expires_at > NOW()
);

-- ========================
-- LISTINGS — Each subscribed seller gets 2 listings
-- ========================
DO $$
DECLARE
  seller RECORD;
  rice_types text[] := ARRAY['Gạo ST25', 'Gạo Jasmine', 'Gạo Nàng Hoa', 'Gạo Tài Nguyên', 'Gạo Lứt Đỏ', 'Gạo Nếp Cái Hoa Vàng', 'Gạo IR 50404', 'Gạo Hữu Cơ', 'Gạo Thơm Lài', 'Gạo Nàng Thơm Chợ Đào', 'Gạo Tấm Thơm', 'Gạo Đài Loan', 'Gạo Nhật', 'Gạo Nếp Than', 'Gạo Lứt Gạo Mầm'];
  titles text[] := ARRAY[
    'Gạo ST25 chính hãng Sóc Trăng — hạt dài thơm dẻo',
    'Gạo Jasmine thượng hạng — thơm tự nhiên',
    'Gạo Nàng Hoa 9 hạt dài mềm dẻo',
    'Gạo Tài Nguyên Chợ Đào đặc sản Long An',
    'Gạo Lứt Đỏ hữu cơ — tốt cho sức khỏe',
    'Gạo Nếp Cái Hoa Vàng — dẻo thơm làm xôi',
    'Gạo IR 50404 xuất khẩu tiêu chuẩn 5%',
    'Gạo Hữu Cơ Đồng Tháp — không thuốc trừ sâu',
    'Gạo Thơm Lài An Giang thượng hạng',
    'Gạo Nàng Thơm Chợ Đào — đặc sản miền Tây',
    'Gạo Tấm Thơm giá sỉ — ngon cơm bình dân',
    'Gạo Đài Loan hạt tròn dẻo mềm',
    'Gạo Nhật Koshihikari trồng tại Việt Nam',
    'Gạo Nếp Than — làm bánh, nấu chè',
    'Gạo Lứt Gạo Mầm — dinh dưỡng cao'
  ];
  descs text[] := ARRAY[
    'Gạo ST25 đạt giải gạo ngon nhất thế giới. Hạt dài, cơm dẻo thơm mùi lá dứa tự nhiên. Sản phẩm được kiểm định chất lượng, đóng gói cẩn thận.',
    'Gạo Jasmine nhập từ vùng nguyên liệu An Giang. Hạt gạo trắng trong, cơm mềm thơm. Phù hợp cho bữa ăn gia đình và nhà hàng.',
    'Gạo Nàng Hoa 9 là giống gạo thơm cao cấp miền Tây. Cơm mềm dẻo, vị ngọt tự nhiên. Đóng gói hút chân không giữ tươi lâu.',
    'Đặc sản Long An nổi tiếng. Gạo Tài Nguyên hạt dài trắng, cơm dẻo vừa, thơm nhẹ. Giá cạnh tranh, giao hàng toàn quốc.',
    'Gạo lứt đỏ hữu cơ giàu chất xơ, vitamin B, khoáng chất. Thích hợp cho người ăn kiêng, tiểu đường. Canh tác không hóa chất.',
    'Nếp cái hoa vàng Bắc Bộ chính gốc. Hạt tròn mẩy, nấu xôi dẻo thơm đặc trưng. Thích hợp làm bánh chưng, xôi các loại.',
    'Gạo IR 50404 tiêu chuẩn xuất khẩu. Hạt dài, cơm khô tơi. Giá sỉ cạnh tranh, đóng bao 50kg. Phù hợp cho cơm bình dân.',
    'Gạo hữu cơ được chứng nhận VietGAP. Trồng trên đất phù sa Đồng Tháp, không thuốc trừ sâu, không phân hóa học. An toàn 100%.',
    'Gạo Thơm Lài An Giang hạt dài trắng đẹp. Cơm mềm, thơm thoang thoảng mùi hoa lài. Đóng gói 5kg, 10kg, 25kg.',
    'Gạo đặc sản nổi tiếng Chợ Đào Long An. Hạt nhỏ dài, cơm dẻo ngọt, thơm mùi lá dứa. Sản phẩm OCOP 4 sao.',
    'Gạo tấm thơm giá tốt nhất thị trường. Hạt gạo gãy đều, cơm mềm thơm. Phù hợp cho quán ăn, bếp ăn công nghiệp.',
    'Giống Đài Loan hạt tròn ngắn, cơm dẻo mịn. Thích hợp nấu cơm nắm, sushi. Trồng tại Lâm Đồng, chất lượng cao.',
    'Gạo Nhật Koshihikari trồng tại Đà Lạt. Hạt ngắn tròn, cơm dẻo mềm thơm. Premium quality, đóng gói 2kg và 5kg.',
    'Nếp than (gạo nếp cẩm) hạt tím đen đặc trưng. Giàu anthocyanin, tốt cho sức khỏe. Làm bánh, nấu chè, xôi gấc.',
    'Gạo lứt nảy mầm giàu GABA, vitamin E. Tốt cho hệ tiêu hóa, tim mạch. Sản phẩm dinh dưỡng cao cấp, đóng gói 1kg.'
  ];
  quantities numeric[] := ARRAY[500, 1000, 2000, 5000, 300, 800, 10000, 1500, 3000, 700, 20000, 400, 200, 600, 250];
  prices numeric[] := ARRAY[28000, 18000, 22000, 15000, 35000, 25000, 12000, 32000, 20000, 24000, 14000, 30000, 45000, 28000, 40000];
  i int := 0;
  j int;
BEGIN
  FOR seller IN
    SELECT id, province FROM users
    WHERE phone LIKE '090300000%' AND phone <= '0903000030'
    ORDER BY phone
  LOOP
    FOR j IN 0..1 LOOP
      INSERT INTO listings (user_id, title, rice_type, province, quantity_kg, price_per_kg, description, status, created_at, updated_at)
      VALUES (
        seller.id,
        titles[(i % 15) + 1],
        rice_types[(i % 15) + 1],
        seller.province,
        quantities[(i % 15) + 1],
        prices[(i % 15) + 1],
        descs[(i % 15) + 1],
        'active',
        NOW() - ((30 - i) || ' days')::interval,
        NOW() - ((30 - i) || ' days')::interval
      );
      i := i + 1;
    END LOOP;
  END LOOP;
END $$;

COMMIT;

-- Summary
DO $$
DECLARE
  v_users int;
  v_sellers int;
  v_buyers int;
  v_subs int;
  v_listings int;
BEGIN
  SELECT COUNT(*) INTO v_users FROM users WHERE phone LIKE '0903%' OR phone LIKE '0904%';
  SELECT COUNT(*) INTO v_sellers FROM users WHERE phone LIKE '0903%' AND role = 'seller';
  SELECT COUNT(*) INTO v_buyers FROM users WHERE phone LIKE '0904%' AND role = 'buyer';
  SELECT COUNT(*) INTO v_subs FROM subscriptions WHERE status = 'active' AND expires_at > NOW();
  SELECT COUNT(*) INTO v_listings FROM listings WHERE status = 'active';
  RAISE NOTICE '========== Seed Summary ==========';
  RAISE NOTICE 'Seeded users: %', v_users;
  RAISE NOTICE 'Sellers: %', v_sellers;
  RAISE NOTICE 'Buyers: %', v_buyers;
  RAISE NOTICE 'Active subscriptions (total): %', v_subs;
  RAISE NOTICE 'Active listings (total): %', v_listings;
  RAISE NOTICE '==================================';
END $$;
