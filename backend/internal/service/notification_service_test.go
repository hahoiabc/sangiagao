package service

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sangiagao/rice-marketplace/internal/model"
	"github.com/sangiagao/rice-marketplace/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockNotifRepo struct{ mock.Mock }

func (m *mockNotifRepo) Create(ctx context.Context, userID, nType, title, body string, data json.RawMessage) (*model.Notification, error) {
	args := m.Called(ctx, userID, nType, title, body, data)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Notification), args.Error(1)
}
func (m *mockNotifRepo) ListByUser(ctx context.Context, userID string, page, limit int) ([]*model.Notification, int, error) {
	args := m.Called(ctx, userID, page, limit)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*model.Notification), args.Int(1), args.Error(2)
}
func (m *mockNotifRepo) MarkRead(ctx context.Context, id, userID string) error {
	return m.Called(ctx, id, userID).Error(0)
}
func (m *mockNotifRepo) RegisterDevice(ctx context.Context, userID, token, platform string) error {
	return m.Called(ctx, userID, token, platform).Error(0)
}
func (m *mockNotifRepo) GetDeviceTokens(ctx context.Context, userID string) ([]string, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}
func (m *mockNotifRepo) UnreadCount(ctx context.Context, userID string) (int, error) {
	args := m.Called(ctx, userID)
	return args.Int(0), args.Error(1)
}
func (m *mockNotifRepo) CreateBatch(ctx context.Context, userIDs []string, nType, title, body string, data json.RawMessage) error {
	return m.Called(ctx, userIDs, nType, title, body, data).Error(0)
}
func (m *mockNotifRepo) GetAllUserIDs(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}
func (m *mockNotifRepo) GetAllDeviceTokens(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func TestNotifRegisterDevice_Success(t *testing.T) {
	repo := new(mockNotifRepo)
	svc := NewNotificationService(repo, nil)

	repo.On("RegisterDevice", mock.Anything, "user-1", "token-abc", "ios").Return(nil)

	err := svc.RegisterDevice(context.Background(), "user-1", "token-abc", "ios")
	assert.NoError(t, err)
}

func TestNotifList_Success(t *testing.T) {
	repo := new(mockNotifRepo)
	svc := NewNotificationService(repo, nil)

	notifications := []*model.Notification{{ID: "n-1", Title: "Test"}}
	repo.On("ListByUser", mock.Anything, "user-1", 1, 20).Return(notifications, 1, nil)

	result, total, err := svc.List(context.Background(), "user-1", 1, 20)
	assert.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, result, 1)
}

func TestNotifMarkRead_Success(t *testing.T) {
	repo := new(mockNotifRepo)
	svc := NewNotificationService(repo, nil)

	repo.On("MarkRead", mock.Anything, "n-1", "user-1").Return(nil)

	err := svc.MarkRead(context.Background(), "n-1", "user-1")
	assert.NoError(t, err)
}

func TestNotifMarkRead_NotFound(t *testing.T) {
	repo := new(mockNotifRepo)
	svc := NewNotificationService(repo, nil)

	repo.On("MarkRead", mock.Anything, "bad", "user-1").Return(repository.ErrNotificationNotFound)

	err := svc.MarkRead(context.Background(), "bad", "user-1")
	assert.ErrorIs(t, err, repository.ErrNotificationNotFound)
}

func TestNotifUnreadCount_Success(t *testing.T) {
	repo := new(mockNotifRepo)
	svc := NewNotificationService(repo, nil)

	repo.On("UnreadCount", mock.Anything, "user-1").Return(5, nil)

	count, err := svc.UnreadCount(context.Background(), "user-1")
	assert.NoError(t, err)
	assert.Equal(t, 5, count)
}
