package handler

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sangiagao/rice-marketplace/internal/model"
)

type FeedbackHandler struct {
	service      FeedbackServiceInterface
	notifService NotificationServiceInterface
}

func NewFeedbackHandler(service FeedbackServiceInterface, notifService NotificationServiceInterface) *FeedbackHandler {
	return &FeedbackHandler{service: service, notifService: notifService}
}

func (h *FeedbackHandler) Create(c *gin.Context) {
	userID := c.GetString("user_id")

	var req model.CreateFeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nội dung góp ý là bắt buộc"})
		return
	}

	feedback, err := h.service.Create(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, feedback)
}

func (h *FeedbackHandler) ListMy(c *gin.Context) {
	userID := c.GetString("user_id")
	page, limit := parsePagination(c, 20)

	items, total, err := h.service.ListByUser(c.Request.Context(), userID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi tải góp ý"})
		return
	}

	totalPages := (total + limit - 1) / limit
	c.JSON(http.StatusOK, model.PaginatedResponse{
		Data: items, Total: total, Page: page, Limit: limit, TotalPages: totalPages,
	})
}

func (h *FeedbackHandler) List(c *gin.Context) {
	page, limit := parsePagination(c, 20)

	items, total, err := h.service.ListAll(c.Request.Context(), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi tải góp ý"})
		return
	}

	totalPages := (total + limit - 1) / limit
	c.JSON(http.StatusOK, model.PaginatedResponse{
		Data: items, Total: total, Page: page, Limit: limit, TotalPages: totalPages,
	})
}

func (h *FeedbackHandler) Reply(c *gin.Context) {
	id := c.Param("id")

	var req model.ReplyFeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nội dung phản hồi là bắt buộc"})
		return
	}

	feedback, err := h.service.Reply(c.Request.Context(), id, req.Reply)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Notify the feedback author about the reply
	if h.notifService != nil && feedback.UserID != "" {
		go func() {
			preview := req.Reply
			if len(preview) > 100 {
				preview = preview[:100] + "..."
			}
			if _, err := h.notifService.Create(context.Background(), feedback.UserID, "feedback_reply", "Phản hồi góp ý", preview, nil); err != nil {
				log.Printf("Failed to send feedback reply notification: %v", err)
			}
		}()
	}

	c.JSON(http.StatusOK, feedback)
}

func (h *FeedbackHandler) CountUnreplied(c *gin.Context) {
	count, err := h.service.CountUnreplied(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi đếm góp ý"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"count": count})
}
