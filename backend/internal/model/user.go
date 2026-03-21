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
	Phone    string `json:"phone" binding:"required"`
	Name     string `json:"name" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
	Province string `json:"province"`
	Ward     string `json:"ward"`
	Code     string `json:"code" binding:"required"`
}

type LoginRequest struct {
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UpdateProfileRequest struct {
	Name        *string `json:"name"`
	Role        *string `json:"role"`
	Address     *string `json:"address"`
	Province    *string `json:"province"`
	Ward        *string `json:"ward"`
	Description *string `json:"description"`
	OrgName     *string `json:"org_name"`
	AcceptTOS   *bool   `json:"accept_tos"`
}
