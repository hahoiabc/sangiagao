package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sangiagao/rice-marketplace/internal/model"
	"github.com/sangiagao/rice-marketplace/internal/repository"
	"github.com/sangiagao/rice-marketplace/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Mock ---

type mockListingService struct{ mock.Mock }

func (m *mockListingService) Create(ctx context.Context, userID string, req *model.CreateListingRequest) (*model.Listing, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Listing), args.Error(1)
}

func (m *mockListingService) GetByID(ctx context.Context, id string) (*model.Listing, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Listing), args.Error(1)
}

func (m *mockListingService) Update(ctx context.Context, userID, id string, req *model.UpdateListingRequest) (*model.Listing, error) {
	args := m.Called(ctx, userID, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Listing), args.Error(1)
}

func (m *mockListingService) Delete(ctx context.Context, userID, id string) error {
	return m.Called(ctx, userID, id).Error(0)
}

func (m *mockListingService) ListByUser(ctx context.Context, userID string, page, limit int) ([]*model.Listing, int, error) {
	args := m.Called(ctx, userID, page, limit)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*model.Listing), args.Int(1), args.Error(2)
}

func (m *mockListingService) AddImage(ctx context.Context, userID, id, imageURL string) (*model.Listing, error) {
	args := m.Called(ctx, userID, id, imageURL)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Listing), args.Error(1)
}

func (m *mockListingService) RemoveImage(ctx context.Context, userID, id, imageURL string) (*model.Listing, error) {
	args := m.Called(ctx, userID, id, imageURL)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Listing), args.Error(1)
}

func (m *mockListingService) Browse(ctx context.Context, page, limit int) ([]*model.Listing, int, error) {
	args := m.Called(ctx, page, limit)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*model.Listing), args.Int(1), args.Error(2)
}

func (m *mockListingService) Search(ctx context.Context, filter *model.ListingFilter) ([]*model.Listing, int, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*model.Listing), args.Int(1), args.Error(2)
}

func (m *mockListingService) GetDetail(ctx context.Context, id string) (*model.ListingDetail, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ListingDetail), args.Error(1)
}

func (m *mockListingService) GetPriceBoard(ctx context.Context) (*model.PriceBoardResponse, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.PriceBoardResponse), args.Error(1)
}
func (m *mockListingService) BatchDeleteOwn(ctx context.Context, userID string, ids []string) (int, error) {
	args := m.Called(ctx, userID, ids)
	return args.Int(0), args.Error(1)
}

// --- Helpers ---

func listingRouter(h *ListingHandler) *gin.Engine {
	r := gin.New()
	g := r.Group("")
	g.Use(func(c *gin.Context) {
		c.Set("user_id", "user-1")
		c.Next()
	})
	g.POST("/listings", h.Create)
	g.GET("/listings/my", h.ListMy)
	g.GET("/listings/:id", h.Get)
	g.PUT("/listings/:id", h.Update)
	g.DELETE("/listings/:id", h.Delete)
	g.POST("/listings/:id/images", h.AddImage)
	g.DELETE("/listings/:id/images", h.RemoveImage)
	return r
}

func doRequest(r *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// --- Create Tests ---

func TestListingCreate_Success(t *testing.T) {
	svc := new(mockListingService)
	h := NewListingHandler(svc)
	r := listingRouter(h)

	listing := &model.Listing{ID: "l-1", Title: "Test", Images: []string{}}
	svc.On("Create", mock.Anything, "user-1", mock.AnythingOfType("*model.CreateListingRequest")).Return(listing, nil)

	w := doRequest(r, "POST", "/listings", `{"title":"Test","category":"gao_deo_thom","rice_type":"st_25","province":"HCM","quantity_kg":10,"price_per_kg":1000}`)

	assert.Equal(t, 201, w.Code)
	assert.Contains(t, w.Body.String(), "l-1")
}

func TestListingCreate_InvalidBody(t *testing.T) {
	h := NewListingHandler(new(mockListingService))
	r := listingRouter(h)

	w := doRequest(r, "POST", "/listings", `{"title":"Test"}`)
	assert.Equal(t, 400, w.Code)
}

func TestListingCreate_ServerError(t *testing.T) {
	svc := new(mockListingService)
	h := NewListingHandler(svc)
	r := listingRouter(h)

	svc.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(nil, assert.AnError)

	w := doRequest(r, "POST", "/listings", `{"title":"Test","category":"gao_deo_thom","rice_type":"st_25","province":"HCM","quantity_kg":10,"price_per_kg":1000}`)
	assert.Equal(t, 500, w.Code)
}

// --- Get Tests ---

func TestListingGet_Success(t *testing.T) {
	svc := new(mockListingService)
	h := NewListingHandler(svc)
	r := listingRouter(h)

	listing := &model.Listing{ID: "l-1", Title: "Test", Images: []string{}}
	svc.On("GetByID", mock.Anything, "l-1").Return(listing, nil)

	w := doRequest(r, "GET", "/listings/l-1", "")
	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "l-1")
}

