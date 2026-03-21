package model

import "time"

type Feedback struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	UserName  string     `json:"user_name,omitempty"`
	UserPhone string     `json:"user_phone,omitempty"`
	Content   string     `json:"content"`
	Reply     *string    `json:"reply"`
	RepliedAt *time.Time `json:"replied_at"`
	CreatedAt time.Time  `json:"created_at"`
}

type CreateFeedbackRequest struct {
	Content string `json:"content" binding:"required"`
}

type ReplyFeedbackRequest struct {
	Reply string `json:"reply" binding:"required"`
}
