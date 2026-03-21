package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sangiagao/rice-marketplace/internal/model"
)

type MarketplaceHandler struct {
	listingService ListingServiceInterface
	catalogService CatalogServiceInterface
}

func NewMarketplaceHandler(listingService ListingServiceInterface, catalogService CatalogServiceInterface) *MarketplaceHandler {
	return &MarketplaceHandler{listingService: listingService, catalogService: catalogService}
}

func (h *MarketplaceHandler) Browse(c *gin.Context) {
	page, limit := parsePagination(c, 20)

	listings, total, err := h.listingService.Browse(c.Request.Context(), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to browse listings"})
		return
	}

	totalPages := (total + limit - 1) / limit
	c.JSON(http.StatusOK, model.PaginatedResponse{
		Data: listings, Total: total, Page: page, Limit: limit, TotalPages: totalPages,
	})
}

func (h *MarketplaceHandler) Search(c *gin.Context) {
	filter := &model.ListingFilter{
		Query:    c.Query("q"),
		Category: c.Query("category"),
		RiceType: c.Query("type"),
		Province: c.Query("province"),
		Ward:     c.Query("ward"),
	}

	if v := c.Query("min_price"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			filter.MinPrice = &f
		}
	}
	if v := c.Query("max_price"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			filter.MaxPrice = &f
		}
	}
	if v := c.Query("min_qty"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			filter.MinQty = &f
		}
	}

	filter.Sort = c.Query("sort")
	filter.Page, filter.Limit = parsePagination(c, 20)

	listings, total, err := h.listingService.Search(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to search listings"})
		return
	}

	totalPages := (total + filter.Limit - 1) / filter.Limit
	c.JSON(http.StatusOK, model.PaginatedResponse{
		Data: listings, Total: total, Page: filter.Page, Limit: filter.Limit, TotalPages: totalPages,
	})
}

func (h *MarketplaceHandler) GetPriceBoard(c *gin.Context) {
	board, err := h.listingService.GetPriceBoard(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get price board"})
		return
	}
	c.JSON(http.StatusOK, board)
}

func (h *MarketplaceHandler) GetProductCatalog(c *gin.Context) {
	catalog, err := h.catalogService.GetCatalogForAPI(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get product catalog"})
		return
	}
	c.JSON(http.StatusOK, catalog)
}

func (h *MarketplaceHandler) GetDetail(c *gin.Context) {
	id := c.Param("id")

	detail, err := h.listingService.GetDetail(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "listing not found"})
		return
	}

	c.JSON(http.StatusOK, detail)
}
