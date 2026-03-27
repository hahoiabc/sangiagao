package sms

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"
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

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("zalo oauth parse: %w", err)
	}

	if tokenResp.AccessToken == "" {
		return "", fmt.Errorf("zalo oauth: empty access token")
	}

	z.accessToken = tokenResp.AccessToken
	if tokenResp.RefreshToken != "" {
		z.refreshToken = tokenResp.RefreshToken
	}
	// Refresh 60s trước khi hết hạn
	z.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn-60) * time.Second)

	log.Printf("[ZALO ZNS] Token refreshed, expires in %ds", tokenResp.ExpiresIn)
	return z.accessToken, nil
}
