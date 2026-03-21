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
	ProductKey  string  `json:"product_key" binding:"required"`
	LogoURL     string  `json:"logo_url" binding:"required"`
	SponsorName *string `json:"sponsor_name"`
}

type UpdateSponsorRequest struct {
	LogoURL     *string `json:"logo_url"`
	SponsorName *string `json:"sponsor_name"`
	IsActive    *bool   `json:"is_active"`
}
