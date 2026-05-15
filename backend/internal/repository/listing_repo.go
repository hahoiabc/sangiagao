package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sangiagao/rice-marketplace/internal/model"
)

var (
	ErrListingNotFound     = errors.New("listing not found")
	ErrBumpCooldown        = errors.New("bump cooldown not elapsed")
	ErrBumpQuotaExhausted  = errors.New("bump quota exhausted")
)

type ListingRepo struct {
	pool *pgxpool.Pool
}

func NewListingRepo(pool *pgxpool.Pool) *ListingRepo {
	return &ListingRepo{pool: pool}
}

const listingColumns = `id, user_id, title, category, rice_type, province, district,
	quantity_kg, price_per_kg, harvest_season, description, certifications,
	images, status, view_count, bumped_at, bump_count, created_at, updated_at`

// rankingExpr — score 0-100 cho mỗi tin, dùng làm primary ORDER BY trong marketplace
// search/browse. Phân tách 3 thành phần:
//   - completeness 40%: tin có ảnh + description + district + harvest_season + certs + title đủ dài
//   - freshness 40%: GREATEST(created_at, updated_at, bumped_at) — càng gần now càng cao
//   - engagement 20%: view_count chuẩn hóa, cap ở 100 view
//
// Không tham chiếu user/subscription → công bằng giữa free + premium subscriber, chỉ
// phân biệt theo NỘI DUNG tin đăng.
const rankingExpr = `(
	(CASE WHEN images != '[]'::jsonb THEN 25 ELSE 0 END
	 + CASE WHEN LENGTH(COALESCE(description, '')) >= 50 THEN 20 ELSE 0 END
	 + CASE WHEN district IS NOT NULL THEN 15 ELSE 0 END
	 + CASE WHEN harvest_season IS NOT NULL THEN 15 ELSE 0 END
	 + CASE WHEN certifications IS NOT NULL THEN 15 ELSE 0 END
	 + CASE WHEN LENGTH(title) >= 20 THEN 10 ELSE 0 END
	) * 0.4
	+ CASE
		WHEN GREATEST(created_at, updated_at, COALESCE(bumped_at, '1970-01-01'::timestamptz)) > NOW() - INTERVAL '1 day'   THEN 40
		WHEN GREATEST(created_at, updated_at, COALESCE(bumped_at, '1970-01-01'::timestamptz)) > NOW() - INTERVAL '7 days'  THEN 32
		WHEN GREATEST(created_at, updated_at, COALESCE(bumped_at, '1970-01-01'::timestamptz)) > NOW() - INTERVAL '30 days' THEN 20
		WHEN GREATEST(created_at, updated_at, COALESCE(bumped_at, '1970-01-01'::timestamptz)) > NOW() - INTERVAL '60 days' THEN 8
		ELSE 2
	END
	+ LEAST(view_count::float / 100, 1) * 20
)`

func (r *ListingRepo) scanListing(row pgx.Row) (*model.Listing, error) {
	var l model.Listing
	var imagesJSON []byte
	err := row.Scan(
		&l.ID, &l.UserID, &l.Title, &l.Category, &l.RiceType, &l.Province, &l.Ward,
		&l.QuantityKG, &l.PricePerKG, &l.HarvestSeason, &l.Description, &l.Certifications,
		&imagesJSON, &l.Status, &l.ViewCount, &l.BumpedAt, &l.BumpCount, &l.CreatedAt, &l.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrListingNotFound
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(imagesJSON, &l.Images); err != nil || l.Images == nil {
		l.Images = []string{}
	}
	return &l, nil
}

func (r *ListingRepo) scanListings(rows pgx.Rows) ([]*model.Listing, error) {
	var listings []*model.Listing
	for rows.Next() {
		var l model.Listing
		var imagesJSON []byte
		err := rows.Scan(
			&l.ID, &l.UserID, &l.Title, &l.Category, &l.RiceType, &l.Province, &l.Ward,
			&l.QuantityKG, &l.PricePerKG, &l.HarvestSeason, &l.Description, &l.Certifications,
			&imagesJSON, &l.Status, &l.ViewCount, &l.BumpedAt, &l.BumpCount, &l.CreatedAt, &l.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(imagesJSON, &l.Images); err != nil || l.Images == nil {
			l.Images = []string{}
		}
		listings = append(listings, &l)
	}
	if listings == nil {
		listings = []*model.Listing{}
	}
	return listings, rows.Err()
}

func (r *ListingRepo) CountTodayByUser(ctx context.Context, userID string) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM listings WHERE user_id = $1 AND created_at >= CURRENT_DATE AND status != 'deleted'`,
		userID,
	).Scan(&count)
	return count, err
}

func (r *ListingRepo) CountTodayByUserAndType(ctx context.Context, userID, riceType string) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM listings WHERE user_id = $1 AND rice_type = $2 AND created_at >= CURRENT_DATE AND status != 'deleted'`,
		userID, riceType,
	).Scan(&count)
	return count, err
}