func TestListingGet_NotFound(t *testing.T) {
	svc := new(mockListingService)
	h := NewListingHandler(svc)
	r := listingRouter(h)

	svc.On("GetByID", mock.Anything, "bad").Return(nil, repository.ErrListingNotFound)

	w := doRequest(r, "GET", "/listings/bad", "")
	assert.Equal(t, 404, w.Code)
}

func TestListingGet_ServerError(t *testing.T) {
	svc := new(mockListingService)
	h := NewListingHandler(svc)
	r := listingRouter(h)

	svc.On("GetByID", mock.Anything, "l-1").Return(nil, assert.AnError)

	w := doRequest(r, "GET", "/listings/l-1", "")
	assert.Equal(t, 500, w.Code)
}

// --- Update Tests ---

func TestListingUpdate_Success(t *testing.T) {
	svc := new(mockListingService)
	h := NewListingHandler(svc)
	r := listingRouter(h)

	listing := &model.Listing{ID: "l-1", PricePerKG: 25000, Images: []string{}}
	svc.On("Update", mock.Anything, "user-1", "l-1", mock.Anything).Return(listing, nil)

	w := doRequest(r, "PUT", "/listings/l-1", `{"price_per_kg":25000}`)
	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "25000")
}

func TestListingUpdate_InvalidBody(t *testing.T) {
	h := NewListingHandler(new(mockListingService))
	r := listingRouter(h)

	w := doRequest(r, "PUT", "/listings/l-1", "not json")
	assert.Equal(t, 400, w.Code)
}

func TestListingUpdate_NotFound(t *testing.T) {
	svc := new(mockListingService)
	h := NewListingHandler(svc)
	r := listingRouter(h)

	svc.On("Update", mock.Anything, "user-1", "l-1", mock.Anything).Return(nil, repository.ErrListingNotFound)

	w := doRequest(r, "PUT", "/listings/l-1", `{"price_per_kg":25000}`)
	assert.Equal(t, 404, w.Code)
}

func TestListingUpdate_NotOwner(t *testing.T) {
	svc := new(mockListingService)
	h := NewListingHandler(svc)
	r := listingRouter(h)

	svc.On("Update", mock.Anything, "user-1", "l-1", mock.Anything).Return(nil, service.ErrNotListingOwner)

	w := doRequest(r, "PUT", "/listings/l-1", `{"price_per_kg":25000}`)
	assert.Equal(t, 403, w.Code)
}

func TestListingUpdate_ServerError(t *testing.T) {
	svc := new(mockListingService)
	h := NewListingHandler(svc)
	r := listingRouter(h)

	svc.On("Update", mock.Anything, "user-1", "l-1", mock.Anything).Return(nil, assert.AnError)

	w := doRequest(r, "PUT", "/listings/l-1", `{"price_per_kg":25000}`)
	assert.Equal(t, 500, w.Code)
}

// --- Delete Tests ---

func TestListingDelete_Success(t *testing.T) {
	svc := new(mockListingService)
	h := NewListingHandler(svc)
	r := listingRouter(h)

	svc.On("Delete", mock.Anything, "user-1", "l-1").Return(nil)

	w := doRequest(r, "DELETE", "/listings/l-1", "")
	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "listing deleted")
}

func TestListingDelete_NotFound(t *testing.T) {
	svc := new(mockListingService)
	h := NewListingHandler(svc)
	r := listingRouter(h)

	svc.On("Delete", mock.Anything, "user-1", "l-1").Return(repository.ErrListingNotFound)

	w := doRequest(r, "DELETE", "/listings/l-1", "")
	assert.Equal(t, 404, w.Code)
}

func TestListingDelete_NotOwner(t *testing.T) {
	svc := new(mockListingService)
	h := NewListingHandler(svc)
	r := listingRouter(h)

	svc.On("Delete", mock.Anything, "user-1", "l-1").Return(service.ErrNotListingOwner)

	w := doRequest(r, "DELETE", "/listings/l-1", "")
	assert.Equal(t, 403, w.Code)
}

// --- ListMy Tests ---

func TestListMy_Success(t *testing.T) {
	svc := new(mockListingService)
	h := NewListingHandler(svc)
	r := listingRouter(h)

	listings := []*model.Listing{{ID: "l-1", Images: []string{}}}
	svc.On("ListByUser", mock.Anything, "user-1", 1, 20).Return(listings, 1, nil)

	w := doRequest(r, "GET", "/listings/my", "")
	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "l-1")
	assert.Contains(t, w.Body.String(), `"total":1`)
}

