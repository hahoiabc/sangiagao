package apple

import (
	"crypto/ecdsa"
	"crypto/x509"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

//go:embed certs/AppleRootCA-G3.pem
var appleRootG3PEM []byte

// Notification types per App Store Server Notifications V2.
const (
	NotifSubscribed     = "SUBSCRIBED"
	NotifDidRenew       = "DID_RENEW"
	NotifExpired        = "EXPIRED"
	NotifGracePeriod    = "GRACE_PERIOD_EXPIRED"
	NotifRefund         = "REFUND"
	NotifRevoke         = "REVOKE"
	NotifPriceIncrease  = "PRICE_INCREASE"
	NotifBillingRetry   = "DID_FAIL_TO_RENEW"
	NotifDidChangeStatus = "DID_CHANGE_RENEWAL_STATUS"
	NotifConsumption    = "CONSUMPTION_REQUEST"
	NotifTest           = "TEST"
)

// NotificationPayload is the decoded V2 notification body.
type NotificationPayload struct {
	NotificationType string                 `json:"notificationType"`
	Subtype          string                 `json:"subtype"`
	NotificationUUID string                 `json:"notificationUUID"`
	Version          string                 `json:"version"`
	SignedDate       int64                  `json:"signedDate"`
	Data             NotificationData       `json:"data"`
	Summary          map[string]interface{} `json:"summary,omitempty"`
}

type NotificationData struct {
	BundleID              string `json:"bundleId"`
	Environment           string `json:"environment"`
	SignedTransactionInfo string `json:"signedTransactionInfo"`
	SignedRenewalInfo     string `json:"signedRenewalInfo"`
}

// VerifyAndDecodeNotification accepts the raw signedPayload from Apple webhook,
// verifies the JWS x5c chain against Apple Root CA G3, and returns the decoded
// payload + transaction info. Returns ErrInvalidSignature if any check fails.
func VerifyAndDecodeNotification(signedPayload string) (*NotificationPayload, *TransactionInfo, error) {
	if signedPayload == "" {
		return nil, nil, errors.New("apple: empty signedPayload")
	}

	parts := strings.Split(signedPayload, ".")
	if len(parts) != 3 {
		return nil, nil, errors.New("apple: invalid JWS format")
	}

	// Parse header to get x5c chain.
	headerJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, nil, fmt.Errorf("apple: decode header: %w", err)
	}
	var header struct {
		Alg string   `json:"alg"`
		X5c []string `json:"x5c"`
	}
	if err := json.Unmarshal(headerJSON, &header); err != nil {
		return nil, nil, fmt.Errorf("apple: parse header: %w", err)
	}
	if header.Alg != "ES256" {
		return nil, nil, fmt.Errorf("apple: unexpected alg %s", header.Alg)
	}
	if len(header.X5c) == 0 {
		return nil, nil, errors.New("apple: missing x5c chain")
	}

	// Verify x5c chain.
	leafCert, err := verifyX5cChain(header.X5c)
	if err != nil {
		return nil, nil, fmt.Errorf("apple: verify x5c: %w", err)
	}

	leafPub, ok := leafCert.PublicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, nil, errors.New("apple: leaf cert not ECDSA")
	}

	// Verify JWS signature using leaf cert public key.
	parser := jwt.NewParser(jwt.WithValidMethods([]string{"ES256"}))
	tok, err := parser.Parse(signedPayload, func(t *jwt.Token) (interface{}, error) {
		return leafPub, nil
	})
	if err != nil {
		return nil, nil, fmt.Errorf("apple: verify JWS sig: %w", err)
	}
	if !tok.Valid {
		return nil, nil, errors.New("apple: invalid JWS signature")
	}

	// Decode payload claims.
	claims, ok := tok.Claims.(jwt.MapClaims)
	if !ok {
		return nil, nil, errors.New("apple: invalid claims type")
	}
	raw, _ := json.Marshal(claims)
	var payload NotificationPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, nil, fmt.Errorf("apple: parse payload: %w", err)
	}

	// Decode embedded transaction info if present.
	var tx *TransactionInfo
	if payload.Data.SignedTransactionInfo != "" {
		tx, err = decodeJWSPayload(payload.Data.SignedTransactionInfo)
		if err != nil {
			return nil, nil, fmt.Errorf("apple: decode tx info: %w", err)
		}
	}

	return &payload, tx, nil
}

// verifyX5cChain validates the certificate chain in JWS x5c header against
// Apple Root CA G3. Returns the leaf cert if valid.
func verifyX5cChain(x5c []string) (*x509.Certificate, error) {
	if len(x5c) == 0 {
		return nil, errors.New("empty x5c")
	}

	parsed := make([]*x509.Certificate, 0, len(x5c))
	for i, b64 := range x5c {
		der, err := base64.StdEncoding.DecodeString(b64)
		if err != nil {
			return nil, fmt.Errorf("x5c[%d] base64: %w", i, err)
		}
		c, err := x509.ParseCertificate(der)
		if err != nil {
			return nil, fmt.Errorf("x5c[%d] parse: %w", i, err)
		}
		parsed = append(parsed, c)
	}

	// Apple chain order: [leaf, intermediate, root]
	leaf := parsed[0]

	// Build root pool from embedded Apple Root CA G3.
	roots := x509.NewCertPool()
	block, _ := pem.Decode(appleRootG3PEM)
	if block == nil {
		return nil, errors.New("apple root: invalid PEM")
	}
	rootCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("apple root: parse: %w", err)
	}
	roots.AddCert(rootCert)

	intermediates := x509.NewCertPool()
	for i := 1; i < len(parsed); i++ {
		intermediates.AddCert(parsed[i])
	}

	opts := x509.VerifyOptions{
		Roots:         roots,
		Intermediates: intermediates,
		CurrentTime:   time.Now(),
	}
	if _, err := leaf.Verify(opts); err != nil {
		return nil, fmt.Errorf("verify chain: %w", err)
	}
	return leaf, nil
}
