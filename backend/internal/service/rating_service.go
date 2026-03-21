package service

import (
	"context"
	"errors"

	"github.com/sangiagao/rice-marketplace/internal/model"
	"github.com/sangiagao/rice-marketplace/internal/repository"
)

var (
	ErrCannotRateSelf  = errors.New("cannot rate yourself")
	ErrTargetNotSeller = errors.New("target user is not a seller")
	ErrAlreadyRated    = errors.New("you already rated this seller")
)

type RatingService struct {
	ratingRepo RatingRepository
}

func NewRatingService(ratingRepo RatingRepository) *RatingService {
	return &RatingService{ratingRepo: ratingRepo}
}

func (s *RatingService) Create(ctx context.Context, reviewerID string, req *model.CreateRatingRequest) (*model.Rating, error) {
	if reviewerID == req.SellerID {
		return nil, ErrCannotRateSelf
	}

	// Verify target user exists
	_, err := s.ratingRepo.GetSellerRole(ctx, req.SellerID)
	if err != nil {
		return nil, err
	}

	already, err := s.ratingRepo.HasRated(ctx, reviewerID, req.SellerID)
	if err != nil {
		return nil, err
	}
	if already {
		return nil, ErrAlreadyRated
	}

	rating, err := s.ratingRepo.Create(ctx, reviewerID, req)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateRating) {
			return nil, ErrAlreadyRated
		}
		return nil, err
	}
	return rating, nil
}

func (s *RatingService) ListBySeller(ctx context.Context, sellerID string, page, limit int) ([]*model.Rating, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 20
	}
	return s.ratingRepo.ListBySeller(ctx, sellerID, page, limit)
}

func (s *RatingService) GetSummary(ctx context.Context, sellerID string) (*model.RatingSummary, error) {
	return s.ratingRepo.GetSummary(ctx, sellerID)
}
