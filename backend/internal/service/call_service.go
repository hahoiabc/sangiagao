package service

import (
	"context"
	"fmt"

	"github.com/sangiagao/rice-marketplace/internal/model"
)

type CallService struct {
	callRepo CallRepository
	convRepo ConversationRepository
	userRepo UserRepository
}

func NewCallService(callRepo CallRepository, convRepo ConversationRepository, userRepo UserRepository) *CallService {
	return &CallService{callRepo: callRepo, convRepo: convRepo, userRepo: userRepo}
}

func (s *CallService) InitiateCall(ctx context.Context, callerID, conversationID, calleeID, callType string) (*model.CallLog, error) {
	// Verify caller is participant
	ok, err := s.convRepo.IsParticipant(ctx, conversationID, callerID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("không phải thành viên cuộc hội thoại")
	}

	// Check if caller is blocked
	caller, err := s.userRepo.GetByID(ctx, callerID)
	if err != nil {
		return nil, err
	}
	if caller.IsBlocked {
		return nil, fmt.Errorf("tài khoản đã bị khóa")
	}

	// Check if callee is blocked
	callee, err := s.userRepo.GetByID(ctx, calleeID)
	if err != nil {
		return nil, err
	}
	if callee.IsBlocked {
		return nil, fmt.Errorf("không thể gọi cho người dùng này")
	}

	return s.callRepo.Create(ctx, callerID, calleeID, conversationID, callType)
}

func (s *CallService) AnswerCall(ctx context.Context, callID, userID string) error {
	call, err := s.callRepo.GetByID(ctx, callID)
	if err != nil {
		return err
	}
	if call.CalleeID != userID {
		return fmt.Errorf("không có quyền trả lời cuộc gọi này")
	}
	return s.callRepo.MarkAnswered(ctx, callID)
}

func (s *CallService) EndCall(ctx context.Context, callID, userID string) error {
	call, err := s.callRepo.GetByID(ctx, callID)
	if err != nil {
		return err
	}
	if call.CallerID != userID && call.CalleeID != userID {
		return fmt.Errorf("không có quyền kết thúc cuộc gọi này")
	}
	return s.callRepo.EndCall(ctx, callID)
}

func (s *CallService) RejectCall(ctx context.Context, callID, userID string) error {
	call, err := s.callRepo.GetByID(ctx, callID)
	if err != nil {
		return err
	}
	if call.CalleeID != userID {
		return fmt.Errorf("không có quyền từ chối cuộc gọi này")
	}
	return s.callRepo.UpdateStatus(ctx, callID, "rejected", 0)
}

func (s *CallService) GetCallByID(ctx context.Context, callID string) (*model.CallLog, error) {
	return s.callRepo.GetByID(ctx, callID)
}

func (s *CallService) MissCall(ctx context.Context, callID, userID string) error {
	call, err := s.callRepo.GetByID(ctx, callID)
	if err != nil {
		return err
	}
	if call.CallerID != userID && call.CalleeID != userID {
		return fmt.Errorf("không có quyền đánh dấu cuộc gọi nhỡ")
	}
	return s.callRepo.UpdateStatus(ctx, callID, "missed", 0)
}

func (s *CallService) GetCallHistory(ctx context.Context, userID, conversationID string, page, limit int) ([]*model.CallLog, int, error) {
	ok, err := s.convRepo.IsParticipant(ctx, conversationID, userID)
	if err != nil {
		return nil, 0, err
	}
	if !ok {
		return nil, 0, fmt.Errorf("không phải thành viên cuộc hội thoại")
	}
	return s.callRepo.ListByConversation(ctx, conversationID, page, limit)
}
