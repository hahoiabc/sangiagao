package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sangiagao/rice-marketplace/internal/repository"
)

// SEOHandler exposes public read-only endpoints powering static SEO pages
// (sangiagao.vn/bang-gia-gao/...). No auth required.
type SEOHandler struct {
	listingRepo *repository.ListingRepo
	pool        *pgxpool.Pool
}

func NewSEOHandler(pool *pgxpool.Pool, listingRepo *repository.ListingRepo) *SEOHandler {
	return &SEOHandler{listingRepo: listingRepo, pool: pool}
}

type seoPriceEntry struct {
	Province      string    `json:"province"`
	ProvinceSlug  string    `json:"province_slug"`
	Category      string    `json:"category"`
	CategoryLabel string    `json:"category_label"`
	RiceType      string    `json:"rice_type"`
	RiceTypeLabel string    `json:"rice_type_label"`
	RiceTypeSlug  string    `json:"rice_type_slug"`
	MinPrice      float64   `json:"min_price"`
	AvgPrice      float64   `json:"avg_price"`
	MaxPrice      float64   `json:"max_price"`
	ListingCount  int       `json:"listing_count"`
	LastUpdated   time.Time `json:"last_updated"`
}

// GetPriceBoard returns aggregated price data grouped by (province, category,
// rice_type) for SEO landing pages.
// Endpoint: GET /api/v1/seo/price-board
func (h *SEOHandler) GetPriceBoard(c *gin.Context) {
	rows, err := h.listingRepo.GetSEOPriceBoard(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch SEO price board"})
		return
	}

	out := make([]seoPriceEntry, 0, len(rows))
	for _, r := range rows {
		out = append(out, seoPriceEntry{
			Province:      r.Province,
			ProvinceSlug:  slugify(r.Province),
			Category:      r.Category,
			CategoryLabel: r.CategoryLabel,
			RiceType:      r.RiceType,
			RiceTypeLabel: r.RiceTypeLabel,
			RiceTypeSlug:  slugify(r.RiceType),
			MinPrice:      r.MinPrice,
			AvgPrice:      r.AvgPrice,
			MaxPrice:      r.MaxPrice,
			ListingCount:  r.ListingCount,
			LastUpdated:   r.LastUpdated,
		})
	}

	c.Header("Cache-Control", "public, max-age=600, s-maxage=600")
	c.JSON(http.StatusOK, gin.H{
		"data":         out,
		"total":        len(out),
		"generated_at": time.Now().UTC(),
	})
}

// GetListingsByProvinceAndRiceType returns active listings for SEO detail page.
// Endpoint: GET /api/v1/seo/listings?province={slug}&rice_type={slug}&limit=20
func (h *SEOHandler) GetListingsByProvinceAndRiceType(c *gin.Context) {
	provinceSlug := c.Query("province")
	riceTypeSlug := c.Query("rice_type")
	limit := 20

	if provinceSlug == "" || riceTypeSlug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "province and rice_type required"})
		return
	}

	rows, err := h.pool.Query(c.Request.Context(),
		`SELECT
		    l.id,
		    l.title,
		    l.price_per_kg,
		    l.quantity_kg,
		    l.province,
		    l.ward,
		    l.created_at,
		    COALESCE(u.name, '') AS seller_name
		 FROM listings l
		 LEFT JOIN users u ON u.id = l.user_id
		 WHERE l.status = 'active'
		   AND lower(unaccent(TRIM(regexp_replace(l.province, '^(Tỉnh|Thành phố|TP\.?|TP|Tỉnh\.?)\s+', '', 'i')))) = lower(unaccent($1))
		   AND lower(unaccent(l.rice_type)) = lower(unaccent($2))
		 ORDER BY l.created_at DESC
		 LIMIT $3`,
		unslugify(provinceSlug), unslugify(riceTypeSlug), limit,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed: " + err.Error()})
		return
	}
	defer rows.Close()

	type listingRow struct {
		ID         string    `json:"id"`
		Title      string    `json:"title"`
		PricePerKg float64   `json:"price_per_kg"`
		QuantityKg float64   `json:"quantity_kg"`
		Province   *string   `json:"province"`
		Ward       *string   `json:"ward"`
		CreatedAt  time.Time `json:"created_at"`
		SellerName string    `json:"seller_name"`
	}

	out := make([]listingRow, 0, limit)
	for rows.Next() {
		var l listingRow
		if err := rows.Scan(
			&l.ID, &l.Title, &l.PricePerKg, &l.QuantityKg,
			&l.Province, &l.Ward, &l.CreatedAt, &l.SellerName,
		); err != nil {
			continue
		}
		out = append(out, l)
	}

	c.Header("Cache-Control", "public, max-age=600, s-maxage=600")
	c.JSON(http.StatusOK, gin.H{"data": out, "total": len(out)})
}

