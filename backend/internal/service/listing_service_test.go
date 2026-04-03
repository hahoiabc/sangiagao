package service

import (
	"context"
	"testing"
	"time"

	"github.com/sangiagao/rice-marketplace/internal/model"
	"github.com/sangiagao/rice-marketplace/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Mock ---

type mockListingRepo struct{ mock.Mock }

func (m *mockListingRepo) CountTodayByUser(ctx context.Context, userID string) (int, error) {
	args := m.Called(ctx, userID)
	return args.Int(0), args.Error(1)
}
func (m *mockListingRepo) CountTodayByUserAndType(ctx context.Context, userID, riceType string) (int, error) {
	args := m.Called(ctx, userID, riceType)
	return args.Int(0), args.Error(1)
}

func (m *mockListingRepo) Create(ctx context.Context, userID string, req *model.CreateListingRequest) (*model.Listing, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Listing), args.Error(1)
}

func (m *mockListingRepo) GetByID(ctx context.Context, id string) (*model.Listing, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Listing), args.Error(1)
}

func (m *mockListingRepo) Update(ctx context.Context, id string, req *model.UpdateListingRequest) (*model.Listing, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Listing), args.Error(1)
}

func (m *mockListingRepo) SoftDelete(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockListingRepo) BatchSoftDelete(ctx context.Context, ids []string) (int, error) {
	args := m.Called(ctx, ids)
	return args.Int(0), args.Error(1)
}

func (m *mockListingRepo) ListByUser(ctx context.Context, userID string, page, limit int) ([]*model.Listing, int, error) {
	args := m.Called(ctx, userID, page, limit)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*model.Listing), args.Int(1), args.Error(2)
}

func (m *mockListingRepo) AddImage(ctx context.Context, id, imageURL string) (*model.Listing, error) {
	args := m.Called(ctx, id, imageURL)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Listing), args.Error(1)
}

func (m *mockListingRepo) GetImageCount(ctx context.Context, id string) (int, error) {
	args := m.Called(ctx, id)
	return args.Int(0), args.Error(1)
}

func (m *mockListingRepo) Browse(ctx context.Context, page, limit int) ([]*model.Listing, int, error) {
	args := m.Called(ctx, page, limit)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*model.Listing), args.Int(1), args.Error(2)
}

func (m *mockListingRepo) Search(ctx context.Context, filter *model.ListingFilter) ([]*model.Listing, int, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*model.Listing), args.Int(1), args.Error(2)
}

func (m *mockListingRepo) GetDetailWithSeller(ctx context.Context, id string) (*model.ListingDetail, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ListingDetail), args.Error(1)
}

func (m *mockListingRepo) IncrementViewCount(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mockListingRepo) GetPriceBoardData(ctx context.Context) ([]repository.PriceBoardRow, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]repository.PriceBoardRow), args.Error(1)
}

type mockCatalogRepo struct{ mock.Mock }

func (m *mockCatalogRepo) ListCategories(ctx context.Context) ([]*model.CatalogCategory, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.CatalogCategory), args.Error(1)
}
func (m *mockCatalogRepo) CreateCategory(ctx context.Context, req *model.CreateCategoryRequest) (*model.CatalogCategory, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.CatalogCategory), args.Error(1)
}
func (m *mockCatalogRepo) UpdateCategory(ctx context.Context, id string, req *model.UpdateCategoryRequest) (*model.CatalogCategory, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.CatalogCategory), args.Error(1)
}
func (m *mockCatalogRepo) DeleteCategory(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockCatalogRepo) ListProducts(ctx context.Context) ([]*model.CatalogProduct, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.CatalogProduct), args.Error(1)
}
func (m *mockCatalogRepo) CreateProduct(ctx context.Context, req *model.CreateProductRequest) (*model.CatalogProduct, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.CatalogProduct), args.Error(1)
}
func (m *mockCatalogRepo) UpdateProduct(ctx context.Context, id string, req *model.UpdateProductRequest) (*model.CatalogProduct, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.CatalogProduct), args.Error(1)
}
func (m *mockCatalogRepo) DeleteProduct(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockCatalogRepo) GetCatalogForAPI(ctx context.Context) ([]model.RiceCategory, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.RiceCategory), args.Error(1)
}
func (m *mockCatalogRepo) ValidateCategory(ctx context.Context, categoryKey string) (bool, error) {
	args := m.Called(ctx, categoryKey)
	return args.Bool(0), args.Error(1)
}
func (m *mockCatalogRepo) ValidateProductInCategory(ctx context.Context, categoryKey, productKey string) (bool, error) {
	args := m.Called(ctx, categoryKey, productKey)
	return args.Bool(0), args.Error(1)
}
func (m *mockCatalogRepo) GetProductLabelByKey(ctx context.Context, productKey string) (string, error) {
	args := m.Called(ctx, productKey)
	return args.String(0), args.Error(1)
}

