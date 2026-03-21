package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sangiagao/rice-marketplace/internal/model"
)

type ReportRepo struct {
	pool *pgxpool.Pool
}

func NewReportRepo(pool *pgxpool.Pool) *ReportRepo {
	return &ReportRepo{pool: pool}
}

const reportColumns = `id, reporter_id, target_type, target_id, reason, description,
	status, admin_action, admin_note, resolved_by, resolved_at, created_at`

func scanReport(row interface{ Scan(dest ...any) error }) (*model.Report, error) {
	var rpt model.Report
	err := row.Scan(&rpt.ID, &rpt.ReporterID, &rpt.TargetType, &rpt.TargetID,
		&rpt.Reason, &rpt.Description, &rpt.Status, &rpt.AdminAction,
		&rpt.AdminNote, &rpt.ResolvedBy, &rpt.ResolvedAt, &rpt.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &rpt, nil
}

func (r *ReportRepo) Create(ctx context.Context, reporterID string, req *model.CreateReportRequest) (*model.Report, error) {
	row := r.pool.QueryRow(ctx,
		`INSERT INTO reports (reporter_id, target_type, target_id, reason, description)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING `+reportColumns,
		reporterID, req.TargetType, req.TargetID, req.Reason, req.Description,
	)
	return scanReport(row)
}

func (r *ReportRepo) ListByStatus(ctx context.Context, status string, page, limit int) ([]*model.Report, int, error) {
	offset := (page - 1) * limit

	var total int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM reports WHERE status = $1`, status,
	).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.Query(ctx,
		`SELECT `+reportColumns+`
		 FROM reports WHERE status = $1
		 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		status, limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var reports []*model.Report
	for rows.Next() {
		rpt, err := scanReport(rows)
		if err != nil {
			return nil, 0, err
		}
		reports = append(reports, rpt)
	}
	if reports == nil {
		reports = []*model.Report{}
	}
	return reports, total, rows.Err()
}

func (r *ReportRepo) ListAll(ctx context.Context, page, limit int) ([]*model.Report, int, error) {
	offset := (page - 1) * limit

	var total int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM reports`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.Query(ctx,
		`SELECT `+reportColumns+`
		 FROM reports
		 ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var reports []*model.Report
	for rows.Next() {
		rpt, err := scanReport(rows)
		if err != nil {
			return nil, 0, err
		}
		reports = append(reports, rpt)
	}
	if reports == nil {
		reports = []*model.Report{}
	}
	return reports, total, rows.Err()
}

func (r *ReportRepo) Resolve(ctx context.Context, reportID, adminID, action string, adminNote *string) (*model.Report, error) {
	row := r.pool.QueryRow(ctx,
		`UPDATE reports SET status = 'resolved', admin_action = $2, admin_note = $3, resolved_by = $4, resolved_at = NOW()
		 WHERE id = $1 AND status = 'pending'
		 RETURNING `+reportColumns,
		reportID, action, adminNote, adminID,
	)
	return scanReport(row)
}

func (r *ReportRepo) Dismiss(ctx context.Context, reportID, adminID string, adminNote *string) (*model.Report, error) {
	row := r.pool.QueryRow(ctx,
		`UPDATE reports SET status = 'dismissed', admin_note = $2, resolved_by = $3, resolved_at = NOW()
		 WHERE id = $1 AND status = 'pending'
		 RETURNING `+reportColumns,
		reportID, adminNote, adminID,
	)
	return scanReport(row)
}
