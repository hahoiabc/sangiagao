package service

import (
	"context"
	"testing"

	"github.com/sangiagao/rice-marketplace/internal/model"
	"github.com/sangiagao/rice-marketplace/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockAdminUserRepo implements UserRepository for admin tests
type mockAdminUserRepo struct{ mock.Mock }

func (m *mockAdminUserRepo) Create(ctx context.Context, phone, role string) (*model.User, error) {
	args := m.Called(ctx, phone, role)
	return args.Get(0).(*model.User), args.Error(1)
}
func (m *mockAdminUserRepo) GetByPhone(ctx context.Context, phone string) (*model.User, error) {
	args := m.Called(ctx, phone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}
func (m *mockAdminUserRepo) CreateWithPassword(ctx context.Context, phone, name, passwordHash, province, ward, address string) (*model.User, error) {
	args := m.Called(ctx, phone, name, passwordHash, province, ward, address)
	return args.Get(0).(*model.User), args.Error(1)
}
func (m *mockAdminUserRepo) GetPasswordHash(ctx context.Context, phone string) (string, error) {
	args := m.Called(ctx, phone)
	return args.String(0), args.Error(1)
}
func (m *mockAdminUserRepo) UpdatePassword(ctx context.Context, phone, passwordHash string) error {
	args := m.Called(ctx, phone, passwordHash)
	return args.Error(0)
}
func (m *mockAdminUserRepo) GetByID(ctx context.Context, id string) (*model.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}
func (m *mockAdminUserRepo) UpdateProfile(ctx context.Context, id string, req *model.UpdateProfileRequest) (*model.User, error) {
	args := m.Called(ctx, id, req)
	return args.Get(0).(*model.User), args.Error(1)
}
func (m *mockAdminUserRepo) SetRole(ctx context.Context, id, role string) (*model.User, error) {
	args := m.Called(ctx, id, role)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}
func (m *mockAdminUserRepo) AcceptTOS(ctx context.Context, id string) (*model.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*model.User), args.Error(1)
}
func (m *mockAdminUserRepo) UpdateAvatar(ctx context.Context, id, avatarURL string) (*model.User, error) {
	args := m.Called(ctx, id, avatarURL)
	return args.Get(0).(*model.User), args.Error(1)
}
func (m *mockAdminUserRepo) GetByIDs(ctx context.Context, ids []string) ([]*model.User, error) {
	args := m.Called(ctx, ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.User), args.Error(1)
}
func (m *mockAdminUserRepo) BlockUser(ctx context.Context, id, reason string) (*model.User, error) {
	args := m.Called(ctx, id, reason)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}
func (m *mockAdminUserRepo) BatchBlock(ctx context.Context, ids []string, reason string) (int, error) {
	args := m.Called(ctx, ids, reason)
	return args.Int(0), args.Error(1)
}
func (m *mockAdminUserRepo) UnblockUser(ctx context.Context, id string) (*model.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}
func (m *mockAdminUserRepo) ListUsers(ctx context.Context, search string, page, limit int) ([]*model.User, int, error) {
	args := m.Called(ctx, search, page, limit)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*model.User), args.Int(1), args.Error(2)
}
func (m *mockAdminUserRepo) GetDashboardStats(ctx context.Context) (map[string]int, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]int), args.Error(1)
}
func (m *mockAdminUserRepo) GetDashboardCharts(ctx context.Context) (*repository.DashboardCharts, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.DashboardCharts), args.Error(1)
}
func (m *mockAdminUserRepo) DeleteUser(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *mockAdminUserRepo) GetPasswordHashByID(ctx context.Context, userID string) (string, error) {
	args := m.Called(ctx, userID)
	return args.String(0), args.Error(1)
}
func (m *mockAdminUserRepo) UpdatePasswordByID(ctx context.Context, userID, passwordHash string) error {
	args := m.Called(ctx, userID, passwordHash)
	return args.Error(0)
}
func (m *mockAdminUserRepo) UpdatePhone(ctx context.Context, userID, newPhone string) (*model.User, error) {
	args := m.Called(ctx, userID, newPhone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}
func (m *mockAdminUserRepo) PhoneExists(ctx context.Context, phone string) (bool, error) {
	args := m.Called(ctx, phone)
	return args.Bool(0), args.Error(1)
}

// mockAdminListingRepo for admin listing operations
type mockAdminListingRepo struct{ mock.Mock }

func (m *mockAdminListingRepo) CountTodayByUser(ctx context.Context, userID string) (int, error) {
	args := m.Called(ctx, userID)
	return args.Int(0), args.Error(1)
}
func (m *mockAdminListingRepo) Create(ctx context.Context, userID string, req *model.CreateListingRequest) (*model.Listing, error) {
	args := m.Called(ctx, userID, req)
	return args.Get(0).(*model.Listing), args.Error(1)
}
func (m *mockAdminListingRepo) GetByID(ctx context.Context, id string) (*model.Listing, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*model.Listing), args.Error(1)
}
func (m *mockAdminListingRepo) Update(ctx context.Context, id string, req *model.UpdateListingRequest) (*model.Listing, error) {
	args := m.Called(ctx, id, req)
	return args.Get(0).(*model.Listing), args.Error(1)
}
func (m *mockAdminListingRepo) SoftDelete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *mockAdminListingRepo) BatchSoftDelete(ctx context.Context, ids []string) (int, error) {
	args := m.Called(ctx, ids)
	return args.Int(0), args.Error(1)
}
func (m *mockAdminListingRepo) ListByUser(ctx context.Context, userID string, page, limit int) ([]*model.Listing, int, error) {
	args := m.Called(ctx, userID, page, limit)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*model.Listing), args.Int(1), args.Error(2)
}
func (m *mockAdminListingRepo) AddImage(ctx context.Context, id, imageURL string) (*model.Listing, error) {
	args := m.Called(ctx, id, imageURL)
	return args.Get(0).(*model.Listing), args.Error(1)
}
func (m *mockAdminListingRepo) GetImageCount(ctx context.Context, id string) (int, error) {
	args := m.Called(ctx, id)
	return args.Int(0), args.Error(1)
}
func (m *mockAdminListingRepo) Browse(ctx context.Context, page, limit int) ([]*model.Listing, int, error) {
	args := m.Called(ctx, page, limit)
	return args.Get(0).([]*model.Listing), args.Int(1), args.Error(2)
}
func (m *mockAdminListingRepo) Search(ctx context.Context, filter *model.ListingFilter) ([]*model.Listing, int, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*model.Listing), args.Int(1), args.Error(2)
}
func (m *mockAdminListingRepo) GetDetailWithSeller(ctx context.Context, id string) (*model.ListingDetail, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*model.ListingDetail), args.Error(1)
}
func (m *mockAdminListingRepo) IncrementViewCount(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *mockAdminListingRepo) GetPriceBoardData(ctx context.Context) ([]repository.PriceBoardRow, error) {
	args := m.Called(ctx)
	return args.Get(0).([]repository.PriceBoardRow), args.Error(1)
}

