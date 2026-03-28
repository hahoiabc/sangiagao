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
			if err := s.pushSender.SendToTokens(context.Background(), tokens, title, body, "", pushData); err != nil {
				log.Printf("Push notification error for user %s: %v", userID, err)
			}
		}()
	}

	return notif, nil
}

func (s *NotificationService) SendPushOnly(ctx context.Context, userID, title, body string, data map[string]string) error {
	if s.pushSender == nil {
		return nil
	}
	tokens, err := s.notifRepo.GetDeviceTokens(ctx, userID)
	if err != nil || len(tokens) == 0 {
		return nil
	}
	return s.pushSender.SendToTokens(ctx, tokens, title, body, "", data)
}

func (s *NotificationService) UnreadCount(ctx context.Context, userID string) (int, error) {
	return s.notifRepo.UnreadCount(ctx, userID)
}

// BroadcastNotification creates a notification for all active users and sends push to all devices.
func (s *NotificationService) BroadcastNotification(ctx context.Context, nType, title, body, imageURL string, data json.RawMessage) (int, error) {
	userIDs, err := s.notifRepo.GetAllUserIDs(ctx)
	if err != nil {
		return 0, err
	}
	if len(userIDs) == 0 {
		return 0, nil
	}

	// Batch insert notifications for all users
	if err := s.notifRepo.CreateBatch(ctx, userIDs, nType, title, body, data); err != nil {
		return 0, err
	}

	// Send push notifications asynchronously in batches of 500
	if s.pushSender != nil {
		go func() {
			tokens, err := s.notifRepo.GetAllDeviceTokens(context.Background())
			if err != nil || len(tokens) == 0 {
				return
			}
			pushData := map[string]string{"type": nType}
			const batchSize = 500
			for i := 0; i < len(tokens); i += batchSize {
				end := i + batchSize
				if end > len(tokens) {
					end = len(tokens)
				}
				if err := s.pushSender.SendToTokens(context.Background(), tokens[i:end], title, body, imageURL, pushData); err != nil {
					log.Printf("Broadcast push error (batch %d-%d): %v", i, end, err)
				}
			}
			log.Printf("Broadcast push sent to %d devices for %d users", len(tokens), len(userIDs))
		}()
	}

	return len(userIDs), nil
}
