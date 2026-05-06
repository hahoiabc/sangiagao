package apple

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TransactionInfo represents the decoded payload of a JWS-signed transaction
// from Apple's App Store Server API. Fields per Apple JWSTransactionDecodedPayload.
type TransactionInfo struct {
	TransactionID         string `json:"transactionId"`
	OriginalTransactionID string `json:"originalTransactionId"`
	WebOrderLineItemID    string `json:"webOrderLineItemId"`
	BundleID              string `json:"bundleId"`
	ProductID             string `json:"productId"`
	SubscriptionGroupID   string `json:"subscriptionGroupIdentifier"`
	PurchaseDate          int64  `json:"purchaseDate"`
	OriginalPurchaseDate  int64  `json:"originalPurchaseDate"`
	ExpiresDate           int64  `json:"expiresDate"`
	Quantity              int    `json:"quantity"`
	Type                  string `json:"type"`
	InAppOwnershipType    string `json:"inAppOwnershipType"`
	SignedDate            int64  `json:"signedDate"`
	Environment           string `json:"environment"`
	IsUpgraded            bool   `json:"isUpgraded"`
	OfferType             int    `json:"offerType"`
	RevocationDate        int64  `json:"revocationDate,omitempty"`
	RevocationReason      int    `json:"revocationReason,omitempty"`
	AppAccountToken       string `json:"appAccountToken,omitempty"`
}

// PurchaseTime returns the purchase moment.
func (t *TransactionInfo) PurchaseTime() time.Time {
	return time.UnixMilli(t.PurchaseDate)
}

// ExpiresTime returns the subscription expiry moment.
func (t *TransactionInfo) ExpiresTime() time.Time {
	return time.UnixMilli(t.ExpiresDate)
}

// Client wraps App Store Server API calls.
type Client struct {
	cfg  *Config
	http *http.Client
}

func NewClient(cfg *Config) *Client {
	return &Client{
		cfg:  cfg,
		http: &http.Client{Timeout: 15 * time.Second},
	}
}

// GetTransactionInfo calls GET /inApps/v1/transactions/{transactionId}, returns
// the decoded JWS payload. Apple's response is a JWS string; we decode (without
// verifying signature) because Apple already authenticated the request via JWT.
//
// For webhook notifications (Phase 2B) we will verify the JWS signature
// against Apple root certs because the source is unauthenticated.
func (c *Client) GetTransactionInfo(ctx context.Context, transactionID string) (*TransactionInfo, error) {
	if transactionID == "" {
		return nil, errors.New("apple: empty transactionId")
	}

	bearer, err := generateBearerToken(c.cfg)
	if err != nil {
		return nil, fmt.Errorf("apple: gen bearer token: %w", err)
	}

	url := fmt.Sprintf("%s/inApps/v1/transactions/%s", c.cfg.APIBase(), transactionID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+bearer)
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("apple: HTTP request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("apple: read body: %w", err)
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("apple: transaction not found (%s)", transactionID)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("apple: status %d: %s", resp.StatusCode, string(body))
	}

	var wrapper struct {
		SignedTransactionInfo string `json:"signedTransactionInfo"`
	}
	if err := json.Unmarshal(body, &wrapper); err != nil {
		return nil, fmt.Errorf("apple: parse response: %w", err)
	}
	if wrapper.SignedTransactionInfo == "" {
		return nil, errors.New("apple: empty signedTransactionInfo")
	}

	return decodeJWSPayload(wrapper.SignedTransactionInfo)
}

// decodeJWSPayload extracts the payload from a JWS without verifying signature.
// Apple-side trust is established via the bearer token used to fetch the JWS.
func decodeJWSPayload(jws string) (*TransactionInfo, error) {
	parts := strings.Split(jws, ".")
	if len(parts) != 3 {
		return nil, errors.New("apple: invalid JWS format")
	}
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	tok, _, err := parser.ParseUnverified(jws, jwt.MapClaims{})
	if err != nil {
		return nil, fmt.Errorf("apple: parse JWS: %w", err)
	}
	claims, ok := tok.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("apple: invalid JWS claims")
	}
	raw, err := json.Marshal(claims)
	if err != nil {
		return nil, err
	}
	var info TransactionInfo
	if err := json.Unmarshal(raw, &info); err != nil {
		return nil, fmt.Errorf("apple: unmarshal claims: %w", err)
	}
	return &info, nil
}