// mockAdminSubRepo for admin subscription operations
type mockAdminSubRepo struct{ mock.Mock }

func (m *mockAdminSubRepo) Create(ctx context.Context, userID, plan string, daysValid int) (*model.Subscription, error) {
	args := m.Called(ctx, userID, plan, daysValid)
	return args.Get(0).(*model.Subscription), args.Error(1)
}
func (m *mockAdminSubRepo) GetActiveByUserID(ctx context.Context, userID string) (*model.Subscription, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*model.Subscription), args.Error(1)
}
func (m *mockAdminSubRepo) GetByUserID(ctx context.Context, userID string) (*model.Subscription, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*model.Subscription), args.Error(1)
}
func (m *mockAdminSubRepo) ExpireOverdue(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}
func (m *mockAdminSubRepo) HideListingsForExpired(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}
func (m *mockAdminSubRepo) ActivateByUserID(ctx context.Context, userID string, daysValid int, durationMonths int, amount int64) (*model.Subscription, error) {
	args := m.Called(ctx, userID, daysValid, durationMonths, amount)
	return args.Get(0).(*model.Subscription), args.Error(1)
}
func (m *mockAdminSubRepo) ExtendSubscription(ctx context.Context, subID string, extraDays int, durationMonths int, amount int64) (*model.Subscription, error) {
	args := m.Called(ctx, subID, extraDays, durationMonths, amount)
	return args.Get(0).(*model.Subscription), args.Error(1)
}
func (m *mockAdminSubRepo) RestoreListings(ctx context.Context, userID string) (int, error) {
	args := m.Called(ctx, userID)
	return args.Int(0), args.Error(1)
}
func (m *mockAdminSubRepo) ListByUserID(ctx context.Context, userID string, page, limit int) ([]*model.Subscription, int, error) {
	args := m.Called(ctx, userID, page, limit)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*model.Subscription), args.Int(1), args.Error(2)
}
func (m *mockAdminSubRepo) GetExpiringSoon(ctx context.Context, withinHours int) ([]*model.Subscription, error) {
	args := m.Called(ctx, withinHours)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Subscription), args.Error(1)
}
func (m *mockAdminSubRepo) GetRevenueStats(ctx context.Context) (*repository.SubRevenueStats, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.SubRevenueStats), args.Error(1)
}
func (m *mockAdminSubRepo) GetDailyRevenue(ctx context.Context, from, to string) (*repository.SubDailyRevenueReport, error) {
	args := m.Called(ctx, from, to)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.SubDailyRevenueReport), args.Error(1)
}