func (r *ListingRepo) Create(ctx context.Context, userID string, req *model.CreateListingRequest) (*model.Listing, error) {
	row := r.pool.QueryRow(ctx,
		`INSERT INTO listings (user_id, title, category, rice_type, province, district,
			quantity_kg, price_per_kg, harvest_season, description, certifications)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		 RETURNING `+listingColumns,
		userID, req.Title, req.Category, req.RiceType, req.Province, req.Ward,
		req.QuantityKG, req.PricePerKG, req.HarvestSeason, req.Description, req.Certifications,
	)
	return r.scanListing(row)
}

func (r *ListingRepo) GetByID(ctx context.Context, id string) (*model.Listing, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT `+listingColumns+` FROM listings WHERE id = $1 AND status != 'deleted'`, id,
	)
	return r.scanListing(row)
}

func (r *ListingRepo) Update(ctx context.Context, id string, req *model.UpdateListingRequest) (*model.Listing, error) {
	row := r.pool.QueryRow(ctx,
		`UPDATE listings SET
			title = COALESCE($2, title),
			category = COALESCE($3, category),
			rice_type = COALESCE($4, rice_type),
			province = COALESCE($5, province),
			district = COALESCE($6, district),
			quantity_kg = COALESCE($7, quantity_kg),
			price_per_kg = COALESCE($8, price_per_kg),
			harvest_season = COALESCE($9, harvest_season),
			description = COALESCE($10, description),
			certifications = COALESCE($11, certifications)
		 WHERE id = $1 AND status != 'deleted'
		 RETURNING `+listingColumns,
		id, req.Title, req.Category, req.RiceType, req.Province, req.Ward,
		req.QuantityKG, req.PricePerKG, req.HarvestSeason, req.Description, req.Certifications,
	)
	return r.scanListing(row)
}

func (r *ListingRepo) SoftDelete(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE listings SET status = 'deleted', deleted_at = NOW()
		 WHERE id = $1 AND status != 'deleted'`, id,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrListingNotFound
	}
	return nil
}

func (r *ListingRepo) BatchSoftDelete(ctx context.Context, ids []string) (int, error) {
	tag, err := r.pool.Exec(ctx,
		`UPDATE listings SET status = 'deleted', deleted_at = NOW()
		 WHERE id = ANY($1) AND status != 'deleted'`, ids,
	)
	if err != nil {
		return 0, err
	}
	return int(tag.RowsAffected()), nil
}

func (r *ListingRepo) ListByUser(ctx context.Context, userID string, page, limit int) ([]*model.Listing, int, error) {
	offset := (page - 1) * limit

	var total int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM listings WHERE user_id = $1 AND status != 'deleted'`, userID,
	).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.Query(ctx,
		`SELECT `+listingColumns+` FROM listings
		 WHERE user_id = $1 AND status != 'deleted'
		 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	listings, err := r.scanListings(rows)
	return listings, total, err
}

func (r *ListingRepo) AddImage(ctx context.Context, id, imageURL string) (*model.Listing, error) {
	row := r.pool.QueryRow(ctx,
		`UPDATE listings
		 SET images = images || jsonb_build_array($2::text)
		 WHERE id = $1 AND status != 'deleted' AND jsonb_array_length(images) < 3
		 RETURNING `+listingColumns,
		id, imageURL,
	)
	return r.scanListing(row)
}

