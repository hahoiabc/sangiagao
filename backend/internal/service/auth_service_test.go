package service

import (
	"context"
	"testing"
	"time"

	"github.com/sangiagao/rice-marketplace/internal/model"
	"github.com/sangiagao/rice-marketplace/internal/repository"
	jwtpkg "github.com/sangiagao/rice-marketplace/pkg/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// --- Mocks ---

type mockUserRepo struct{ mock.Mock }

func (m *mockUserRepo) Create(ctx context.Context, phone, role string) (*model.User, error) {
	args := m.Called(ctx, phone, role)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}
func (m *mockUserRepo) GetByPhone(ctx context.Context, phone string) (*model.User, error) {
	args := m.Called(ctx, phone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}
func (m *mockUserRepo) GetByID(ctx context.Context, id string) (*model.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}
func (m *mockUserRepo) UpdateProfile(ctx context.Context, id string, req *model.UpdateProfileRequest) (*model.User, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}
func (m *mockUserRepo) SetRole(ctx context.Context, id, role string) (*model.User, error) {
	args := m.Called(ctx, id, role)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}
func (m *mockUserRepo) AcceptTOS(ctx context.Context, id string) (*model.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}
func (m *mockUserRepo) UpdateAvatar(ctx context.Context, id, url string) (*model.User, error) {
	args := m.Called(ctx, id, url)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}
func (m *mockUserRepo) GetByIDs(ctx context.Context, ids []string) ([]*model.User, error) {
	args := m.Called(ctx, ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.User), args.Error(1)
}
func (m *mockUserRepo) BlockUser(ctx context.Context, id, reason string) (*model.User, error) {
	args := m.Called(ctx, id, reason)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}
func (m *mockUserRepo) BatchBlock(ctx context.Context, ids []string, reason string) (int, error) {
	args := m.Called(ctx, ids, reason)
	return args.Int(0), args.Error(1)
}
func (m *mockUserRepo) UnblockUser(ctx context.Context, id string) (*model.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}
func (m *mockUserRepo) ListUsers(ctx context.Context, search string, page, limit int) ([]*model.User, int, error) {
	args := m.Called(ctx, search, page, limit)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*model.User), args.Int(1), args.Error(2)
}
func (m *mockUserRepo) GetDashboardStats(ctx context.Context) (map[string]int, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]int), args.Error(1)
}
func (m *mockUserRepo) GetDashboardCharts(ctx context.Context) (*repository.DashboardCharts, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.DashboardCharts), args.Error(1)
}
func (m *mockUserRepo) CreateWithPassword(ctx context.Context, phone, name, passwordHash, province, ward, address string) (*model.User, error) {
	args := m.Called(ctx, phone, name, passwordHash, province, ward, address)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}
