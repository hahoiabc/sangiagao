-- Rice categories table
CREATE TABLE IF NOT EXISTS rice_categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    key VARCHAR(50) NOT NULL UNIQUE,
    label VARCHAR(100) NOT NULL,
    icon VARCHAR(50) DEFAULT 'category',
    sort_order INTEGER NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_rice_categories_active ON rice_categories(is_active, sort_order);

-- Rice products table
CREATE TABLE IF NOT EXISTS rice_products (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    key VARCHAR(50) NOT NULL UNIQUE,
    label VARCHAR(100) NOT NULL,
    category_id UUID NOT NULL REFERENCES rice_categories(id) ON DELETE CASCADE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_rice_products_category ON rice_products(category_id);
CREATE INDEX IF NOT EXISTS idx_rice_products_active ON rice_products(is_active, sort_order);

-- Auto update triggers
DROP TRIGGER IF EXISTS rice_categories_updated_at ON rice_categories;
CREATE TRIGGER rice_categories_updated_at BEFORE UPDATE ON rice_categories
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

DROP TRIGGER IF EXISTS rice_products_updated_at ON rice_products;
CREATE TRIGGER rice_products_updated_at BEFORE UPDATE ON rice_products
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- Seed existing catalog data
-- Categories
INSERT INTO rice_categories (key, label, icon, sort_order) VALUES
('gao_deo_thom', 'Gạo dẻo thơm', 'rice_bowl', 1),
('gao_kho', 'Gạo khô', 'grass', 2),
('tam_deo_thom', 'Tấm dẻo thơm', 'grain', 3),
('tam_kho', 'Tấm khô', 'scatter_plot', 4),
('nep', 'Nếp', 'spa', 5)
ON CONFLICT (key) DO NOTHING;

-- Products for gao_deo_thom
INSERT INTO rice_products (key, label, category_id, sort_order)
SELECT p.key, p.label, c.id, p.sort_order FROM (VALUES
    ('st_25', 'ST 25', 1), ('st_24', 'ST 24', 2), ('st_21', 'ST 21', 3),
    ('om_18', 'OM 18', 4), ('om_49', 'OM 49', 5), ('om_5451', 'OM 5451', 6),
    ('dai_thom_8', 'Đài Thơm 8', 7), ('om_6976', 'OM 6976', 8),
    ('nhat', 'Nhật', 9), ('lien_huong', 'Liên Hương', 10),
    ('mien', 'Miên', 11), ('dai_loan', 'Đài Loan', 12)
) AS p(key, label, sort_order)
CROSS JOIN rice_categories c WHERE c.key = 'gao_deo_thom'
ON CONFLICT (key) DO NOTHING;

-- Products for gao_kho
INSERT INTO rice_products (key, label, category_id, sort_order)
SELECT p.key, p.label, c.id, p.sort_order FROM (VALUES
    ('tai_nguyen', 'Tài Nguyên', 1), ('soc', 'Sóc', 2), ('so_ri', 'Sơ Ri', 3),
    ('mong_chim', 'Móng Chim', 4), ('ham_chau_sieu', 'Hàm Châu siêu', 5),
    ('ir_504', 'IR 504', 6), ('q5', 'Q5', 7), ('an_no', 'Ấn nở', 8),
    ('myanmar', 'Myanmar', 9)
) AS p(key, label, sort_order)
CROSS JOIN rice_categories c WHERE c.key = 'gao_kho'
ON CONFLICT (key) DO NOTHING;

-- Products for tam_deo_thom
INSERT INTO rice_products (key, label, category_id, sort_order)
SELECT p.key, p.label, c.id, p.sort_order FROM (VALUES
    ('tam_st_25', 'Tấm ST 25', 1), ('tam_st_24', 'Tấm ST 24', 2), ('tam_st_21', 'Tấm ST 21', 3),
    ('tam_om_18', 'Tấm OM 18', 4), ('tam_om_49', 'Tấm OM 49', 5), ('tam_om_5451', 'Tấm OM 5451', 6),
    ('tam_dai_thom_8', 'Tấm Đài Thơm 8', 7), ('tam_om_6976', 'Tấm OM 6976', 8),
    ('tam_nhat', 'Tấm Nhật', 9), ('tam_lien_huong', 'Tấm Liên Hương', 10),
    ('tam_mien', 'Tấm Miên', 11), ('tam_dai_loan', 'Tấm Đài Loan', 12)
) AS p(key, label, sort_order)
CROSS JOIN rice_categories c WHERE c.key = 'tam_deo_thom'
ON CONFLICT (key) DO NOTHING;

-- Products for tam_kho
INSERT INTO rice_products (key, label, category_id, sort_order)
SELECT p.key, p.label, c.id, p.sort_order FROM (VALUES
    ('tam_tai_nguyen', 'Tấm Tài Nguyên', 1), ('tam_soc', 'Tấm Sóc', 2), ('tam_so_ri', 'Tấm Sơ Ri', 3),
    ('tam_mong_chim', 'Tấm Móng Chim', 4), ('tam_ham_chau_sieu', 'Tấm Hàm Châu siêu', 5),
    ('tam_ir_504', 'Tấm IR 504', 6), ('tam_q5', 'Tấm Q5', 7), ('tam_an_no', 'Tấm Ấn nở', 8),
    ('tam_myanmar', 'Tấm Myanmar', 9)
) AS p(key, label, sort_order)
CROSS JOIN rice_categories c WHERE c.key = 'tam_kho'
ON CONFLICT (key) DO NOTHING;

-- Products for nep
INSERT INTO rice_products (key, label, category_id, sort_order)
SELECT p.key, p.label, c.id, p.sort_order FROM (VALUES
    ('sap_moi', 'Sáp Mới', 1), ('sap_cu', 'Sáp cũ', 2),
    ('nep_la_moi', 'Nếp Lá mới', 3), ('nep_la_cu', 'Nếp Lá cũ', 4),
    ('bac_hat_lon', 'Bắc Hạt Lớn', 5), ('bac_hat_nho', 'Bắc Hạt Nhỏ', 6),
    ('nep_thai', 'Nếp Thái', 7), ('nep_than', 'Nếp Than', 8),
    ('huyet_rong', 'Huyết Rồng', 9)
) AS p(key, label, sort_order)
CROSS JOIN rice_categories c WHERE c.key = 'nep'
ON CONFLICT (key) DO NOTHING;
