package service

import (
	"context"
	"testing"

	"github.com/sangiagao/rice-marketplace/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockRatingRepo struct{ mock.Mock }

func (m *mockRatingRepo) Create(ctx context.Context, reviewerID string, req *model.CreateRatingRequest) (*model.Rating, error) {
	args := m.Called(ctx, reviewerID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Rating), args.Error(1)
}
func (m *mockRatingRepo) ListBySeller(ctx context.Context, sellerID string, page, limit int) ([]*model.Rating, int, error) {
	args := m.Called(ctx, sellerID, page, limit)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*model.Rating), args.Int(1), args.Error(2)
}
func (m *mockRatingRepo) GetSummary(ctx context.Context, sellerID string) (*model.RatingSummary, error) {
	args := m.Called(ctx, sellerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.RatingSummary), args.Error(1)
}
func (m *mockRatingRepo) HasRated(ctx context.Context, reviewerID, sellerID string) (bool, error) {
	args := m.Called(ctx, reviewerID, sellerID)
	return args.Bool(0), args.Error(1)
}
func (m *mockRatingRepo) GetSellerRole(ctx context.Context, userID string) (string, error) {
	args := m.Called(ctx, userID)
	return args.String(0), args.Error(1)
}

// --- Tests ---

func TestRatingCreate_Success(t *testing.T) {
	repo := new(mockRatingRepo)
	svc := NewRatingService(repo)

	req := &model.CreateRatingRequest{SellerID: "seller-1", Stars: 5, Comment: "Great seller!!!!"}
	repo.On("GetSellerRole", mock.Anything, "seller-1").Return("member", nil)
	repo.On("HasRated", mock.Anything, "buyer-1", "seller-1").Return(false, nil)
	repo.On("Create", mock.Anything, "buyer-1", req).Return(&model.Rating{ID: "r-1", Stars: 5}, nil)

	rating, err := svc.Create(context.Background(), "buyer-1", req)
	assert.NoError(t, err)
	assert.Equal(t, 5, rating.Stars)
}

func TestRatingCreate_SelfRate(t *testing.T) {
	repo := new(mockRatingRepo)
	svc := NewRatingService(repo)

	req := &model.CreateRatingRequest{SellerID: "user-1", Stars: 5, Comment: "Rating myself!!!!"}
	_, err := svc.Create(context.Background(), "user-1", req)
	assert.ErrorIs(t, err, ErrCannotRateSelf)
}

func TestRatingCreate_MemberCanReceiveRating(t *testing.T) {
	repo := new(mockRatingRepo)
	svc := NewRatingService(repo)

	req := &model.CreateRatingRequest{SellerID: "user-2", Stars: 5, Comment: "Great member!"}
	repo.On("GetSellerRole", mock.Anything, "user-2").Return("member", nil)
	repo.On("HasRated", mock.Anything, "user-1", "user-2").Return(false, nil)
	repo.On("Create", mock.Anything, "user-1", req).Return(&model.Rating{ID: "r-2", Stars: 5}, nil)

	rating, err := svc.Create(context.Background(), "user-1", req)
	assert.NoError(t, err)
	assert.Equal(t, 5, rating.Stars)
}

func TestRatingCreate_AlreadyRated(t *testing.T) {
	repo := new(mockRatingRepo)
	svc := NewRatingService(repo)

	req := &model.CreateRatingRequest{SellerID: "seller-1", Stars: 4, Comment: "Duplicate attempt!!!!"}
	repo.On("GetSellerRole", mock.Anything, "seller-1").Return("member", nil)
	repo.On("HasRated", mock.Anything, "buyer-1", "seller-1").Return(true, nil)

	_, err := svc.Create(context.Background(), "buyer-1", req)
	assert.ErrorIs(t, err, ErrAlreadyRated)
}

func TestRatingListBySeller_Success(t *testing.T) {
	repo := new(mockRatingRepo)
	svc := NewRatingService(repo)

	ratings := []*model.Rating{{ID: "r-1", Stars: 5}}
	repo.On("ListBySeller", mock.Anything, "seller-1", 1, 20).Return(ratings, 1, nil)

	result, total, err := svc.ListBySeller(context.Background(), "seller-1", 1, 20)
	assert.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, result, 1)
}

func TestRatingGetSummary_Success(t *testing.T) {
	repo := new(mockRatingRepo)
	svc := NewRatingService(repo)

	repo.On("GetSummary", mock.Anything, "seller-1").Return(&model.RatingSummary{Average: 4.5, Count: 10}, nil)

	summary, err := svc.GetSummary(context.Background(), "seller-1")
	assert.NoError(t, err)
	assert.Equal(t, 4.5, summary.Average)
	assert.Equal(t, 10, summary.Count)
}
