package handler

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sangiagao/rice-marketplace/internal/model"
	"github.com/sangiagao/rice-marketplace/internal/repository"
	"github.com/sangiagao/rice-marketplace/internal/service"
	"github.com/sangiagao/rice-marketplace/pkg/workerpool"
)

type ConversationHandler struct {
	chatService  ChatServiceInterface
	notifService NotificationServiceInterface
	pool         *workerpool.Pool
}

func NewConversationHandler(chatService ChatServiceInterface, notifService NotificationServiceInterface) *ConversationHandler {
	return &ConversationHandler{chatService: chatService, notifService: notifService}
}

func (h *ConversationHandler) SetPool(p *workerpool.Pool) {
	h.pool = p
}

func (h *ConversationHandler) Create(c *gin.Context) {
	userID := requireUserID(c)
	if c.IsAborted() {
		return
	}
	var req model.CreateConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
		return
	}

	conv, err := h.chatService.CreateConversation(c.Request.Context(), userID, &req)
	if err != nil {
		switch err {
		case service.ErrChatWithSelf:
			c.JSON(http.StatusBadRequest, gin.H{"error": "chat_with_self", "message": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal", "message": "failed to create conversation"})
		}
		return
	}

	c.JSON(http.StatusCreated, conv)
}

func (h *ConversationHandler) List(c *gin.Context) {
	userID := c.GetString("user_id")
	page, limit := parsePagination(c, 20)

	convs, total, err := h.chatService.ListConversations(c.Request.Context(), userID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal", "message": "failed to list conversations"})
		return
	}

	totalPages := (total + limit - 1) / limit
	c.JSON(http.StatusOK, model.PaginatedResponse{
		Data:       convs,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	})
}

func (h *ConversationHandler) GetMessages(c *gin.Context) {
	userID := c.GetString("user_id")
	conversationID := c.Param("id")
	page, limit := parsePagination(c, 30)

	messages, total, err := h.chatService.GetMessages(c.Request.Context(), userID, conversationID, page, limit)
	if err != nil {
		switch err {
		case repository.ErrConversationNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "conversation not found"})
		case service.ErrNotParticipant:
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden", "message": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal", "message": "failed to get messages"})
		}
		return
	}

	totalPages := (total + limit - 1) / limit
	c.JSON(http.StatusOK, model.PaginatedResponse{
		Data:       messages,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	})
}

func (h *ConversationHandler) SendMessage(c *gin.Context) {
	userID := c.GetString("user_id")
	conversationID := c.Param("id")

	var req model.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
		return
	}

	msg, err := h.chatService.SendMessage(c.Request.Context(), userID, conversationID, &req)
	if err != nil {
		switch err {
		case repository.ErrConversationNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "conversation not found"})
		case service.ErrNotParticipant:
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden", "message": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal", "message": "failed to send message"})
		}
		return
	}

	// Send push notification (FCM only, no DB record) to the other participant
	if h.notifService != nil {
		submitFn := func(fn func()) {
			if h.pool != nil {
				h.pool.Submit(fn)
			} else {
				go fn()
			}
		}
		submitFn(func() {
			ctx := context.Background()
			conv, err := h.chatService.GetConversation(ctx, conversationID)
			if err != nil {
				return
			}
			recipientID := conv.MemberID
			if recipientID == userID {
				recipientID = conv.SellerID
			}

			// Use sender name as push title (like Zalo)
			senderName, err := h.chatService.GetUserName(ctx, userID)
			if err != nil {
				senderName = "Tin nhắn mới"
			}

			preview := req.Content
			msgType := req.Type
			if msgType == "" {
				msgType = "text"
			}
			switch msgType {
			case "image":
				preview = "[Hình ảnh]"
			case "audio":
				preview = "[Tin nhắn thoại]"
			case "listing_link":
				preview = "[Tin đăng]"
			default:
				if len(preview) > 100 {
					preview = preview[:100] + "..."
				}
			}
			pushData := map[string]string{
				"conversation_id": conversationID,
				"type":            "new_message",
				"sender_id":       userID,
			}
			if err := h.notifService.SendPushOnly(ctx, recipientID, senderName, preview, pushData); err != nil {
				log.Printf("Failed to send chat push: %v", err)
			}
		})
	}

	c.JSON(http.StatusCreated, msg)
}

