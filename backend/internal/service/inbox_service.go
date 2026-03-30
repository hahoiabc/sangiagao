package service

import (
	"context"
	"log"

	"github.com/sangiagao/rice-marketplace/internal/model"
)

type InboxService struct {
	inboxRepo    InboxRepository
	notifService *NotificationService
}

func NewInboxService(inboxRepo InboxRepository, notifService *NotificationService) *InboxService {
	return &InboxService{inboxRepo: inboxRepo, notifService: notifService}
}

func (s *InboxService) Create(ctx context.Context, adminID string, req *model.CreateInboxRequest) (*model.InboxMessage, error) {
	msg, err := s.inboxRepo.Create(ctx, adminID, req)
	if err != nil {
		return nil, err
	}

	// Async: send light push "Bạn có thông báo mới" to target users
	if s.notifService != nil {
		go func() {
			bgCtx := context.Background()
			target := req.Target
			if target == "" {
				target = "all_users"
			}
			userIDs, err := s.inboxRepo.GetTargetUserIDs(bgCtx, target)
			if err != nil || len(userIDs) == 0 {
				return
			}
			pushData := map[string]string{"type": "system_inbox", "inbox_id": msg.ID}
			for _, uid := range userIDs {
				if err := s.notifService.SendPushOnly(bgCtx, uid, "Sàn Giá Gạo", "Bạn có thông báo mới", pushData); err != nil {
					log.Printf("Inbox push error for user %s: %v", uid, err)
				}
			}
			log.Printf("Inbox push sent to %d users for inbox %s", len(userIDs), msg.ID)
		}()
	}

	return msg, nil
}

func (s *InboxService) Update(ctx context.Context, id string, req *model.UpdateInboxRequest) (*model.InboxMessage, error) {
	return s.inboxRepo.Update(ctx, id, req)
}

func (s *InboxService) Delete(ctx context.Context, id string) error {
	return s.inboxRepo.Delete(ctx, id)
}

func (s *InboxService) ListForUser(ctx context.Context, userID, userRole string, page, limit int) ([]*model.InboxMessage, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 20
	}
	return s.inboxRepo.ListForUser(ctx, userID, userRole, page, limit)
}

func (s *InboxService) GetByID(ctx context.Context, id, userID string) (*model.InboxMessage, error) {
	return s.inboxRepo.GetByID(ctx, id, userID)
}

func (s *InboxService) MarkRead(ctx context.Context, userID, inboxID string) error {
	return s.inboxRepo.MarkRead(ctx, userID, inboxID)
}

func (s *InboxService) UnreadCount(ctx context.Context, userID, userRole string) (int, error) {
	return s.inboxRepo.UnreadCount(ctx, userID, userRole)
}

func (s *InboxService) ListAll(ctx context.Context, page, limit int) ([]*model.InboxMessage, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 20
	}
	return s.inboxRepo.ListAll(ctx, page, limit)
}
