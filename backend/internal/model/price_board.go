package model

type PriceBoardEntry struct {
	ProductKey   string   `json:"product_key"`
	ProductLabel string   `json:"product_label"`
	MinPrice     *float64 `json:"min_price"`
	ListingCount int      `json:"listing_count"`
	SponsorLogo  *string  `json:"sponsor_logo,omitempty"`
	ImageURL     *string  `json:"image_url,omitempty"` // ảnh tin RẺ NHẤT (có ảnh) của loại này
}

type PriceBoardCategory struct {
	CategoryKey   string            `json:"category_key"`
	CategoryLabel string            `json:"category_label"`
	Kind          string            `json:"kind"` // nong_san (gộp giá đ/unit) | mat_hang (Cách B: đếm tin)
	Unit          string            `json:"unit"` // đơn vị giá: kg | cái | km...
	Products      []PriceBoardEntry `json:"products"`
}

type PriceBoardResponse struct {
	Categories []PriceBoardCategory `json:"categories"`
	UpdatedAt  string               `json:"updated_at"`
}
