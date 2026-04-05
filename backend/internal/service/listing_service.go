package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/sangiagao/rice-marketplace/internal/model"
	"github.com/sangiagao/rice-marketplace/pkg/cache"
)

var (
	ErrNotListingOwner      = errors.New("you don't own this listing")
	ErrMaxImages            = errors.New("maximum 1 image per listing")
	ErrListingDeleted       = errors.New("listing has been deleted")
	ErrSubscriptionRequired = errors.New("bạn cần có gói đăng ký còn hiệu lực để đăng tin")
)

const (
	marketplaceCacheTTL = 5 * time.Minute
	marketplaceCachePrefix = "marketplace:"
)

var ErrInvalidCategory = errors.New("phân loại gạo không hợp lệ")
var ErrInvalidProduct = errors.New("loại gạo không hợp lệ hoặc không thuộc phân loại đã chọn")
var ErrDailyLimitReached = errors.New("loại gạo này đã đăng tối đa 3 lần hôm nay")

const maxPerProductPerDay = 3

// catalogCache holds in-memory catalog data to avoid DB queries on every listing create.
type catalogCache struct {
	mu            sync.RWMutex
	categories    map[string]bool            // category_key → exists
	products      map[string]map[string]bool // category_key → product_key → exists
	productLabels map[string]string          // product_key → label
	loadedAt      time.Time
}

const catalogCacheTTL = 10 * time.Minute

type ListingService struct {
	listingRepo ListingRepository
	sponsorRepo SponsorRepository
	userRepo    UserRepository
	catalogRepo CatalogRepository
	cache       cache.Cache
	catCache    catalogCache
}

func NewListingService(listingRepo ListingRepository, sponsorRepo SponsorRepository, userRepo UserRepository, catalogRepo CatalogRepository) *ListingService {
	return &ListingService{listingRepo: listingRepo, sponsorRepo: sponsorRepo, userRepo: userRepo, catalogRepo: catalogRepo}
}

// SetCache enables caching for marketplace queries (optional).
func (s *ListingService) SetCache(c cache.Cache) {
	s.cache = c
}

// loadCatalogCache loads catalog data into memory. Thread-safe, skips if cache is fresh.
func (s *ListingService) loadCatalogCache(ctx context.Context) {
	s.catCache.mu.RLock()
	if !s.catCache.loadedAt.IsZero() && time.Since(s.catCache.loadedAt) < catalogCacheTTL {
		s.catCache.mu.RUnlock()
		return
	}
	s.catCache.mu.RUnlock()

	// Double-check under write lock
	s.catCache.mu.Lock()
	defer s.catCache.mu.Unlock()
	if !s.catCache.loadedAt.IsZero() && time.Since(s.catCache.loadedAt) < catalogCacheTTL {
		return
	}

	catalog, err := s.catalogRepo.GetCatalogForAPI(ctx)
	if err != nil {
		log.Printf("[CATALOG CACHE] Failed to load: %v", err)
		return
	}

	categories := make(map[string]bool)
	products := make(map[string]map[string]bool)
	labels := make(map[string]string)

	for _, cat := range catalog {
		categories[cat.Key] = true
		products[cat.Key] = make(map[string]bool)
		for _, p := range cat.Products {
			products[cat.Key][p.Key] = true
			labels[p.Key] = p.Label
		}
	}

	s.catCache.categories = categories
	s.catCache.products = products
	s.catCache.productLabels = labels
	s.catCache.loadedAt = time.Now()
}

// validateCatalog checks category+product from in-memory cache, falls back to DB on cache miss.
func (s *ListingService) validateCatalog(ctx context.Context, categoryKey, productKey string) (bool, bool, string) {
	s.loadCatalogCache(ctx)

	s.catCache.mu.RLock()
	defer s.catCache.mu.RUnlock()

	if s.catCache.categories == nil {
		// Cache failed to load — caller should fall back to DB
		return false, false, ""
	}

	catOK := s.catCache.categories[categoryKey]
	prodOK := false
	if catOK {
		if prods, ok := s.catCache.products[categoryKey]; ok {
			prodOK = prods[productKey]
		}
	}
	label := s.catCache.productLabels[productKey]
	return catOK, prodOK, label
}

// --- Seller operations ---

