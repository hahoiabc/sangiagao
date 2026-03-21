package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sangiagao/rice-marketplace/internal/model"
)

var ErrSponsorNotFound = errors.New("sponsor not found")

type SponsorRepo struct {
	pool *pgxpool.Pool
}

func NewSponsorRepo(pool *pgxpool.Pool) *SponsorRepo {
	return &SponsorRepo{pool: pool}
}

func (r *SponsorRepo) GetAllActive(ctx context.Context) ([]*model.ProductSponsor, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, product_key, logo_url, sponsor_name, is_active, created_at, updated_at
		 FROM product_sponsors WHERE is_active = true`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSponsors(rows)
}

func (r *SponsorRepo) Create(ctx context.Context, req *model.CreateSponsorRequest) (*model.ProductSponsor, error) {
	var s model.ProductSponsor
	err := r.pool.QueryRow(ctx,
		`INSERT INTO product_sponsors (product_key, logo_url, sponsor_name)
		 VALUES ($1, $2, $3)
		 RETURNING id, product_key, logo_url, sponsor_name, is_active, created_at, updated_at`,
		req.ProductKey, req.LogoURL, req.SponsorName,
	).Scan(&s.ID, &s.ProductKey, &s.LogoURL, &s.SponsorName, &s.IsActive, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *SponsorRepo) Update(ctx context.Context, id string, req *model.UpdateSponsorRequest) (*model.ProductSponsor, error) {
	var s model.ProductSponsor
	err := r.pool.QueryRow(ctx,
		`UPDATE product_sponsors SET
			logo_url = COALESCE($2, logo_url),
			sponsor_name = COALESCE($3, sponsor_name),
			is_active = COALESCE($4, is_active)
		 WHERE id = $1
		 RETURNING id, product_key, logo_url, sponsor_name, is_active, created_at, updated_at`,
		id, req.LogoURL, req.SponsorName, req.IsActive,
	).Scan(&s.ID, &s.ProductKey, &s.LogoURL, &s.SponsorName, &s.IsActive, &s.CreatedAt, &s.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrSponsorNotFound
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *SponsorRepo) Delete(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM product_sponsors WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrSponsorNotFound
	}
	return nil
}

func (r *SponsorRepo) List(ctx context.Context, page, limit int) ([]*model.ProductSponsor, int, error) {
	offset := (page - 1) * limit
	var total int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM product_sponsors`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.Query(ctx,
		`SELECT id, product_key, logo_url, sponsor_name, is_active, created_at, updated_at
		 FROM product_sponsors ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	sponsors, err := scanSponsors(rows)
	return sponsors, total, err
}

func scanSponsors(rows pgx.Rows) ([]*model.ProductSponsor, error) {
	var sponsors []*model.ProductSponsor
	for rows.Next() {
		var s model.ProductSponsor
		if err := rows.Scan(&s.ID, &s.ProductKey, &s.LogoURL, &s.SponsorName, &s.IsActive, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		sponsors = append(sponsors, &s)
	}
	if sponsors == nil {
		sponsors = []*model.ProductSponsor{}
	}
	return sponsors, rows.Err()
}