func (m *mockUserRepo) GetPasswordHash(ctx context.Context, phone string) (string, error) {
	args := m.Called(ctx, phone)
	return args.String(0), args.Error(1)
}
func (m *mockUserRepo) UpdatePassword(ctx context.Context, phone, passwordHash string) error {
	return m.Called(ctx, phone, passwordHash).Error(0)
}
func (m *mockUserRepo) DeleteUser(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockUserRepo) GetPasswordHashByID(ctx context.Context, userID string) (string, error) {
	args := m.Called(ctx, userID)
	return args.String(0), args.Error(1)
}
func (m *mockUserRepo) UpdatePasswordByID(ctx context.Context, userID, passwordHash string) error {
	return m.Called(ctx, userID, passwordHash).Error(0)
}
func (m *mockUserRepo) UpdatePhone(ctx context.Context, userID, newPhone string) (*model.User, error) {
	args := m.Called(ctx, userID, newPhone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}
func (m *mockUserRepo) PhoneExists(ctx context.Context, phone string) (bool, error) {
	args := m.Called(ctx, phone)
	return args.Bool(0), args.Error(1)
}

type mockOTPRepo struct{ mock.Mock }

func (m *mockOTPRepo) Create(ctx context.Context, phone, code string, expiresAt time.Time) error {
	args := m.Called(ctx, phone, mock.AnythingOfType("string"), mock.AnythingOfType("time.Time"))
	return args.Error(0)
}
func (m *mockOTPRepo) GetLatest(ctx context.Context, phone string) (*repository.OTPRecord, error) {
	args := m.Called(ctx, phone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.OTPRecord), args.Error(1)
}
func (m *mockOTPRepo) IncrementAttempts(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *mockOTPRepo) MarkVerified(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *mockOTPRepo) CountRecent(ctx context.Context, phone string, since time.Time) (int, error) {
	args := m.Called(ctx, phone, mock.AnythingOfType("time.Time"))
	return args.Int(0), args.Error(1)
}

type mockSubRepo struct{ mock.Mock }

func (m *mockSubRepo) Create(ctx context.Context, userID, plan string, days int) (*model.Subscription, error) {
	args := m.Called(ctx, userID, plan, days)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Subscription), args.Error(1)
}
func (m *mockSubRepo) GetActiveByUserID(ctx context.Context, userID string) (*model.Subscription, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Subscription), args.Error(1)
}
func (m *mockSubRepo) GetByUserID(ctx context.Context, userID string) (*model.Subscription, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Subscription), args.Error(1)
}
func (m *mockSubRepo) ExpireOverdue(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}
func (m *mockSubRepo) HideListingsForExpired(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}
func (m *mockSubRepo) ActivateByUserID(ctx context.Context, userID string, days int, durationMonths int, amount int64, plan string) (*model.Subscription, error) {
	args := m.Called(ctx, userID, days, durationMonths, amount, plan)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Subscription), args.Error(1)
}
func (m *mockSubRepo) ExtendSubscription(ctx context.Context, subID string, extraDays int, durationMonths int, amount int64) (*model.Subscription, error) {
	args := m.Called(ctx, subID, extraDays, durationMonths, amount)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Subscription), args.Error(1)
}
func (m *mockSubRepo) RestoreListings(ctx context.Context, userID string) (int, error) {
	args := m.Called(ctx, userID)
	return args.Int(0), args.Error(1)
}
func (m *mockSubRepo) ListByUserID(ctx context.Context, userID string, page, limit int) ([]*model.Subscription, int, error) {
	args := m.Called(ctx, userID, page, limit)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*model.Subscription), args.Int(1), args.Error(2)
}
func (m *mockSubRepo) GetExpiringSoon(ctx context.Context, withinHours int) ([]*model.Subscription, error) {
	args := m.Called(ctx, withinHours)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Subscription), args.Error(1)
}
func (m *mockSubRepo) GetRevenueStats(ctx context.Context) (*repository.SubRevenueStats, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.SubRevenueStats), args.Error(1)
}
func (m *mockSubRepo) GetDailyRevenue(ctx context.Context, from, to string) (*repository.SubDailyRevenueReport, error) {
	args := m.Called(ctx, from, to)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.SubDailyRevenueReport), args.Error(1)
}

type mockSMS struct{ mock.Mock }

func (m *mockSMS) SendOTP(phone, code string) error {
	args := m.Called(phone, code)
	return args.Error(0)
}

// --- Helpers ---

func newTestJWT() *jwtpkg.Manager {
	return jwtpkg.NewManager("test-secret-at-least-32-chars-long", 15*time.Minute, 720*time.Hour)
}

func testUser() *model.User {
	return &model.User{
		ID:    "user-123",
		Phone: "0901234567",
		Role:  "member",
	}
}

// --- SendOTP Tests ---

