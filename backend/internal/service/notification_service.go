package service

import (
	"context"
	"encoding/json"
	"log"

	"github.com/sangiagao/rice-marketplace/internal/model"
	"github.com/sangiagao/rice-marketplace/pkg/firebase"
)

type NotificationService struct {
	notifRepo  NotificationRepository
	pushSender firebase.PushSender
}

func NewNotificationService(notifRepo NotificationRepository, pushSender firebase.PushSender) *NotificationService {
	return &NotificationService{notifRepo: notifRepo, pushSender: pushSender}
}

func (s *NotificationService) RegisterDevice(ctx context.Context, userID, token, platform string) error {
	return s.notifRepo.RegisterDevice(ctx, userID, token, platform)
}

func (s *NotificationService) List(ctx context.Context, userID string, page, limit int) ([]*model.Notification, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 20
	}
	return s.notifRepo.ListByUser(ctx, userID, page, limit)
}

func (s *NotificationService) MarkRead(ctx context.Context, id, userID string) error {
	return s.notifRepo.MarkRead(ctx, id, userID)
}

func (s *NotificationService) Create(ctx context.Context, userID, nType, title, body string, data json.RawMessage) (*model.Notification, error) {
	notif, err := s.notifRepo.Create(ctx, userID, nType, title, body, data)
	if err != nil {
		return nil, err
	}

	// Send push notification asynchronously
	if s.pushSender != nil {
		go func() {
			tokens, err := s.notifRepo.GetDeviceTokens(context.Background(), userID)
			if err != nil || len(tokens) == 0 {
				return
			}
			pushData := map[string]string{
				"type":            nType,
				"notification_id": notif.ID,
			}
			if err := s.pushSender.SendToTokens(context.Background(), tokens, title, body, pushData); err != nil {
				log.Printf("Push notification error for user %s: %v", userID, err)
			}
		}()
	}

	return notif, nil
}

func (s *NotificationService) UnreadCount(ctx context.Context, userID string) (int, error) {
	return s.notifRepo.UnreadCount(ctx, userID)
}