// --- Helpers ---

func sampleListing(userID string) *model.Listing {
	return &model.Listing{
		ID: "listing-1", UserID: userID, Title: "Gạo ST25",
		RiceType: "ST25", Province: strPtr("Long An"), QuantityKG: 500,
		PricePerKG: 28000, Images: []string{}, Status: "active",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
}

// --- Create Tests ---

// testCatalog returns catalog data used by in-memory cache in tests.
func testCatalog() []model.RiceCategory {
	return []model.RiceCategory{
		{Key: "gao_deo_thom", Label: "Gạo dẻo thơm", Products: []model.RiceProduct{
			{Key: "st_25", Label: "ST25", Category: "gao_deo_thom"},
		}},
	}
}

func TestListingCreate_Success(t *testing.T) {
	repo := new(mockListingRepo)
	catRepo := new(mockCatalogRepo)
	svc := NewListingService(repo, nil, nil, catRepo)

	req := &model.CreateListingRequest{Title: "Gạo ST25", Category: "gao_deo_thom", RiceType: "st_25", QuantityKG: 500, PricePerKG: 28000}
	expected := sampleListing("user-1")
	repo.On("CountTodayByUserAndType", mock.Anything, "user-1", "st_25").Return(0, nil)
	catRepo.On("GetCatalogForAPI", mock.Anything).Return(testCatalog(), nil)
	repo.On("Create", mock.Anything, "user-1", req).Return(expected, nil)

	result, err := svc.Create(context.Background(), "user-1", req)
	assert.NoError(t, err)
	assert.Equal(t, expected.ID, result.ID)
}

func TestListingCreate_DailyLimitReached(t *testing.T) {
	repo := new(mockListingRepo)
	svc := NewListingService(repo, nil, nil, nil)

	req := &model.CreateListingRequest{Title: "Test", Category: "gao_deo_thom", RiceType: "st_25", QuantityKG: 1, PricePerKG: 1}
	repo.On("CountTodayByUserAndType", mock.Anything, "user-1", "st_25").Return(3, nil)

	_, err := svc.Create(context.Background(), "user-1", req)
	assert.ErrorIs(t, err, ErrDailyLimitReached)
}

func TestListingCreate_RepoError(t *testing.T) {
	repo := new(mockListingRepo)
	catRepo := new(mockCatalogRepo)
	svc := NewListingService(repo, nil, nil, catRepo)

	req := &model.CreateListingRequest{Title: "Test", Category: "gao_deo_thom", RiceType: "st_25", QuantityKG: 1, PricePerKG: 1}
	repo.On("CountTodayByUserAndType", mock.Anything, "user-1", "st_25").Return(0, nil)
	catRepo.On("GetCatalogForAPI", mock.Anything).Return(testCatalog(), nil)
	repo.On("Create", mock.Anything, "user-1", req).Return(nil, assert.AnError)

	_, err := svc.Create(context.Background(), "user-1", req)
	assert.Error(t, err)
}

// --- GetByID Tests ---

func TestListingGetByID_Success(t *testing.T) {
	repo := new(mockListingRepo)
	svc := NewListingService(repo, nil, nil, nil)

	expected := sampleListing("user-1")
	repo.On("GetByID", mock.Anything, "listing-1").Return(expected, nil)

	result, err := svc.GetByID(context.Background(), "listing-1")
	assert.NoError(t, err)
	assert.Equal(t, "listing-1", result.ID)
}

func TestListingGetByID_NotFound(t *testing.T) {
	repo := new(mockListingRepo)
	svc := NewListingService(repo, nil, nil, nil)

	repo.On("GetByID", mock.Anything, "bad-id").Return(nil, repository.ErrListingNotFound)

	_, err := svc.GetByID(context.Background(), "bad-id")
	assert.ErrorIs(t, err, repository.ErrListingNotFound)
}

// --- Update Tests ---

func TestListingUpdate_Success(t *testing.T) {
	repo := new(mockListingRepo)
	svc := NewListingService(repo, nil, nil, nil)

	existing := sampleListing("user-1")
	repo.On("GetByID", mock.Anything, "listing-1").Return(existing, nil)

	newPrice := float64(25000)
	req := &model.UpdateListingRequest{PricePerKG: &newPrice}
	updated := sampleListing("user-1")
	updated.PricePerKG = 25000
	repo.On("Update", mock.Anything, "listing-1", req).Return(updated, nil)

	result, err := svc.Update(context.Background(), "user-1", "listing-1", req)
	assert.NoError(t, err)
	assert.Equal(t, float64(25000), result.PricePerKG)
}

func TestListingUpdate_NotOwner(t *testing.T) {
	repo := new(mockListingRepo)
	svc := NewListingService(repo, nil, nil, nil)

	existing := sampleListing("user-1")
	repo.On("GetByID", mock.Anything, "listing-1").Return(existing, nil)

	req := &model.UpdateListingRequest{}
	_, err := svc.Update(context.Background(), "user-2", "listing-1", req)
	assert.ErrorIs(t, err, ErrNotListingOwner)
}

func TestListingUpdate_NotFound(t *testing.T) {
	repo := new(mockListingRepo)
	svc := NewListingService(repo, nil, nil, nil)

	repo.On("GetByID", mock.Anything, "bad-id").Return(nil, repository.ErrListingNotFound)

	req := &model.UpdateListingRequest{}
	_, err := svc.Update(context.Background(), "user-1", "bad-id", req)
	assert.ErrorIs(t, err, repository.ErrListingNotFound)
}

// --- Delete Tests ---

func TestListingDelete_Success(t *testing.T) {
	repo := new(mockListingRepo)
	svc := NewListingService(repo, nil, nil, nil)

	existing := sampleListing("user-1")
	repo.On("GetByID", mock.Anything, "listing-1").Return(existing, nil)
	repo.On("SoftDelete", mock.Anything, "listing-1").Return(nil)

	err := svc.Delete(context.Background(), "user-1", "listing-1")
	assert.NoError(t, err)
}

func TestListingDelete_NotOwner(t *testing.T) {
	repo := new(mockListingRepo)
	svc := NewListingService(repo, nil, nil, nil)

	existing := sampleListing("user-1")
	repo.On("GetByID", mock.Anything, "listing-1").Return(existing, nil)

	err := svc.Delete(context.Background(), "user-2", "listing-1")
	assert.ErrorIs(t, err, ErrNotListingOwner)
}

// --- ListByUser Tests ---

func TestListByUser_Success(t *testing.T) {
	repo := new(mockListingRepo)
	svc := NewListingService(repo, nil, nil, nil)

	listings := []*model.Listing{sampleListing("user-1")}
	repo.On("ListByUser", mock.Anything, "user-1", 1, 20).Return(listings, 1, nil)

	result, total, err := svc.ListByUser(context.Background(), "user-1", 1, 20)
	assert.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, result, 1)
}