func TestSendOTP_Success(t *testing.T) {
	userRepo := new(mockUserRepo)
	otpRepo := new(mockOTPRepo)
	subRepo := new(mockSubRepo)
	smsMock := new(mockSMS)
	svc := NewAuthService(userRepo, otpRepo, subRepo, newTestJWT(), smsMock)

	otpRepo.On("CountRecent", mock.Anything, "0901234567", mock.Anything).Return(0, nil)
	otpRepo.On("Create", mock.Anything, "0901234567", mock.Anything, mock.Anything).Return(nil)
	smsMock.On("SendOTP", "0901234567", mock.AnythingOfType("string")).Return(nil)

	err := svc.SendOTP(context.Background(), "0901234567")

	require.NoError(t, err)
	otpRepo.AssertExpectations(t)
	smsMock.AssertExpectations(t)
}

func TestSendOTP_InvalidPhone(t *testing.T) {
	svc := NewAuthService(nil, nil, nil, newTestJWT(), nil)

	tests := []string{"", "123", "090123456", "09012345678", "1901234567", "abc"}
	for _, phone := range tests {
		err := svc.SendOTP(context.Background(), phone)
		assert.ErrorIs(t, err, ErrInvalidPhone, "phone: %s", phone)
	}
}

func TestSendOTP_RateLimited(t *testing.T) {
	otpRepo := new(mockOTPRepo)
	svc := NewAuthService(nil, otpRepo, nil, newTestJWT(), nil)

	otpRepo.On("CountRecent", mock.Anything, "0901234567", mock.Anything).Return(5, nil)

	err := svc.SendOTP(context.Background(), "0901234567")

	assert.ErrorIs(t, err, ErrRateLimited)
}

// --- VerifyOTP Tests ---

func TestVerifyOTP_Success_NewUser(t *testing.T) {
	userRepo := new(mockUserRepo)
	otpRepo := new(mockOTPRepo)
	subRepo := new(mockSubRepo)
	svc := NewAuthService(userRepo, otpRepo, subRepo, newTestJWT(), nil)

	otp := &repository.OTPRecord{
		ID:        "otp-1",
		Phone:     "0901234567",
		Code:      "123456",
		Attempts:  0,
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}
	user := testUser()

	otpRepo.On("GetLatest", mock.Anything, "0901234567").Return(otp, nil)
	otpRepo.On("MarkVerified", mock.Anything, "otp-1").Return(nil)
	userRepo.On("GetByPhone", mock.Anything, "0901234567").Return(nil, repository.ErrUserNotFound)
	userRepo.On("Create", mock.Anything, "0901234567", "member").Return(user, nil)

	result, err := svc.VerifyOTP(context.Background(), "0901234567", "123456")

	require.NoError(t, err)
	assert.True(t, result.IsNewUser)
	assert.Equal(t, "user-123", result.User.ID)
	assert.NotEmpty(t, result.Tokens.AccessToken)
}

func TestVerifyOTP_Success_ExistingUser(t *testing.T) {
	userRepo := new(mockUserRepo)
	otpRepo := new(mockOTPRepo)
	svc := NewAuthService(userRepo, otpRepo, nil, newTestJWT(), nil)

	otp := &repository.OTPRecord{
		ID: "otp-1", Phone: "0901234567", Code: "123456",
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}
	otpRepo.On("GetLatest", mock.Anything, "0901234567").Return(otp, nil)
	otpRepo.On("MarkVerified", mock.Anything, "otp-1").Return(nil)
	userRepo.On("GetByPhone", mock.Anything, "0901234567").Return(testUser(), nil)

	result, err := svc.VerifyOTP(context.Background(), "0901234567", "123456")

	require.NoError(t, err)
	assert.False(t, result.IsNewUser)
}

func TestVerifyOTP_WrongCode(t *testing.T) {
	otpRepo := new(mockOTPRepo)
	svc := NewAuthService(nil, otpRepo, nil, newTestJWT(), nil)

	otp := &repository.OTPRecord{
		ID: "otp-1", Phone: "0901234567", Code: "123456",
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}
	otpRepo.On("GetLatest", mock.Anything, "0901234567").Return(otp, nil)
	otpRepo.On("IncrementAttempts", mock.Anything, "otp-1").Return(nil)

	_, err := svc.VerifyOTP(context.Background(), "0901234567", "999999")

	assert.ErrorIs(t, err, ErrInvalidOTP)
}

