package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sangiagao/rice-marketplace/internal/model"
)

type CatalogRepo struct {
	pool *pgxpool.Pool
}

func NewCatalogRepo(pool *pgxpool.Pool) *CatalogRepo {
	return &CatalogRepo{pool: pool}
}

// --- Categories ---

func (r *CatalogRepo) ListCategories(ctx context.Context) ([]*model.CatalogCategory, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, key, label, icon, sort_order, is_active, created_at, updated_at
		 FROM rice_categories ORDER BY sort_order, key`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cats []*model.CatalogCategory
	for rows.Next() {
		c := &model.CatalogCategory{}
		if err := rows.Scan(&c.ID, &c.Key, &c.Label, &c.Icon, &c.SortOrder, &c.IsActive, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		cats = append(cats, c)
	}
	return cats, nil
}

func (r *CatalogRepo) CreateCategory(ctx context.Context, req *model.CreateCategoryRequest) (*model.CatalogCategory, error) {
	// Auto sort_order = max + 1
	var maxOrder int
	_ = r.pool.QueryRow(ctx, `SELECT COALESCE(MAX(sort_order), 0) FROM rice_categories`).Scan(&maxOrder)

	icon := "category"
	if req.Icon != "" {
		icon = req.Icon
	}

	c := &model.CatalogCategory{}
	err := r.pool.QueryRow(ctx,
		`INSERT INTO rice_categories (key, label, icon, sort_order)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, key, label, icon, sort_order, is_active, created_at, updated_at`,
		req.Key, req.Label, icon, maxOrder+1,
	).Scan(&c.ID, &c.Key, &c.Label, &c.Icon, &c.SortOrder, &c.IsActive, &c.CreatedAt, &c.UpdatedAt)
	return c, err
}

func (r *CatalogRepo) UpdateCategory(ctx context.Context, id string, req *model.UpdateCategoryRequest) (*model.CatalogCategory, error) {
	c := &model.CatalogCategory{}
	err := r.pool.QueryRow(ctx,
		`UPDATE rice_categories SET
			label = COALESCE($2, label),
			icon = COALESCE($3, icon),
			sort_order = COALESCE($4, sort_order),
			is_active = COALESCE($5, is_active)
		 WHERE id = $1
		 RETURNING id, key, label, icon, sort_order, is_active, created_at, updated_at`,
		id, req.Label, req.Icon, req.SortOrder, req.IsActive,
	).Scan(&c.ID, &c.Key, &c.Label, &c.Icon, &c.SortOrder, &c.IsActive, &c.CreatedAt, &c.UpdatedAt)
	return c, err
}

func (r *CatalogRepo) DeleteCategory(ctx context.Context, id string) error {
	ct, err := r.pool.Exec(ctx, `DELETE FROM rice_categories WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("category not found")
	}
	return nil
}

// --- Products ---

func (r *CatalogRepo) ListProducts(ctx context.Context) ([]*model.CatalogProduct, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, key, label, category_id, sort_order, is_active, created_at, updated_at
		 FROM rice_products ORDER BY sort_order, key`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*model.CatalogProduct
	for rows.Next() {
		p := &model.CatalogProduct{}
		if err := rows.Scan(&p.ID, &p.Key, &p.Label, &p.CategoryID, &p.SortOrder, &p.IsActive, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, nil
}

func (r *CatalogRepo) CreateProduct(ctx context.Context, req *model.CreateProductRequest) (*model.CatalogProduct, error) {
	var maxOrder int
	_ = r.pool.QueryRow(ctx, `SELECT COALESCE(MAX(sort_order), 0) FROM rice_products WHERE category_id = $1`, req.CategoryID).Scan(&maxOrder)

	p := &model.CatalogProduct{}
	err := r.pool.QueryRow(ctx,
		`INSERT INTO rice_products (key, label, category_id, sort_order)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, key, label, category_id, sort_order, is_active, created_at, updated_at`,
		req.Key, req.Label, req.CategoryID, maxOrder+1,
	).Scan(&p.ID, &p.Key, &p.Label, &p.CategoryID, &p.SortOrder, &p.IsActive, &p.CreatedAt, &p.UpdatedAt)
	return p, err
}

func (r *CatalogRepo) UpdateProduct(ctx context.Context, id string, req *model.UpdateProductRequest) (*model.CatalogProduct, error) {
	p := &model.CatalogProduct{}
	err := r.pool.QueryRow(ctx,
		`UPDATE rice_products SET
			label = COALESCE($2, label),
			category_id = COALESCE($3, category_id),
			sort_order = COALESCE($4, sort_order),
			is_active = COALESCE($5, is_active)
		 WHERE id = $1
		 RETURNING id, key, label, category_id, sort_order, is_active, created_at, updated_at`,
		id, req.Label, req.CategoryID, req.SortOrder, req.IsActive,
	).Scan(&p.ID, &p.Key, &p.Label, &p.CategoryID, &p.SortOrder, &p.IsActive, &p.CreatedAt, &p.UpdatedAt)
	return p, err
}

func (r *CatalogRepo) DeleteProduct(ctx context.Context, id string) error {
	ct, err := r.pool.Exec(ctx, `DELETE FROM rice_products WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("product not found")
	}
	return nil
}

// ValidateCategory checks if a category key exists and is active.
func (r *CatalogRepo) ValidateCategory(ctx context.Context, categoryKey string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM rice_categories WHERE key = $1 AND is_active = true)`,
		categoryKey,
	).Scan(&exists)
	return exists, err
}

// ValidateProductInCategory checks if a product key belongs to the given category and both are active.
func (r *CatalogRepo) ValidateProductInCategory(ctx context.Context, categoryKey, productKey string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM rice_products p
			JOIN rice_categories c ON c.id = p.category_id
			WHERE p.key = $1 AND c.key = $2 AND p.is_active = true AND c.is_active = true
		)`,
		productKey, categoryKey,
	).Scan(&exists)
	return exists, err
}

// GetProductLabelByKey returns the label of a product by its key.
func (r *CatalogRepo) GetProductLabelByKey(ctx context.Context, productKey string) (string, error) {
	var label string
	err := r.pool.QueryRow(ctx,
		`SELECT label FROM rice_products WHERE key = $1 AND is_active = true`,
		productKey,
	).Scan(&label)
	return label, err
}

// GetCatalogForAPI returns categories with nested products (only active) for public API.
func (r *CatalogRepo) GetCatalogForAPI(ctx context.Context) ([]model.RiceCategory, error) {
	cats, err := r.ListCategories(ctx)
	if err != nil {
		return nil, err
	}
	prods, err := r.ListProducts(ctx)
	if err != nil {
		return nil, err
	}

	// Group products by category
	prodByCat := make(map[string][]*model.CatalogProduct)
	for _, p := range prods {
		if p.IsActive {
			prodByCat[p.CategoryID] = append(prodByCat[p.CategoryID], p)
		}
	}

	var result []model.RiceCategory
	for _, c := range cats {
		if !c.IsActive {
			continue
		}
		var products []model.RiceProduct
		for _, p := range prodByCat[c.ID] {
			products = append(products, model.RiceProduct{
				Key:      p.Key,
				Label:    p.Label,
				Category: c.Key,
			})
		}
		result = append(result, model.RiceCategory{
			Key:      c.Key,
			Label:    c.Label,
			Products: products,
		})
	}
	return result, nil
}