func TestAdminGetDashboardStats_Success(t *testing.T) {
	userRepo := new(mockAdminUserRepo)
	svc := NewAdminService(userRepo, new(mockAdminListingRepo), new(mockAdminSubRepo))

	stats := map[string]int{"total_users": 100, "total_listings": 50}
	userRepo.On("GetDashboardStats", mock.Anything).Return(stats, nil)

	result, err := svc.GetDashboardStats(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 100, result["total_users"])
	assert.Equal(t, 50, result["total_listings"])
}

func TestAdminListUsers_Success(t *testing.T) {
	userRepo := new(mockAdminUserRepo)
	svc := NewAdminService(userRepo, new(mockAdminListingRepo), new(mockAdminSubRepo))

	users := []*model.User{{ID: "u-1", Name: strPtr("Nguyen Van A")}}
	userRepo.On("ListUsers", mock.Anything, "nguyen", 1, 20).Return(users, 1, nil)

	result, total, err := svc.ListUsers(context.Background(), "nguyen", 1, 20)
	assert.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, result, 1)
}

func TestAdminListUsers_DefaultsInvalidPage(t *testing.T) {
	userRepo := new(mockAdminUserRepo)
	svc := NewAdminService(userRepo, new(mockAdminListingRepo), new(mockAdminSubRepo))

	users := []*model.User{{ID: "u-1"}}
	userRepo.On("ListUsers", mock.Anything, "", 1, 20).Return(users, 1, nil)

	result, total, err := svc.ListUsers(context.Background(), "", 0, 0)
	assert.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, result, 1)
}

func TestAdminBlockUser_Success(t *testing.T) {
	userRepo := new(mockAdminUserRepo)
	svc := NewAdminService(userRepo, new(mockAdminListingRepo), new(mockAdminSubRepo))

	userRepo.On("GetByID", mock.Anything, "u-1").Return(
		&model.User{ID: "u-1", Role: "member"}, nil)
	userRepo.On("BlockUser", mock.Anything, "u-1", "vi phạm").Return(
		&model.User{ID: "u-1", IsBlocked: true}, nil)

	user, err := svc.BlockUser(context.Background(), "u-1", "vi phạm", "admin")
	assert.NoError(t, err)
	assert.True(t, user.IsBlocked)
}

func TestAdminBlockUser_CannotBlockAdmin(t *testing.T) {
	userRepo := new(mockAdminUserRepo)
	svc := NewAdminService(userRepo, new(mockAdminListingRepo), new(mockAdminSubRepo))

	userRepo.On("GetByID", mock.Anything, "u-admin").Return(
		&model.User{ID: "u-admin", Role: "admin"}, nil)

	user, err := svc.BlockUser(context.Background(), "u-admin", "test", "admin")
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, ErrCannotModifyAdmin, err)
}

func TestAdminBlockUser_OwnerCanBlockAdmin(t *testing.T) {
	userRepo := new(mockAdminUserRepo)
	svc := NewAdminService(userRepo, new(mockAdminListingRepo), new(mockAdminSubRepo))

	userRepo.On("GetByID", mock.Anything, "u-admin").Return(
		&model.User{ID: "u-admin", Role: "admin"}, nil)
	userRepo.On("BlockUser", mock.Anything, "u-admin", "test").Return(
		&model.User{ID: "u-admin", IsBlocked: true}, nil)

	user, err := svc.BlockUser(context.Background(), "u-admin", "test", "owner")
	assert.NoError(t, err)
	assert.True(t, user.IsBlocked)
}

