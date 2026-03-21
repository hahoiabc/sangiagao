package service

import (
	"context"
	"errors"

	"github.com/sangiagao/rice-marketplace/internal/model"
)

var ErrInvalidProductKey = errors.New("product_key không hợp lệ")

type SponsorService struct {
	sponsorRepo SponsorRepository
}

func NewSponsorService(sponsorRepo SponsorRepository) *SponsorService {
	return &SponsorService{sponsorRepo: sponsorRepo}
}

func (s *SponsorService) Create(ctx context.Context, req *model.CreateSponsorRequest) (*model.ProductSponsor, error) {
	if !model.AllProductKeys()[req.ProductKey] {
		return nil, ErrInvalidProductKey
	}
	return s.sponsorRepo.Create(ctx, req)
}

func (s *SponsorService) Update(ctx context.Context, id string, req *model.UpdateSponsorRequest) (*model.ProductSponsor, error) {
	return s.sponsorRepo.Update(ctx, id, req)
}

func (s *SponsorService) Delete(ctx context.Context, id string) error {
	return s.sponsorRepo.Delete(ctx, id)
}

func (s *SponsorService) List(ctx context.Context, page, limit int) ([]*model.ProductSponsor, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 20
	}
	return s.sponsorRepo.List(ctx, page, limit)
}