// slugify converts Vietnamese place/rice name to URL-safe slug.
// Example: "Đồng Nai" → "dong-nai", "ST 25" → "st-25", "Nàng Hoa 9" → "nang-hoa-9"
func slugify(s string) string {
	s = stripDiacritics(s)
	out := make([]byte, 0, len(s))
	prevDash := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c >= 'A' && c <= 'Z':
			out = append(out, c+32)
			prevDash = false
		case (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9'):
			out = append(out, c)
			prevDash = false
		default:
			if !prevDash && len(out) > 0 {
				out = append(out, '-')
				prevDash = true
			}
		}
	}
	// trim trailing dash
	for len(out) > 0 && out[len(out)-1] == '-' {
		out = out[:len(out)-1]
	}
	return string(out)
}

// unslugify is best-effort reverse of slugify, returning name with dashes
// replaced by spaces. Used in SQL LIKE-style queries (with unaccent).
func unslugify(slug string) string {
	out := make([]byte, len(slug))
	for i := 0; i < len(slug); i++ {
		if slug[i] == '-' {
			out[i] = ' '
		} else {
			out[i] = slug[i]
		}
	}
	return string(out)
}

// stripDiacritics removes Vietnamese tone marks.
// Maps: à→a, ố→o, ư→u, đ→d, etc. Basic mapping, enough for slug generation.
func stripDiacritics(s string) string {
	replacer := map[rune]rune{
		'à': 'a', 'á': 'a', 'ả': 'a', 'ã': 'a', 'ạ': 'a',
		'ă': 'a', 'ằ': 'a', 'ắ': 'a', 'ẳ': 'a', 'ẵ': 'a', 'ặ': 'a',
		'â': 'a', 'ầ': 'a', 'ấ': 'a', 'ẩ': 'a', 'ẫ': 'a', 'ậ': 'a',
		'è': 'e', 'é': 'e', 'ẻ': 'e', 'ẽ': 'e', 'ẹ': 'e',
		'ê': 'e', 'ề': 'e', 'ế': 'e', 'ể': 'e', 'ễ': 'e', 'ệ': 'e',
		'ì': 'i', 'í': 'i', 'ỉ': 'i', 'ĩ': 'i', 'ị': 'i',
		'ò': 'o', 'ó': 'o', 'ỏ': 'o', 'õ': 'o', 'ọ': 'o',
		'ô': 'o', 'ồ': 'o', 'ố': 'o', 'ổ': 'o', 'ỗ': 'o', 'ộ': 'o',
		'ơ': 'o', 'ờ': 'o', 'ớ': 'o', 'ở': 'o', 'ỡ': 'o', 'ợ': 'o',
		'ù': 'u', 'ú': 'u', 'ủ': 'u', 'ũ': 'u', 'ụ': 'u',
		'ư': 'u', 'ừ': 'u', 'ứ': 'u', 'ử': 'u', 'ữ': 'u', 'ự': 'u',
		'ỳ': 'y', 'ý': 'y', 'ỷ': 'y', 'ỹ': 'y', 'ỵ': 'y',
		'đ': 'd',
		'À': 'A', 'Á': 'A', 'Ả': 'A', 'Ã': 'A', 'Ạ': 'A',
		'Ă': 'A', 'Ằ': 'A', 'Ắ': 'A', 'Ẳ': 'A', 'Ẵ': 'A', 'Ặ': 'A',
		'Â': 'A', 'Ầ': 'A', 'Ấ': 'A', 'Ẩ': 'A', 'Ẫ': 'A', 'Ậ': 'A',
		'È': 'E', 'É': 'E', 'Ẻ': 'E', 'Ẽ': 'E', 'Ẹ': 'E',
		'Ê': 'E', 'Ề': 'E', 'Ế': 'E', 'Ể': 'E', 'Ễ': 'E', 'Ệ': 'E',
		'Ì': 'I', 'Í': 'I', 'Ỉ': 'I', 'Ĩ': 'I', 'Ị': 'I',
		'Ò': 'O', 'Ó': 'O', 'Ỏ': 'O', 'Õ': 'O', 'Ọ': 'O',
		'Ô': 'O', 'Ồ': 'O', 'Ố': 'O', 'Ổ': 'O', 'Ỗ': 'O', 'Ộ': 'O',
		'Ơ': 'O', 'Ờ': 'O', 'Ớ': 'O', 'Ở': 'O', 'Ỡ': 'O', 'Ợ': 'O',
		'Ù': 'U', 'Ú': 'U', 'Ủ': 'U', 'Ũ': 'U', 'Ụ': 'U',
		'Ư': 'U', 'Ừ': 'U', 'Ứ': 'U', 'Ử': 'U', 'Ữ': 'U', 'Ự': 'U',
		'Ỳ': 'Y', 'Ý': 'Y', 'Ỷ': 'Y', 'Ỹ': 'Y', 'Ỵ': 'Y',
		'Đ': 'D',
	}
	out := make([]rune, 0, len(s))
	for _, r := range s {
		if mapped, ok := replacer[r]; ok {
			out = append(out, mapped)
		} else {
			out = append(out, r)
		}
	}
	return string(out)
}
