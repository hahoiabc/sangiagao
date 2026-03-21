package service

import (
	"context"
	"errors"
	"testing"

	"github.com/sangiagao/rice-marketplace/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockSponsorRepo struct{ mock.Mock }

func (m *mockSponsorRepo) GetAllActive(ctx context.Context) ([]*model.ProductSponsor, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.ProductSponsor), args.Error(1)
}

func (m *mockSponsorRepo) Create(ctx context.Context, req *model.CreateSponsorRequest) (*model.ProductSponsor, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ProductSponsor), args.Error(1)
}

func (m *mockSponsorRepo) Update(ctx context.Context, id string, req *model.UpdateSponsorRequest) (*model.ProductSponsor, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ProductSponsor), args.Error(1)
}

func (m *mockSponsorRepo) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockSponsorRepo) List(ctx context.Context, page, limit int) ([]*model.ProductSponsor, int, error) {
	args := m.Called(ctx, page, limit)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*model.ProductSponsor), args.Int(1), args.Error(2)
}

func TestSponsorCreate_Success(t *testing.T) {
	repo := new(mockSponsorRepo)
	svc := NewSponsorService(repo)

	// Get a valid product key from the catalog
	keys := model.AllProductKeys()
	var validKey string
	for k := range keys {
		validKey = k
		break
	}

	req := &model.CreateSponsorRequest{ProductKey: validKey, LogoURL: "https://example.com/logo.png"}
	repo.On("Create", mock.Anything, req).Return(
		&model.ProductSponsor{ID: "sp-1", ProductKey: validKey, LogoURL: req.LogoURL, IsActive: true}, nil)

	sp, err := svc.Create(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, "sp-1", sp.ID)
	assert.Equal(t, validKey, sp.ProductKey)
}

func TestSponsorCreate_InvalidProductKey(t *testing.T) {
	repo := new(mockSponsorRepo)
	svc := NewSponsorService(repo)

	req := &model.CreateSponsorRequest{ProductKey: "invalid_key", LogoURL: "https://example.com/logo.png"}
	sp, err := svc.Create(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, sp)
	assert.Equal(t, ErrInvalidProductKey, err)
}

func TestSponsorUpdate_Success(t *testing.T) {
	repo := new(mockSponsorRepo)
	svc := NewSponsorService(repo)

	newURL := "https://example.com/new-logo.png"
	req := &model.UpdateSponsorRequest{LogoURL: &newURL}
	repo.On("Update", mock.Anything, "sp-1", req).Return(
		&model.ProductSponsor{ID: "sp-1", LogoURL: newURL}, nil)

	sp, err := svc.Update(context.Background(), "sp-1", req)
	assert.NoError(t, err)
	assert.Equal(t, newURL, sp.LogoURL)
}

func TestSponsorDelete_Success(t *testing.T) {
	repo := new(mockSponsorRepo)
	svc := NewSponsorService(repo)

	repo.On("Delete", mock.Anything, "sp-1").Return(nil)

	err := svc.Delete(context.Background(), "sp-1")
	assert.NoError(t, err)
}

func TestSponsorDelete_NotFound(t *testing.T) {
	repo := new(mockSponsorRepo)
	svc := NewSponsorService(repo)

	repo.On("Delete", mock.Anything, "sp-999").Return(errors.New("not found"))

	err := svc.Delete(context.Background(), "sp-999")
	assert.Error(t, err)
}

func TestSponsorList_Success(t *testing.T) {
	repo := new(mockSponsorRepo)
	svc := NewSponsorService(repo)

	items := []*model.ProductSponsor{{ID: "sp-1"}, {ID: "sp-2"}}
	repo.On("List", mock.Anything, 1, 20).Return(items, 2, nil)

	result, total, err := svc.List(context.Background(), 1, 20)
	assert.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, result, 2)
}

func TestSponsorList_DefaultsInvalidPage(t *testing.T) {
	repo := new(mockSponsorRepo)
	svc := NewSponsorService(repo)

	items := []*model.ProductSponsor{{ID: "sp-1"}}
	repo.On("List", mock.Anything, 1, 20).Return(items, 1, nil)

	result, total, err := svc.List(context.Background(), 0, 0)
	assert.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, result, 1)
}
