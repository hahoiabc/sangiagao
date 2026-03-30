package model

import "time"

type InboxMessage struct {
	ID        string     `json:"id"`
	Title     string     `json:"title"`
	Body      string     `json:"body"`
	ImageURL  *string    `json:"image_url,omitempty"`
	Target    string     `json:"target"`
	IsPinned  bool       `json:"is_pinned"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	CreatedBy *string    `json:"created_by,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	IsRead    bool       `json:"is_read"`
}

type CreateInboxRequest struct {
	Title    string  `json:"title" binding:"required,max=200"`
	Body     string  `json:"body" binding:"required"`
	ImageURL *string `json:"image_url,omitempty"`
	Target   string  `json:"target" binding:"omitempty,max=30"`
	IsPinned bool    `json:"is_pinned"`
	// ExpiresAt as string to parse flexibly
	ExpiresAt *string `json:"expires_at,omitempty"`
}

type UpdateInboxRequest struct {
	Title     *string `json:"title,omitempty" binding:"omitempty,max=200"`
	Body      *string `json:"body,omitempty"`
	ImageURL  *string `json:"image_url,omitempty"`
	IsPinned  *bool   `json:"is_pinned,omitempty"`
	ExpiresAt *string `json:"expires_at,omitempty"`
}
