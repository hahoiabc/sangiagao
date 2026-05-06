package apple

import (
	"errors"
	"fmt"
	"os"
)

// Config holds App Store Connect API credentials.
// Loaded from env vars APP_STORE_KEY_ID, APP_STORE_ISSUER_ID, APP_STORE_KEY_PATH,
// APP_STORE_BUNDLE_ID, APP_STORE_ENV.
type Config struct {
	KeyID      string
	IssuerID   string
	KeyPEM     []byte
	BundleID   string
	Env        string // "Sandbox" or "Production"
}

func LoadConfig() (*Config, error) {
	keyID := os.Getenv("APP_STORE_KEY_ID")
	issuerID := os.Getenv("APP_STORE_ISSUER_ID")
	keyPath := os.Getenv("APP_STORE_KEY_PATH")
	bundleID := os.Getenv("APP_STORE_BUNDLE_ID")
	env := os.Getenv("APP_STORE_ENV")
	if env == "" {
		env = "Sandbox"
	}
	if keyID == "" || issuerID == "" || keyPath == "" || bundleID == "" {
		return nil, errors.New("apple: missing one of APP_STORE_KEY_ID, APP_STORE_ISSUER_ID, APP_STORE_KEY_PATH, APP_STORE_BUNDLE_ID")
	}
	pem, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("apple: read key file %s: %w", keyPath, err)
	}
	return &Config{
		KeyID:    keyID,
		IssuerID: issuerID,
		KeyPEM:   pem,
		BundleID: bundleID,
		Env:      env,
	}, nil
}

func (c *Config) APIBase() string {
	if c.Env == "Production" {
		return "https://api.storekit.itunes.apple.com"
	}
	return "https://api.storekit-sandbox.itunes.apple.com"
}
