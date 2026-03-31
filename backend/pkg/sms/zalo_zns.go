package sms

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/sangiagao/rice-marketplace/pkg/cache"
)

const (
	zaloRefreshTokenKey = "zalo:refresh_token"
	zaloRefreshTokenTTL = 90 * 24 * time.Hour // 90 days
)

// ZaloZNSSender sends OTP via Zalo Notification Service (ZNS).
type ZaloZNSSender struct {
	appID        string
	appSecret    string
	templateID   string
	accessToken  string
	refreshToken string
	tokenExpiry  time.Time
	mu           sync.Mutex
	client       *http.Client
	cache        cache.Cache
}

type znsRequest struct {
	Phone        string            `json:"phone"`
	TemplateID   string            `json:"template_id"`
	TemplateData map[string]string `json:"template_data"`
}

type znsResponse struct {
	Error   int    `json:"error"`
	Message string `json:"message"`
}

func NewZaloZNSSender(appID, appSecret, templateID, refreshToken string) *ZaloZNSSender {
	return &ZaloZNSSender{
		appID:        appID,
		appSecret:    appSecret,
		templateID:   templateID,
		refreshToken: refreshToken,
		client:       &http.Client{Timeout: 10 * time.Second},
	}
}

// SetCache enables Redis persistence for refresh tokens.
// On startup, loads the latest refresh token from Redis (if available).
func (z *ZaloZNSSender) SetCache(c cache.Cache) {
	z.cache = c
	// Load persisted refresh token (newer than env file)
	if c != nil {
		if raw, err := c.Get(context.Background(), zaloRefreshTokenKey); err == nil && len(raw) > 0 {
			z.refreshToken = string(raw)
			log.Printf("[ZALO ZNS] Loaded refresh token from Redis")
		}
	}
}

// Status returns current ZNS configuration status (secrets masked).
func (z *ZaloZNSSender) Status() map[string]interface{} {
	z.mu.Lock()
	defer z.mu.Unlock()

	mask := func(s string) string {
		if len(s) <= 8 {
			return "***"
		}
		return s[:4] + "..." + s[len(s)-4:]
	}

	tokenStatus := "no_token"
	if z.accessToken != "" {
		if time.Now().Before(z.tokenExpiry) {
			tokenStatus = "valid"
		} else {
			tokenStatus = "expired"
		}
	}

	refreshSource := "env"
	if z.cache != nil {
		if raw, err := z.cache.Get(context.Background(), zaloRefreshTokenKey); err == nil && len(raw) > 0 {
			refreshSource = "redis"
		}
	}

	return map[string]interface{}{
		"app_id":          mask(z.appID),
		"app_secret":      mask(z.appSecret),
		"template_id":     z.templateID,
		"refresh_token":   mask(z.refreshToken),
		"access_token":    tokenStatus,
		"token_expiry":    z.tokenExpiry.Format(time.RFC3339),
		"refresh_source":  refreshSource,
		"redis_connected": z.cache != nil,
	}
}

// UpdateRefreshToken sets a new refresh token and persists to Redis.
func (z *ZaloZNSSender) UpdateRefreshToken(token string) error {
	z.mu.Lock()
	defer z.mu.Unlock()

	z.refreshToken = token
	// Invalidate current access token to force re-auth with new refresh token
	z.accessToken = ""
	z.tokenExpiry = time.Time{}

	if z.cache != nil {
		if err := z.cache.Set(context.Background(), zaloRefreshTokenKey, []byte(token), zaloRefreshTokenTTL); err != nil {
			return fmt.Errorf("persist to redis: %w", err)
		}
	}

	log.Printf("[ZALO ZNS] Refresh token updated manually")
	return nil
}

func (z *ZaloZNSSender) SendOTP(phone, code string) error {
	// Convert 0xxx → 84xxx
	if len(phone) > 0 && phone[0] == '0' {
		phone = "84" + phone[1:]
	}

	token, err := z.getAccessToken()
	if err != nil {
		return fmt.Errorf("zalo token: %w", err)
	}

	body, _ := json.Marshal(znsRequest{
		Phone:      phone,
		TemplateID: z.templateID,
		TemplateData: map[string]string{
			"otp": code,
		},
	})

	req, _ := http.NewRequest("POST",
		"https://business.openapi.zalo.me/message/template",
		bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("access_token", token)

	resp, err := z.client.Do(req)
	if err != nil {
		return fmt.Errorf("zalo request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var znsResp znsResponse
	if err := json.Unmarshal(respBody, &znsResp); err != nil {
		return fmt.Errorf("zalo response parse: %w", err)
	}

	if znsResp.Error != 0 {
		return fmt.Errorf("zalo ZNS error %d: %s", znsResp.Error, znsResp.Message)
	}

	log.Printf("[ZALO ZNS] OTP sent to %s", phone)
	return nil
}

func (z *ZaloZNSSender) getAccessToken() (string, error) {
	z.mu.Lock()
	defer z.mu.Unlock()

	if z.accessToken != "" && time.Now().Before(z.tokenExpiry) {
		return z.accessToken, nil
	}

	data := url.Values{}
	data.Set("refresh_token", z.refreshToken)
	data.Set("app_id", z.appID)
	data.Set("grant_type", "refresh_token")

	req, _ := http.NewRequest("POST",
		"https://oauth.zaloapp.com/v4/oa/access_token",
		bytes.NewBufferString(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("secret_key", z.appSecret)

	resp, err := z.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("zalo oauth: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	log.Printf("[ZALO ZNS] OAuth response (HTTP %d): %s", resp.StatusCode, string(respBody))

	var tokenResp struct {
		AccessToken  string      `json:"access_token"`
		RefreshToken string      `json:"refresh_token"`
		ExpiresIn    json.Number `json:"expires_in"`
		Error        json.Number `json:"error"`
		Message      string      `json:"message"`
	}
	if err := json.Unmarshal(respBody, &tokenResp); err != nil {
		return "", fmt.Errorf("zalo oauth parse: %w", err)
	}

	if errCode, _ := tokenResp.Error.Int64(); errCode != 0 {
		return "", fmt.Errorf("zalo oauth error %d: %s", errCode, tokenResp.Message)
	}

	if tokenResp.AccessToken == "" {
		return "", fmt.Errorf("zalo oauth: empty access token, raw: %s", string(respBody))
	}

	z.accessToken = tokenResp.AccessToken
	if tokenResp.RefreshToken != "" {
		z.refreshToken = tokenResp.RefreshToken
		// Persist new refresh token to Redis so it survives container restarts
		if z.cache != nil {
			if err := z.cache.Set(context.Background(), zaloRefreshTokenKey, []byte(z.refreshToken), zaloRefreshTokenTTL); err != nil {
				log.Printf("[ZALO ZNS] Failed to persist refresh token to Redis: %v", err)
			} else {
				log.Printf("[ZALO ZNS] Refresh token persisted to Redis")
			}
		}
	}
	// Refresh 60s trước khi hết hạn
	expiresIn, _ := tokenResp.ExpiresIn.Int64()
	if expiresIn <= 0 {
		expiresIn = 3600 // default 1h
	}
	z.tokenExpiry = time.Now().Add(time.Duration(expiresIn-60) * time.Second)

	log.Printf("[ZALO ZNS] Token refreshed, expires in %ds", expiresIn)
	return z.accessToken, nil
}
