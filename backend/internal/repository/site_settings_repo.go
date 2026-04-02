package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sangiagao/rice-marketplace/internal/model"
)

type SiteSettingsRepo struct {
	pool *pgxpool.Pool
}

func NewSiteSettingsRepo(pool *pgxpool.Pool) *SiteSettingsRepo {
	return &SiteSettingsRepo{pool: pool}
}

func (r *SiteSettingsRepo) Get(ctx context.Context, key string) (*model.SiteSetting, error) {
	var s model.SiteSetting
	err := r.pool.QueryRow(ctx,
		`SELECT key, value, updated_at FROM site_settings WHERE key = $1`, key,
	).Scan(&s.Key, &s.Value, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *SiteSettingsRepo) Set(ctx context.Context, key, value string) (*model.SiteSetting, error) {
	var s model.SiteSetting
	err := r.pool.QueryRow(ctx,
		`INSERT INTO site_settings (key, value, updated_at)
		 VALUES ($1, $2, NOW())
		 ON CONFLICT (key) DO UPDATE SET value = $2, updated_at = NOW()
		 RETURNING key, value, updated_at`, key, value,
	).Scan(&s.Key, &s.Value, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}
