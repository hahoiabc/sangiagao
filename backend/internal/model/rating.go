package model

import "time"

type Rating struct {
	ID         string    `json:"id"`
	ReviewerID string    `json:"reviewer_id"`
	SellerID   string    `json:"seller_id"`
	Stars      int       `json:"stars"`
	Comment    string    `json:"comment"`
	CreatedAt  time.Time `json:"created_at"`
}

type CreateRatingRequest struct {
	SellerID string `json:"seller_id" binding:"required"`
	Stars    int    `json:"stars" binding:"required,min=1,max=5"`
	Comment  string `json:"comment" binding:"omitempty,min=10"`
}

type RatingSummary struct {
	Average float64 `json:"average"`
	Count   int     `json:"count"`
}
