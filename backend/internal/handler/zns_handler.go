package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sangiagao/rice-marketplace/pkg/sms"
)

type ZNSHandler struct {
	sender *sms.ZaloZNSSender
}

func NewZNSHandler(sender *sms.ZaloZNSSender) *ZNSHandler {
	return &ZNSHandler{sender: sender}
}

func (h *ZNSHandler) GetStatus(c *gin.Context) {
	if h.sender == nil {
		c.JSON(http.StatusOK, gin.H{
			"enabled": false,
			"message": "Zalo ZNS chưa được cấu hình (SMS_PROVIDER != zalo/zalo+mock)",
		})
		return
	}

	status := h.sender.Status()
	status["enabled"] = true
	c.JSON(http.StatusOK, status)
}

func (h *ZNSHandler) UpdateRefreshToken(c *gin.Context) {
	if h.sender == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Zalo ZNS chưa được cấu hình"})
		return
	}

	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required,min=10"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "refresh_token là bắt buộc (ít nhất 10 ký tự)"})
		return
	}

	if err := h.sender.UpdateRefreshToken(req.RefreshToken); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lưu refresh token thất bại: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Cập nhật refresh token thành công. Token mới sẽ được sử dụng cho lần gửi OTP tiếp theo.",
	})
}

func (h *ZNSHandler) TestSend(c *gin.Context) {
	if h.sender == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Zalo ZNS chưa được cấu hình"})
		return
	}

	var req struct {
		Phone string `json:"phone" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "phone là bắt buộc"})
		return
	}

	if err := h.sender.SendOTP(req.Phone, "123456"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gửi test OTP thất bại",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Gửi test OTP thành công tới " + req.Phone + " (mã: 123456)",
	})
}
