package handler

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sangiagao/rice-marketplace/internal/model"
	"github.com/sangiagao/rice-marketplace/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func marketplaceRouter(h *MarketplaceHandler) *gin.Engine {
	r := gin.New()
	r.GET("/marketplace", h.Browse)
	r.GET("/marketplace/search", h.Search)
	r.GET("/marketplace/:id", h.GetDetail)
	return r
}

func doGet(r *gin.Engine, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("GET", path, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// --- Browse Tests ---

func TestBrowse_Success(t *testing.T) {
	svc := new(mockListingService)
	h := NewMarketplaceHandler(svc, nil)
	r := marketplaceRouter(h)

	listings := []*model.Listing{{ID: "l-1", Images: []string{}}, {ID: "l-2", Images: []string{}}}
	svc.On("Browse", mock.Anything, 1, 20).Return(listings, 2, nil)

	w := doGet(r, "/marketplace")
	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "l-1")
	assert.Contains(t, w.Body.String(), "l-2")
	assert.Contains(t, w.Body.String(), `"total":2`)
}

func TestBrowse_WithPagination(t *testing.T) {
	svc := new(mockListingService)
	h := NewMarketplaceHandler(svc, nil)
	r := marketplaceRouter(h)

	svc.On("Browse", mock.Anything, 2, 10).Return([]*model.Listing{}, 0, nil)

	w := doGet(r, "/marketplace?page=2&limit=10")
	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), `"page":2`)
}

func TestBrowse_ServerError(t *testing.T) {
	svc := new(mockListingService)
	h := NewMarketplaceHandler(svc, nil)
	r := marketplaceRouter(h)

	svc.On("Browse", mock.Anything, mock.Anything, mock.Anything).Return(nil, 0, assert.AnError)

	w := doGet(r, "/marketplace")
	assert.Equal(t, 500, w.Code)
}

// --- Search Tests ---

func TestSearch_ByQuery(t *testing.T) {
	svc := new(mockListingService)
	h := NewMarketplaceHandler(svc, nil)
	r := marketplaceRouter(h)

	listings := []*model.Listing{{ID: "l-1", Images: []string{}}}
	svc.On("Search", mock.Anything, mock.MatchedBy(func(f *model.ListingFilter) bool {
		return f.Query == "ST25"
	})).Return(listings, 1, nil)

	w := doGet(r, "/marketplace/search?q=ST25")
	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "l-1")
}

func TestSearch_ByFilters(t *testing.T) {
	svc := new(mockListingService)
	h := NewMarketplaceHandler(svc, nil)
	r := marketplaceRouter(h)

	svc.On("Search", mock.Anything, mock.MatchedBy(func(f *model.ListingFilter) bool {
		return f.RiceType == "Jasmine" && f.Province == "HCM" &&
			f.MinPrice != nil && *f.MinPrice == 10000 &&
			f.MaxPrice != nil && *f.MaxPrice == 30000
	})).Return([]*model.Listing{}, 0, nil)

	w := doGet(r, "/marketplace/search?type=Jasmine&province=HCM&min_price=10000&max_price=30000")
	assert.Equal(t, 200, w.Code)
}

func TestSearch_WithMinQty(t *testing.T) {
	svc := new(mockListingService)
	h := NewMarketplaceHandler(svc, nil)
	r := marketplaceRouter(h)

	svc.On("Search", mock.Anything, mock.MatchedBy(func(f *model.ListingFilter) bool {
		return f.MinQty != nil && *f.MinQty == 100
	})).Return([]*model.Listing{}, 0, nil)

	w := doGet(r, "/marketplace/search?min_qty=100")
	assert.Equal(t, 200, w.Code)
}

func TestSearch_ServerError(t *testing.T) {
	svc := new(mockListingService)
	h := NewMarketplaceHandler(svc, nil)
	r := marketplaceRouter(h)

	svc.On("Search", mock.Anything, mock.Anything).Return(nil, 0, assert.AnError)

	w := doGet(r, "/marketplace/search?q=test")
	assert.Equal(t, 500, w.Code)
}

// --- GetDetail Tests ---

func TestGetDetail_Success(t *testing.T) {
	svc := new(mockListingService)
	h := NewMarketplaceHandler(svc, nil)
	r := marketplaceRouter(h)

	detail := &model.ListingDetail{
		Listing: model.Listing{ID: "l-1", Title: "Test", Images: []string{}},
		Seller:  &model.PublicProfile{ID: "seller-1", Role: "member"},
	}
	svc.On("GetDetail", mock.Anything, "l-1").Return(detail, nil)

	w := doGet(r, "/marketplace/l-1")
	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "l-1")
	assert.Contains(t, w.Body.String(), "seller-1")
}

func TestGetDetail_NotFound(t *testing.T) {
	svc := new(mockListingService)
	h := NewMarketplaceHandler(svc, nil)
	r := marketplaceRouter(h)

	svc.On("GetDetail", mock.Anything, "bad").Return(nil, repository.ErrListingNotFound)

	w := doGet(r, "/marketplace/bad")
	assert.Equal(t, 404, w.Code)
}
