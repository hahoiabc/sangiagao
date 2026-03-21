package model

type PriceBoardEntry struct {
	ProductKey   string   `json:"product_key"`
	ProductLabel string   `json:"product_label"`
	MinPrice     *float64 `json:"min_price"`
	ListingCount int      `json:"listing_count"`
	SponsorLogo  *string  `json:"sponsor_logo,omitempty"`
}

type PriceBoardCategory struct {
	CategoryKey   string            `json:"category_key"`
	CategoryLabel string            `json:"category_label"`
	Products      []PriceBoardEntry `json:"products"`
}

type PriceBoardResponse struct {
	Categories []PriceBoardCategory `json:"categories"`
	UpdatedAt  string               `json:"updated_at"`
}
