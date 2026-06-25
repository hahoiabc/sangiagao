package service

import (
	"context"
	"errors"

	"github.com/sangiagao/rice-marketplace/internal/model"
	"github.com/sangiagao/rice-marketplace/internal/repository"
)

var (
	ErrCannotRateSelf          = errors.New("cannot rate yourself")
	ErrTargetNotSeller         = errors.New("target user is not a seller")
	ErrAlreadyRated            = errors.New("you already rated this seller")
	ErrInsufficientInteraction = errors.New("cần ít nhất 5 tin nhắn qua lại với người bán trước khi đánh giá")
)

// minMessagesForRating — ngưỡng tương tác tối thiểu để được đánh giá (PRD FR-009 AC-1).
const minMessagesForRating = 5

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

	// BUG #16 fix (PRD FR-009 AC-1): chỉ cho đánh giá khi đã đủ tương tác thực
	// (≥5 tin nhắn qua lại) → chống rating brigading bằng nhiều tài khoản.
	msgCount, err := s.ratingRepo.CountMessagesBetween(ctx, reviewerID, req.SellerID)
	if err != nil {
		return nil, err
	}
	if msgCount < minMessagesForRating {
		return nil, ErrInsufficientInteraction
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
