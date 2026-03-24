package model

import "time"

type ProductSponsor struct {
	ID          string    `json:"id"`
	ProductKey  string    `json:"product_key"`
	LogoURL     string    `json:"logo_url"`
	SponsorName *string   `json:"sponsor_name,omitempty"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateSponsorRequest struct {
	ProductKey  string  `json:"product_key" binding:"required,max=50"`
	LogoURL     string  `json:"logo_url" binding:"required,url,max=500"`
	SponsorName *string `json:"sponsor_name" binding:"omitempty,max=200"`
}

type UpdateSponsorRequest struct {
	LogoURL     *string `json:"logo_url" binding:"omitempty,url,max=500"`
	SponsorName *string `json:"sponsor_name" binding:"omitempty,max=200"`
	IsActive    *bool   `json:"is_active"`
}
