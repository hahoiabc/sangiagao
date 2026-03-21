package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sangiagao/rice-marketplace/internal/model"
	"github.com/sangiagao/rice-marketplace/internal/service"
)

type RatingHandler struct {
	ratingService RatingServiceInterface
}

func NewRatingHandler(ratingService RatingServiceInterface) *RatingHandler {
	return &RatingHandler{ratingService: ratingService}
}

func (h *RatingHandler) Create(c *gin.Context) {
	userID := c.GetString("user_id")

	var req model.CreateRatingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "seller_id, stars (1-5), and comment (min 10 chars) are required"})
		return
	}

	rating, err := h.ratingService.Create(c.Request.Context(), userID, &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrCannotRateSelf):
			c.JSON(http.StatusBadRequest, gin.H{"error": "cannot rate yourself"})
		case errors.Is(err, service.ErrTargetNotSeller):
			c.JSON(http.StatusBadRequest, gin.H{"error": "target user is not a seller"})
		case errors.Is(err, service.ErrAlreadyRated):
			c.JSON(http.StatusConflict, gin.H{"error": "you already rated this seller"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create rating"})
		}
		return
	}

	c.JSON(http.StatusCreated, rating)
}

func (h *RatingHandler) ListBySeller(c *gin.Context) {
	sellerID := c.Param("id")
	page, limit := parsePagination(c, 20)

	ratings, total, err := h.ratingService.ListBySeller(c.Request.Context(), sellerID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list ratings"})
		return
	}

	totalPages := (total + limit - 1) / limit
	c.JSON(http.StatusOK, model.PaginatedResponse{
		Data: ratings, Total: total, Page: page, Limit: limit, TotalPages: totalPages,
	})
}

func (h *RatingHandler) GetSummary(c *gin.Context) {
	sellerID := c.Param("id")

	summary, err := h.ratingService.GetSummary(c.Request.Context(), sellerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get rating summary"})
		return
	}

	c.JSON(http.StatusOK, summary)
}
