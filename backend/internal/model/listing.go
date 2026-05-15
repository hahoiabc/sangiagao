package model

import "time"

type Listing struct {
	ID             string     `json:"id"`
	UserID         string     `json:"user_id"`
	Title          string     `json:"title"`
	Category       *string    `json:"category,omitempty"`
	RiceType       string     `json:"rice_type"`
	Province       *string    `json:"province,omitempty"`
	Ward           *string    `json:"ward,omitempty"`
	QuantityKG     float64    `json:"quantity_kg"`
	PricePerKG     float64    `json:"price_per_kg"`
	HarvestSeason  *string    `json:"harvest_season,omitempty"`
	Description    *string    `json:"description,omitempty"`
	Certifications *string    `json:"certifications,omitempty"`
	Images         []string   `json:"images"`
	Status         string     `json:"status"`
	ViewCount      int        `json:"view_count"`
	BumpedAt       *time.Time `json:"bumped_at,omitempty"`
	BumpCount      int        `json:"bump_count"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// BumpLifetimeCap — số lần "Làm mới tin đăng" tối đa cho 1 tin (60 ngày × 4/ngày).
const BumpLifetimeCap = 240

// BumpCooldown — khoảng thời gian phải đợi giữa 2 lần "Làm mới tin đăng".
// 5,9h = 354 phút → cho phép tối đa ~4 lần/ngày.
const BumpCooldownMinutes = 354

type ListingDetail struct {
	Listing
	Seller *PublicProfile `json:"seller"`
}

type CreateListingRequest struct {
	Title          string  `json:"title" binding:"omitempty,max=200"`
	Category       string  `json:"category" binding:"required,max=50"`
	RiceType       string  `json:"rice_type" binding:"required,max=50"`
	Province       *string `json:"province" binding:"omitempty,max=100"`
	Ward           *string `json:"ward" binding:"omitempty,max=100"`
	QuantityKG     float64 `json:"quantity_kg" binding:"required,gt=0"`
	PricePerKG     float64 `json:"price_per_kg" binding:"required,gt=0"`
	HarvestSeason  *string `json:"harvest_season" binding:"omitempty,max=100"`
	Description    *string `json:"description" binding:"omitempty,max=2000"`
	Certifications *string `json:"certifications" binding:"omitempty,max=500"`
}

type UpdateListingRequest struct {
	Title          *string  `json:"title" binding:"omitempty,max=200"`
	Category       *string  `json:"category" binding:"omitempty,max=50"`
	RiceType       *string  `json:"rice_type" binding:"omitempty,max=50"`
	Province       *string  `json:"province" binding:"omitempty,max=100"`
	Ward           *string  `json:"ward" binding:"omitempty,max=100"`
	QuantityKG     *float64 `json:"quantity_kg"`
	PricePerKG     *float64 `json:"price_per_kg"`
	HarvestSeason  *string  `json:"harvest_season" binding:"omitempty,max=100"`
	Description    *string  `json:"description" binding:"omitempty,max=2000"`
	Certifications *string  `json:"certifications" binding:"omitempty,max=500"`
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
	// HasPhoto: nếu true → chỉ trả tin có ít nhất 1 ảnh.
	HasPhoto bool
	// PostedWithinDays: 0 = không filter; 1/7/30 = chỉ tin được tạo/cập nhật/bump trong N ngày.
	PostedWithinDays int
	Sort             string
	Page             int
	Limit            int
}

type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Total      int         `json:"total"`
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
	TotalPages int         `json:"total_pages"`
}
