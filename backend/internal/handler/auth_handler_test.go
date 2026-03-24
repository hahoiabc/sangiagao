package handler

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sangiagao/rice-marketplace/internal/model"
	"github.com/sangiagao/rice-marketplace/internal/service"
	jwtpkg "github.com/sangiagao/rice-marketplace/pkg/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// --- Mock ---

type mockAuthService struct{ mock.Mock }

func (m *mockAuthService) SendOTP(ctx context.Context, phone string) error {
	args := m.Called(ctx, phone)
	return args.Error(0)
}
func (m *mockAuthService) VerifyOTP(ctx context.Context, phone, code string) (*service.VerifyOTPResult, error) {
	args := m.Called(ctx, phone, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.VerifyOTPResult), args.Error(1)
}
func (m *mockAuthService) RefreshToken(ctx context.Context, token string) (*jwtpkg.TokenPair, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*jwtpkg.TokenPair), args.Error(1)
}

func (m *mockAuthService) CompleteRegister(ctx context.Context, phone, code, name, password, province, ward, address string) (*service.RegisterResult, error) {
	args := m.Called(ctx, phone, code, name, password, province, ward, address)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.RegisterResult), args.Error(1)
}
func (m *mockAuthService) LoginPassword(ctx context.Context, phone, password string) (*service.RegisterResult, error) {
	args := m.Called(ctx, phone, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.RegisterResult), args.Error(1)
}
func (m *mockAuthService) ResetPassword(ctx context.Context, phone, code, newPassword string) error {
	args := m.Called(ctx, phone, code, newPassword)
	return args.Error(0)
}
func (m *mockAuthService) CheckPhoneRegistered(ctx context.Context, phone string) error {
	args := m.Called(ctx, phone)
	return args.Error(0)
}

// --- Helpers ---

func authRouter(h *AuthHandler) *gin.Engine {
	r := gin.New()
	r.POST("/auth/send-otp", h.SendOTP)
	r.POST("/auth/verify-otp", h.VerifyOTP)
	r.POST("/auth/refresh", h.Refresh)
	return r
}

func doPost(r *gin.Engine, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("POST", path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// --- SendOTP Tests ---

func TestSendOTP_Success(t *testing.T) {
	svc := new(mockAuthService)
	h := NewAuthHandler(svc, "localhost", false)
	r := authRouter(h)

	svc.On("SendOTP", mock.Anything, "0901234567").Return(nil)

	w := doPost(r, "/auth/send-otp", `{"phone":"0901234567"}`)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "OTP sent")
	assert.Contains(t, w.Body.String(), "300")
}

func TestSendOTP_MissingPhone(t *testing.T) {
	h := NewAuthHandler(new(mockAuthService), "localhost", false)
	r := authRouter(h)

	w := doPost(r, "/auth/send-otp", `{}`)

	assert.Equal(t, 400, w.Code)
	assert.Contains(t, w.Body.String(), "phone is required")
}

func TestSendOTP_InvalidPhone(t *testing.T) {
	svc := new(mockAuthService)
	h := NewAuthHandler(svc, "localhost", false)
	r := authRouter(h)

	svc.On("SendOTP", mock.Anything, "123").Return(service.ErrInvalidPhone)

	w := doPost(r, "/auth/send-otp", `{"phone":"123"}`)

	assert.Equal(t, 400, w.Code)
	assert.Contains(t, w.Body.String(), "invalid phone number format")
}

func TestSendOTP_RateLimited(t *testing.T) {
	svc := new(mockAuthService)
	h := NewAuthHandler(svc, "localhost", false)
	r := authRouter(h)

	svc.On("SendOTP", mock.Anything, "0901234567").Return(service.ErrRateLimited)

	w := doPost(r, "/auth/send-otp", `{"phone":"0901234567"}`)

	assert.Equal(t, 429, w.Code)
}

// --- VerifyOTP Tests ---

func TestVerifyOTP_Success(t *testing.T) {
	svc := new(mockAuthService)
	h := NewAuthHandler(svc, "localhost", false)
	r := authRouter(h)

	result := &service.VerifyOTPResult{
		User:      &model.User{ID: "user-123", Phone: "0901234567", Role: "member"},
		Tokens:    &jwtpkg.TokenPair{AccessToken: "at", RefreshToken: "rt", ExpiresIn: 900},
		IsNewUser: true,
	}
	svc.On("VerifyOTP", mock.Anything, "0901234567", "123456").Return(result, nil)

	w := doPost(r, "/auth/verify-otp", `{"phone":"0901234567","code":"123456"}`)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "user-123")
	assert.Contains(t, w.Body.String(), "is_new_user")
}

func TestVerifyOTP_MissingFields(t *testing.T) {
	h := NewAuthHandler(new(mockAuthService), "localhost", false)
	r := authRouter(h)

	w := doPost(r, "/auth/verify-otp", `{"phone":"0901234567"}`)
	assert.Equal(t, 400, w.Code)

	w = doPost(r, "/auth/verify-otp", `{"code":"123456"}`)
	assert.Equal(t, 400, w.Code)
}