func (h *ConversationHandler) MarkRead(c *gin.Context) {
	userID := c.GetString("user_id")
	conversationID := c.Param("id")

	err := h.chatService.MarkConversationRead(c.Request.Context(), userID, conversationID)
	if err != nil {
		switch err {
		case service.ErrNotParticipant:
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden", "message": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal", "message": "failed to mark read"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func (h *ConversationHandler) DeleteMessage(c *gin.Context) {
	userID := c.GetString("user_id")
	conversationID := c.Param("id")
	messageID := c.Param("msgId")

	err := h.chatService.DeleteMessage(c.Request.Context(), userID, conversationID, messageID)
	if err != nil {
		switch err {
		case service.ErrMessageNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": err.Error()})
		case service.ErrNotParticipant:
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden", "message": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal", "message": "xóa tin nhắn thất bại"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Đã xóa tin nhắn"})
}

func (h *ConversationHandler) RecallMessage(c *gin.Context) {
	userID := c.GetString("user_id")
	conversationID := c.Param("id")
	messageID := c.Param("msgId")

	msg, err := h.chatService.RecallMessage(c.Request.Context(), userID, conversationID, messageID)
	if err != nil {
		switch err {
		case service.ErrMessageNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": err.Error()})
		case service.ErrNotMessageOwner:
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden", "message": err.Error()})
		case service.ErrRecallExpired:
			c.JSON(http.StatusBadRequest, gin.H{"error": "recall_expired", "message": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal", "message": "thu hồi tin nhắn thất bại"})
		}
		return
	}

	c.JSON(http.StatusOK, msg)
}

func (h *ConversationHandler) BatchDeleteMessages(c *gin.Context) {
	userID := c.GetString("user_id")
	conversationID := c.Param("id")

	var req struct {
		MessageIDs []string `json:"message_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
		return
	}

	if len(req.MessageIDs) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "too_many_items", "message": "tối đa 100 tin nhắn mỗi lần"})
		return
	}

	err := h.chatService.DeleteMessages(c.Request.Context(), userID, conversationID, req.MessageIDs)
	if err != nil {
		switch err {
		case service.ErrMessageNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": err.Error()})
		case service.ErrNotParticipant:
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden", "message": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal", "message": "xóa tin nhắn thất bại"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Đã xóa tin nhắn"})
}

func (h *ConversationHandler) BatchRecallMessages(c *gin.Context) {
	userID := c.GetString("user_id")
	conversationID := c.Param("id")

	var req struct {
		MessageIDs []string `json:"message_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
		return
	}

	if len(req.MessageIDs) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "too_many_items", "message": "tối đa 100 tin nhắn mỗi lần"})
		return
	}

	err := h.chatService.RecallMessages(c.Request.Context(), userID, conversationID, req.MessageIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal", "message": "thu hồi tin nhắn thất bại"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Đã thu hồi tin nhắn"})
}

func (h *ConversationHandler) ToggleReaction(c *gin.Context) {
	userID := c.GetString("user_id")
	conversationID := c.Param("id")
	messageID := c.Param("msgId")

	var req model.ToggleReactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
		return
	}

	reactions, err := h.chatService.ToggleReaction(c.Request.Context(), userID, conversationID, messageID, req.Emoji)
	if err != nil {
		switch err {
		case service.ErrMessageNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": err.Error()})
		case service.ErrNotParticipant:
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden", "message": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal", "message": "thả cảm xúc thất bại"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"reactions": reactions})
}

func (h *ConversationHandler) DeleteConversation(c *gin.Context) {
	userID := c.GetString("user_id")
	conversationID := c.Param("id")

	err := h.chatService.DeleteConversation(c.Request.Context(), userID, conversationID)
	if err != nil {
		switch err {
		case repository.ErrConversationNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "cuộc trò chuyện không tồn tại"})
		case service.ErrNotParticipant:
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden", "message": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal", "message": "xóa cuộc trò chuyện thất bại"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Đã xóa cuộc trò chuyện"})
}

func (h *ConversationHandler) UnreadTotal(c *gin.Context) {
	userID := c.GetString("user_id")
	total, err := h.chatService.TotalUnreadCount(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal", "message": "failed to get unread count"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"total": total})
}

func (h *ConversationHandler) SearchByPhone(c *gin.Context) {
	phone := c.Query("phone")
	if phone == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "phone is required"})
		return
	}

	profile, err := h.chatService.SearchUserByPhone(c.Request.Context(), phone)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "không tìm thấy người dùng"})
		return
	}

	c.JSON(http.StatusOK, profile)
}
