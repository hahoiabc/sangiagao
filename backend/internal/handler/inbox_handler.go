package handler

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sangiagao/rice-marketplace/internal/model"
	"github.com/sangiagao/rice-marketplace/internal/repository"
)

type InboxHandler struct {
	inboxService InboxServiceInterface
}

func NewInboxHandler(inboxService InboxServiceInterface) *InboxHandler {
	return &InboxHandler{inboxService: inboxService}
}

// --- Public API (auth required) ---

func (h *InboxHandler) List(c *gin.Context) {
	userID := c.GetString("user_id")
	userRole := c.GetString("user_role")
	page, limit := parsePagination(c, 20)

	items, total, err := h.inboxService.ListForUser(c.Request.Context(), userID, userRole, page, limit)
	if err != nil {
		log.Printf("[INBOX] ListForUser error: userID=%s role=%s err=%v", userID, userRole, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal", "message": "Lỗi tải hộp thư"})
		return
	}

	unread, _ := h.inboxService.UnreadCount(c.Request.Context(), userID, userRole)

	totalPages := (total + limit - 1) / limit
	c.JSON(http.StatusOK, gin.H{
		"data":         items,
		"total":        total,
		"page":         page,
		"limit":        limit,
		"total_pages":  totalPages,
		"unread_count": unread,
	})
}

func (h *InboxHandler) GetByID(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")

	msg, err := h.inboxService.GetByID(c.Request.Context(), id, userID)
	if err != nil {
		if errors.Is(err, repository.ErrInboxNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "Không tìm thấy thông báo"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal", "message": "Lỗi tải thông báo"})
		return
	}

	// Auto mark read
	_ = h.inboxService.MarkRead(c.Request.Context(), userID, id)
	msg.IsRead = true

	c.JSON(http.StatusOK, msg)
}

func (h *InboxHandler) MarkRead(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")

	if err := h.inboxService.MarkRead(c.Request.Context(), userID, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal", "message": "Lỗi đánh dấu đã đọc"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func (h *InboxHandler) UnreadCount(c *gin.Context) {
	userID := c.GetString("user_id")
	userRole := c.GetString("user_role")

	count, err := h.inboxService.UnreadCount(c.Request.Context(), userID, userRole)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal", "message": "Lỗi đếm chưa đọc"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"unread_count": count})
}

// --- Admin API ---

func (h *InboxHandler) AdminList(c *gin.Context) {
	page, limit := parsePagination(c, 20)

	items, total, err := h.inboxService.ListAll(c.Request.Context(), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal", "message": "Lỗi tải danh sách"})
		return
	}

	totalPages := (total + limit - 1) / limit
	c.JSON(http.StatusOK, model.PaginatedResponse{
		Data:       items,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	})
}

func (h *InboxHandler) AdminCreate(c *gin.Context) {
	adminID := c.GetString("user_id")

	var req model.CreateInboxRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "Tiêu đề và nội dung là bắt buộc"})
		return
	}

	msg, err := h.inboxService.Create(c.Request.Context(), adminID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal", "message": "Tạo thông báo thất bại"})
		return
	}

	c.JSON(http.StatusCreated, msg)
}

func (h *InboxHandler) AdminUpdate(c *gin.Context) {
	id := c.Param("id")

	var req model.UpdateInboxRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
		return
	}

	msg, err := h.inboxService.Update(c.Request.Context(), id, &req)
	if err != nil {
		if errors.Is(err, repository.ErrInboxNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "Không tìm thấy thông báo"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal", "message": "Cập nhật thất bại"})
		return
	}

	c.JSON(http.StatusOK, msg)
}

func (h *InboxHandler) AdminDelete(c *gin.Context) {
	id := c.Param("id")

	if err := h.inboxService.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, repository.ErrInboxNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "Không tìm thấy thông báo"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal", "message": "Xóa thất bại"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Đã xóa thông báo"})
}
