package firebase

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// PushSender sends push notifications via FCM HTTP v1 API.
type PushSender interface {
	SendToTokens(ctx context.Context, tokens []string, title, body string, data map[string]string) error
}

// FCMSender implements PushSender using Firebase Cloud Messaging Legacy HTTP API.
type FCMSender struct {
	serverKey  string
	httpClient *http.Client
}

// NewFCMSender creates a new FCM sender. Pass empty serverKey for mock mode.
func NewFCMSender(serverKey string) PushSender {
	if serverKey == "" {
		return &MockPushSender{}
	}
	return &FCMSender{
		serverKey: serverKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type fcmMessage struct {
	To           string            `json:"to,omitempty"`
	Notification *fcmNotification  `json:"notification"`
	Data         map[string]string `json:"data,omitempty"`
}

type fcmNotification struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

func (s *FCMSender) SendToTokens(ctx context.Context, tokens []string, title, body string, data map[string]string) error {
	for _, token := range tokens {
		msg := fcmMessage{
			To: token,
			Notification: &fcmNotification{
				Title: title,
				Body:  body,
			},
			Data: data,
		}

		payload, err := json.Marshal(msg)
		if err != nil {
			continue
		}

		req, err := http.NewRequestWithContext(ctx, "POST", "https://fcm.googleapis.com/fcm/send", bytes.NewReader(payload))
		if err != nil {
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "key="+s.serverKey)

		resp, err := s.httpClient.Do(req)
		if err != nil {
			log.Printf("FCM send error for token %s: %v", token[:10], err)
			continue
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Printf("FCM send failed for token %s: status %d", token[:10], resp.StatusCode)
		}
	}
	return nil
}

// MockPushSender logs push notifications instead of sending them.
type MockPushSender struct{}

func (m *MockPushSender) SendToTokens(ctx context.Context, tokens []string, title, body string, data map[string]string) error {
	if len(tokens) > 0 {
		fmt.Printf("[PUSH MOCK] To %d device(s): %s — %s\n", len(tokens), title, body)
	}
	return nil
}
