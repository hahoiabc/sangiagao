package service

import (
	"context"
	"time"

	"github.com/sangiagao/rice-marketplace/internal/model"
	"github.com/sangiagao/rice-marketplace/pkg/cache"
)

const sloganCacheKey = "site:slogan"
const sloganColorCacheKey = "site:slogan_color"
const guideVideoCacheKey = "site:guide_video"
const sloganCacheTTL = 10 * time.Minute

type SiteSettingsService struct {
	repo  SiteSettingsRepository
	cache cache.Cache
}

func NewSiteSettingsService(repo SiteSettingsRepository) *SiteSettingsService {
	return &SiteSettingsService{repo: repo}
}

func (s *SiteSettingsService) SetCache(c cache.Cache) {
	s.cache = c
}

func (s *SiteSettingsService) GetSlogan(ctx context.Context) (*model.SiteSetting, error) {
	// Try cache first
	if s.cache != nil {
		if cached, err := s.cache.Get(ctx, sloganCacheKey); err == nil && cached != nil {
			return &model.SiteSetting{Key: "slogan", Value: string(cached)}, nil
		}
	}

	setting, err := s.repo.Get(ctx, "slogan")
	if err != nil {
		return nil, err
	}

	// Cache the result
	if s.cache != nil {
		_ = s.cache.Set(ctx, sloganCacheKey, []byte(setting.Value), sloganCacheTTL)
	}

	return setting, nil
}

func (s *SiteSettingsService) UpdateSlogan(ctx context.Context, value string) (*model.SiteSetting, error) {
	setting, err := s.repo.Set(ctx, "slogan", value)
	if err != nil {
		return nil, err
	}

	// Invalidate cache
	if s.cache != nil {
		_ = s.cache.Delete(ctx, sloganCacheKey)
	}

	return setting, nil
}

func (s *SiteSettingsService) GetSloganColor(ctx context.Context) (*model.SiteSetting, error) {
	if s.cache != nil {
		if cached, err := s.cache.Get(ctx, sloganColorCacheKey); err == nil && cached != nil {
			return &model.SiteSetting{Key: "slogan_color", Value: string(cached)}, nil
		}
	}

	setting, err := s.repo.Get(ctx, "slogan_color")
	if err != nil {
		// Default color if not set
		return &model.SiteSetting{Key: "slogan_color", Value: "#4F46E5"}, nil
	}

	if s.cache != nil {
		_ = s.cache.Set(ctx, sloganColorCacheKey, []byte(setting.Value), sloganCacheTTL)
	}

	return setting, nil
}

func (s *SiteSettingsService) UpdateSloganColor(ctx context.Context, value string) (*model.SiteSetting, error) {
	setting, err := s.repo.Set(ctx, "slogan_color", value)
	if err != nil {
		return nil, err
	}

	if s.cache != nil {
		_ = s.cache.Delete(ctx, sloganColorCacheKey)
	}

	return setting, nil
}

func (s *SiteSettingsService) GetGuideVideo(ctx context.Context) (*model.SiteSetting, error) {
	if s.cache != nil {
		if data, err := s.cache.Get(ctx, guideVideoCacheKey); err == nil && data != nil {
			return &model.SiteSetting{Key: "guide_video", Value: string(data)}, nil
		}
	}
	setting, err := s.repo.Get(ctx, "guide_video")
	if err != nil {
		return &model.SiteSetting{Key: "guide_video", Value: ""}, nil
	}
	if s.cache != nil {
		_ = s.cache.Set(ctx, guideVideoCacheKey, []byte(setting.Value), sloganCacheTTL)
	}
	return setting, nil
}

func (s *SiteSettingsService) UpdateGuideVideo(ctx context.Context, value string) (*model.SiteSetting, error) {
	setting, err := s.repo.Set(ctx, "guide_video", value)
	if err != nil {
		return nil, err
	}
	if s.cache != nil {
		_ = s.cache.Delete(ctx, guideVideoCacheKey)
	}
	return setting, nil
}
