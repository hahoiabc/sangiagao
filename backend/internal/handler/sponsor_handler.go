package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sangiagao/rice-marketplace/internal/model"
	"github.com/sangiagao/rice-marketplace/internal/service"
)

type SponsorHandler struct {
	sponsorService SponsorServiceInterface
}

func NewSponsorHandler(sponsorService SponsorServiceInterface) *SponsorHandler {
	return &SponsorHandler{sponsorService: sponsorService}
}

func (h *SponsorHandler) Create(c *gin.Context) {
	var req model.CreateSponsorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Vui lòng điền đầy đủ thông tin"})
		return
	}

	sponsor, err := h.sponsorService.Create(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidProductKey) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Tạo tài trợ thất bại"})
		return
	}

	c.JSON(http.StatusCreated, sponsor)
}

func (h *SponsorHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req model.UpdateSponsorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	sponsor, err := h.sponsorService.Update(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cập nhật tài trợ thất bại"})
		return
	}

	c.JSON(http.StatusOK, sponsor)
}

func (h *SponsorHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.sponsorService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Xóa tài trợ thất bại"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Đã xóa tài trợ"})
}

func (h *SponsorHandler) List(c *gin.Context) {
	page, limit := parsePagination(c, 20)

	sponsors, total, err := h.sponsorService.List(c.Request.Context(), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lấy danh sách tài trợ thất bại"})
		return
	}

	totalPages := (total + limit - 1) / limit
	c.JSON(http.StatusOK, model.PaginatedResponse{
		Data: sponsors, Total: total, Page: page, Limit: limit, TotalPages: totalPages,
	})
}
