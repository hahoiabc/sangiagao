package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/sangiagao/rice-marketplace/internal/model"
	"github.com/sangiagao/rice-marketplace/pkg/cache"
)

var (
	ErrNotListingOwner = errors.New("you don't own this listing")
	ErrMaxImages       = errors.New("maximum 3 images per listing")
	ErrListingDeleted  = errors.New("listing has been deleted")
)

const (
	marketplaceCacheTTL = 5 * time.Minute
	marketplaceCachePrefix = "marketplace:"
)

var ErrInvalidCategory = errors.New("phân loại gạo không hợp lệ")
var ErrInvalidProduct = errors.New("loại gạo không hợp lệ hoặc không thuộc phân loại đã chọn")
var ErrDailyLimitReached = errors.New("bạn đã đạt giới hạn 50 tin đăng/ngày. Vui lòng thử lại vào ngày mai")

const maxListingsPerDay = 50

type ListingService struct {
	listingRepo ListingRepository
	sponsorRepo SponsorRepository
	userRepo    UserRepository
	catalogRepo CatalogRepository
	cache       cache.Cache
}

func NewListingService(listingRepo ListingRepository, sponsorRepo SponsorRepository, userRepo UserRepository, catalogRepo CatalogRepository) *ListingService {
	return &ListingService{listingRepo: listingRepo, sponsorRepo: sponsorRepo, userRepo: userRepo, catalogRepo: catalogRepo}
}

// SetCache enables caching for marketplace queries (optional).
func (s *ListingService) SetCache(c cache.Cache) {
	s.cache = c
}

// --- Seller operations ---

func (s *ListingService) Create(ctx context.Context, userID string, req *model.CreateListingRequest) (*model.Listing, error) {
	// Check daily listing limit
	todayCount, err := s.listingRepo.CountTodayByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("check daily limit: %w", err)
	}
	if todayCount >= maxListingsPerDay {
		return nil, ErrDailyLimitReached
	}

	// Validate category and product against database catalog
	validCat, err := s.catalogRepo.ValidateCategory(ctx, req.Category)
	if err != nil {
		return nil, fmt.Errorf("validate category: %w", err)
	}
	if !validCat {
		return nil, ErrInvalidCategory
	}
	validProd, err := s.catalogRepo.ValidateProductInCategory(ctx, req.Category, req.RiceType)
	if err != nil {
		return nil, fmt.Errorf("validate product: %w", err)
	}
	if !validProd {
		return nil, ErrInvalidProduct
	}
	// Auto-generate title from product label if not provided
	if req.Title == "" {
		if label, err := s.catalogRepo.GetProductLabelByKey(ctx, req.RiceType); err == nil && label != "" {
			req.Title = label
		} else {
			req.Title = req.RiceType
		}
	}
	// Auto-fill province/ward from user profile if not provided
	if s.userRepo != nil && (req.Province == nil || req.Ward == nil) {
		if user, err := s.userRepo.GetByID(ctx, userID); err == nil {
			if req.Province == nil {
				req.Province = user.Province
			}
			if req.Ward == nil {
				req.Ward = user.Ward
			}
		}
	}
	listing, err := s.listingRepo.Create(ctx, userID, req)
	if err == nil {
		s.invalidateMarketplaceCache(ctx)
	}
	return listing, err
}

func (s *ListingService) GetByID(ctx context.Context, id string) (*model.Listing, error) {
	return s.listingRepo.GetByID(ctx, id)
}

func (s *ListingService) Update(ctx context.Context, userID, id string, req *model.UpdateListingRequest) (*model.Listing, error) {
	listing, err := s.listingRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if listing.UserID != userID {
		return nil, ErrNotListingOwner
	}
	updated, err := s.listingRepo.Update(ctx, id, req)
	if err == nil {
		s.invalidateMarketplaceCache(ctx)
	}
	return updated, err
}

func (s *ListingService) Delete(ctx context.Context, userID, id string) error {
	listing, err := s.listingRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if listing.UserID != userID {
		return ErrNotListingOwner
	}
	if err := s.listingRepo.SoftDelete(ctx, id); err != nil {
		return err
	}
	s.invalidateMarketplaceCache(ctx)
	return nil
}

func (s *ListingService) invalidateMarketplaceCache(ctx context.Context) {
	if s.cache != nil {
		_ = s.cache.DeleteByPrefix(ctx, marketplaceCachePrefix)
		_ = s.cache.Delete(ctx, "priceboard:v1")
	}
}

func (s *ListingService) ListByUser(ctx context.Context, userID string, page, limit int) ([]*model.Listing, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 20
	}
	return s.listingRepo.ListByUser(ctx, userID, page, limit)
}

