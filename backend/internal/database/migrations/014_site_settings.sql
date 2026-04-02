-- Site settings (key-value store for admin-configurable settings)
CREATE TABLE IF NOT EXISTS site_settings (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Default slogan
INSERT INTO site_settings (key, value) VALUES ('slogan', 'Kết nối ngành gạo')
ON CONFLICT (key) DO NOTHING;
