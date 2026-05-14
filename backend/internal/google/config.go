package google

import (
	"errors"
	"os"
)

// Config holds Google Play Developer API credentials.
//
// Required env vars (production):
//   GOOGLE_IAP_PACKAGE_NAME   — e.g. "com.sangiagao.rice_marketplace"
//   GOOGLE_IAP_SERVICE_ACCOUNT_JSON — full JSON content of service account key
//   (or GOOGLE_IAP_SERVICE_ACCOUNT_FILE pointing at .json on disk)
//   GOOGLE_IAP_PUBSUB_AUDIENCE — optional, the OIDC audience expected on
//     incoming Pub/Sub push tokens (Google's identity verification).
//     Usually set to the webhook URL itself or a custom string.
type Config struct {
	PackageName        string
	ServiceAccountJSON []byte
	PubSubAudience     string
}

var ErrConfigMissing = errors.New("google_iap: not configured (missing env vars)")

// LoadConfig reads env. Returns ErrConfigMissing if not set so caller can
// gracefully skip wiring Google IAP without crashing.
func LoadConfig() (*Config, error) {
	pkg := os.Getenv("GOOGLE_IAP_PACKAGE_NAME")
	if pkg == "" {
		return nil, ErrConfigMissing
	}

	var saJSON []byte
	if raw := os.Getenv("GOOGLE_IAP_SERVICE_ACCOUNT_JSON"); raw != "" {
		saJSON = []byte(raw)
	} else if path := os.Getenv("GOOGLE_IAP_SERVICE_ACCOUNT_FILE"); path != "" {
		b, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		saJSON = b
	} else {
		return nil, ErrConfigMissing
	}

	return &Config{
		PackageName:        pkg,
		ServiceAccountJSON: saJSON,
		PubSubAudience:     os.Getenv("GOOGLE_IAP_PUBSUB_AUDIENCE"),
	}, nil
}
