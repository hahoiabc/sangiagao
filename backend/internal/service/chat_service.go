package service

import (
	"context"
	"errors"
	"time"

	"github.com/sangiagao/rice-marketplace/internal/model"
	"github.com/sangiagao/rice-marketplace/internal/repository"
	"github.com/sangiagao/rice-marketplace/pkg/cache"
)

var (
	ErrChatWithSelf    = errors.New("cannot chat with yourself")
	ErrNotParticipant  = errors.New("not a participant in this conversation")
	ErrNotMessageOwner = errors.New("bạn không phải chủ tin nhắn này")
	ErrMessageNotFound = errors.New("tin nhắn không tồn tại")
	ErrRecallExpired   = errors.New("chỉ được thu hồi tin nhắn trong vòng 24 giờ")
)

type ChatService struct {
	convRepo ConversationRepository
	userRepo UserRepository
	cache    cache.Cache
}

func NewChatService(convRepo ConversationRepository, userRepo UserRepository) *ChatService {
	return &ChatService{convRepo: convRepo, userRepo: userRepo}
}

func (s *ChatService) SetCache(c cache.Cache) {
	s.cache = c
}

func (s *ChatService) GetConversation(ctx context.Context, id string) (*model.Conversation, error) {
	return s.convRepo.GetByID(ctx, id)
}

func (s *ChatService) CreateConversation(ctx context.Context, buyerID string, req *model.CreateConversationRequest) (*model.Conversation, error) {
	if buyerID == req.SellerID {
		return nil, ErrChatWithSelf
	}
	return s.convRepo.FindOrCreate(ctx, buyerID, req.SellerID, req.ListingID)
}

func (s *ChatService) ListConversations(ctx context.Context, userID string, page, limit int) ([]*model.Conversation, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 20
	}
	convs, total, err := s.convRepo.ListByUser(ctx, userID, page, limit)
	if err != nil {
		return nil, 0, err
	}
	// Check online status + last seen for other users
	if s.cache != nil {
		for _, conv := range convs {
			if conv.OtherUser != nil {
				online, cErr := s.cache.Exists(ctx, "online:"+conv.OtherUser.ID)
				if cErr == nil {
					conv.OtherUser.IsOnline = &online
				}
				if !online {
					if raw, err := s.cache.Get(ctx, "lastseen:"+conv.OtherUser.ID); err == nil && raw != nil {
						if t, tErr := time.Parse(time.RFC3339, string(raw)); tErr == nil {
							conv.OtherUser.LastSeenAt = &t
						}
					}
				}
			}
		}
	}
	return convs, total, nil
}

func (s *ChatService) SendMessage(ctx context.Context, userID, conversationID string, req *model.SendMessageRequest) (*model.Message, error) {
	// Check if sender is blocked
	if s.userRepo != nil {
		user, err := s.userRepo.GetByID(ctx, userID)
		if err != nil {
			return nil, err
		}
		if user.IsBlocked {
			return nil, ErrUserBlocked
		}
	}

	ok, err := s.convRepo.IsParticipant(ctx, conversationID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrConversationNotFound) {
			return nil, repository.ErrConversationNotFound
		}
		return nil, err
	}
	if !ok {
		return nil, ErrNotParticipant
	}

	msgType := req.Type
	if msgType == "" {
		msgType = "text"
	}
	return s.convRepo.SendMessage(ctx, conversationID, userID, req.Content, msgType)
}

func (s *ChatService) GetMessages(ctx context.Context, userID, conversationID string, page, limit int) ([]*model.Message, int, error) {
	ok, err := s.convRepo.IsParticipant(ctx, conversationID, userID)
	if err != nil {
		return nil, 0, err
	}
	if !ok {
		return nil, 0, ErrNotParticipant
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 30
	}

	// Mark as read
	_ = s.convRepo.MarkRead(ctx, conversationID, userID)

	return s.convRepo.GetMessages(ctx, conversationID, userID, page, limit)
}

func (s *ChatService) MarkConversationRead(ctx context.Context, userID, conversationID string) error {
	ok, err := s.convRepo.IsParticipant(ctx, conversationID, userID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrNotParticipant
	}
	return s.convRepo.MarkRead(ctx, conversationID, userID)
}

func (s *ChatService) DeleteMessage(ctx context.Context, userID, conversationID, messageID string) error {
	msg, err := s.convRepo.GetMessageByID(ctx, messageID)
	if err != nil {
		return ErrMessageNotFound
	}
	if msg.ConversationID != conversationID {
		return ErrMessageNotFound
	}
	if msg.SenderID != userID {
		return ErrNotMessageOwner
	}
	return s.convRepo.DeleteMessage(ctx, messageID)
}

func (s *ChatService) RecallMessage(ctx context.Context, userID, conversationID, messageID string) (*model.Message, error) {
	msg, err := s.convRepo.GetMessageByID(ctx, messageID)
	if err != nil {
		return nil, ErrMessageNotFound
	}
	if msg.ConversationID != conversationID {
		return nil, ErrMessageNotFound
	}
	if msg.SenderID != userID {
		return nil, ErrNotMessageOwner
	}
	if time.Since(msg.CreatedAt) > 24*time.Hour {
		return nil, ErrRecallExpired
	}
	if err := s.convRepo.RecallMessage(ctx, messageID); err != nil {
		return nil, err
	}
	return s.convRepo.GetMessageByID(ctx, messageID)
}

func (s *ChatService) DeleteMessages(ctx context.Context, userID, conversationID string, messageIDs []string) error {
	ok, err := s.convRepo.IsParticipant(ctx, conversationID, userID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrNotParticipant
	}
	// Verify ownership of each message
	for _, msgID := range messageIDs {
		msg, err := s.convRepo.GetMessageByID(ctx, msgID)
		if err != nil {
			return ErrMessageNotFound
		}
		if msg.ConversationID != conversationID || msg.SenderID != userID {
			return ErrNotMessageOwner
		}
	}
	return s.convRepo.DeleteMessages(ctx, messageIDs)
}

func (s *ChatService) IsParticipant(ctx context.Context, conversationID, userID string) (bool, error) {
	return s.convRepo.IsParticipant(ctx, conversationID, userID)
}

func (s *ChatService) GetUserName(ctx context.Context, userID string) (string, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return "", err
	}
	if user.Name != nil {
		return *user.Name, nil
	}
	return "Người dùng", nil
}

func (s *ChatService) RecallMessages(ctx context.Context, userID, conversationID string, messageIDs []string) error {
	ok, err := s.convRepo.IsParticipant(ctx, conversationID, userID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrNotParticipant
	}
	// Verify ownership and recall window for each message
	for _, msgID := range messageIDs {
		msg, err := s.convRepo.GetMessageByID(ctx, msgID)
		if err != nil {
			return ErrMessageNotFound
		}
		if msg.ConversationID != conversationID || msg.SenderID != userID {
			return ErrNotMessageOwner
		}
		if time.Since(msg.CreatedAt) > 24*time.Hour {
			return ErrRecallExpired
		}
	}
	return s.convRepo.RecallMessages(ctx, messageIDs)
}
