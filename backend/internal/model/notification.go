package model

import (
	"encoding/json"
	"time"
)

type Notification struct {
	ID        string          `json:"id"`
	UserID    string          `json:"user_id"`
	Type      string          `json:"type"`
	Title     string          `json:"title"`
	Body      string          `json:"body"`
	Data      json.RawMessage `json:"data,omitempty"`
	IsRead    bool            `json:"is_read"`
	CreatedAt time.Time       `json:"created_at"`
}

type RegisterDeviceRequest struct {
	Token    string `json:"token" binding:"required,min=32,max=512"`
	Platform string `json:"platform" binding:"required,oneof=ios android"`
}
