package service

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/sangiagao/rice-marketplace/internal/model"
	"github.com/sangiagao/rice-marketplace/pkg/cache"
)

const sloganCacheKey = "site:slogan"
const sloganColorCacheKey = "site:slogan_color"
const guideVideoCacheKey = "site:guide_video"
const aboutPageCacheKey = "site:about_page"
const listingDisplayDaysCacheKey = "site:listing_display_days"
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

// GetListingDisplayDays trả về số ngày tin đăng được hiển thị trên sàn.
// 0 = KHÔNG giới hạn (mặc định). Thiếu key / lỗi parse → 0 (an toàn, không ẩn tin).
// Chữ ký func(ctx) int khớp callback tiêm vào ListingRepo.SetDisplayDaysFn.
func (s *SiteSettingsService) GetListingDisplayDays(ctx context.Context) int {
	if s.cache != nil {
		if cached, err := s.cache.Get(ctx, listingDisplayDaysCacheKey); err == nil && cached != nil {
			if n, e := strconv.Atoi(string(cached)); e == nil {
				return n
			}
		}
	}
	setting, err := s.repo.Get(ctx, "listing_display_days")
	if err != nil || setting == nil {
		return 0
	}
	n, err := strconv.Atoi(strings.TrimSpace(setting.Value))
	if err != nil || n < 0 {
		n = 0
	}
	if s.cache != nil {
		_ = s.cache.Set(ctx, listingDisplayDaysCacheKey, []byte(strconv.Itoa(n)), sloganCacheTTL)
	}
	return n
}

// UpdateListingDisplayDays lưu số ngày (kẹp 0-365; 0 = không giới hạn).
func (s *SiteSettingsService) UpdateListingDisplayDays(ctx context.Context, days int) (*model.SiteSetting, error) {
	if days < 0 {
		days = 0
	}
	if days > 365 {
		days = 365
	}
	setting, err := s.repo.Set(ctx, "listing_display_days", strconv.Itoa(days))
	if err != nil {
		return nil, err
	}
	if s.cache != nil {
		_ = s.cache.Delete(ctx, listingDisplayDaysCacheKey)
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

func (s *SiteSettingsService) GetAboutPage(ctx context.Context) (*model.SiteSetting, error) {
	if s.cache != nil {
		if data, err := s.cache.Get(ctx, aboutPageCacheKey); err == nil && data != nil {
			return &model.SiteSetting{Key: "about_page", Value: string(data)}, nil
		}
	}
	setting, err := s.repo.Get(ctx, "about_page")
	if err != nil {
		return &model.SiteSetting{Key: "about_page", Value: ""}, nil
	}
	if s.cache != nil {
		_ = s.cache.Set(ctx, aboutPageCacheKey, []byte(setting.Value), sloganCacheTTL)
	}
	return setting, nil
}

func (s *SiteSettingsService) UpdateAboutPage(ctx context.Context, value string) (*model.SiteSetting, error) {
	setting, err := s.repo.Set(ctx, "about_page", value)
	if err != nil {
		return nil, err
	}
	if s.cache != nil {
		_ = s.cache.Delete(ctx, aboutPageCacheKey)
	}
	return setting, nil
}
