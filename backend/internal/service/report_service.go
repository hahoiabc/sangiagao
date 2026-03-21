package service

import (
	"context"

	"github.com/sangiagao/rice-marketplace/internal/model"
)

type ReportService struct {
	reportRepo ReportRepository
}

func NewReportService(reportRepo ReportRepository) *ReportService {
	return &ReportService{reportRepo: reportRepo}
}

func (s *ReportService) Create(ctx context.Context, reporterID string, req *model.CreateReportRequest) (*model.Report, error) {
	return s.reportRepo.Create(ctx, reporterID, req)
}

func (s *ReportService) ListPending(ctx context.Context, page, limit int) ([]*model.Report, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 20
	}
	return s.reportRepo.ListByStatus(ctx, "pending", page, limit)
}

func (s *ReportService) ListByStatus(ctx context.Context, status string, page, limit int) ([]*model.Report, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 20
	}
	if status == "all" {
		return s.reportRepo.ListAll(ctx, page, limit)
	}
	return s.reportRepo.ListByStatus(ctx, status, page, limit)
}

func (s *ReportService) Resolve(ctx context.Context, reportID, adminID, action string, adminNote *string) (*model.Report, error) {
	return s.reportRepo.Resolve(ctx, reportID, adminID, action, adminNote)
}

func (s *ReportService) Dismiss(ctx context.Context, reportID, adminID string, adminNote *string) (*model.Report, error) {
	return s.reportRepo.Dismiss(ctx, reportID, adminID, adminNote)
}