func TestVerifyOTP_Expired(t *testing.T) {
	otpRepo := new(mockOTPRepo)
	svc := NewAuthService(nil, otpRepo, nil, newTestJWT(), nil)

	otp := &repository.OTPRecord{
		ID: "otp-1", Code: "123456",
		ExpiresAt: time.Now().Add(-1 * time.Minute),
	}
	otpRepo.On("GetLatest", mock.Anything, "0901234567").Return(otp, nil)

	_, err := svc.VerifyOTP(context.Background(), "0901234567", "123456")

	assert.ErrorIs(t, err, ErrInvalidOTP)
}

func TestVerifyOTP_TooManyAttempts(t *testing.T) {
	otpRepo := new(mockOTPRepo)
	svc := NewAuthService(nil, otpRepo, nil, newTestJWT(), nil)

	otp := &repository.OTPRecord{
		ID: "otp-1", Code: "123456", Attempts: 5,
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}
	otpRepo.On("GetLatest", mock.Anything, "0901234567").Return(otp, nil)

	_, err := svc.VerifyOTP(context.Background(), "0901234567", "123456")

	assert.ErrorIs(t, err, ErrTooManyAttempts)
}

func TestVerifyOTP_BlockedUser(t *testing.T) {
	userRepo := new(mockUserRepo)
	otpRepo := new(mockOTPRepo)
	svc := NewAuthService(userRepo, otpRepo, nil, newTestJWT(), nil)

	otp := &repository.OTPRecord{
		ID: "otp-1", Code: "123456",
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}
	blockedUser := testUser()
	blockedUser.IsBlocked = true

	otpRepo.On("GetLatest", mock.Anything, "0901234567").Return(otp, nil)
	otpRepo.On("MarkVerified", mock.Anything, "otp-1").Return(nil)
	userRepo.On("GetByPhone", mock.Anything, "0901234567").Return(blockedUser, nil)

	_, err := svc.VerifyOTP(context.Background(), "0901234567", "123456")

	assert.ErrorIs(t, err, ErrUserBlocked)
}

func TestVerifyOTP_InvalidPhone(t *testing.T) {
	svc := NewAuthService(nil, nil, nil, newTestJWT(), nil)

	_, err := svc.VerifyOTP(context.Background(), "invalid", "123456")

	assert.ErrorIs(t, err, ErrInvalidPhone)
}

// --- RefreshToken Tests ---

func TestRefreshToken_Success(t *testing.T) {
	userRepo := new(mockUserRepo)
	jm := newTestJWT()
	svc := NewAuthService(userRepo, nil, nil, jm, nil)

	pair, _ := jm.GenerateTokenPair("user-123", "member")
	userRepo.On("GetByID", mock.Anything, "user-123").Return(testUser(), nil)

	tokens, err := svc.RefreshToken(context.Background(), pair.RefreshToken)

	require.NoError(t, err)
	assert.NotEmpty(t, tokens.AccessToken)
}

func TestRefreshToken_InvalidToken(t *testing.T) {
	svc := NewAuthService(nil, nil, nil, newTestJWT(), nil)

	_, err := svc.RefreshToken(context.Background(), "invalid-token")

	assert.ErrorIs(t, err, jwtpkg.ErrInvalidToken)
}

func TestRefreshToken_BlockedUser(t *testing.T) {
	userRepo := new(mockUserRepo)
	jm := newTestJWT()
	svc := NewAuthService(userRepo, nil, nil, jm, nil)

	pair, _ := jm.GenerateTokenPair("user-123", "member")
	blocked := testUser()
	blocked.IsBlocked = true
	userRepo.On("GetByID", mock.Anything, "user-123").Return(blocked, nil)

	_, err := svc.RefreshToken(context.Background(), pair.RefreshToken)

	assert.ErrorIs(t, err, ErrUserBlocked)
}
