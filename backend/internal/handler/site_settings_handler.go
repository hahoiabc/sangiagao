package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sangiagao/rice-marketplace/internal/model"
)

type SiteSettingsHandler struct {
	service SiteSettingsServiceInterface
}

func NewSiteSettingsHandler(service SiteSettingsServiceInterface) *SiteSettingsHandler {
	return &SiteSettingsHandler{service: service}
}

// GetSlogan — public, no auth required
func (h *SiteSettingsHandler) GetSlogan(c *gin.Context) {
	setting, err := h.service.GetSlogan(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể lấy slogan"})
		return
	}
	c.JSON(http.StatusOK, setting)
}

// UpdateSlogan — admin only
func (h *SiteSettingsHandler) UpdateSlogan(c *gin.Context) {
	var req model.UpdateSiteSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nội dung slogan không hợp lệ (tối đa 500 ký tự)"})
		return
	}

	setting, err := h.service.UpdateSlogan(c.Request.Context(), req.Value)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cập nhật slogan thất bại"})
		return
	}

	c.JSON(http.StatusOK, setting)
}

// GetSloganColor — public, no auth required
func (h *SiteSettingsHandler) GetSloganColor(c *gin.Context) {
	setting, err := h.service.GetSloganColor(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể lấy màu slogan"})
		return
	}
	c.JSON(http.StatusOK, setting)
}

// UpdateSloganColor — admin only
func (h *SiteSettingsHandler) UpdateSloganColor(c *gin.Context) {
	var req model.UpdateSiteSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Mã màu không hợp lệ"})
		return
	}

	setting, err := h.service.UpdateSloganColor(c.Request.Context(), req.Value)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cập nhật màu slogan thất bại"})
		return
	}

	c.JSON(http.StatusOK, setting)
}
