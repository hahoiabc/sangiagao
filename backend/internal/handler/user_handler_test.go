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

type mockUserService struct{ mock.Mock }

func (m *mockUserService) GetMe(ctx context.Context, userID string) (*model.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}
func (m *mockUserService) GetPublicProfile(ctx context.Context, userID string) (*model.PublicProfile, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.PublicProfile), args.Error(1)
}
func (m *mockUserService) UpdateProfile(ctx context.Context, userID string, req *model.UpdateProfileRequest) (*model.User, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}
func (m *mockUserService) UpdateAvatar(ctx context.Context, userID, url string) (*model.User, error) {
	args := m.Called(ctx, userID, url)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

// --- Helpers ---

func strPtr(s string) *string { return &s }

func userRouter(h *UserHandler) *gin.Engine {
	r := gin.New()
	// Simulate auth middleware by setting user_id in context
	authed := r.Group("", func(c *gin.Context) {
		c.Set("user_id", "user-123")
		c.Set("user_role", "member")
		c.Next()
	})
	authed.GET("/users/me", h.GetMe)
	authed.PUT("/users/me", h.UpdateMe)
	authed.POST("/users/me/avatar", h.UploadAvatar)
	r.GET("/users/:id/profile", h.GetProfile)
	return r
}

func testUser() *model.User {
	return &model.User{
		ID:    "user-123",
		Phone: "0901234567",
		Role:  "member",
		Name:  strPtr("Test User"),
	}
}

// --- GetMe Tests ---

func TestGetMe_Success(t *testing.T) {
	svc := new(mockUserService)
	h := NewUserHandler(svc)
	r := userRouter(h)

	svc.On("GetMe", mock.Anything, "user-123").Return(testUser(), nil)

	req := httptest.NewRequest("GET", "/users/me", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "user-123")
	assert.Contains(t, w.Body.String(), "Test User")
}

func TestGetMe_Error(t *testing.T) {
	svc := new(mockUserService)
	h := NewUserHandler(svc)
	r := userRouter(h)

	svc.On("GetMe", mock.Anything, "user-123").Return(nil, assert.AnError)

	req := httptest.NewRequest("GET", "/users/me", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 500, w.Code)
	assert.Contains(t, w.Body.String(), "failed to get profile")
}

// --- UpdateMe Tests ---

func TestUpdateMe_Success(t *testing.T) {
	svc := new(mockUserService)
	h := NewUserHandler(svc)
	r := userRouter(h)

	updated := testUser()
	updated.Name = strPtr("New Name")
	svc.On("UpdateProfile", mock.Anything, "user-123", mock.Anything).Return(updated, nil)

	body := `{"name":"New Name"}`
	req := httptest.NewRequest("PUT", "/users/me", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "New Name")
}

func TestUpdateMe_InvalidBody(t *testing.T) {
	h := NewUserHandler(new(mockUserService))
	r := userRouter(h)

	req := httptest.NewRequest("PUT", "/users/me", strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)
}

func TestUpdateMe_InvalidRole(t *testing.T) {
	svc := new(mockUserService)
	h := NewUserHandler(svc)
	r := userRouter(h)

	svc.On("UpdateProfile", mock.Anything, "user-123", mock.Anything).Return(nil, service.ErrInvalidRole)

	body := `{"role":"admin"}`
	req := httptest.NewRequest("PUT", "/users/me", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)
}

func TestUpdateMe_RoleAlreadySet(t *testing.T) {
	svc := new(mockUserService)
	h := NewUserHandler(svc)
	r := userRouter(h)

	svc.On("UpdateProfile", mock.Anything, "user-123", mock.Anything).Return(nil, service.ErrRoleAlreadySet)

	body := `{"role":"member"}`
	req := httptest.NewRequest("PUT", "/users/me", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 409, w.Code)
}

func TestUpdateMe_ServerError(t *testing.T) {
	svc := new(mockUserService)
	h := NewUserHandler(svc)
	r := userRouter(h)

	svc.On("UpdateProfile", mock.Anything, "user-123", mock.Anything).Return(nil, assert.AnError)

	body := `{"name":"X"}`
	req := httptest.NewRequest("PUT", "/users/me", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 500, w.Code)
}

// --- GetProfile Tests ---

func TestGetProfile_Success(t *testing.T) {
	svc := new(mockUserService)
	h := NewUserHandler(svc)
	r := userRouter(h)

	profile := &model.PublicProfile{
		ID:   "user-456",
		Role: "member",
		Name: strPtr("Seller A"),
	}
	svc.On("GetPublicProfile", mock.Anything, "user-456").Return(profile, nil)

	req := httptest.NewRequest("GET", "/users/user-456/profile", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "Seller A")
	assert.NotContains(t, w.Body.String(), "phone") // no phone in public profile
}

func TestGetProfile_NotFound(t *testing.T) {
	svc := new(mockUserService)
	h := NewUserHandler(svc)
	r := userRouter(h)

	svc.On("GetPublicProfile", mock.Anything, "nonexistent").Return(nil, repository.ErrUserNotFound)

	req := httptest.NewRequest("GET", "/users/nonexistent/profile", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 404, w.Code)
	assert.Contains(t, w.Body.String(), "user not found")
}

func TestGetProfile_ServerError(t *testing.T) {
	svc := new(mockUserService)
	h := NewUserHandler(svc)
	r := userRouter(h)

	svc.On("GetPublicProfile", mock.Anything, "user-456").Return(nil, assert.AnError)

	req := httptest.NewRequest("GET", "/users/user-456/profile", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 500, w.Code)
}

// --- UploadAvatar Tests ---

func TestUploadAvatar_Success(t *testing.T) {
	svc := new(mockUserService)
	h := NewUserHandler(svc)
	r := userRouter(h)

	user := testUser()
	user.AvatarURL = strPtr("https://cdn.example.com/avatar.jpg")
	svc.On("UpdateAvatar", mock.Anything, "user-123", "https://cdn.example.com/avatar.jpg").Return(user, nil)

	body := `{"url":"https://cdn.example.com/avatar.jpg"}`
	req := httptest.NewRequest("POST", "/users/me/avatar", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "avatar.jpg")
}

func TestUploadAvatar_MissingURL(t *testing.T) {
	h := NewUserHandler(new(mockUserService))
	r := userRouter(h)

	body := `{}`
	req := httptest.NewRequest("POST", "/users/me/avatar", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)
	assert.Contains(t, w.Body.String(), "url is required")
}

func TestUploadAvatar_ServerError(t *testing.T) {
	svc := new(mockUserService)
	h := NewUserHandler(svc)
	r := userRouter(h)

	svc.On("UpdateAvatar", mock.Anything, "user-123", "https://x.com/a.jpg").Return(nil, assert.AnError)

	body := `{"url":"https://x.com/a.jpg"}`
	req := httptest.NewRequest("POST", "/users/me/avatar", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 500, w.Code)
}

// --- Empty body edge case ---

func TestUpdateMe_EmptyBody(t *testing.T) {
	h := NewUserHandler(new(mockUserService))
	r := userRouter(h)

	req := httptest.NewRequest("PUT", "/users/me", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)
}

func TestUploadAvatar_EmptyBody(t *testing.T) {
	h := NewUserHandler(new(mockUserService))
	r := userRouter(h)

	req := httptest.NewRequest("POST", "/users/me/avatar", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
