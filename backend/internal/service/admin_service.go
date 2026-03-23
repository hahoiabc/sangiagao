package service

import (
	"context"
	"errors"

	"github.com/sangiagao/rice-marketplace/internal/model"
	"github.com/sangiagao/rice-marketplace/internal/repository"
)

var (
	ErrInvalidAdminRole  = errors.New("invalid role")
	ErrCannotModifyAdmin = errors.New("không thể thao tác trên tài khoản quản trị viên")
)

type AdminService struct {
	userRepo    UserRepository
	listingRepo ListingRepository
	subRepo     SubscriptionRepository
}

func NewAdminService(userRepo UserRepository, listingRepo ListingRepository, subRepo SubscriptionRepository) *AdminService {
	return &AdminService{userRepo: userRepo, listingRepo: listingRepo, subRepo: subRepo}
}

func (s *AdminService) GetDashboardStats(ctx context.Context) (map[string]int, error) {
	return s.userRepo.GetDashboardStats(ctx)
}

func (s *AdminService) GetDashboardCharts(ctx context.Context) (*repository.DashboardCharts, error) {
	return s.userRepo.GetDashboardCharts(ctx)
}

func (s *AdminService) ListUsers(ctx context.Context, search string, page, limit int) ([]*model.User, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 20
	}
	return s.userRepo.ListUsers(ctx, search, page, limit)
}

func (s *AdminService) GetUserByID(ctx context.Context, userID string) (*model.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

func (s *AdminService) ListUserListings(ctx context.Context, userID string, page, limit int) ([]*model.Listing, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 10
	}
	return s.listingRepo.ListByUser(ctx, userID, page, limit)
}

func (s *AdminService) BlockUser(ctx context.Context, userID, reason, callerRole string) (*model.User, error) {
	target, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if target.Role == "admin" && callerRole != "owner" {
		return nil, ErrCannotModifyAdmin
	}
	if target.Role == "owner" {
		return nil, ErrCannotModifyAdmin
	}
	return s.userRepo.BlockUser(ctx, userID, reason)
}

func (s *AdminService) UnblockUser(ctx context.Context, userID, callerRole string) (*model.User, error) {
	target, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if target.Role == "admin" && callerRole != "owner" {
		return nil, ErrCannotModifyAdmin
	}
	if target.Role == "owner" {
		return nil, ErrCannotModifyAdmin
	}
	return s.userRepo.UnblockUser(ctx, userID)
}

func (s *AdminService) ListUserSubscriptions(ctx context.Context, userID string, page, limit int) ([]*model.Subscription, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 10
	}
	return s.subRepo.ListByUserID(ctx, userID, page, limit)
}

func (s *AdminService) ChangeUserRole(ctx context.Context, userID, role, callerRole string) (*model.User, error) {
	validRoles := map[string]bool{"owner": true, "admin": true, "editor": true, "member": true}
	if !validRoles[role] {
		return nil, ErrInvalidAdminRole
	}
	target, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if target.Role == "admin" && callerRole != "owner" {
		return nil, ErrCannotModifyAdmin
	}
	if target.Role == "owner" {
		return nil, ErrCannotModifyAdmin
	}
	return s.userRepo.SetRole(ctx, userID, role)
}

func (s *AdminService) DeleteUser(ctx context.Context, userID, callerRole string) error {
	target, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if target.Role == "owner" {
		return ErrCannotModifyAdmin
	}
	if target.Role == "admin" && callerRole != "owner" {
		return ErrCannotModifyAdmin
	}
	return s.userRepo.DeleteUser(ctx, userID)
}

func (s *AdminService) DeleteListing(ctx context.Context, listingID string) error {
	return s.listingRepo.SoftDelete(ctx, listingID)
}

type BatchItemError struct {
	ID    string `json:"id"`
	Error string `json:"error"`
}

type BatchBlockResult struct {
	Blocked int              `json:"blocked"`
	Errors  []BatchItemError `json:"errors"`
}

type BatchDeleteResult struct {
	Deleted int              `json:"deleted"`
	Errors  []BatchItemError `json:"errors"`
}

func (s *AdminService) BatchBlockUsers(ctx context.Context, userIDs []string, reason, callerRole string) (*BatchBlockResult, error) {
	result := &BatchBlockResult{Errors: []BatchItemError{}}
	for _, id := range userIDs {
		target, err := s.userRepo.GetByID(ctx, id)
		if err != nil {
			result.Errors = append(result.Errors, BatchItemError{ID: id, Error: err.Error()})
			continue
		}
		if target.Role == "owner" || (target.Role == "admin" && callerRole != "owner") {
			result.Errors = append(result.Errors, BatchItemError{ID: id, Error: ErrCannotModifyAdmin.Error()})
			continue
		}
		_, err = s.userRepo.BlockUser(ctx, id, reason)
		if err != nil {
			result.Errors = append(result.Errors, BatchItemError{ID: id, Error: err.Error()})
			continue
		}
		result.Blocked++
	}
	return result, nil
}

func (s *AdminService) BatchDeleteListings(ctx context.Context, listingIDs []string) (*BatchDeleteResult, error) {
	result := &BatchDeleteResult{Errors: []BatchItemError{}}
	for _, id := range listingIDs {
		err := s.listingRepo.SoftDelete(ctx, id)
		if err != nil {
			result.Errors = append(result.Errors, BatchItemError{ID: id, Error: err.Error()})
			continue
		}
		result.Deleted++
	}
	return result, nil
}
