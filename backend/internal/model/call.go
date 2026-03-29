package model

import "time"

type CallLog struct {
	ID               string         `json:"id"`
	ConversationID   string         `json:"conversation_id"`
	CallerID         string         `json:"caller_id"`
	CalleeID         string         `json:"callee_id"`
	CallType         string         `json:"call_type"`
	Status           string         `json:"status"`
	DurationSeconds  int            `json:"duration_seconds"`
	StartedAt        *time.Time     `json:"started_at,omitempty"`
	EndedAt          *time.Time     `json:"ended_at,omitempty"`
	CreatedAt        time.Time      `json:"created_at"`
	CallerName       string         `json:"caller_name,omitempty"`
	CalleeName       string         `json:"callee_name,omitempty"`
}

type CreateCallLogRequest struct {
	ConversationID string `json:"conversation_id" binding:"required"`
	CalleeID       string `json:"callee_id" binding:"required"`
	CallType       string `json:"call_type" binding:"required,oneof=audio video"`
}

type UpdateCallStatusRequest struct {
	Status          string `json:"status" binding:"required,oneof=answered rejected busy failed"`
	DurationSeconds int    `json:"duration_seconds"`
}