func (s *ListingService) Create(ctx context.Context, userID string, req *model.CreateListingRequest) (*model.Listing, error) {
	// Check if user is blocked or has active subscription
	var seller *model.User
	if s.userRepo != nil {
		var err error
		seller, err = s.userRepo.GetByID(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("check user: %w", err)
		}
		if seller.IsBlocked {
			return nil, ErrUserBlocked
		}
		if seller.SubscriptionExpiresAt == nil || seller.SubscriptionExpiresAt.Before(time.Now()) {
			return nil, ErrSubscriptionRequired
		}
	}

	// Check per-product daily limit (3 per rice_type per day)
	typeCount, err := s.listingRepo.CountTodayByUserAndType(ctx, userID, req.RiceType)
	if err != nil {
		return nil, fmt.Errorf("check daily limit: %w", err)
	}
	if typeCount >= maxPerProductPerDay {
		return nil, ErrDailyLimitReached
	}

	// Validate category and product from in-memory cache (avoids 2-3 DB queries per create)
	catOK, prodOK, productLabel := s.validateCatalog(ctx, req.Category, req.RiceType)
	if !catOK {
		// Fallback to DB if cache empty
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
		if req.Title == "" {
			if label, err := s.catalogRepo.GetProductLabelByKey(ctx, req.RiceType); err == nil && label != "" {
				productLabel = label
			}
		}
	} else if !prodOK {
		return nil, ErrInvalidProduct
	}

	// Auto-generate title from product label if not provided
	if req.Title == "" {
		if productLabel != "" {
			req.Title = productLabel
		} else {
			req.Title = req.RiceType
		}
	}
	// Auto-fill province/ward from user profile (reuse already-fetched seller)
	if seller != nil && (req.Province == nil || req.Ward == nil) {
		if req.Province == nil {
			req.Province = seller.Province
		}
		if req.Ward == nil {
			req.Ward = seller.Ward
		}
	}
	listing, err := s.listingRepo.Create(ctx, userID, req)
	if err == nil {
		s.InvalidateMarketplaceCache(ctx)
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
		s.InvalidateMarketplaceCache(ctx)
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
	s.InvalidateMarketplaceCache(ctx)
	return nil
}

func (s *ListingService) BatchDeleteOwn(ctx context.Context, userID string, ids []string) (int, error) {
	// Verify all listings belong to user
	for _, id := range ids {
		listing, err := s.listingRepo.GetByID(ctx, id)
		if err != nil {
			return 0, err
		}
		if listing.UserID != userID {
			return 0, ErrNotListingOwner
		}
	}
	deleted, err := s.listingRepo.BatchSoftDelete(ctx, ids)
	if err != nil {
		return 0, err
	}
	if deleted > 0 {
		s.InvalidateMarketplaceCache(ctx)
	}
	return deleted, nil
}

func (s *ListingService) InvalidateMarketplaceCache(ctx context.Context) {
	if s.cache != nil {
		// Only invalidate priceboard (aggregate data changes).
		// Marketplace listing pages use TTL-based expiry (5 min) — no need to
		// scan+delete all marketplace:* keys on every create/update/delete.
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

func (s *ListingService) RemoveImage(ctx context.Context, userID, id, imageURL string) (*model.Listing, error) {
	listing, err := s.listingRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if listing.UserID != userID {
		return nil, ErrNotListingOwner
	}
	return s.listingRepo.RemoveImage(ctx, id, imageURL)
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

	// Try cache
	if s.cache != nil {
		cacheKey := s.searchCacheKey(filter)
		if data, err := s.cache.Get(ctx, cacheKey); err == nil && data != nil {
			var result struct {
				Data  []*model.Listing `json:"data"`
				Total int              `json:"total"`
			}
			if json.Unmarshal(data, &result) == nil {
				return result.Data, result.Total, nil
			}
		}

		listings, total, err := s.listingRepo.Search(ctx, filter)
		if err != nil {
			return nil, 0, err
		}
		payload := struct {
			Data  []*model.Listing `json:"data"`
			Total int              `json:"total"`
		}{Data: listings, Total: total}
		if encoded, e := json.Marshal(payload); e == nil {
			_ = s.cache.Set(ctx, cacheKey, encoded, 2*time.Minute)
		}
		return listings, total, nil
	}

	return s.listingRepo.Search(ctx, filter)
}

func (s *ListingService) searchCacheKey(f *model.ListingFilter) string {
	key := fmt.Sprintf("%ssearch:q=%s:cat=%s:rt=%s:prov=%s:ward=%s:sort=%s:p=%d:l=%d",
		marketplaceCachePrefix, f.Query, f.Category, f.RiceType, f.Province, f.Ward, f.Sort, f.Page, f.Limit)
	if f.MinPrice != nil {
		key += fmt.Sprintf(":minP=%.0f", *f.MinPrice)
	}
	if f.MaxPrice != nil {
		key += fmt.Sprintf(":maxP=%.0f", *f.MaxPrice)
	}
	if f.MinQty != nil {
		key += fmt.Sprintf(":minQ=%.0f", *f.MinQty)
	}
	return key
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
