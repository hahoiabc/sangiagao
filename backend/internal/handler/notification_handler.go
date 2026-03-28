package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sangiagao/rice-marketplace/internal/model"
	"github.com/sangiagao/rice-marketplace/internal/repository"
)

type NotificationHandler struct {
	notifService NotificationServiceInterface
}

func NewNotificationHandler(notifService NotificationServiceInterface) *NotificationHandler {
	return &NotificationHandler{notifService: notifService}
}

func (h *NotificationHandler) RegisterDevice(c *gin.Context) {
	userID := c.GetString("user_id")

	var req model.RegisterDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token and platform (ios/android) are required"})
		return
	}

	if err := h.notifService.RegisterDevice(c.Request.Context(), userID, req.Token, req.Platform); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register device"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "device registered"})
}

func (h *NotificationHandler) List(c *gin.Context) {
	userID := c.GetString("user_id")
	page, limit := parsePagination(c, 20)

	notifications, total, err := h.notifService.List(c.Request.Context(), userID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list notifications"})
		return
	}

	unread, _ := h.notifService.UnreadCount(c.Request.Context(), userID)

	totalPages := (total + limit - 1) / limit
	c.JSON(http.StatusOK, gin.H{
		"data":        notifications,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": totalPages,
		"unread":      unread,
	})
}

func (h *NotificationHandler) MarkRead(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")

	if err := h.notifService.MarkRead(c.Request.Context(), id, userID); err != nil {
		if errors.Is(err, repository.ErrNotificationNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "notification not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to mark notification as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "notification marked as read"})
}

type broadcastRequest struct {
	Title    string          `json:"title" binding:"required,max=200"`
	Body     string          `json:"body" binding:"required"`
	ImageURL string          `json:"image_url,omitempty"`
	Data     json.RawMessage `json:"data,omitempty"`
}

func (h *NotificationHandler) Broadcast(c *gin.Context) {
	var req broadcastRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tiêu đề và nội dung là bắt buộc"})
		return
	}

	count, err := h.notifService.BroadcastNotification(c.Request.Context(), "broadcast", req.Title, req.Body, req.ImageURL, req.Data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gửi thông báo thất bại"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Đã gửi thông báo",
		"sent_to": count,
	})
}
