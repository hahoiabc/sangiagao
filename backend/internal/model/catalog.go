package model

import "time"

// Kiểu danh mục — quyết định luồng đăng tin + hiển thị bảng giá.
const (
	CategoryKindCommodity = "nong_san" // gạo/nếp/tấm: giá đ/kg, số lượng kg, có mùa vụ/chứng nhận
	CategoryKindItem      = "mat_hang" // máy móc/xe/ghe: giá đ/<unit>, khoảng giá rộng, ẩn kg/mùa vụ/chứng nhận
)

type CatalogCategory struct {
	ID        string    `json:"id"`
	Key       string    `json:"key"`
	Label     string    `json:"label"`
	Icon      string    `json:"icon"`
	Kind      string    `json:"kind"`            // nong_san | mat_hang
	Unit      string    `json:"unit"`            // kg | cái | km | chiếc...
	AggregatePrice bool  `json:"aggregate_price"` // true=hiện min đ/unit; false=đếm tin (Cách B)
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
	Kind  string `json:"kind" binding:"omitempty,oneof=nong_san mat_hang"` // mặc định nong_san
	Unit  string `json:"unit"`                                             // mặc định kg
	AggregatePrice *bool `json:"aggregate_price"`                           // mặc định true
}

type UpdateCategoryRequest struct {
	Label     *string `json:"label"`
	Icon      *string `json:"icon"`
	Kind      *string `json:"kind" binding:"omitempty,oneof=nong_san mat_hang"`
	Unit      *string `json:"unit"`
	AggregatePrice *bool `json:"aggregate_price"`
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