func (r *ListingRepo) RemoveImage(ctx context.Context, id, imageURL string) (*model.Listing, error) {
	row := r.pool.QueryRow(ctx,
		`UPDATE listings
		 SET images = (SELECT COALESCE(jsonb_agg(elem), '[]'::jsonb) FROM jsonb_array_elements(images) AS elem WHERE elem #>> '{}' != $2)
		 WHERE id = $1 AND status != 'deleted'
		 RETURNING `+listingColumns,
		id, imageURL,
	)
	return r.scanListing(row)
}

func (r *ListingRepo) GetImageCount(ctx context.Context, id string) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT jsonb_array_length(images) FROM listings WHERE id = $1`, id,
	).Scan(&count)
	return count, err
}

// --- Marketplace ---

func (r *ListingRepo) Browse(ctx context.Context, page, limit int) ([]*model.Listing, int, error) {
	offset := (page - 1) * limit

	var total int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM listings WHERE status = 'active'`,
	).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.Query(ctx,
		`SELECT `+listingColumns+` FROM listings
		 WHERE status = 'active'
		 ORDER BY `+rankingExpr+` DESC, created_at DESC
		 LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	listings, err := r.scanListings(rows)
	return listings, total, err
}

func (r *ListingRepo) Search(ctx context.Context, filter *model.ListingFilter) ([]*model.Listing, int, error) {
	offset := (filter.Page - 1) * filter.Limit

	where := []string{"status = 'active'"}
	args := []interface{}{}
	argIdx := 1

	if filter.Query != "" {
		// search_vector được rebuild với unaccent ở mig 032 → match cả "gao st"
		// (không dấu) lẫn "Gạo ST25" (có dấu). Apply unaccent lên user query cho
		// đối xứng.
		where = append(where, fmt.Sprintf(
			"search_vector @@ plainto_tsquery('simple', unaccent($%d))", argIdx))
		args = append(args, filter.Query)
		argIdx++
	}

	if filter.Category != "" {
		where = append(where, fmt.Sprintf("category = $%d", argIdx))
		args = append(args, filter.Category)
		argIdx++
	}

	if filter.RiceType != "" {
		where = append(where, fmt.Sprintf("rice_type = $%d", argIdx))
		args = append(args, filter.RiceType)
		argIdx++
	}

	if filter.Province != "" {
		where = append(where, fmt.Sprintf("province = $%d", argIdx))
		args = append(args, filter.Province)
		argIdx++
	}

	if filter.Ward != "" {
		where = append(where, fmt.Sprintf("district = $%d", argIdx))
		args = append(args, filter.Ward)
		argIdx++
	}

	if filter.MinPrice != nil {
		where = append(where, fmt.Sprintf("price_per_kg >= $%d", argIdx))
		args = append(args, *filter.MinPrice)
		argIdx++
	}

	if filter.MaxPrice != nil {
		where = append(where, fmt.Sprintf("price_per_kg <= $%d", argIdx))
		args = append(args, *filter.MaxPrice)
		argIdx++
	}

	if filter.MinQty != nil {
		where = append(where, fmt.Sprintf("quantity_kg >= $%d", argIdx))
		args = append(args, *filter.MinQty)
		argIdx++
	}

	if filter.HasPhoto {
		where = append(where, "images != '[]'::jsonb")
	}

	if filter.PostedWithinDays > 0 {
		where = append(where, fmt.Sprintf(
			"GREATEST(created_at, updated_at, COALESCE(bumped_at, '1970-01-01'::timestamptz)) > NOW() - ($%d::text || ' days')::interval",
			argIdx))
		args = append(args, fmt.Sprintf("%d", filter.PostedWithinDays))
		argIdx++
	}

	whereClause := strings.Join(where, " AND ")

	// Count
	var total int
	countQuery := "SELECT COUNT(*) FROM listings WHERE " + whereClause
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Data — default ordering uses content_score (relevance ranking), user can
	// override with explicit sort. quantity_desc & quantity_asc added.
	orderBy := rankingExpr + " DESC, created_at DESC"
	switch filter.Sort {
	case "price_asc":
		orderBy = "price_per_kg ASC"
	case "price_desc":
		orderBy = "price_per_kg DESC"
	case "name_asc":
		orderBy = "rice_type ASC"
	case "name_desc":
		orderBy = "rice_type DESC"
	case "quantity_desc":
		orderBy = "quantity_kg DESC"
	case "quantity_asc":
		orderBy = "quantity_kg ASC"
	case "newest":
		orderBy = "created_at DESC"
	}
	dataQuery := fmt.Sprintf(
		"SELECT %s FROM listings WHERE %s ORDER BY %s LIMIT $%d OFFSET $%d",
		listingColumns, whereClause, orderBy, argIdx, argIdx+1,
	)
	args = append(args, filter.Limit, offset)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	listings, err := r.scanListings(rows)
	return listings, total, err
}

func (r *ListingRepo) GetDetailWithSeller(ctx context.Context, id string) (*model.ListingDetail, error) {
	var d model.ListingDetail
	d.Seller = &model.PublicProfile{}
	var imagesJSON []byte
	var discardUserDistrict *string
	err := r.pool.QueryRow(ctx,
		`SELECT l.id, l.user_id, l.title, l.category, l.rice_type, l.province, l.district,
			l.quantity_kg, l.price_per_kg, l.harvest_season, l.description, l.certifications,
			l.images, l.status, l.view_count, l.bumped_at, l.bump_count, l.created_at, l.updated_at,
			u.id, u.phone, u.role, u.name, u.avatar_url, u.province, u.district, u.ward, u.description, u.org_name, u.created_at
		 FROM listings l
		 JOIN users u ON u.id = l.user_id
		 WHERE l.id = $1 AND l.status = 'active'`, id,
	).Scan(
		&d.ID, &d.UserID, &d.Title, &d.Category, &d.RiceType, &d.Province, &d.Ward,
		&d.QuantityKG, &d.PricePerKG, &d.HarvestSeason, &d.Description, &d.Certifications,
		&imagesJSON, &d.Status, &d.ViewCount, &d.BumpedAt, &d.BumpCount, &d.CreatedAt, &d.UpdatedAt,
		&d.Seller.ID, &d.Seller.Phone, &d.Seller.Role, &d.Seller.Name, &d.Seller.AvatarURL,
		&d.Seller.Province, &discardUserDistrict, &d.Seller.Ward, &d.Seller.Description, &d.Seller.OrgName, &d.Seller.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrListingNotFound
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(imagesJSON, &d.Images); err != nil || d.Images == nil {
		d.Images = []string{}
	}
	return &d, nil
}

func (r *ListingRepo) IncrementViewCount(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE listings SET view_count = view_count + 1 WHERE id = $1`, id,
	)
	return err
}

