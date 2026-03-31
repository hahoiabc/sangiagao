package firebase

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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
	// SendDataOnly sends a data-only FCM message (no notification field).
	// Used for incoming calls so onBackgroundMessage always fires on Android.
	SendDataOnly(ctx context.Context, tokens []string, data map[string]string) error
}

// FCMSender implements PushSender using Firebase Cloud Messaging HTTP v1 API.
type FCMSender struct {
	projectID   string
	tokenSource oauth2.TokenSource
	httpClient  *http.Client

	// OnInvalidToken is called when FCM returns 404 (token expired/unregistered).
	// Set this to auto-delete stale tokens from the database.
	OnInvalidToken func(deviceToken string)
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
	Notification *fcmNotification  `json:"notification,omitempty"`
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

// sendFCM is the shared send logic for both SendToTokens and SendDataOnly.
func (s *FCMSender) sendFCM(ctx context.Context, tokens []string, msg func(deviceToken string) fcmV1Request, label string) error {
	token, err := s.tokenSource.Token()
	if err != nil {
		return fmt.Errorf("FCM: failed to get access token: %w", err)
	}

	url := fmt.Sprintf("https://fcm.googleapis.com/v1/projects/%s/messages:send", s.projectID)
	var failCount int
	var successCount int

	for _, deviceToken := range tokens {
		fcmReq := msg(deviceToken)

		payload, err := json.Marshal(fcmReq)
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
			log.Printf("FCM %s error for token %s...: %v", label, deviceToken[:min(10, len(deviceToken))], err)
			failCount++
			continue
		}

		if resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			successCount++
			continue
		}

		// Read error response body for debugging
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		resp.Body.Close()

		log.Printf("FCM %s failed for token %s...: status %d, body: %s",
			label, deviceToken[:min(10, len(deviceToken))], resp.StatusCode, string(body))

		// 404 = token expired/unregistered — auto-cleanup
		if resp.StatusCode == http.StatusNotFound && s.OnInvalidToken != nil {
			log.Printf("FCM: removing invalid token %s...", deviceToken[:min(10, len(deviceToken))])
			s.OnInvalidToken(deviceToken)
		}

		failCount++
	}

	log.Printf("FCM %s: %d/%d succeeded", label, successCount, len(tokens))

	if failCount > 0 && failCount == len(tokens) {
		return fmt.Errorf("FCM: all %d %s notifications failed", failCount, label)
	}
	return nil
}

func (s *FCMSender) SendToTokens(ctx context.Context, tokens []string, title, body, imageURL string, data map[string]string) error {
	return s.sendFCM(ctx, tokens, func(deviceToken string) fcmV1Request {
		return fcmV1Request{
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
	}, "notification")
}

// SendDataOnly sends a data-only high-priority FCM message for incoming calls.
// MUST NOT have a "notification" field — otherwise Android OS intercepts it and
// shows a text notification instead of firing onBackgroundMessage/native handler.
//
// Delivery chain on Android:
// 1. Native: CallFirebaseService.onMessageReceived → shows CallKit directly (Kotlin)
// 2. Dart backup: firebaseMessagingBackgroundHandler → shows CallKit (if native missed)
func (s *FCMSender) SendDataOnly(ctx context.Context, tokens []string, data map[string]string) error {
	return s.sendFCM(ctx, tokens, func(deviceToken string) fcmV1Request {
		return fcmV1Request{
			Message: fcmV1Message{
				Token: deviceToken,
				// NO Notification field — data-only ensures native handler fires
				Data: data,
				Android: &fcmAndroid{
					Priority: "high",
				},
				APNS: &fcmAPNS{
					Headers: map[string]string{
						"apns-priority":  "10",
						"apns-push-type": "voip",
					},
					Payload: &fcmAPNSPayload{
						APS: &fcmAPS{
							ContentAvailable: 1,
							Sound:            "default",
						},
					},
				},
			},
		}
	}, "incoming-call")
}

// MockPushSender logs push notifications instead of sending them.
type MockPushSender struct{}

func (m *MockPushSender) SendToTokens(ctx context.Context, tokens []string, title, body, imageURL string, data map[string]string) error {
	if len(tokens) > 0 {
		fmt.Printf("[PUSH MOCK] To %d device(s): %s — %s\n", len(tokens), title, body)
	}
	return nil
}

func (m *MockPushSender) SendDataOnly(ctx context.Context, tokens []string, data map[string]string) error {
	if len(tokens) > 0 {
		fmt.Printf("[PUSH MOCK] Data-only to %d device(s): %v\n", len(tokens), data)
	}
	return nil
}
