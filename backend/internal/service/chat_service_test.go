package service

import (
	"context"
	"testing"

	"github.com/sangiagao/rice-marketplace/internal/model"
	"github.com/sangiagao/rice-marketplace/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockConvRepo struct{ mock.Mock }

func (m *mockConvRepo) FindOrCreate(ctx context.Context, buyerID, sellerID string, listingID *string) (*model.Conversation, error) {
	args := m.Called(ctx, buyerID, sellerID, listingID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Conversation), args.Error(1)
}
func (m *mockConvRepo) GetByID(ctx context.Context, id string) (*model.Conversation, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Conversation), args.Error(1)
}
func (m *mockConvRepo) ListByUser(ctx context.Context, userID string, page, limit int) ([]*model.Conversation, int, error) {
	args := m.Called(ctx, userID, page, limit)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*model.Conversation), args.Int(1), args.Error(2)
}
func (m *mockConvRepo) SendMessage(ctx context.Context, conversationID, senderID, content, msgType string, replyToID *string) (*model.Message, error) {
	args := m.Called(ctx, conversationID, senderID, content, msgType, replyToID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Message), args.Error(1)
}
func (m *mockConvRepo) GetMessages(ctx context.Context, conversationID, readerID string, page, limit int) ([]*model.Message, int, error) {
	args := m.Called(ctx, conversationID, readerID, page, limit)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*model.Message), args.Int(1), args.Error(2)
}
func (m *mockConvRepo) MarkRead(ctx context.Context, conversationID, readerID string) error {
	args := m.Called(ctx, conversationID, readerID)
	return args.Error(0)
}
func (m *mockConvRepo) IsParticipant(ctx context.Context, conversationID, userID string) (bool, error) {
	args := m.Called(ctx, conversationID, userID)
	return args.Bool(0), args.Error(1)
}
func (m *mockConvRepo) GetMessageByID(ctx context.Context, messageID string) (*model.Message, error) {
	args := m.Called(ctx, messageID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Message), args.Error(1)
}
func (m *mockConvRepo) DeleteMessage(ctx context.Context, messageID string, asSender bool) error {
	return m.Called(ctx, messageID, asSender).Error(0)
}
func (m *mockConvRepo) DeleteMessages(ctx context.Context, messageIDs []string, asSender bool) error {
	return m.Called(ctx, messageIDs, asSender).Error(0)
}
func (m *mockConvRepo) RecallMessage(ctx context.Context, messageID string) error {
	return m.Called(ctx, messageID).Error(0)
}
func (m *mockConvRepo) RecallMessages(ctx context.Context, messageIDs []string) error {
	return m.Called(ctx, messageIDs).Error(0)
}
func (m *mockConvRepo) ToggleReaction(ctx context.Context, messageID, userID, emoji string) (bool, error) {
	args := m.Called(ctx, messageID, userID, emoji)
	return args.Bool(0), args.Error(1)
}
func (m *mockConvRepo) GetReactionsByMessage(ctx context.Context, messageID string) ([]model.MessageReaction, error) {
	args := m.Called(ctx, messageID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.MessageReaction), args.Error(1)
}

// --- Tests ---

func TestChatCreateConversation_Success(t *testing.T) {
	repo := new(mockConvRepo)
	svc := NewChatService(repo, nil)

	req := &model.CreateConversationRequest{SellerID: "seller-1"}
	repo.On("FindOrCreate", mock.Anything, "buyer-1", "seller-1", (*string)(nil)).
		Return(&model.Conversation{ID: "conv-1", MemberID: "buyer-1", SellerID: "seller-1"}, nil)

	conv, err := svc.CreateConversation(context.Background(), "buyer-1", req)
	assert.NoError(t, err)
	assert.Equal(t, "conv-1", conv.ID)
}

func TestChatCreateConversation_SelfChat(t *testing.T) {
	repo := new(mockConvRepo)
	svc := NewChatService(repo, nil)

	req := &model.CreateConversationRequest{SellerID: "user-1"}
	_, err := svc.CreateConversation(context.Background(), "user-1", req)
	assert.ErrorIs(t, err, ErrChatWithSelf)
}

func TestChatCreateConversation_WithListingID(t *testing.T) {
	repo := new(mockConvRepo)
	svc := NewChatService(repo, nil)

	listingID := "listing-1"
	req := &model.CreateConversationRequest{SellerID: "seller-1", ListingID: &listingID}
	repo.On("FindOrCreate", mock.Anything, "buyer-1", "seller-1", &listingID).
		Return(&model.Conversation{ID: "conv-2", ListingID: &listingID}, nil)

	conv, err := svc.CreateConversation(context.Background(), "buyer-1", req)
	assert.NoError(t, err)
	assert.Equal(t, &listingID, conv.ListingID)
}

func TestChatListConversations_Success(t *testing.T) {
	repo := new(mockConvRepo)
	svc := NewChatService(repo, nil)

	convs := []*model.Conversation{{ID: "conv-1"}, {ID: "conv-2"}}
	repo.On("ListByUser", mock.Anything, "user-1", 1, 20).Return(convs, 2, nil)

	result, total, err := svc.ListConversations(context.Background(), "user-1", 1, 20)
	assert.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, result, 2)
}

