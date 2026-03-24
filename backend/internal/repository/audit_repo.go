package repository

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AuditRepo struct {
	pool *pgxpool.Pool
}

func NewAuditRepo(pool *pgxpool.Pool) *AuditRepo {
	return &AuditRepo{pool: pool}
}

func (r *AuditRepo) Log(ctx context.Context, adminID, action, targetType, targetID string, details map[string]interface{}) error {
	var detailsJSON []byte
	if details != nil {
		detailsJSON, _ = json.Marshal(details)
	}
	_, err := r.pool.Exec(ctx,
		`INSERT INTO admin_audit_logs (admin_id, action, target_type, target_id, details)
		 VALUES ($1, $2, $3, $4, $5)`,
		adminID, action, targetType, targetID, detailsJSON,
	)
	return err
}
