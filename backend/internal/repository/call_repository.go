package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sangiagao/rice-marketplace/internal/model"
)

type CallRepository struct {
	db *pgxpool.Pool
}

func NewCallRepository(db *pgxpool.Pool) *CallRepository {
	return &CallRepository{db: db}
}

func (r *CallRepository) Create(ctx context.Context, callerID, calleeID, conversationID, callType string) (*model.CallLog, error) {
	var call model.CallLog
	err := r.db.QueryRow(ctx, `
		INSERT INTO call_logs (caller_id, callee_id, conversation_id, call_type)
		VALUES ($1, $2, $3, $4)
		RETURNING id, conversation_id, caller_id, callee_id, call_type, status, duration_seconds, started_at, ended_at, created_at
	`, callerID, calleeID, conversationID, callType).Scan(
		&call.ID, &call.ConversationID, &call.CallerID, &call.CalleeID,
		&call.CallType, &call.Status, &call.DurationSeconds,
		&call.StartedAt, &call.EndedAt, &call.CreatedAt,
	)
	return &call, err
}

func (r *CallRepository) UpdateStatus(ctx context.Context, id, status string, duration int) error {
	var endedAt *time.Time
	if status == "answered" {
		now := time.Now()
		endedAt = &now
	}

	_, err := r.db.Exec(ctx, `
		UPDATE call_logs SET status = $2, duration_seconds = $3, ended_at = $4,
			started_at = CASE WHEN started_at IS NULL AND $2 = 'answered' THEN created_at ELSE started_at END
		WHERE id = $1
	`, id, status, duration, endedAt)
	return err
}

func (r *CallRepository) MarkAnswered(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE call_logs SET status = 'answered', started_at = NOW()
		WHERE id = $1
	`, id)
	return err
}

func (r *CallRepository) EndCall(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE call_logs SET ended_at = NOW(),
			duration_seconds = EXTRACT(EPOCH FROM (NOW() - COALESCE(started_at, created_at)))::INT
		WHERE id = $1 AND ended_at IS NULL
	`, id)
	return err
}

func (r *CallRepository) GetByID(ctx context.Context, id string) (*model.CallLog, error) {
	var call model.CallLog
	err := r.db.QueryRow(ctx, `
		SELECT id, conversation_id, caller_id, callee_id, call_type, status, duration_seconds, started_at, ended_at, created_at
		FROM call_logs WHERE id = $1
	`, id).Scan(
		&call.ID, &call.ConversationID, &call.CallerID, &call.CalleeID,
		&call.CallType, &call.Status, &call.DurationSeconds,
		&call.StartedAt, &call.EndedAt, &call.CreatedAt,
	)
	return &call, err
}

func (r *CallRepository) ListByConversation(ctx context.Context, conversationID string, page, limit int) ([]*model.CallLog, int, error) {
	offset := (page - 1) * limit

	var total int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM call_logs WHERE conversation_id = $1`, conversationID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(ctx, `
		SELECT cl.id, cl.conversation_id, cl.caller_id, cl.callee_id, cl.call_type, cl.status,
			cl.duration_seconds, cl.started_at, cl.ended_at, cl.created_at,
			COALESCE(caller.name, caller.phone) as caller_name,
			COALESCE(callee.name, callee.phone) as callee_name
		FROM call_logs cl
		JOIN users caller ON caller.id = cl.caller_id
		JOIN users callee ON callee.id = cl.callee_id
		WHERE cl.conversation_id = $1
		ORDER BY cl.created_at DESC
		LIMIT $2 OFFSET $3
	`, conversationID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var calls []*model.CallLog
	for rows.Next() {
		var c model.CallLog
		if err := rows.Scan(
			&c.ID, &c.ConversationID, &c.CallerID, &c.CalleeID,
			&c.CallType, &c.Status, &c.DurationSeconds,
			&c.StartedAt, &c.EndedAt, &c.CreatedAt,
			&c.CallerName, &c.CalleeName,
		); err != nil {
			return nil, 0, err
		}
		calls = append(calls, &c)
	}
	return calls, total, nil
}