func (s *ListingService) AddImage(ctx context.Context, userID, id, imageURL string) (*model.Listing, error) {
	listing, err := s.listingRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if listing.UserID != userID {
		return nil, ErrNotListingOwner
	}
	if len(listing.Images) >= 3 {
		return nil, ErrMaxImages
	}
	return s.listingRepo.AddImage(ctx, id, imageURL)
}

// --- Marketplace operations ---

func (s *ListingService) Browse(ctx context.Context, page, limit int) ([]*model.Listing, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 20
	}

	// Try cache
	if s.cache != nil {
		cacheKey := fmt.Sprintf("%spage:%d:limit:%d", marketplaceCachePrefix, page, limit)
		if data, err := s.cache.Get(ctx, cacheKey); err == nil && data != nil {
			var result struct {
				Data  []*model.Listing `json:"data"`
				Total int              `json:"total"`
			}
			if json.Unmarshal(data, &result) == nil {
				return result.Data, result.Total, nil
			}
		}

		// Cache miss — fetch from DB and cache
		listings, total, err := s.listingRepo.Browse(ctx, page, limit)
		if err != nil {
			return nil, 0, err
		}
		payload := struct {
			Data  []*model.Listing `json:"data"`
			Total int              `json:"total"`
		}{Data: listings, Total: total}
		if encoded, e := json.Marshal(payload); e == nil {
			_ = s.cache.Set(ctx, cacheKey, encoded, marketplaceCacheTTL)
		}
		return listings, total, nil
	}

	return s.listingRepo.Browse(ctx, page, limit)
}

func (s *ListingService) Search(ctx context.Context, filter *model.ListingFilter) ([]*model.Listing, int, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 || filter.Limit > 50 {
		filter.Limit = 20
	}
	return s.listingRepo.Search(ctx, filter)
}

func (s *ListingService) GetDetail(ctx context.Context, id string) (*model.ListingDetail, error) {
	detail, err := s.listingRepo.GetDetailWithSeller(ctx, id)
	if err != nil {
		return nil, err
	}
	_ = s.listingRepo.IncrementViewCount(ctx, id)
	return detail, nil
}

func (s *ListingService) GetPriceBoard(ctx context.Context) (*model.PriceBoardResponse, error) {
	// Try cache
	if s.cache != nil {
		if data, err := s.cache.Get(ctx, "priceboard:v1"); err == nil && data != nil {
			var result model.PriceBoardResponse
			if json.Unmarshal(data, &result) == nil {
				return &result, nil
			}
		}
	}

	// Get aggregated data from DB
	rows, err := s.listingRepo.GetPriceBoardData(ctx)
	if err != nil {
		return nil, fmt.Errorf("get price board data: %w", err)
	}

	// Build lookup: "category:rice_type" → {MinPrice, ListingCount}
	type priceInfo struct {
		MinPrice     float64
		ListingCount int
	}
	lookup := make(map[string]priceInfo)
	for _, r := range rows {
		key := r.Category + ":" + r.RiceType
		lookup[key] = priceInfo{MinPrice: r.MinPrice, ListingCount: r.ListingCount}
	}

	// Get active sponsors
	var sponsorMap map[string]string // product_key → logo_url
	if s.sponsorRepo != nil {
		sponsors, err := s.sponsorRepo.GetAllActive(ctx)
		if err == nil {
			sponsorMap = make(map[string]string)
			for _, sp := range sponsors {
				sponsorMap[sp.ProductKey] = sp.LogoURL
			}
		}
	}

	// Get catalog from database
	dbCatalog, err := s.catalogRepo.GetCatalogForAPI(ctx)
	if err != nil {
		return nil, fmt.Errorf("get catalog: %w", err)
	}

	// Build response from DB catalog
	var categories []model.PriceBoardCategory
	for _, cat := range dbCatalog {
		var products []model.PriceBoardEntry
		for _, p := range cat.Products {
			entry := model.PriceBoardEntry{
				ProductKey:   p.Key,
				ProductLabel: p.Label,
			}
			key := cat.Key + ":" + p.Key
			if info, ok := lookup[key]; ok {
				price := info.MinPrice
				entry.MinPrice = &price
				entry.ListingCount = info.ListingCount
			}
			if sponsorMap != nil {
				if logo, ok := sponsorMap[p.Key]; ok {
					entry.SponsorLogo = &logo
				}
			}
			products = append(products, entry)
		}
		categories = append(categories, model.PriceBoardCategory{
			CategoryKey:   cat.Key,
			CategoryLabel: cat.Label,
			Products:      products,
		})
	}

	result := &model.PriceBoardResponse{
		Categories: categories,
		UpdatedAt:  time.Now().Format(time.RFC3339),
	}

	// Cache result
	if s.cache != nil {
		if encoded, e := json.Marshal(result); e == nil {
			_ = s.cache.Set(ctx, "priceboard:v1", encoded, marketplaceCacheTTL)
		}
	}

	return result, nil
}
