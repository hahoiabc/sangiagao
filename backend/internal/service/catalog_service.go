package service

import (
	"context"
	"errors"

	"github.com/sangiagao/rice-marketplace/internal/model"
)

var (
	ErrCategoryKeyExists = errors.New("mã danh mục đã tồn tại")
	ErrProductKeyExists  = errors.New("mã sản phẩm đã tồn tại")
)

type CatalogService struct {
	catalogRepo CatalogRepository
}

func NewCatalogService(catalogRepo CatalogRepository) *CatalogService {
	return &CatalogService{catalogRepo: catalogRepo}
}

// --- Categories ---

func (s *CatalogService) ListCategories(ctx context.Context) ([]*model.CatalogCategory, error) {
	return s.catalogRepo.ListCategories(ctx)
}

func (s *CatalogService) CreateCategory(ctx context.Context, req *model.CreateCategoryRequest) (*model.CatalogCategory, error) {
	return s.catalogRepo.CreateCategory(ctx, req)
}

func (s *CatalogService) UpdateCategory(ctx context.Context, id string, req *model.UpdateCategoryRequest) (*model.CatalogCategory, error) {
	return s.catalogRepo.UpdateCategory(ctx, id, req)
}

func (s *CatalogService) DeleteCategory(ctx context.Context, id string) error {
	return s.catalogRepo.DeleteCategory(ctx, id)
}

// --- Products ---

func (s *CatalogService) ListProducts(ctx context.Context) ([]*model.CatalogProduct, error) {
	return s.catalogRepo.ListProducts(ctx)
}

func (s *CatalogService) CreateProduct(ctx context.Context, req *model.CreateProductRequest) (*model.CatalogProduct, error) {
	return s.catalogRepo.CreateProduct(ctx, req)
}

func (s *CatalogService) UpdateProduct(ctx context.Context, id string, req *model.UpdateProductRequest) (*model.CatalogProduct, error) {
	return s.catalogRepo.UpdateProduct(ctx, id, req)
}

func (s *CatalogService) DeleteProduct(ctx context.Context, id string) error {
	return s.catalogRepo.DeleteProduct(ctx, id)
}

// --- Public API ---

func (s *CatalogService) GetCatalogForAPI(ctx context.Context) ([]model.RiceCategory, error) {
	return s.catalogRepo.GetCatalogForAPI(ctx)
}
