package service

import (
	"context"
	"testing"
	"time"

	"github.com/sangiagao/rice-marketplace/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockPlanRepo struct{ mock.Mock }

func (m *mockPlanRepo) ListActive(ctx context.Context) ([]model.SubscriptionPlan, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.SubscriptionPlan), args.Error(1)
}
func (m *mockPlanRepo) ListAll(ctx context.Context) ([]model.SubscriptionPlan, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.SubscriptionPlan), args.Error(1)
}
func (m *mockPlanRepo) GetByMonths(ctx context.Context, months int) (*model.SubscriptionPlan, error) {
	args := m.Called(ctx, months)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.SubscriptionPlan), args.Error(1)
}
func (m *mockPlanRepo) Create(ctx context.Context, req *model.CreatePlanRequest) (*model.SubscriptionPlan, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.SubscriptionPlan), args.Error(1)
}
func (m *mockPlanRepo) Update(ctx context.Context, id string, req *model.UpdatePlanRequest) (*model.SubscriptionPlan, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.SubscriptionPlan), args.Error(1)
}
func (m *mockPlanRepo) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestSubscriptionGetStatus_Active(t *testing.T) {
	subRepo := new(mockSubRepo)
	planRepo := new(mockPlanRepo)
	svc := NewSubscriptionService(subRepo, planRepo)

	sub := &model.Subscription{
		ID: "sub-1", UserID: "user-1", Plan: "free_trial",
		Status: "active", ExpiresAt: time.Now().Add(15 * 24 * time.Hour),
	}
	subRepo.On("GetByUserID", ctx, "user-1").Return(sub, nil)

	status, err := svc.GetStatus(context.Background(), "user-1")
	assert.NoError(t, err)
	assert.True(t, status.IsActive)
	assert.Equal(t, 15, status.DaysLeft)
}

func TestSubscriptionGetStatus_Expired(t *testing.T) {
	subRepo := new(mockSubRepo)
	planRepo := new(mockPlanRepo)
	svc := NewSubscriptionService(subRepo, planRepo)

	sub := &model.Subscription{
		ID: "sub-1", UserID: "user-1", Plan: "free_trial",
		Status: "expired", ExpiresAt: time.Now().Add(-1 * time.Hour),
	}
	subRepo.On("GetByUserID", ctx, "user-1").Return(sub, nil)

	status, err := svc.GetStatus(context.Background(), "user-1")
	assert.NoError(t, err)
	assert.False(t, status.IsActive)
	assert.Equal(t, 0, status.DaysLeft)
}

func TestSubscriptionGetStatus_NoSubscription(t *testing.T) {
	subRepo := new(mockSubRepo)
	planRepo := new(mockPlanRepo)
	svc := NewSubscriptionService(subRepo, planRepo)

	subRepo.On("GetByUserID", ctx, "user-1").Return(nil, nil)

	status, err := svc.GetStatus(context.Background(), "user-1")
	assert.NoError(t, err)
	assert.False(t, status.IsActive)
	assert.Equal(t, 0, status.DaysLeft)
}

func TestSubscriptionAdminActivate_NewSubscription(t *testing.T) {
	subRepo := new(mockSubRepo)
	planRepo := new(mockPlanRepo)
	svc := NewSubscriptionService(subRepo, planRepo)

	plan := &model.SubscriptionPlan{Months: 12, Amount: 300000, Label: "12 tháng", IsActive: true}
	planRepo.On("GetByMonths", ctx, 12).Return(plan, nil)
	subRepo.On("GetActiveByUserID", ctx, "user-1").Return(nil, nil)
	sub := &model.Subscription{ID: "sub-1", UserID: "user-1", Plan: "paid", Status: "active"}
	subRepo.On("ActivateByUserID", ctx, "user-1", 360, 12, int64(300000), "paid").Return(sub, nil)
	subRepo.On("RestoreListings", ctx, "user-1").Return(2, nil)

	result, err := svc.AdminActivate(context.Background(), "user-1", 12)
	assert.NoError(t, err)
	assert.Equal(t, "paid", result.Plan)
	subRepo.AssertCalled(t, "ActivateByUserID", ctx, "user-1", 360, 12, int64(300000), "paid")
}

func TestSubscriptionAdminActivate_ExtendExisting(t *testing.T) {
	subRepo := new(mockSubRepo)
	planRepo := new(mockPlanRepo)
	svc := NewSubscriptionService(subRepo, planRepo)

	plan := &model.SubscriptionPlan{Months: 1, Amount: 35000, Label: "1 tháng", IsActive: true}
	planRepo.On("GetByMonths", ctx, 1).Return(plan, nil)
	existing := &model.Subscription{ID: "sub-existing", UserID: "user-1", Plan: "paid", Status: "active"}
	subRepo.On("GetActiveByUserID", ctx, "user-1").Return(existing, nil)
	extended := &model.Subscription{ID: "sub-existing", UserID: "user-1", Plan: "paid", Status: "active"}
	subRepo.On("ExtendSubscription", ctx, "sub-existing", 30, 1, int64(35000)).Return(extended, nil)
	subRepo.On("RestoreListings", ctx, "user-1").Return(0, nil)

	result, err := svc.AdminActivate(context.Background(), "user-1", 1)
	assert.NoError(t, err)
	assert.Equal(t, "sub-existing", result.ID)
	subRepo.AssertNotCalled(t, "ActivateByUserID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestSubscriptionAdminActivate_InvalidPlan(t *testing.T) {
	subRepo := new(mockSubRepo)
	planRepo := new(mockPlanRepo)
	svc := NewSubscriptionService(subRepo, planRepo)

	planRepo.On("GetByMonths", ctx, 0).Return(nil, nil)
	planRepo.On("GetByMonths", ctx, 5).Return(nil, nil)

	_, err := svc.AdminActivate(context.Background(), "user-1", 0)
	assert.ErrorIs(t, err, ErrInvalidPlan)

	_, err = svc.AdminActivate(context.Background(), "user-1", 5)
	assert.ErrorIs(t, err, ErrInvalidPlan)
}

var ctx = context.Background()
