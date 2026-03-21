package service

import (
	"context"
	"testing"

	"github.com/sangiagao/rice-marketplace/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockReportRepo struct{ mock.Mock }

func (m *mockReportRepo) Create(ctx context.Context, reporterID string, req *model.CreateReportRequest) (*model.Report, error) {
	args := m.Called(ctx, reporterID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Report), args.Error(1)
}
func (m *mockReportRepo) ListByStatus(ctx context.Context, status string, page, limit int) ([]*model.Report, int, error) {
	args := m.Called(ctx, status, page, limit)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*model.Report), args.Int(1), args.Error(2)
}
func (m *mockReportRepo) ListAll(ctx context.Context, page, limit int) ([]*model.Report, int, error) {
	args := m.Called(ctx, page, limit)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*model.Report), args.Int(1), args.Error(2)
}
func (m *mockReportRepo) Resolve(ctx context.Context, reportID, adminID, action string, adminNote *string) (*model.Report, error) {
	args := m.Called(ctx, reportID, adminID, action, adminNote)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Report), args.Error(1)
}
func (m *mockReportRepo) Dismiss(ctx context.Context, reportID, adminID string, adminNote *string) (*model.Report, error) {
	args := m.Called(ctx, reportID, adminID, adminNote)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Report), args.Error(1)
}

func TestReportCreate_Success(t *testing.T) {
	repo := new(mockReportRepo)
	svc := NewReportService(repo)

	req := &model.CreateReportRequest{TargetType: "listing", TargetID: "l-1", Reason: "spam"}
	repo.On("Create", mock.Anything, "user-1", req).Return(&model.Report{ID: "rpt-1", Status: "pending"}, nil)

	report, err := svc.Create(context.Background(), "user-1", req)
	assert.NoError(t, err)
	assert.Equal(t, "pending", report.Status)
}

func TestReportListPending_Success(t *testing.T) {
	repo := new(mockReportRepo)
	svc := NewReportService(repo)

	reports := []*model.Report{{ID: "rpt-1"}}
	repo.On("ListByStatus", mock.Anything, "pending", 1, 20).Return(reports, 1, nil)

	result, total, err := svc.ListPending(context.Background(), 1, 20)
	assert.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, result, 1)
}

func TestReportResolve_Success(t *testing.T) {
	repo := new(mockReportRepo)
	svc := NewReportService(repo)

	repo.On("Resolve", mock.Anything, "rpt-1", "admin-1", "delete_listing", (*string)(nil)).Return(
		&model.Report{ID: "rpt-1", Status: "resolved"}, nil)

	report, err := svc.Resolve(context.Background(), "rpt-1", "admin-1", "delete_listing", nil)
	assert.NoError(t, err)
	assert.Equal(t, "resolved", report.Status)
}

func TestReportDismiss_Success(t *testing.T) {
	repo := new(mockReportRepo)
	svc := NewReportService(repo)

	repo.On("Dismiss", mock.Anything, "rpt-1", "admin-1", (*string)(nil)).Return(
		&model.Report{ID: "rpt-1", Status: "dismissed"}, nil)

	report, err := svc.Dismiss(context.Background(), "rpt-1", "admin-1", nil)
	assert.NoError(t, err)
	assert.Equal(t, "dismissed", report.Status)
}
