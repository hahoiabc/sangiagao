package google

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

// RTDN notification types (subscriptionNotification.notificationType).
// https://developer.android.com/google/play/billing/rtdn-reference
const (
	NotifSubscriptionRecovered      = 1
	NotifSubscriptionRenewed        = 2
	NotifSubscriptionCanceled       = 3
	NotifSubscriptionPurchased      = 4
	NotifSubscriptionOnHold         = 5
	NotifSubscriptionInGracePeriod  = 6
	NotifSubscriptionRestarted      = 7
	NotifSubscriptionPriceChange    = 8
	NotifSubscriptionDeferred       = 9
	NotifSubscriptionPaused         = 10
	NotifSubscriptionPauseScheduleChange = 11
	NotifSubscriptionRevoked        = 12 // refund
	NotifSubscriptionExpired        = 13
)

// PubSubMessage is the envelope Google Pub/Sub HTTP push uses.
type PubSubMessage struct {
	Message struct {
		Data        string            `json:"data"` // base64-encoded RTDNPayload JSON
		MessageID   string            `json:"messageId"`
		PublishTime string            `json:"publishTime"`
		Attributes  map[string]string `json:"attributes"`
	} `json:"message"`
	Subscription string `json:"subscription"`
}

// RTDNPayload is the JSON inside .message.data (after base64 decode).
type RTDNPayload struct {
	Version                 string `json:"version"`
	PackageName             string `json:"packageName"`
	EventTimeMillis         string `json:"eventTimeMillis"`
	SubscriptionNotification *struct {
		Version          string `json:"version"`
		NotificationType int    `json:"notificationType"`
		PurchaseToken    string `json:"purchaseToken"`
		SubscriptionID   string `json:"subscriptionId"` // product_id e.g. com.sangiagao.premium.1m
	} `json:"subscriptionNotification,omitempty"`
	TestNotification *struct {
		Version string `json:"version"`
	} `json:"testNotification,omitempty"`
}

// DecodePubSubMessage parses the Pub/Sub envelope and returns the decoded
// RTDN payload + raw bytes (for audit logging).
func DecodePubSubMessage(body []byte) (*RTDNPayload, []byte, error) {
	var env PubSubMessage
	if err := json.Unmarshal(body, &env); err != nil {
		return nil, nil, fmt.Errorf("google_iap: parse envelope: %w", err)
	}
	if env.Message.Data == "" {
		return nil, nil, fmt.Errorf("google_iap: empty data field")
	}
	raw, err := base64.StdEncoding.DecodeString(env.Message.Data)
	if err != nil {
		return nil, nil, fmt.Errorf("google_iap: base64 decode: %w", err)
	}
	var p RTDNPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, nil, fmt.Errorf("google_iap: parse payload: %w", err)
	}
	return &p, raw, nil
}
