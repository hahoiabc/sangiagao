package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sangiagao/rice-marketplace/internal/model"
)

var ErrListingNotFound = errors.New("listing not found")

type ListingRepo struct {
	pool *pgxpool.Pool
}

func NewListingRepo(pool *pgxpool.Pool) *ListingRepo {
	return &ListingRepo{pool: pool}
}

const listingColumns = `id, user_id, title, category, rice_type, province, district,
	quantity_kg, price_per_kg, harvest_season, description, certifications,
	images, status, view_count, created_at, updated_at`

func (r *ListingRepo) scanListing(row pgx.Row) (*model.Listing, error) {
	var l model.Listing
	var imagesJSON []byte
	err := row.Scan(
		&l.ID, &l.UserID, &l.Title, &l.Category, &l.RiceType, &l.Province, &l.Ward,
		&l.QuantityKG, &l.PricePerKG, &l.HarvestSeason, &l.Description, &l.Certifications,
		&imagesJSON, &l.Status, &l.ViewCount, &l.CreatedAt, &l.UpdatedAt,
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
			&imagesJSON, &l.Status, &l.ViewCount, &l.CreatedAt, &l.UpdatedAt,
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
		 ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
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
		// Convert search query to tsquery format
		words := strings.Fields(filter.Query)
		tsquery := strings.Join(words, " & ")
		where = append(where, fmt.Sprintf("search_vector @@ to_tsquery('simple', $%d)", argIdx))
		args = append(args, tsquery)
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

	whereClause := strings.Join(where, " AND ")

	// Count
	var total int
	countQuery := "SELECT COUNT(*) FROM listings WHERE " + whereClause
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Data
	orderBy := "created_at DESC"
	if filter.Sort == "price_asc" {
		orderBy = "price_per_kg ASC"
	} else if filter.Sort == "price_desc" {
		orderBy = "price_per_kg DESC"
	} else if filter.Sort == "name_asc" {
		orderBy = "rice_type ASC"
	} else if filter.Sort == "name_desc" {
		orderBy = "rice_type DESC"
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
			l.images, l.status, l.view_count, l.created_at, l.updated_at,
			u.id, u.phone, u.role, u.name, u.avatar_url, u.province, u.district, u.ward, u.description, u.org_name, u.created_at
		 FROM listings l
		 JOIN users u ON u.id = l.user_id
		 WHERE l.id = $1 AND l.status = 'active'`, id,
	).Scan(
		&d.ID, &d.UserID, &d.Title, &d.Category, &d.RiceType, &d.Province, &d.Ward,
		&d.QuantityKG, &d.PricePerKG, &d.HarvestSeason, &d.Description, &d.Certifications,
		&imagesJSON, &d.Status, &d.ViewCount, &d.CreatedAt, &d.UpdatedAt,
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