func TestAdminUnblockUser_Success(t *testing.T) {
	userRepo := new(mockAdminUserRepo)
	svc := NewAdminService(userRepo, new(mockAdminListingRepo), new(mockAdminSubRepo))

	userRepo.On("GetByID", mock.Anything, "u-1").Return(
		&model.User{ID: "u-1", Role: "member", IsBlocked: true}, nil)
	userRepo.On("UnblockUser", mock.Anything, "u-1").Return(
		&model.User{ID: "u-1", IsBlocked: false}, nil)

	user, err := svc.UnblockUser(context.Background(), "u-1", "admin")
	assert.NoError(t, err)
	assert.False(t, user.IsBlocked)
}

func TestAdminChangeUserRole_Success(t *testing.T) {
	userRepo := new(mockAdminUserRepo)
	svc := NewAdminService(userRepo, new(mockAdminListingRepo), new(mockAdminSubRepo))

	userRepo.On("GetByID", mock.Anything, "u-1").Return(
		&model.User{ID: "u-1", Role: "member"}, nil)
	userRepo.On("SetRole", mock.Anything, "u-1", "member").Return(
		&model.User{ID: "u-1", Role: "member"}, nil)

	user, err := svc.ChangeUserRole(context.Background(), "u-1", "member", "admin")
	assert.NoError(t, err)
	assert.Equal(t, "member", user.Role)
}

func TestAdminChangeUserRole_CannotChangeAdmin(t *testing.T) {
	userRepo := new(mockAdminUserRepo)
	svc := NewAdminService(userRepo, new(mockAdminListingRepo), new(mockAdminSubRepo))

	userRepo.On("GetByID", mock.Anything, "u-admin").Return(
		&model.User{ID: "u-admin", Role: "admin"}, nil)

	user, err := svc.ChangeUserRole(context.Background(), "u-admin", "member", "admin")
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, ErrCannotModifyAdmin, err)
}

func TestAdminChangeUserRole_OwnerCanChangeAdmin(t *testing.T) {
	userRepo := new(mockAdminUserRepo)
	svc := NewAdminService(userRepo, new(mockAdminListingRepo), new(mockAdminSubRepo))

	userRepo.On("GetByID", mock.Anything, "u-admin").Return(
		&model.User{ID: "u-admin", Role: "admin"}, nil)
	userRepo.On("SetRole", mock.Anything, "u-admin", "editor").Return(
		&model.User{ID: "u-admin", Role: "editor"}, nil)

	user, err := svc.ChangeUserRole(context.Background(), "u-admin", "editor", "owner")
	assert.NoError(t, err)
	assert.Equal(t, "editor", user.Role)
}

func TestAdminChangeUserRole_InvalidRole(t *testing.T) {
	userRepo := new(mockAdminUserRepo)
	svc := NewAdminService(userRepo, new(mockAdminListingRepo), new(mockAdminSubRepo))

	user, err := svc.ChangeUserRole(context.Background(), "u-1", "superadmin", "admin")
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, ErrInvalidAdminRole, err)
}

func TestAdminDeleteListing_Success(t *testing.T) {
	listingRepo := new(mockAdminListingRepo)
	svc := NewAdminService(new(mockAdminUserRepo), listingRepo, new(mockAdminSubRepo))

	listingRepo.On("SoftDelete", mock.Anything, "lst-1").Return(nil)

	err := svc.DeleteListing(context.Background(), "lst-1")
	assert.NoError(t, err)
}

func TestAdminListUserListings_Success(t *testing.T) {
	listingRepo := new(mockAdminListingRepo)
	svc := NewAdminService(new(mockAdminUserRepo), listingRepo, new(mockAdminSubRepo))

	listings := []*model.Listing{{ID: "lst-1"}, {ID: "lst-2"}}
	listingRepo.On("ListByUser", mock.Anything, "u-1", 1, 10).Return(listings, 2, nil)

	result, total, err := svc.ListUserListings(context.Background(), "u-1", 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, result, 2)
}

func TestAdminListUserSubscriptions_Success(t *testing.T) {
	subRepo := new(mockAdminSubRepo)
	svc := NewAdminService(new(mockAdminUserRepo), new(mockAdminListingRepo), subRepo)

	subs := []*model.Subscription{{ID: "sub-1", Status: "active"}}
	subRepo.On("ListByUserID", mock.Anything, "u-1", 1, 10).Return(subs, 1, nil)

	result, total, err := svc.ListUserSubscriptions(context.Background(), "u-1", 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, result, 1)
}
