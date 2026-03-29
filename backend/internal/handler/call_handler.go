package handler

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sangiagao/rice-marketplace/internal/model"
)

type CallServiceInterface interface {
	InitiateCall(ctx context.Context, callerID, conversationID, calleeID, callType string) (*model.CallLog, error)
	AnswerCall(ctx context.Context, callID, userID string) error
	EndCall(ctx context.Context, callID, userID string) error
	RejectCall(ctx context.Context, callID, userID string) error
	MissCall(ctx context.Context, callID, userID string) error
	GetCallByID(ctx context.Context, callID string) (*model.CallLog, error)
	GetCallHistory(ctx context.Context, userID, conversationID string, page, limit int) ([]*model.CallLog, int, error)
}

type CallHandler struct {
	callService  CallServiceInterface
	notifService NotificationServiceInterface
	chatService  ChatServiceInterface
}

func NewCallHandler(callService CallServiceInterface, notifService NotificationServiceInterface, chatService ChatServiceInterface) *CallHandler {
	return &CallHandler{callService: callService, notifService: notifService, chatService: chatService}
}

func (h *CallHandler) InitiateCall(c *gin.Context) {
	userID := c.GetString("user_id")
	conversationID := c.Param("id")

	var req model.CreateCallLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "callee_id và call_type là bắt buộc"})
		return
	}

	call, err := h.callService.InitiateCall(c.Request.Context(), userID, conversationID, req.CalleeID, req.CallType)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Send FCM data-only push to callee to show incoming call UI
	go func() {
		callerName, _ := h.chatService.GetUserName(context.Background(), userID)
		if callerName == "" {
			callerName = "Người gọi"
		}
		pushData := map[string]string{
			"type":            "incoming_call",
			"call_id":         call.ID,
			"conversation_id": conversationID,
			"caller_id":       userID,
			"caller_name":     callerName,
			"call_type":       req.CallType,
		}
		if err := h.notifService.SendPushOnly(context.Background(), req.CalleeID, callerName, "Cuộc gọi đến", pushData); err != nil {
			log.Printf("Failed to send call push to %s: %v", req.CalleeID, err)
		}
	}()

	c.JSON(http.StatusCreated, call)
}

func (h *CallHandler) AnswerCall(c *gin.Context) {
	userID := c.GetString("user_id")
	callID := c.Param("call_id")

	if err := h.callService.AnswerCall(c.Request.Context(), callID, userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Đã trả lời cuộc gọi"})
}

func (h *CallHandler) EndCall(c *gin.Context) {
	userID := c.GetString("user_id")
	callID := c.Param("call_id")

	if err := h.callService.EndCall(c.Request.Context(), callID, userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Đã kết thúc cuộc gọi"})
}

func (h *CallHandler) RejectCall(c *gin.Context) {
	userID := c.GetString("user_id")
	callID := c.Param("call_id")

	if err := h.callService.RejectCall(c.Request.Context(), callID, userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Đã từ chối cuộc gọi"})
}

func (h *CallHandler) MissCall(c *gin.Context) {
	userID := c.GetString("user_id")
	callID := c.Param("call_id")

	if err := h.callService.MissCall(c.Request.Context(), callID, userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Send missed call push to callee
	go func() {
		call, err := h.callService.GetCallByID(c.Request.Context(), callID)
		if err != nil || call == nil {
			return
		}
		callerName, _ := h.chatService.GetUserName(context.Background(), call.CallerID)
		if callerName == "" {
			callerName = "Người gọi"
		}
		pushData := map[string]string{
			"type":            "missed_call",
			"call_id":         callID,
			"conversation_id": call.ConversationID,
			"caller_id":       call.CallerID,
			"caller_name":     callerName,
		}
		_ = h.notifService.SendPushOnly(context.Background(), call.CalleeID, callerName, "Cuộc gọi nhỡ", pushData)
	}()

	c.JSON(http.StatusOK, gin.H{"message": "Đã đánh dấu cuộc gọi nhỡ"})
}

func (h *CallHandler) GetCallHistory(c *gin.Context) {
	userID := c.GetString("user_id")
	conversationID := c.Param("id")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 20
	}

	calls, total, err := h.callService.GetCallHistory(c.Request.Context(), userID, conversationID, page, limit)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  calls,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// GetTURNCredentials returns time-limited TURN server credentials
func (h *CallHandler) GetTURNCredentials(c *gin.Context) {
	turnHost := os.Getenv("TURN_HOST")
	if turnHost == "" {
		turnHost = "sangiagao.vn"
	}
	turnPort := os.Getenv("TURN_PORT")
	if turnPort == "" {
		turnPort = "3478"
	}
	turnUser := os.Getenv("TURN_USER")
	if turnUser == "" {
		turnUser = "riceturn"
	}
	turnPass := os.Getenv("TURN_PASSWORD")
	if turnPass == "" {
		turnPass = "riceturn_secret"
	}

	c.JSON(http.StatusOK, gin.H{
		"ice_servers": []gin.H{
			{"urls": "stun:stun.l.google.com:19302"},
			{"urls": "stun:" + turnHost + ":" + turnPort},
			{
				"urls":       "turn:" + turnHost + ":" + turnPort,
				"username":   turnUser,
				"credential": turnPass,
			},
		},
		"ttl": 86400,
		"expires_at": time.Now().Add(24 * time.Hour).Unix(),
	})
}
