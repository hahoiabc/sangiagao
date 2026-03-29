package firebase

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// PushSender sends push notifications via FCM HTTP v1 API.
type PushSender interface {
	SendToTokens(ctx context.Context, tokens []string, title, body, imageURL string, data map[string]string) error
}

// FCMSender implements PushSender using Firebase Cloud Messaging HTTP v1 API.
type FCMSender struct {
	projectID   string
	tokenSource oauth2.TokenSource
	httpClient  *http.Client
}

// NewFCMSender creates a new FCM sender. Pass empty credPath for mock mode.
// credPath should point to a Firebase service account JSON file.
func NewFCMSender(credPath string) PushSender {
	if credPath == "" {
		return &MockPushSender{}
	}

	jsonKey, err := os.ReadFile(credPath)
	if err != nil {
		log.Printf("Firebase: failed to read credentials at %s: %v — using mock", credPath, err)
		return &MockPushSender{}
	}

	var sa struct {
		ProjectID string `json:"project_id"`
	}
	if err := json.Unmarshal(jsonKey, &sa); err != nil || sa.ProjectID == "" {
		log.Printf("Firebase: failed to parse project_id from credentials — using mock")
		return &MockPushSender{}
	}

	creds, err := google.CredentialsFromJSON(context.Background(), jsonKey,
		"https://www.googleapis.com/auth/firebase.messaging",
	)
	if err != nil {
		log.Printf("Firebase: failed to create credentials: %v — using mock", err)
		return &MockPushSender{}
	}

	log.Printf("Firebase: FCM v1 sender initialized for project %s", sa.ProjectID)
	return &FCMSender{
		projectID:   sa.ProjectID,
		tokenSource: creds.TokenSource,
		httpClient:  &http.Client{Timeout: 10 * time.Second},
	}
}

type fcmV1Request struct {
	Message fcmV1Message `json:"message"`
}

type fcmV1Message struct {
	Token        string            `json:"token"`
	Notification *fcmNotification  `json:"notification"`
	Data         map[string]string `json:"data,omitempty"`
	Android      *fcmAndroid       `json:"android,omitempty"`
	APNS         *fcmAPNS          `json:"apns,omitempty"`
}

type fcmNotification struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	Image string `json:"image,omitempty"`
}

type fcmAndroid struct {
	Priority     string                 `json:"priority,omitempty"`
	Notification *fcmAndroidNotification `json:"notification,omitempty"`
}

type fcmAndroidNotification struct {
	ChannelID    string `json:"channel_id,omitempty"`
	Sound        string `json:"sound,omitempty"`
	DefaultSound bool   `json:"default_sound,omitempty"`
}

type fcmAPNS struct {
	Headers map[string]string `json:"headers,omitempty"`
	Payload *fcmAPNSPayload   `json:"payload,omitempty"`
}

type fcmAPNSPayload struct {
	APS *fcmAPS `json:"aps,omitempty"`
}

type fcmAPS struct {
	Sound            string `json:"sound,omitempty"`
	ContentAvailable int    `json:"content-available,omitempty"`
}

func (s *FCMSender) SendToTokens(ctx context.Context, tokens []string, title, body, imageURL string, data map[string]string) error {
	token, err := s.tokenSource.Token()
	if err != nil {
		return fmt.Errorf("FCM: failed to get access token: %w", err)
	}

	url := fmt.Sprintf("https://fcm.googleapis.com/v1/projects/%s/messages:send", s.projectID)
	var failCount int

	for _, deviceToken := range tokens {
		msg := fcmV1Request{
			Message: fcmV1Message{
				Token: deviceToken,
				Notification: &fcmNotification{
					Title: title,
					Body:  body,
					Image: imageURL,
				},
				Data: data,
				Android: &fcmAndroid{
					Priority: "high",
					Notification: &fcmAndroidNotification{
						ChannelID:    "sangiagao_notifications",
						Sound:        "default",
						DefaultSound: true,
					},
				},
				APNS: &fcmAPNS{
					Headers: map[string]string{
						"apns-priority":  "10",
						"apns-push-type": "alert",
					},
					Payload: &fcmAPNSPayload{
						APS: &fcmAPS{
							Sound:            "default",
							ContentAvailable: 1,
						},
					},
				},
			},
		}

		payload, err := json.Marshal(msg)
		if err != nil {
			failCount++
			continue
		}

		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(payload))
		if err != nil {
			failCount++
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token.AccessToken)

		resp, err := s.httpClient.Do(req)
		if err != nil {
			log.Printf("FCM send error for token %s...: %v", deviceToken[:min(10, len(deviceToken))], err)
			failCount++
			continue
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Printf("FCM send failed for token %s...: status %d", deviceToken[:min(10, len(deviceToken))], resp.StatusCode)
			failCount++
		}
	}

	if failCount > 0 && failCount == len(tokens) {
		return fmt.Errorf("FCM: all %d notifications failed to send", failCount)
	}
	if failCount > 0 {
		log.Printf("FCM: %d/%d notifications failed", failCount, len(tokens))
	}
	return nil
}

// MockPushSender logs push notifications instead of sending them.
type MockPushSender struct{}

func (m *MockPushSender) SendToTokens(ctx context.Context, tokens []string, title, body, imageURL string, data map[string]string) error {
	if len(tokens) > 0 {
		fmt.Printf("[PUSH MOCK] To %d device(s): %s — %s\n", len(tokens), title, body)
	}
	return nil
}
