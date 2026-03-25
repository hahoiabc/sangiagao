package model

import "time"

type User struct {
	ID            string     `json:"id"`
	Phone         string     `json:"phone,omitempty"`
	Role          string     `json:"role"`
	Name          *string    `json:"name,omitempty"`
	AvatarURL     *string    `json:"avatar_url,omitempty"`
	Address       *string    `json:"address,omitempty"`
	Province      *string    `json:"province,omitempty"`
	Ward          *string    `json:"ward,omitempty"`
	Description   *string    `json:"description,omitempty"`
	OrgName       *string    `json:"org_name,omitempty"`
	IsBlocked              bool       `json:"is_blocked"`
	BlockReason            *string    `json:"block_reason,omitempty"`
	AcceptedTOSAt          *time.Time `json:"accepted_tos_at,omitempty"`
	SubscriptionExpiresAt  *time.Time `json:"subscription_expires_at,omitempty"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
}

type PublicProfile struct {
	ID          string     `json:"id"`
	Phone       string     `json:"phone"`
	Role        string     `json:"role"`
	Name        *string    `json:"name,omitempty"`
	AvatarURL   *string    `json:"avatar_url,omitempty"`
	Province    *string    `json:"province,omitempty"`
	Ward        *string    `json:"ward,omitempty"`
	Description *string    `json:"description,omitempty"`
	OrgName     *string    `json:"org_name,omitempty"`
	IsOnline    *bool      `json:"is_online,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// MaskPhone returns a masked version of a phone number, e.g. "****799".
func MaskPhone(phone string) string {
	if len(phone) <= 3 {
		return "****"
	}
	return "****" + phone[len(phone)-3:]
}

func (u *User) ToPublicProfile() *PublicProfile {
	return &PublicProfile{
		ID:          u.ID,
		Phone:       u.Phone,
		Role:        u.Role,
		Name:        u.Name,
		AvatarURL:   u.AvatarURL,
		Province:    u.Province,
		Ward:        u.Ward,
		Description: u.Description,
		OrgName:     u.OrgName,
		CreatedAt:   u.CreatedAt,
	}
}

type RegisterRequest struct {
	Phone    string `json:"phone" binding:"required,max=15"`
	Name     string `json:"name" binding:"required,max=100"`
	Password string `json:"password" binding:"required,min=6,max=128"`
	Province string `json:"province" binding:"omitempty,max=100"`
	Ward     string `json:"ward" binding:"omitempty,max=100"`
	Code     string `json:"code" binding:"required,max=10"`
}

type LoginRequest struct {
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UpdateProfileRequest struct {
	Name        *string `json:"name" binding:"omitempty,max=100"`
	Role        *string `json:"role" binding:"omitempty,max=20"`
	Address     *string `json:"address" binding:"omitempty,max=200"`
	Province    *string `json:"province" binding:"omitempty,max=100"`
	Ward        *string `json:"ward" binding:"omitempty,max=100"`
	Description *string `json:"description" binding:"omitempty,max=2000"`
	OrgName     *string `json:"org_name" binding:"omitempty,max=200"`
	AcceptTOS   *bool   `json:"accept_tos"`
}