// Bump — atomic "Làm mới tin đăng". Trả về row mới NẾU thoả mãn cooldown
// (bumped_at NULL hoặc > 5h54m trước) VÀ chưa hết lifetime quota (bump_count < 240).
// Trả ErrBumpCooldown hoặc ErrBumpQuotaExhausted tuỳ nguyên nhân fail. Tiếp tục
// chạy 1 query GetByID sau đó để service biết tin có tồn tại không.
func (r *ListingRepo) Bump(ctx context.Context, listingID, userID string) (*model.Listing, error) {
	cooldownInterval := fmt.Sprintf("%d minutes", model.BumpCooldownMinutes)
	row := r.pool.QueryRow(ctx,
		`UPDATE listings
		    SET bumped_at = NOW(),
		        bump_count = bump_count + 1
		  WHERE id = $1
		    AND user_id = $2
		    AND status = 'active'
		    AND bump_count < $3
		    AND (bumped_at IS NULL OR bumped_at < NOW() - ($4::text)::interval)
		  RETURNING `+listingColumns,
		listingID, userID, model.BumpLifetimeCap, cooldownInterval,
	)
	updated, err := r.scanListing(row)
	if err == nil {
		return updated, nil
	}
	if !errors.Is(err, ErrListingNotFound) {
		return nil, err
	}
	// UPDATE didn't match — distinguish reasons.
	current, getErr := r.GetByID(ctx, listingID)
	if getErr != nil {
		return nil, getErr
	}
	if current.UserID != userID {
		return nil, ErrListingNotFound
	}
	if current.Status != "active" {
		return nil, ErrListingNotFound
	}
	if current.BumpCount >= model.BumpLifetimeCap {
		return nil, ErrBumpQuotaExhausted
	}
	return nil, ErrBumpCooldown
}

