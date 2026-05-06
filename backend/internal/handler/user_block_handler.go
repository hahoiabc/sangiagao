package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sangiagao/rice-marketplace/internal/service"
)

type UserBlockHandler struct {
	svc *service.UserBlockService
}

func NewUserBlockHandler(svc *service.UserBlockService) *UserBlockHandler {
	return &UserBlockHandler{svc: svc}
}

type meBlockRequest struct {
	BlockedID string `json:"blocked_id" binding:"required"`
	Reason    string `json:"reason"`
}

// Block: POST /api/v1/me/blocks
func (h *UserBlockHandler) Block(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req meBlockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "blocked_id is required"})
		return
	}
	if err := h.svc.Block(c.Request.Context(), userID, req.BlockedID, req.Reason); err != nil {
		switch {
		case errors.Is(err, service.ErrCannotBlockSelf):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "blocked"})
}

// Unblock: DELETE /api/v1/me/blocks/:blocked_id
func (h *UserBlockHandler) Unblock(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	blockedID := c.Param("blocked_id")
	if blockedID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "blocked_id required"})
		return
	}
	if err := h.svc.Unblock(c.Request.Context(), userID, blockedID); err != nil {
		switch {
		case errors.Is(err, service.ErrNotBlocked):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "unblocked"})
}

// List: GET /api/v1/me/blocks
func (h *UserBlockHandler) List(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	out, err := h.svc.List(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": out, "total": len(out)})
}
