-- Anti-spam: track auth attempts by IP and device
CREATE TABLE IF NOT EXISTS auth_attempts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    ip_address VARCHAR(45) NOT NULL,
    device_id VARCHAR(128),
    phone VARCHAR(15),
    action VARCHAR(20) NOT NULL,
    success BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_auth_attempts_ip_action ON auth_attempts(ip_address, action, created_at);
CREATE INDEX idx_auth_attempts_device ON auth_attempts(device_id, action, created_at);
CREATE INDEX idx_auth_attempts_cleanup ON auth_attempts(created_at);
