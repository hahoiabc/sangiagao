package model

import "time"

type Conversation struct {
	ID            string     `json:"id"`
	MemberID      string     `json:"member_id"`
	SellerID      string     `json:"seller_id"`
	ListingID     *string    `json:"listing_id,omitempty"`
	LastMessageAt time.Time  `json:"last_message_at"`
	CreatedAt     time.Time  `json:"created_at"`
	// Joined fields
	OtherUser     *PublicProfile `json:"other_user,omitempty"`
	LastMessage   *Message       `json:"last_message,omitempty"`
	UnreadCount   int            `json:"unread_count"`
}

type Message struct {
	ID             string            `json:"id"`
	ConversationID string            `json:"conversation_id"`
	SenderID       string            `json:"sender_id"`
	Content        string            `json:"content"`
	Type           string            `json:"type"`
	ReadAt         *time.Time        `json:"read_at,omitempty"`
	ReplyToID      *string           `json:"reply_to_id,omitempty"`
	ReplyTo        *ReplyMessage     `json:"reply_to,omitempty"`
	Reactions      []MessageReaction `json:"reactions,omitempty"`
	CreatedAt      time.Time         `json:"created_at"`
}

type ReplyMessage struct {
	ID       string `json:"id"`
	SenderID string `json:"sender_id"`
	Content  string `json:"content"`
	Type     string `json:"type"`
}

type MessageReaction struct {
	ID        string    `json:"id"`
	MessageID string    `json:"message_id"`
	UserID    string    `json:"user_id"`
	Emoji     string    `json:"emoji"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateConversationRequest struct {
	SellerID  string  `json:"seller_id" binding:"required"`
	ListingID *string `json:"listing_id"`
}

type SendMessageRequest struct {
	Content   string  `json:"content" binding:"required,max=5000"`
	Type      string  `json:"type" binding:"omitempty,max=50"`
	ReplyToID *string `json:"reply_to_id"`
}

type ToggleReactionRequest struct {
	Emoji string `json:"emoji" binding:"required,max=10"`
}