func TestChatListConversations_DefaultPagination(t *testing.T) {
	repo := new(mockConvRepo)
	svc := NewChatService(repo, nil)

	repo.On("ListByUser", mock.Anything, "user-1", 1, 20).Return([]*model.Conversation{}, 0, nil)

	_, _, err := svc.ListConversations(context.Background(), "user-1", 0, 0)
	assert.NoError(t, err)
	repo.AssertCalled(t, "ListByUser", mock.Anything, "user-1", 1, 20)
}

func TestChatSendMessage_Success(t *testing.T) {
	repo := new(mockConvRepo)
	svc := NewChatService(repo, nil)

	repo.On("IsParticipant", mock.Anything, "conv-1", "user-1").Return(true, nil)
	repo.On("SendMessage", mock.Anything, "conv-1", "user-1", "Hello", "text", (*string)(nil)).
		Return(&model.Message{ID: "msg-1", Content: "Hello", Type: "text"}, nil)

	req := &model.SendMessageRequest{Content: "Hello"}
	msg, err := svc.SendMessage(context.Background(), "user-1", "conv-1", req)
	assert.NoError(t, err)
	assert.Equal(t, "msg-1", msg.ID)
	assert.Equal(t, "text", msg.Type)
}

func TestChatSendMessage_NotParticipant(t *testing.T) {
	repo := new(mockConvRepo)
	svc := NewChatService(repo, nil)

	repo.On("IsParticipant", mock.Anything, "conv-1", "user-1").Return(false, nil)

	req := &model.SendMessageRequest{Content: "Hello"}
	_, err := svc.SendMessage(context.Background(), "user-1", "conv-1", req)
	assert.ErrorIs(t, err, ErrNotParticipant)
}

func TestChatSendMessage_ConversationNotFound(t *testing.T) {
	repo := new(mockConvRepo)
	svc := NewChatService(repo, nil)

	repo.On("IsParticipant", mock.Anything, "conv-999", "user-1").
		Return(false, repository.ErrConversationNotFound)

	req := &model.SendMessageRequest{Content: "Hello"}
	_, err := svc.SendMessage(context.Background(), "user-1", "conv-999", req)
	assert.ErrorIs(t, err, repository.ErrConversationNotFound)
}

func TestChatGetMessages_Success(t *testing.T) {
	repo := new(mockConvRepo)
	svc := NewChatService(repo, nil)

	repo.On("IsParticipant", mock.Anything, "conv-1", "user-1").Return(true, nil)
	repo.On("MarkRead", mock.Anything, "conv-1", "user-1").Return(nil)
	msgs := []*model.Message{{ID: "msg-1"}, {ID: "msg-2"}}
	repo.On("GetMessages", mock.Anything, "conv-1", "user-1", 1, 30).Return(msgs, 2, nil)

	result, total, err := svc.GetMessages(context.Background(), "user-1", "conv-1", 1, 30)
	assert.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, result, 2)
}

func TestChatGetMessages_NotParticipant(t *testing.T) {
	repo := new(mockConvRepo)
	svc := NewChatService(repo, nil)

	repo.On("IsParticipant", mock.Anything, "conv-1", "user-1").Return(false, nil)

	_, _, err := svc.GetMessages(context.Background(), "user-1", "conv-1", 1, 30)
	assert.ErrorIs(t, err, ErrNotParticipant)
}

func TestChatGetMessages_DefaultPagination(t *testing.T) {
	repo := new(mockConvRepo)
	svc := NewChatService(repo, nil)

	repo.On("IsParticipant", mock.Anything, "conv-1", "user-1").Return(true, nil)
	repo.On("MarkRead", mock.Anything, "conv-1", "user-1").Return(nil)
	repo.On("GetMessages", mock.Anything, "conv-1", "user-1", 1, 30).Return([]*model.Message{}, 0, nil)

	_, _, err := svc.GetMessages(context.Background(), "user-1", "conv-1", -1, 100)
	assert.NoError(t, err)
	repo.AssertCalled(t, "GetMessages", mock.Anything, "conv-1", "user-1", 1, 30)
}

func TestChatSendMessage_ImageType(t *testing.T) {
	repo := new(mockConvRepo)
	svc := NewChatService(repo, nil)

	repo.On("IsParticipant", mock.Anything, "conv-1", "user-1").Return(true, nil)
	repo.On("SendMessage", mock.Anything, "conv-1", "user-1", "https://img.com/photo.jpg", "image", (*string)(nil)).
		Return(&model.Message{ID: "msg-2", Type: "image"}, nil)

	req := &model.SendMessageRequest{Content: "https://img.com/photo.jpg", Type: "image"}
	msg, err := svc.SendMessage(context.Background(), "user-1", "conv-1", req)
	assert.NoError(t, err)
	assert.Equal(t, "image", msg.Type)
}
