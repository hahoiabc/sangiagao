package model

import "time"

type Listing struct {
	ID             string    `json:"id"`
	UserID         string    `json:"user_id"`
	Title          string    `json:"title"`
	Category       *string   `json:"category,omitempty"`
	RiceType       string    `json:"rice_type"`
	Province       *string   `json:"province,omitempty"`
	Ward           *string   `json:"ward,omitempty"`
	QuantityKG     float64   `json:"quantity_kg"`
	PricePerKG     float64   `json:"price_per_kg"`
	HarvestSeason  *string   `json:"harvest_season,omitempty"`
	Description    *string   `json:"description,omitempty"`
	Certifications *string   `json:"certifications,omitempty"`
	Images         []string  `json:"images"`
	Status         string    `json:"status"`
	ViewCount      int       `json:"view_count"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type ListingDetail struct {
	Listing
	Seller *PublicProfile `json:"seller"`
}

type CreateListingRequest struct {
	Title          string  `json:"title"`
	Category       string  `json:"category" binding:"required"`
	RiceType       string  `json:"rice_type" binding:"required"`
	Province       *string `json:"province"`
	Ward           *string `json:"ward"`
	QuantityKG     float64 `json:"quantity_kg" binding:"required,gt=0"`
	PricePerKG     float64 `json:"price_per_kg" binding:"required,gt=0"`
	HarvestSeason  *string `json:"harvest_season"`
	Description    *string `json:"description"`
	Certifications *string `json:"certifications"`
}

type UpdateListingRequest struct {
	Title          *string  `json:"title"`
	Category       *string  `json:"category"`
	RiceType       *string  `json:"rice_type"`
	Province       *string  `json:"province"`
	Ward           *string  `json:"ward"`
	QuantityKG     *float64 `json:"quantity_kg"`
	PricePerKG     *float64 `json:"price_per_kg"`
	HarvestSeason  *string  `json:"harvest_season"`
	Description    *string  `json:"description"`
	Certifications *string  `json:"certifications"`
}

type ListingFilter struct {
	Query    string
	Category string
	RiceType string
	Province string
	Ward     string
	MinPrice *float64
	MaxPrice *float64
	MinQty   *float64
	Sort     string
	Page     int
	Limit    int
}

type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Total      int         `json:"total"`
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
	TotalPages int         `json:"total_pages"`
}