// PriceBoardRow holds aggregated price data for one product.
type PriceBoardRow struct {
	Category     string
	RiceType     string
	MinPrice     float64
	ListingCount int
}

// GetPriceBoardData returns MIN(price_per_kg) and COUNT grouped by (category, rice_type).
func (r *ListingRepo) GetPriceBoardData(ctx context.Context) ([]PriceBoardRow, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT category, rice_type, MIN(price_per_kg), COUNT(*)
		 FROM listings
		 WHERE status = 'active' AND category IS NOT NULL
		 GROUP BY category, rice_type`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []PriceBoardRow
	for rows.Next() {
		var row PriceBoardRow
		if err := rows.Scan(&row.Category, &row.RiceType, &row.MinPrice, &row.ListingCount); err != nil {
			return nil, err
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

// SEOPriceRow holds aggregated price data per (province, category, rice_type)
// for SEO landing pages. Province name is normalized (strip "Tỉnh "/"Thành phố ").
type SEOPriceRow struct {
	Province      string // normalized name, e.g. "Đồng Nai"
	Category      string // enum key, e.g. "gao_deo_thom"
	CategoryLabel string // human-readable, e.g. "Gạo dẻo thơm"
	RiceType      string // enum key, e.g. "dai_loan"
	RiceTypeLabel string // human-readable, e.g. "Đài Loan"
	MinPrice      float64
	AvgPrice      float64
	MaxPrice      float64
	ListingCount  int
	LastUpdated   time.Time
}

// GetSEOPriceBoard returns price aggregation grouped by (province, category,
// rice_type) for static SEO landing pages. Province name is normalized
// to deduplicate "Tỉnh X" and "X" entries from inconsistent user input.
func (r *ListingRepo) GetSEOPriceBoard(ctx context.Context) ([]SEOPriceRow, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT
		    TRIM(regexp_replace(l.province, '^(Tỉnh|Thành phố|TP\.?|TP|Tỉnh\.?)\s+', '', 'i')) AS province_norm,
		    l.category,
		    COALESCE(c.label, l.category) AS category_label,
		    l.rice_type,
		    COALESCE(p.label, l.rice_type) AS rice_type_label,
		    MIN(l.price_per_kg) AS min_price,
		    ROUND(AVG(l.price_per_kg))::float8 AS avg_price,
		    MAX(l.price_per_kg) AS max_price,
		    COUNT(*) AS listing_count,
		    MAX(l.updated_at) AS last_updated
		 FROM listings l
		 LEFT JOIN rice_categories c ON c.key = l.category
		 LEFT JOIN rice_products p ON p.key = l.rice_type
		 WHERE l.status = 'active'
		   AND l.category IS NOT NULL
		   AND l.province IS NOT NULL
		   AND TRIM(l.province) <> ''
		 GROUP BY province_norm, l.category, c.label, l.rice_type, p.label
		 ORDER BY province_norm, l.category, l.rice_type`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []SEOPriceRow
	for rows.Next() {
		var row SEOPriceRow
		if err := rows.Scan(
			&row.Province, &row.Category, &row.CategoryLabel,
			&row.RiceType, &row.RiceTypeLabel,
			&row.MinPrice, &row.AvgPrice, &row.MaxPrice,
			&row.ListingCount, &row.LastUpdated,
		); err != nil {
			return nil, err
		}
		result = append(result, row)
	}
	return result, rows.Err()
}
