package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/sangiagao/rice-marketplace/internal/model"
	"github.com/sangiagao/rice-marketplace/internal/repository"
)

var ErrNoSubscription = errors.New("no subscription found")

// NotificationCreator allows sending notifications without importing the full notification service.
type NotificationCreator interface {
	Create(ctx context.Context, userID, nType, title, body string, data json.RawMessage) (*model.Notification, error)
}

type SubscriptionService struct {
	subRepo   SubscriptionRepository
	planRepo  PlanRepository
	notifier  NotificationCreator
	onExpiry  func(ctx context.Context) // called when listings are hidden due to expiry
}

func NewSubscriptionService(subRepo SubscriptionRepository, planRepo PlanRepository) *SubscriptionService {
	return &SubscriptionService{subRepo: subRepo, planRepo: planRepo}
}

func (s *SubscriptionService) SetNotifier(n NotificationCreator) {
	s.notifier = n
}

func (s *SubscriptionService) SetOnExpiry(fn func(ctx context.Context)) {
	s.onExpiry = fn
}

type SubscriptionStatus struct {
	Subscription *model.Subscription `json:"subscription,omitempty"`
	DaysLeft     int                 `json:"days_left"`
	IsActive     bool                `json:"is_active"`
}

func (s *SubscriptionService) GetStatus(ctx context.Context, userID string) (*SubscriptionStatus, error) {
	sub, err := s.subRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if sub == nil {
		return &SubscriptionStatus{IsActive: false, DaysLeft: 0}, nil
	}

	now := time.Now()
	isActive := sub.Status == "active" && sub.ExpiresAt.After(now)
	daysLeft := 0
	if isActive {
		daysLeft = int(math.Ceil(sub.ExpiresAt.Sub(now).Hours() / 24))
	}

	return &SubscriptionStatus{
		Subscription: sub,
		DaysLeft:     daysLeft,
		IsActive:     isActive,
	}, nil
}

var ErrInvalidPlan = errors.New("invalid subscription plan")

func (s *SubscriptionService) AdminActivate(ctx context.Context, userID string, months int) (*model.Subscription, error) {
	// Look up plan from database
	plan, err := s.planRepo.GetByMonths(ctx, months)
	if err != nil {
		return nil, err
	}
	if plan == nil {
		return nil, ErrInvalidPlan
	}

	days := months * 30 // approximate

	// Check for existing active subscription — extend it instead of creating duplicate
	existing, err := s.subRepo.GetActiveByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var sub *model.Subscription
	if existing != nil {
		sub, err = s.subRepo.ExtendSubscription(ctx, existing.ID, days, months, plan.Amount)
	} else {
		sub, err = s.subRepo.ActivateByUserID(ctx, userID, days, months, plan.Amount)
	}
	if err != nil {
		return nil, err
	}

	restored, err := s.subRepo.RestoreListings(ctx, userID)
	if err != nil {
		log.Printf("Failed to restore listings for user %s: %v", userID, err)
	} else if restored > 0 {
		log.Printf("Restored %d listings for user %s", restored, userID)
	}

	return sub, nil
}

func (s *SubscriptionService) GetPlans(ctx context.Context) ([]model.SubscriptionPlan, error) {
	plans, err := s.planRepo.ListActive(ctx)
	if err != nil {
		return nil, err
	}
	if len(plans) == 0 {
		return model.SubscriptionPlans, nil
	}
	return plans, nil
}

func (s *SubscriptionService) ListAllPlans(ctx context.Context) ([]model.SubscriptionPlan, error) {
	return s.planRepo.ListAll(ctx)
}

func (s *SubscriptionService) CreatePlan(ctx context.Context, req *model.CreatePlanRequest) (*model.SubscriptionPlan, error) {
	return s.planRepo.Create(ctx, req)
}

func (s *SubscriptionService) UpdatePlan(ctx context.Context, id string, req *model.UpdatePlanRequest) (*model.SubscriptionPlan, error) {
	return s.planRepo.Update(ctx, id, req)
}

func (s *SubscriptionService) DeletePlan(ctx context.Context, id string) error {
	return s.planRepo.Delete(ctx, id)
}

func (s *SubscriptionService) GetMyHistory(ctx context.Context, userID string, page, limit int) ([]*model.Subscription, int, error) {
	return s.subRepo.ListByUserID(ctx, userID, page, limit)
}

func (s *SubscriptionService) GetRevenueStats(ctx context.Context) (*repository.SubRevenueStats, error) {
	return s.subRepo.GetRevenueStats(ctx)
}

func (s *SubscriptionService) GetDailyRevenue(ctx context.Context, from, to string) (*repository.SubDailyRevenueReport, error) {
	return s.subRepo.GetDailyRevenue(ctx, from, to)
}

func (s *SubscriptionService) RunExpiryCron(ctx context.Context) {
	expired, err := s.subRepo.ExpireOverdue(ctx)
	if err != nil {
		log.Printf("Subscription expiry cron error: %v", err)
		return
	}
	if expired > 0 {
		log.Printf("Expired %d subscriptions", expired)
	}

	hidden, err := s.subRepo.HideListingsForExpired(ctx)
	if err != nil {
		log.Printf("Hide listings cron error: %v", err)
		return
	}
	if hidden > 0 {
		log.Printf("Hidden %d listings due to expired subscriptions", hidden)
		if s.onExpiry != nil {
			s.onExpiry(ctx)
		}
	}

	// Notify users whose subscriptions expire within 72 hours
	if s.notifier != nil {
		expiring, err := s.subRepo.GetExpiringSoon(ctx, 72)
		if err != nil {
			log.Printf("GetExpiringSoon error: %v", err)
			return
		}
		for _, sub := range expiring {
			daysLeft := int(math.Ceil(sub.ExpiresAt.Sub(time.Now()).Hours() / 24))
			body := fmt.Sprintf("Gói đăng ký của bạn sẽ hết hạn sau %d ngày. Hãy gia hạn để tránh gián đoạn.", daysLeft)
			if _, err := s.notifier.Create(ctx, sub.UserID, "subscription_expiring", "Sắp hết hạn gói đăng ký", body, nil); err != nil {
				log.Printf("Failed to notify user %s about expiring sub: %v", sub.UserID, err)
			}
		}
	}
}