func TestListByUser_DefaultsPagination(t *testing.T) {
	repo := new(mockListingRepo)
	svc := NewListingService(repo, nil, nil, nil)

	repo.On("ListByUser", mock.Anything, "user-1", 1, 20).Return([]*model.Listing{}, 0, nil)

	_, _, err := svc.ListByUser(context.Background(), "user-1", 0, 0)
	assert.NoError(t, err)
	repo.AssertCalled(t, "ListByUser", mock.Anything, "user-1", 1, 20)
}

func TestListByUser_MaxLimit(t *testing.T) {
	repo := new(mockListingRepo)
	svc := NewListingService(repo, nil, nil, nil)

	repo.On("ListByUser", mock.Anything, "user-1", 1, 20).Return([]*model.Listing{}, 0, nil)

	_, _, err := svc.ListByUser(context.Background(), "user-1", 1, 100)
	assert.NoError(t, err)
	repo.AssertCalled(t, "ListByUser", mock.Anything, "user-1", 1, 20)
}

// --- AddImage Tests ---

func TestAddImage_Success(t *testing.T) {
	repo := new(mockListingRepo)
	svc := NewListingService(repo, nil, nil, nil)

	existing := sampleListing("user-1")
	existing.Images = []string{}
	repo.On("GetByID", mock.Anything, "listing-1").Return(existing, nil)

	updated := sampleListing("user-1")
	updated.Images = []string{"img1.jpg"}
	repo.On("AddImage", mock.Anything, "listing-1", "img1.jpg").Return(updated, nil)

	result, err := svc.AddImage(context.Background(), "user-1", "listing-1", "img1.jpg")
	assert.NoError(t, err)
	assert.Len(t, result.Images, 1)
}

