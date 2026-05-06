package apple

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// generateBearerToken creates a JWT for App Store Server API auth.
// Apple requires ES256 signed token with kid=KeyID, iss=IssuerID, aud="appstoreconnect-v1",
// max lifetime 20 minutes.
func generateBearerToken(cfg *Config) (string, error) {
	priv, err := parseECPrivateKey(cfg.KeyPEM)
	if err != nil {
		return "", err
	}

	now := time.Now()
	claims := jwt.MapClaims{
		"iss":   cfg.IssuerID,
		"iat":   now.Unix(),
		"exp":   now.Add(15 * time.Minute).Unix(),
		"aud":   "appstoreconnect-v1",
		"bid":   cfg.BundleID,
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	tok.Header["kid"] = cfg.KeyID
	tok.Header["typ"] = "JWT"
	return tok.SignedString(priv)
}

func parseECPrivateKey(pemBytes []byte) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, errors.New("apple: invalid PEM block")
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		// Fallback to EC-specific PEM.
		ec, err2 := x509.ParseECPrivateKey(block.Bytes)
		if err2 != nil {
			return nil, fmt.Errorf("apple: parse EC key: %v / %v", err, err2)
		}
		return ec, nil
	}
	priv, ok := key.(*ecdsa.PrivateKey)
	if !ok {
		return nil, errors.New("apple: key is not ECDSA")
	}
	return priv, nil
}
