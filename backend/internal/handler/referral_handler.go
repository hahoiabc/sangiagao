package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sangiagao/rice-marketplace/internal/service"
)

// ReferralHandler exposes endpoints for the affiliate / referral program.
type ReferralHandler struct {
	svc *service.ReferralService
}

func NewReferralHandler(svc *service.ReferralService) *ReferralHandler {
	return &ReferralHandler{svc: svc}
}

// GET /api/v1/me/referral
// Returns the caller's referral code + aggregated stats.
func (h *ReferralHandler) GetMyReferral(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	stats, err := h.svc.GetStats(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "không tải được dữ liệu giới thiệu"})
		return
	}
	c.JSON(http.StatusOK, stats)
}

// GET /api/v1/me/referral/history?limit=20&offset=0
func (h *ReferralHandler) GetMyHistory(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	records, err := h.svc.ListHistory(c.Request.Context(), userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "không tải được lịch sử"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": records})
}

// GET /api/v1/referral/resolve/:code
// Public endpoint used by web /r/{code} landing to validate a code before redirect.
func (h *ReferralHandler) Resolve(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "code required"})
		return
	}
	referrerID, err := h.svc.ResolveReferrer(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "lookup failed"})
		return
	}
	if referrerID == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "mã không tồn tại"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"valid": true})
}
