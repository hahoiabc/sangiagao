package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sangiagao/rice-marketplace/internal/model"
	wsPkg "github.com/sangiagao/rice-marketplace/internal/ws"
	jwtpkg "github.com/sangiagao/rice-marketplace/pkg/jwt"
)

type WSHandler struct {
	hub            *wsPkg.Hub
	jwtManager     *jwtpkg.Manager
	chatService    ChatServiceInterface
	upgrader       websocket.Upgrader
	allowedOrigins []string
}

func NewWSHandler(hub *wsPkg.Hub, jwtManager *jwtpkg.Manager, chatService ChatServiceInterface, allowedOrigins string) *WSHandler {
	h := &WSHandler{
		hub:         hub,
		jwtManager:  jwtManager,
		chatService: chatService,
	}

	// Parse allowed origins
	if allowedOrigins == "" || allowedOrigins == "*" {
		h.allowedOrigins = nil // nil means allow all (dev mode)
	} else {
		for _, o := range strings.Split(allowedOrigins, ",") {
			o = strings.TrimSpace(o)
			if o != "" {
				h.allowedOrigins = append(h.allowedOrigins, o)
			}
		}
	}

	h.upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			if h.allowedOrigins == nil {
				return true
			}
			origin := r.Header.Get("Origin")
			for _, allowed := range h.allowedOrigins {
				if origin == allowed {
					return true
				}
			}
			return false
		},
	}

	return h
}

// Connect handles WebSocket upgrade.
// Query params: token (JWT), conversation_id
func (h *WSHandler) Connect(c *gin.Context) {
	token := c.Query("token")
	conversationID := c.Query("conversation_id")

	if token == "" || conversationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing token or conversation_id"})
		return
	}

	claims, err := h.jwtManager.ValidateToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WS upgrade error: %v", err)
		return
	}

	client := &wsPkg.Client{
		UserID:         claims.UserID,
		ConversationID: conversationID,
		Conn:           conn,
		Send:           make(chan []byte, 256),
		Hub:            h.hub,
		OnMessage:      h.handleMessage,
	}

	h.hub.Join(conversationID, client)
	go client.WritePump()
	go client.ReadPump()
}

// handleMessage processes incoming WebSocket messages.
func (h *WSHandler) handleMessage(client *wsPkg.Client, rawMsg []byte) {
	var wsMsg wsPkg.WSMessage
	if err := json.Unmarshal(rawMsg, &wsMsg); err != nil {
		log.Printf("WS: invalid message format from user %s: %v", client.UserID, err)
		return
	}

	switch wsMsg.Event {
	case "send_message":
		h.handleSendMessage(client, wsMsg.Data)
	case "mark_read":
		h.handleMarkRead(client)
	default:
		log.Printf("WS: unknown event %q from user %s", wsMsg.Event, client.UserID)
	}
}

func (h *WSHandler) handleSendMessage(client *wsPkg.Client, data json.RawMessage) {
	var req model.SendMessageRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return
	}
	if req.Content == "" {
		return
	}

	msg, err := h.chatService.SendMessage(
		context.Background(),
		client.UserID,
		client.ConversationID,
		&req,
	)
	if err != nil {
		log.Printf("WS: send_message error: %v", err)
		return
	}

	h.hub.Broadcast(client.ConversationID, wsPkg.NewMessageEvent{
		Event:   "new_message",
		Message: msg,
	})
}

func (h *WSHandler) handleMarkRead(client *wsPkg.Client) {
	// The GetMessages service already marks as read; this is an explicit mark_read
	// We call the underlying chat service — but we need a MarkRead method.
	// For now, broadcast read receipt to the room.
	h.hub.Broadcast(client.ConversationID, wsPkg.ReadReceiptEvent{
		Event:          "read_receipt",
		ConversationID: client.ConversationID,
		ReaderID:       client.UserID,
	})
}
