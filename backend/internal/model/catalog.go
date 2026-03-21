package model

import "time"

type CatalogCategory struct {
	ID        string    `json:"id"`
	Key       string    `json:"key"`
	Label     string    `json:"label"`
	Icon      string    `json:"icon"`
	SortOrder int       `json:"sort_order"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CatalogProduct struct {
	ID         string    `json:"id"`
	Key        string    `json:"key"`
	Label      string    `json:"label"`
	CategoryID string    `json:"category_id"`
	SortOrder  int       `json:"sort_order"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type CreateCategoryRequest struct {
	Key   string `json:"key" binding:"required"`
	Label string `json:"label" binding:"required"`
	Icon  string `json:"icon"`
}

type UpdateCategoryRequest struct {
	Label     *string `json:"label"`
	Icon      *string `json:"icon"`
	SortOrder *int    `json:"sort_order"`
	IsActive  *bool   `json:"is_active"`
}

type CreateProductRequest struct {
	Key        string `json:"key" binding:"required"`
	Label      string `json:"label" binding:"required"`
	CategoryID string `json:"category_id" binding:"required"`
}

type UpdateProductRequest struct {
	Label      *string `json:"label"`
	CategoryID *string `json:"category_id"`
	SortOrder  *int    `json:"sort_order"`
	IsActive   *bool   `json:"is_active"`
}