func TestVerifyOTP_InvalidOTP(t *testing.T) {
	svc := new(mockAuthService)
	h := NewAuthHandler(svc, "localhost", false)
	r := authRouter(h)

	svc.On("VerifyOTP", mock.Anything, "0901234567", "999999").Return(nil, service.ErrInvalidOTP)

	w := doPost(r, "/auth/verify-otp", `{"phone":"0901234567","code":"999999"}`)

	assert.Equal(t, 401, w.Code)
	assert.Contains(t, w.Body.String(), "invalid or expired OTP")
}

func TestVerifyOTP_TooManyAttempts(t *testing.T) {
	svc := new(mockAuthService)
	h := NewAuthHandler(svc, "localhost", false)
	r := authRouter(h)

	svc.On("VerifyOTP", mock.Anything, mock.Anything, mock.Anything).Return(nil, service.ErrTooManyAttempts)

	w := doPost(r, "/auth/verify-otp", `{"phone":"0901234567","code":"123456"}`)

	assert.Equal(t, 429, w.Code)
}

func TestVerifyOTP_BlockedUser(t *testing.T) {
	svc := new(mockAuthService)
	h := NewAuthHandler(svc, "localhost", false)
	r := authRouter(h)

	svc.On("VerifyOTP", mock.Anything, mock.Anything, mock.Anything).Return(nil, service.ErrUserBlocked)

	w := doPost(r, "/auth/verify-otp", `{"phone":"0901234567","code":"123456"}`)

	assert.Equal(t, 403, w.Code)
	assert.Contains(t, w.Body.String(), "account is blocked")
}

// --- Refresh Tests ---

func TestRefresh_Success(t *testing.T) {
	svc := new(mockAuthService)
	h := NewAuthHandler(svc, "localhost", false)
	r := authRouter(h)

	tokens := &jwtpkg.TokenPair{AccessToken: "new-at", RefreshToken: "new-rt", ExpiresIn: 900}
	svc.On("RefreshToken", mock.Anything, "valid-rt").Return(tokens, nil)

	w := doPost(r, "/auth/refresh", `{"refresh_token":"valid-rt"}`)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "new-at")
}

func TestRefresh_MissingToken(t *testing.T) {
	h := NewAuthHandler(new(mockAuthService), "localhost", false)
	r := authRouter(h)

	w := doPost(r, "/auth/refresh", `{}`)

	assert.Equal(t, 400, w.Code)
}

func TestRefresh_InvalidToken(t *testing.T) {
	svc := new(mockAuthService)
	h := NewAuthHandler(svc, "localhost", false)
	r := authRouter(h)

	svc.On("RefreshToken", mock.Anything, "bad-token").Return(nil, jwtpkg.ErrInvalidToken)

	w := doPost(r, "/auth/refresh", `{"refresh_token":"bad-token"}`)

	assert.Equal(t, 401, w.Code)
}

func TestRefresh_BlockedUser(t *testing.T) {
	svc := new(mockAuthService)
	h := NewAuthHandler(svc, "localhost", false)
	r := authRouter(h)

	svc.On("RefreshToken", mock.Anything, "rt").Return(nil, service.ErrUserBlocked)

	w := doPost(r, "/auth/refresh", `{"refresh_token":"rt"}`)

	assert.Equal(t, 403, w.Code)
}

// --- Invalid JSON ---

func TestSendOTP_InvalidJSON(t *testing.T) {
	h := NewAuthHandler(new(mockAuthService), "localhost", false)
	r := authRouter(h)

	w := doPost(r, "/auth/send-otp", `not json`)

	assert.Equal(t, 400, w.Code)
}

// --- Edge: empty body ---

func TestVerifyOTP_EmptyBody(t *testing.T) {
	h := NewAuthHandler(new(mockAuthService), "localhost", false)
	r := authRouter(h)

	req := httptest.NewRequest("POST", "/auth/verify-otp", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)
}

// --- Timing: VerifyOTP with InvalidPhone ---

func TestVerifyOTP_InvalidPhone(t *testing.T) {
	svc := new(mockAuthService)
	h := NewAuthHandler(svc, "localhost", false)
	r := authRouter(h)

	svc.On("VerifyOTP", mock.Anything, "bad", "123456").Return(nil, service.ErrInvalidPhone)

	w := doPost(r, "/auth/verify-otp", `{"phone":"bad","code":"123456"}`)

	assert.Equal(t, 400, w.Code)
}

// --- Refresh: ServerError ---

func TestRefresh_ServerError(t *testing.T) {
	svc := new(mockAuthService)
	h := NewAuthHandler(svc, "localhost", false)
	r := authRouter(h)

	svc.On("RefreshToken", mock.Anything, "rt").Return(nil, assert.AnError)

	w := doPost(r, "/auth/refresh", `{"refresh_token":"rt"}`)

	assert.Equal(t, 500, w.Code)
	assert.Contains(t, w.Body.String(), "refresh failed")
}

// Ensure expired token uses the same manager for consistency
func TestVerifyOTP_InternalError(t *testing.T) {
	svc := new(mockAuthService)
	h := NewAuthHandler(svc, "localhost", false)
	r := authRouter(h)

	svc.On("VerifyOTP", mock.Anything, mock.Anything, mock.Anything).Return(nil, assert.AnError)

	w := doPost(r, "/auth/verify-otp", `{"phone":"0901234567","code":"123456"}`)

	assert.Equal(t, 500, w.Code)
	assert.Contains(t, w.Body.String(), "verification failed")
}
