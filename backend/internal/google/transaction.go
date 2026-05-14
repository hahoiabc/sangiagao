package google

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
)

// SubscriptionPurchase is the relevant subset of Google Play
// SubscriptionPurchaseV2 response. Field reference:
// https://developers.google.com/android-publisher/api-ref/rest/v3/purchases.subscriptions
type SubscriptionPurchase struct {
	Kind                 string `json:"kind"`
	StartTimeMillis      string `json:"startTimeMillis"`
	ExpiryTimeMillis     string `json:"expiryTimeMillis"`
	AutoRenewing         bool   `json:"autoRenewing"`
	PriceCurrencyCode    string `json:"priceCurrencyCode"`
	PriceAmountMicros    string `json:"priceAmountMicros"`
	CountryCode          string `json:"countryCode"`
	OrderID              string `json:"orderId"`
	PurchaseType         *int   `json:"purchaseType,omitempty"` // 0=Test, 1=Promo, 2=Rewarded
	AcknowledgementState int    `json:"acknowledgementState"`   // 0=Yet, 1=Acknowledged
	CancelReason         *int   `json:"cancelReason,omitempty"`
	PaymentState         *int   `json:"paymentState,omitempty"` // 0=Pending, 1=Received, 2=Free trial, 3=Pending deferred upgrade
	UserCancellationTimeMillis string `json:"userCancellationTimeMillis,omitempty"`
}

// ExpiresTime returns the expiry as time.Time.
func (s *SubscriptionPurchase) ExpiresTime() time.Time {
	return parseMillis(s.ExpiryTimeMillis)
}

// StartTime returns the start as time.Time.
func (s *SubscriptionPurchase) StartTime() time.Time {
	return parseMillis(s.StartTimeMillis)
}

func parseMillis(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	var ms int64
	_, _ = fmt.Sscanf(s, "%d", &ms)
	return time.UnixMilli(ms).UTC()
}

// Client wraps Google Play Developer API calls.
type Client struct {
	cfg  *Config
	http *http.Client
}

// NewClient builds an OAuth2-authenticated HTTP client using the service
// account JSON.
func NewClient(cfg *Config) (*Client, error) {
	saConfig, err := google.JWTConfigFromJSON(cfg.ServiceAccountJSON,
		"https://www.googleapis.com/auth/androidpublisher")
	if err != nil {
		return nil, fmt.Errorf("google_iap: parse service account: %w", err)
	}
	httpClient := saConfig.Client(context.Background())
	httpClient.Timeout = 15 * time.Second
	_ = (*jwt.Config)(nil) // silence import if oauth2/jwt unused
	_ = (*oauth2.Token)(nil)
	return &Client{cfg: cfg, http: httpClient}, nil
}

// GetSubscriptionPurchase fetches the current state of a subscription purchase
// by purchase token. Used both on /verify (mobile-initiated) and on RTDN
// webhook (server-initiated).
func (c *Client) GetSubscriptionPurchase(ctx context.Context, subscriptionID, purchaseToken string) (*SubscriptionPurchase, error) {
	url := fmt.Sprintf(
		"https://androidpublisher.googleapis.com/androidpublisher/v3/applications/%s/purchases/subscriptions/%s/tokens/%s",
		c.cfg.PackageName, subscriptionID, purchaseToken,
	)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	res, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("google_iap: fetch: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		var apiErr struct {
			Error struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
			} `json:"error"`
		}
		_ = json.NewDecoder(res.Body).Decode(&apiErr)
		return nil, fmt.Errorf("google_iap: %d %s", apiErr.Error.Code, apiErr.Error.Message)
	}
	var p SubscriptionPurchase
	if err := json.NewDecoder(res.Body).Decode(&p); err != nil {
		return nil, fmt.Errorf("google_iap: decode: %w", err)
	}
	return &p, nil
}

// AcknowledgePurchase marks the purchase as acknowledged. Google requires
// acknowledgment within 3 days or the purchase is auto-refunded.
func (c *Client) AcknowledgePurchase(ctx context.Context, subscriptionID, purchaseToken string) error {
	url := fmt.Sprintf(
		"https://androidpublisher.googleapis.com/androidpublisher/v3/applications/%s/purchases/subscriptions/%s/tokens/%s:acknowledge",
		c.cfg.PackageName, subscriptionID, purchaseToken,
	)
	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return err
	}
	res, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		return errors.New("google_iap: acknowledge failed: " + res.Status)
	}
	return nil
}
