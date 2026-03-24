package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sangiagao/rice-marketplace/internal/model"
	"github.com/sangiagao/rice-marketplace/internal/repository"
	"github.com/sangiagao/rice-marketplace/internal/service"
)

type ListingHandler struct {
	listingService ListingServiceInterface
}

func NewListingHandler(listingService ListingServiceInterface) *ListingHandler {
	return &ListingHandler{listingService: listingService}
}

func (h *ListingHandler) Create(c *gin.Context) {
	userID := requireUserID(c)
	if c.IsAborted() {
		return
	}

	var req model.CreateListingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: category, rice_type, quantity_kg, price_per_kg are required"})
		return
	}

	listing, err := h.listingService.Create(c.Request.Context(), userID, &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCategory):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrInvalidProduct):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrDailyLimitReached):
			c.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create listing"})
		}
		return
	}

	c.JSON(http.StatusCreated, listing)
}

func (h *ListingHandler) BatchCreate(c *gin.Context) {
	userID := requireUserID(c)
	if c.IsAborted() {
		return
	}

	var items []model.CreateListingRequest
	if err := c.ShouldBindJSON(&items); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "expected array of listing objects"})
		return
	}
	if len(items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one listing is required"})
		return
	}
	if len(items) > 20 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "maximum 20 listings per batch"})
		return
	}

	ctx := c.Request.Context()
	var created []*model.Listing
	var errs []string
	for i := range items {
		listing, err := h.listingService.Create(ctx, userID, &items[i])
		if err != nil {
			errs = append(errs, err.Error())
		} else {
			created = append(created, listing)
		}
	}

	c.JSON(http.StatusCreated, gin.H{"created": created, "errors": errs})
}

func (h *ListingHandler) Get(c *gin.Context) {
	id := c.Param("id")

	listing, err := h.listingService.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrListingNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "listing not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get listing"})
		return
	}

	c.JSON(http.StatusOK, listing)
}

func (h *ListingHandler) Update(c *gin.Context) {
	userID := requireUserID(c)
	if c.IsAborted() {
		return
	}
	id := c.Param("id")

	var req model.UpdateListingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	listing, err := h.listingService.Update(c.Request.Context(), userID, id, &req)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrListingNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "listing not found"})
		case errors.Is(err, service.ErrNotListingOwner):
			c.JSON(http.StatusForbidden, gin.H{"error": "you don't own this listing"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update listing"})
		}
		return
	}

	c.JSON(http.StatusOK, listing)
}

func (h *ListingHandler) Delete(c *gin.Context) {
	userID := requireUserID(c)
	if c.IsAborted() {
		return
	}
	id := c.Param("id")

	err := h.listingService.Delete(c.Request.Context(), userID, id)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrListingNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "listing not found"})
		case errors.Is(err, service.ErrNotListingOwner):
			c.JSON(http.StatusForbidden, gin.H{"error": "you don't own this listing"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete listing"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "listing deleted"})
}

func (h *ListingHandler) ListMy(c *gin.Context) {
	userID := c.GetString("user_id")
	page, limit := parsePagination(c, 20)

	listings, total, err := h.listingService.ListByUser(c.Request.Context(), userID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list listings"})
		return
	}

	totalPages := (total + limit - 1) / limit
	c.JSON(http.StatusOK, model.PaginatedResponse{
		Data: listings, Total: total, Page: page, Limit: limit, TotalPages: totalPages,
	})
}

type addImageRequest struct {
	URL string `json:"url" binding:"required"`
}

func (h *ListingHandler) AddImage(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")

	var req addImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "url is required"})
		return
	}

	listing, err := h.listingService.AddImage(c.Request.Context(), userID, id, req.URL)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrListingNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "listing not found"})
		case errors.Is(err, service.ErrNotListingOwner):
			c.JSON(http.StatusForbidden, gin.H{"error": "you don't own this listing"})
		case errors.Is(err, service.ErrMaxImages):
			c.JSON(http.StatusConflict, gin.H{"error": "Tối đa 3 ảnh cho mỗi tin đăng"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add image"})
		}
		return
	}

	c.JSON(http.StatusOK, listing)
}
