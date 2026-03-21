package service

import (
	"context"
	"testing"
	"time"

	"github.com/sangiagao/rice-marketplace/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockFeedbackRepo struct{ mock.Mock }

func (m *mockFeedbackRepo) Create(ctx context.Context, userID, content string) (*model.Feedback, error) {
	args := m.Called(ctx, userID, content)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Feedback), args.Error(1)
}

func (m *mockFeedbackRepo) ListByUser(ctx context.Context, userID string, page, limit int) ([]*model.Feedback, int, error) {
	args := m.Called(ctx, userID, page, limit)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*model.Feedback), args.Int(1), args.Error(2)
}

func (m *mockFeedbackRepo) ListAll(ctx context.Context, page, limit int) ([]*model.Feedback, int, error) {
	args := m.Called(ctx, page, limit)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*model.Feedback), args.Int(1), args.Error(2)
}

func (m *mockFeedbackRepo) Reply(ctx context.Context, id, reply string) (*model.Feedback, error) {
	args := m.Called(ctx, id, reply)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Feedback), args.Error(1)
}

func (m *mockFeedbackRepo) CountUnreplied(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

func TestFeedbackCreate_Success(t *testing.T) {
	repo := new(mockFeedbackRepo)
	svc := NewFeedbackService(repo)

	content := "Ứng dụng rất hay, mong có thêm tính năng lọc theo vùng miền"
	repo.On("Create", mock.Anything, "user-1", content).Return(
		&model.Feedback{ID: "fb-1", UserID: "user-1", Content: content}, nil)

	req := &model.CreateFeedbackRequest{Content: content}
	fb, err := svc.Create(context.Background(), "user-1", req)
	assert.NoError(t, err)
	assert.Equal(t, "fb-1", fb.ID)
	assert.Equal(t, content, fb.Content)
}

func TestFeedbackCreate_TooShort(t *testing.T) {
	repo := new(mockFeedbackRepo)
	svc := NewFeedbackService(repo)

	req := &model.CreateFeedbackRequest{Content: "ngắn"}
	fb, err := svc.Create(context.Background(), "user-1", req)
	assert.Error(t, err)
	assert.Nil(t, fb)
	assert.Contains(t, err.Error(), "ít nhất 10 ký tự")
}

func TestFeedbackCreate_TooLong(t *testing.T) {
	repo := new(mockFeedbackRepo)
	svc := NewFeedbackService(repo)

	longContent := make([]byte, 2001)
	for i := range longContent {
		longContent[i] = 'a'
	}
	req := &model.CreateFeedbackRequest{Content: string(longContent)}
	fb, err := svc.Create(context.Background(), "user-1", req)
	assert.Error(t, err)
	assert.Nil(t, fb)
	assert.Contains(t, err.Error(), "quá 2000 ký tự")
}

func TestFeedbackListByUser_Success(t *testing.T) {
	repo := new(mockFeedbackRepo)
	svc := NewFeedbackService(repo)

	items := []*model.Feedback{{ID: "fb-1"}, {ID: "fb-2"}}
	repo.On("ListByUser", mock.Anything, "user-1", 1, 20).Return(items, 2, nil)

	result, total, err := svc.ListByUser(context.Background(), "user-1", 1, 20)
	assert.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, result, 2)
}

func TestFeedbackListAll_Success(t *testing.T) {
	repo := new(mockFeedbackRepo)
	svc := NewFeedbackService(repo)

	items := []*model.Feedback{{ID: "fb-1"}}
	repo.On("ListAll", mock.Anything, 1, 20).Return(items, 1, nil)

	result, total, err := svc.ListAll(context.Background(), 1, 20)
	assert.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, result, 1)
}

func TestFeedbackReply_Success(t *testing.T) {
	repo := new(mockFeedbackRepo)
	svc := NewFeedbackService(repo)

	reply := "Cảm ơn bạn đã góp ý, chúng tôi sẽ xem xét!"
	now := time.Now()
	repo.On("Reply", mock.Anything, "fb-1", reply).Return(
		&model.Feedback{ID: "fb-1", Reply: &reply, RepliedAt: &now}, nil)

	fb, err := svc.Reply(context.Background(), "fb-1", reply)
	assert.NoError(t, err)
	assert.NotNil(t, fb.Reply)
	assert.Equal(t, reply, *fb.Reply)
}

func TestFeedbackReply_Empty(t *testing.T) {
	repo := new(mockFeedbackRepo)
	svc := NewFeedbackService(repo)

	fb, err := svc.Reply(context.Background(), "fb-1", "")
	assert.Error(t, err)
	assert.Nil(t, fb)
	assert.Contains(t, err.Error(), "bắt buộc")
}

func TestFeedbackReply_TooLong(t *testing.T) {
	repo := new(mockFeedbackRepo)
	svc := NewFeedbackService(repo)

	longReply := make([]byte, 2001)
	for i := range longReply {
		longReply[i] = 'a'
	}
	fb, err := svc.Reply(context.Background(), "fb-1", string(longReply))
	assert.Error(t, err)
	assert.Nil(t, fb)
	assert.Contains(t, err.Error(), "quá 2000 ký tự")
}

func TestFeedbackCountUnreplied_Success(t *testing.T) {
	repo := new(mockFeedbackRepo)
	svc := NewFeedbackService(repo)

	repo.On("CountUnreplied", mock.Anything).Return(5, nil)

	count, err := svc.CountUnreplied(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 5, count)
}
