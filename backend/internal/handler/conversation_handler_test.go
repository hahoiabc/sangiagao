package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sangiagao/rice-marketplace/internal/model"
	"github.com/sangiagao/rice-marketplace/internal/repository"
	"github.com/sangiagao/rice-marketplace/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockChatService struct{ mock.Mock }

func (m *mockChatService) CreateConversation(ctx context.Context, buyerID string, req *model.CreateConversationRequest) (*model.Conversation, error) {
	args := m.Called(ctx, buyerID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Conversation), args.Error(1)
}
func (m *mockChatService) ListConversations(ctx context.Context, userID string, page, limit int) ([]*model.Conversation, int, error) {
	args := m.Called(ctx, userID, page, limit)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*model.Conversation), args.Int(1), args.Error(2)
}
func (m *mockChatService) SendMessage(ctx context.Context, userID, conversationID string, req *model.SendMessageRequest) (*model.Message, error) {
	args := m.Called(ctx, userID, conversationID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Message), args.Error(1)
}
func (m *mockChatService) GetMessages(ctx context.Context, userID, conversationID string, page, limit int) ([]*model.Message, int, error) {
	args := m.Called(ctx, userID, conversationID, page, limit)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*model.Message), args.Int(1), args.Error(2)
}
func (m *mockChatService) DeleteMessage(ctx context.Context, userID, conversationID, messageID string) error {
	args := m.Called(ctx, userID, conversationID, messageID)
	return args.Error(0)
}
func (m *mockChatService) RecallMessage(ctx context.Context, userID, conversationID, messageID string) (*model.Message, error) {
	args := m.Called(ctx, userID, conversationID, messageID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Message), args.Error(1)
}
func (m *mockChatService) DeleteMessages(ctx context.Context, userID, conversationID string, messageIDs []string) error {
	args := m.Called(ctx, userID, conversationID, messageIDs)
	return args.Error(0)
}
func (m *mockChatService) RecallMessages(ctx context.Context, userID, conversationID string, messageIDs []string) error {
	args := m.Called(ctx, userID, conversationID, messageIDs)
	return args.Error(0)
}
func (m *mockChatService) GetConversation(ctx context.Context, id string) (*model.Conversation, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Conversation), args.Error(1)
}

func setupConvRouter(h *ConversationHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	g := r.Group("/api/v1")
	g.Use(func(c *gin.Context) {
		c.Set("user_id", "user-1")
		c.Next()
	})
	convs := g.Group("/conversations")
	convs.GET("", h.List)
	convs.POST("", h.Create)
	convs.GET("/:id/messages", h.GetMessages)
	convs.POST("/:id/messages", h.SendMessage)
	return r
}

func TestConvHandler_Create_Success(t *testing.T) {
	svc := new(mockChatService)
	h := NewConversationHandler(svc, nil)
	r := setupConvRouter(h)

	svc.On("CreateConversation", mock.Anything, "user-1", mock.AnythingOfType("*model.CreateConversationRequest")).
		Return(&model.Conversation{ID: "conv-1", BuyerID: "user-1", SellerID: "seller-1"}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/conversations", strings.NewReader(`{"seller_id":"seller-1"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var conv model.Conversation
	json.Unmarshal(w.Body.Bytes(), &conv)
	assert.Equal(t, "conv-1", conv.ID)
}

func TestConvHandler_Create_SelfChat(t *testing.T) {
	svc := new(mockChatService)
	h := NewConversationHandler(svc, nil)
	r := setupConvRouter(h)

	svc.On("CreateConversation", mock.Anything, "user-1", mock.Anything).
		Return(nil, service.ErrChatWithSelf)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/conversations", strings.NewReader(`{"seller_id":"user-1"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "chat_with_self")
}

func TestConvHandler_Create_InvalidBody(t *testing.T) {
	svc := new(mockChatService)
	h := NewConversationHandler(svc, nil)
	r := setupConvRouter(h)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/conversations", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestConvHandler_List_Success(t *testing.T) {
	svc := new(mockChatService)
	h := NewConversationHandler(svc, nil)
	r := setupConvRouter(h)

	convs := []*model.Conversation{{ID: "conv-1"}, {ID: "conv-2"}}
	svc.On("ListConversations", mock.Anything, "user-1", 1, 20).Return(convs, 2, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/conversations", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp model.PaginatedResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, 2, resp.Total)
}

func TestConvHandler_List_WithPagination(t *testing.T) {
	svc := new(mockChatService)
	h := NewConversationHandler(svc, nil)
	r := setupConvRouter(h)

	svc.On("ListConversations", mock.Anything, "user-1", 2, 10).Return([]*model.Conversation{}, 0, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/conversations?page=2&limit=10", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestConvHandler_GetMessages_Success(t *testing.T) {
	svc := new(mockChatService)
	h := NewConversationHandler(svc, nil)
	r := setupConvRouter(h)

	msgs := []*model.Message{{ID: "msg-1"}, {ID: "msg-2"}}
	svc.On("GetMessages", mock.Anything, "user-1", "conv-1", 1, 30).Return(msgs, 2, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/conversations/conv-1/messages", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp model.PaginatedResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, 2, resp.Total)
}

func TestConvHandler_GetMessages_NotFound(t *testing.T) {
	svc := new(mockChatService)
	h := NewConversationHandler(svc, nil)
	r := setupConvRouter(h)

	svc.On("GetMessages", mock.Anything, "user-1", "conv-999", 1, 30).
		Return(nil, 0, repository.ErrConversationNotFound)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/conversations/conv-999/messages", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestConvHandler_GetMessages_Forbidden(t *testing.T) {
	svc := new(mockChatService)
	h := NewConversationHandler(svc, nil)
	r := setupConvRouter(h)

	svc.On("GetMessages", mock.Anything, "user-1", "conv-1", 1, 30).
		Return(nil, 0, service.ErrNotParticipant)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/conversations/conv-1/messages", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestConvHandler_SendMessage_Success(t *testing.T) {
	svc := new(mockChatService)
	h := NewConversationHandler(svc, nil)
	r := setupConvRouter(h)

	svc.On("SendMessage", mock.Anything, "user-1", "conv-1", mock.AnythingOfType("*model.SendMessageRequest")).
		Return(&model.Message{ID: "msg-1", Content: "Hello"}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/conversations/conv-1/messages", strings.NewReader(`{"content":"Hello"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var msg model.Message
	json.Unmarshal(w.Body.Bytes(), &msg)
	assert.Equal(t, "msg-1", msg.ID)
}

func TestConvHandler_SendMessage_InvalidBody(t *testing.T) {
	svc := new(mockChatService)
	h := NewConversationHandler(svc, nil)
	r := setupConvRouter(h)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/conversations/conv-1/messages", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestConvHandler_SendMessage_NotFound(t *testing.T) {
	svc := new(mockChatService)
	h := NewConversationHandler(svc, nil)
	r := setupConvRouter(h)

	svc.On("SendMessage", mock.Anything, "user-1", "conv-999", mock.Anything).
		Return(nil, repository.ErrConversationNotFound)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/conversations/conv-999/messages", strings.NewReader(`{"content":"Hello"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestConvHandler_SendMessage_Forbidden(t *testing.T) {
	svc := new(mockChatService)
	h := NewConversationHandler(svc, nil)
	r := setupConvRouter(h)

	svc.On("SendMessage", mock.Anything, "user-1", "conv-1", mock.Anything).
		Return(nil, service.ErrNotParticipant)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/conversations/conv-1/messages", strings.NewReader(`{"content":"Hello"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}
