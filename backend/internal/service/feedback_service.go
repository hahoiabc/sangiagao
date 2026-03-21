package service

import (
	"context"
	"errors"

	"github.com/sangiagao/rice-marketplace/internal/model"
)

type FeedbackService struct {
	repo FeedbackRepository
}

func NewFeedbackService(repo FeedbackRepository) *FeedbackService {
	return &FeedbackService{repo: repo}
}

func (s *FeedbackService) Create(ctx context.Context, userID string, req *model.CreateFeedbackRequest) (*model.Feedback, error) {
	content := req.Content
	if len(content) < 10 {
		return nil, errors.New("góp ý phải có ít nhất 10 ký tự")
	}
	if len(content) > 2000 {
		return nil, errors.New("góp ý không được quá 2000 ký tự")
	}
	return s.repo.Create(ctx, userID, content)
}

func (s *FeedbackService) ListByUser(ctx context.Context, userID string, page, limit int) ([]*model.Feedback, int, error) {
	return s.repo.ListByUser(ctx, userID, page, limit)
}

func (s *FeedbackService) ListAll(ctx context.Context, page, limit int) ([]*model.Feedback, int, error) {
	return s.repo.ListAll(ctx, page, limit)
}

func (s *FeedbackService) Reply(ctx context.Context, id, reply string) (*model.Feedback, error) {
	if len(reply) < 1 {
		return nil, errors.New("nội dung phản hồi là bắt buộc")
	}
	if len(reply) > 2000 {
		return nil, errors.New("phản hồi không được quá 2000 ký tự")
	}
	return s.repo.Reply(ctx, id, reply)
}

func (s *FeedbackService) CountUnreplied(ctx context.Context) (int, error) {
	return s.repo.CountUnreplied(ctx)
}