func TestAddImage_MaxImages(t *testing.T) {
	repo := new(mockListingRepo)
	svc := NewListingService(repo, nil, nil, nil)

	existing := sampleListing("user-1")
	existing.Images = []string{"a.jpg", "b.jpg", "c.jpg"}
	repo.On("GetByID", mock.Anything, "listing-1").Return(existing, nil)

	_, err := svc.AddImage(context.Background(), "user-1", "listing-1", "d.jpg")
	assert.ErrorIs(t, err, ErrMaxImages)
}

func TestAddImage_NotOwner(t *testing.T) {
	repo := new(mockListingRepo)
	svc := NewListingService(repo, nil, nil, nil)

	existing := sampleListing("user-1")
	repo.On("GetByID", mock.Anything, "listing-1").Return(existing, nil)

	_, err := svc.AddImage(context.Background(), "user-2", "listing-1", "img.jpg")
	assert.ErrorIs(t, err, ErrNotListingOwner)
}

// --- Browse Tests ---

func TestBrowse_Success(t *testing.T) {
	repo := new(mockListingRepo)
	svc := NewListingService(repo, nil, nil, nil)

	listings := []*model.Listing{sampleListing("user-1")}
	repo.On("Browse", mock.Anything, 1, 20).Return(listings, 1, nil)

	result, total, err := svc.Browse(context.Background(), 1, 20)
	assert.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, result, 1)
}

func TestBrowse_DefaultsPagination(t *testing.T) {
	repo := new(mockListingRepo)
	svc := NewListingService(repo, nil, nil, nil)

	repo.On("Browse", mock.Anything, 1, 20).Return([]*model.Listing{}, 0, nil)

	_, _, err := svc.Browse(context.Background(), -1, 999)
	assert.NoError(t, err)
	repo.AssertCalled(t, "Browse", mock.Anything, 1, 20)
}

// --- Search Tests ---

func TestSearch_Success(t *testing.T) {
	repo := new(mockListingRepo)
	svc := NewListingService(repo, nil, nil, nil)

	filter := &model.ListingFilter{Query: "ST25", Page: 1, Limit: 20}
	listings := []*model.Listing{sampleListing("user-1")}
	repo.On("Search", mock.Anything, filter).Return(listings, 1, nil)

	result, total, err := svc.Search(context.Background(), filter)
	assert.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, result, 1)
}

func TestSearch_DefaultsPagination(t *testing.T) {
	repo := new(mockListingRepo)
	svc := NewListingService(repo, nil, nil, nil)

	filter := &model.ListingFilter{Query: "test", Page: 0, Limit: 0}
	repo.On("Search", mock.Anything, filter).Return([]*model.Listing{}, 0, nil)

	_, _, err := svc.Search(context.Background(), filter)
	assert.NoError(t, err)
	assert.Equal(t, 1, filter.Page)
	assert.Equal(t, 20, filter.Limit)
}

// --- GetDetail Tests ---

func TestGetDetail_Success(t *testing.T) {
	repo := new(mockListingRepo)
	svc := NewListingService(repo, nil, nil, nil)

	detail := &model.ListingDetail{
		Listing: *sampleListing("user-1"),
		Seller:  &model.PublicProfile{ID: "user-1", Role: "seller"},
	}
	repo.On("GetDetailWithSeller", mock.Anything, "listing-1").Return(detail, nil)
	repo.On("IncrementViewCount", mock.Anything, "listing-1").Return(nil)

	result, err := svc.GetDetail(context.Background(), "listing-1")
	assert.NoError(t, err)
	assert.Equal(t, "listing-1", result.ID)
	assert.NotNil(t, result.Seller)
}

func TestGetDetail_NotFound(t *testing.T) {
	repo := new(mockListingRepo)
	svc := NewListingService(repo, nil, nil, nil)

	repo.On("GetDetailWithSeller", mock.Anything, "bad-id").Return(nil, repository.ErrListingNotFound)

	_, err := svc.GetDetail(context.Background(), "bad-id")
	assert.ErrorIs(t, err, repository.ErrListingNotFound)
}