func TestListMy_WithPagination(t *testing.T) {
	svc := new(mockListingService)
	h := NewListingHandler(svc)
	r := listingRouter(h)

	svc.On("ListByUser", mock.Anything, "user-1", 2, 10).Return([]*model.Listing{}, 0, nil)

	w := doRequest(r, "GET", "/listings/my?page=2&limit=10", "")
	assert.Equal(t, 200, w.Code)
}

func TestListMy_ServerError(t *testing.T) {
	svc := new(mockListingService)
	h := NewListingHandler(svc)
	r := listingRouter(h)

	svc.On("ListByUser", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, 0, assert.AnError)

	w := doRequest(r, "GET", "/listings/my", "")
	assert.Equal(t, 500, w.Code)
}

// --- AddImage Tests ---

func TestAddImage_Success(t *testing.T) {
	svc := new(mockListingService)
	h := NewListingHandler(svc)
	r := listingRouter(h)

	listing := &model.Listing{ID: "l-1", Images: []string{"img1.jpg", "img2.jpg"}}
	svc.On("AddImage", mock.Anything, "user-1", "l-1", "img2.jpg").Return(listing, nil)

	w := doRequest(r, "POST", "/listings/l-1/images", `{"url":"img2.jpg"}`)
	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "img2.jpg")
}

func TestAddImage_MissingURL(t *testing.T) {
	h := NewListingHandler(new(mockListingService))
	r := listingRouter(h)

	w := doRequest(r, "POST", "/listings/l-1/images", `{}`)
	assert.Equal(t, 400, w.Code)
}

func TestAddImage_MaxImages(t *testing.T) {
	svc := new(mockListingService)
	h := NewListingHandler(svc)
	r := listingRouter(h)

	svc.On("AddImage", mock.Anything, "user-1", "l-1", "img4.jpg").Return(nil, service.ErrMaxImages)

	w := doRequest(r, "POST", "/listings/l-1/images", `{"url":"img4.jpg"}`)
	assert.Equal(t, 409, w.Code)
	assert.Contains(t, w.Body.String(), "Tối đa")
}

func TestAddImage_NotOwner(t *testing.T) {
	svc := new(mockListingService)
	h := NewListingHandler(svc)
	r := listingRouter(h)

	svc.On("AddImage", mock.Anything, "user-1", "l-1", "img.jpg").Return(nil, service.ErrNotListingOwner)

	w := doRequest(r, "POST", "/listings/l-1/images", `{"url":"img.jpg"}`)
	assert.Equal(t, 403, w.Code)
}

func TestAddImage_NotFound(t *testing.T) {
	svc := new(mockListingService)
	h := NewListingHandler(svc)
	r := listingRouter(h)

	svc.On("AddImage", mock.Anything, "user-1", "l-1", "img.jpg").Return(nil, repository.ErrListingNotFound)

	w := doRequest(r, "POST", "/listings/l-1/images", `{"url":"img.jpg"}`)
	assert.Equal(t, 404, w.Code)
}

// --- RemoveImage Tests ---

func TestRemoveImage_Success(t *testing.T) {
	svc := new(mockListingService)
	h := NewListingHandler(svc)
	r := listingRouter(h)

	listing := &model.Listing{ID: "l-1", Images: []string{"img1.jpg"}}
	svc.On("RemoveImage", mock.Anything, "user-1", "l-1", "img2.jpg").Return(listing, nil)

	w := doRequest(r, "DELETE", "/listings/l-1/images", `{"url":"img2.jpg"}`)
	assert.Equal(t, 200, w.Code)
}

func TestRemoveImage_MissingURL(t *testing.T) {
	svc := new(mockListingService)
	h := NewListingHandler(svc)
	r := listingRouter(h)

	w := doRequest(r, "DELETE", "/listings/l-1/images", `{}`)
	assert.Equal(t, 400, w.Code)
}

func TestRemoveImage_NotOwner(t *testing.T) {
	svc := new(mockListingService)
	h := NewListingHandler(svc)
	r := listingRouter(h)

	svc.On("RemoveImage", mock.Anything, "user-1", "l-1", "img.jpg").Return(nil, service.ErrNotListingOwner)

	w := doRequest(r, "DELETE", "/listings/l-1/images", `{"url":"img.jpg"}`)
	assert.Equal(t, 403, w.Code)
}

func TestRemoveImage_NotFound(t *testing.T) {
	svc := new(mockListingService)
	h := NewListingHandler(svc)
	r := listingRouter(h)

	svc.On("RemoveImage", mock.Anything, "user-1", "l-1", "img.jpg").Return(nil, repository.ErrListingNotFound)

	w := doRequest(r, "DELETE", "/listings/l-1/images", `{"url":"img.jpg"}`)
	assert.Equal(t, 404, w.Code)
}
