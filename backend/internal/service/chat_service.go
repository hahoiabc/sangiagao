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
	// Batch check online status + last seen (2 Redis calls instead of N*2)
	if s.cache != nil && len(convs) > 0 {
		var onlineKeys, lastseenKeys []string
		var userIndexes []int // track which convs have OtherUser
		for i, conv := range convs {
			if conv.OtherUser != nil {
				onlineKeys = append(onlineKeys, "online:"+conv.OtherUser.ID)
				lastseenKeys = append(lastseenKeys, "lastseen:"+conv.OtherUser.ID)
				userIndexes = append(userIndexes, i)
			}
		}
		if len(onlineKeys) > 0 {
			onlineVals, _ := s.cache.MGet(ctx, onlineKeys)
			lastseenVals, _ := s.cache.MGet(ctx, lastseenKeys)
			for j, ci := range userIndexes {
				online := onlineVals != nil && j < len(onlineVals) && onlineVals[j] != nil
				convs[ci].OtherUser.IsOnline = &online
				if !online && lastseenVals != nil && j < len(lastseenVals) && lastseenVals[j] != nil {
					if t, err := time.Parse(time.RFC3339, string(lastseenVals[j])); err == nil {
						convs[ci].OtherUser.LastSeenAt = &t
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
	return s.convRepo.SendMessage(ctx, conversationID, userID, req.Content, msgType, req.ReplyToID)
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
	// Verify user is participant (sender or receiver)
	ok, err := s.convRepo.IsParticipant(ctx, conversationID, userID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrNotParticipant
	}
	isSender := msg.SenderID == userID
	return s.convRepo.DeleteMessage(ctx, messageID, isSender)
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
	// Split messages into sender/receiver groups
	var senderIDs, receiverIDs []string
	for _, msgID := range messageIDs {
		msg, err := s.convRepo.GetMessageByID(ctx, msgID)
		if err != nil {
			return ErrMessageNotFound
		}
		if msg.ConversationID != conversationID {
			return ErrMessageNotFound
		}
		if msg.SenderID == userID {
			senderIDs = append(senderIDs, msgID)
		} else {
			receiverIDs = append(receiverIDs, msgID)
		}
	}
	if len(senderIDs) > 0 {
		if err := s.convRepo.DeleteMessages(ctx, senderIDs, true); err != nil {
			return err
		}
	}
	if len(receiverIDs) > 0 {
		if err := s.convRepo.DeleteMessages(ctx, receiverIDs, false); err != nil {
			return err
		}
	}
	return nil
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

func (s *ChatService) ToggleReaction(ctx context.Context, userID, conversationID, messageID, emoji string) ([]model.MessageReaction, error) {
	msg, err := s.convRepo.GetMessageByID(ctx, messageID)
	if err != nil {
		return nil, ErrMessageNotFound
	}
	if msg.ConversationID != conversationID {
		return nil, ErrMessageNotFound
	}
	ok, err := s.convRepo.IsParticipant(ctx, conversationID, userID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrNotParticipant
	}
	if _, err := s.convRepo.ToggleReaction(ctx, messageID, userID, emoji); err != nil {
		return nil, err
	}
	return s.convRepo.GetReactionsByMessage(ctx, messageID)
}

func (s *ChatService) DeleteConversation(ctx context.Context, userID, conversationID string) error {
	ok, err := s.convRepo.IsParticipant(ctx, conversationID, userID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrNotParticipant
	}
	return s.convRepo.DeleteConversation(ctx, conversationID, userID)
}

func (s *ChatService) TotalUnreadCount(ctx context.Context, userID string) (int, error) {
	return s.convRepo.TotalUnreadCount(ctx, userID)
}

func (s *ChatService) SearchUserByPhone(ctx context.Context, phone string) (*model.PublicProfile, error) {
	if s.userRepo == nil {
		return nil, errors.New("user repository not available")
	}
	user, err := s.userRepo.GetByPhone(ctx, phone)
	if err != nil {
		return nil, err
	}
	return &model.PublicProfile{
		ID:          user.ID,
		Role:        user.Role,
		Name:        user.Name,
		AvatarURL:   user.AvatarURL,
		Province:    user.Province,
		Description: user.Description,
		OrgName:     user.OrgName,
		CreatedAt:   user.CreatedAt,
	}, nil
}
