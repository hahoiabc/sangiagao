package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sangiagao/rice-marketplace/internal/model"
)

type FeedbackRepo struct {
	pool *pgxpool.Pool
}

func NewFeedbackRepo(pool *pgxpool.Pool) *FeedbackRepo {
	return &FeedbackRepo{pool: pool}
}

func (r *FeedbackRepo) Create(ctx context.Context, userID, content string) (*model.Feedback, error) {
	var f model.Feedback
	err := r.pool.QueryRow(ctx,
		`INSERT INTO feedbacks (user_id, content) VALUES ($1, $2)
		 RETURNING id, user_id, content, reply, replied_at, created_at`,
		userID, content,
	).Scan(&f.ID, &f.UserID, &f.Content, &f.Reply, &f.RepliedAt, &f.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *FeedbackRepo) ListByUser(ctx context.Context, userID string, page, limit int) ([]*model.Feedback, int, error) {
	var total int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM feedbacks WHERE user_id = $1`, userID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, content, reply, replied_at, created_at
		 FROM feedbacks WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []*model.Feedback
	for rows.Next() {
		var f model.Feedback
		if err := rows.Scan(&f.ID, &f.UserID, &f.Content, &f.Reply, &f.RepliedAt, &f.CreatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, &f)
	}
	return items, total, nil
}

func (r *FeedbackRepo) ListAll(ctx context.Context, page, limit int) ([]*model.Feedback, int, error) {
	var total int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM feedbacks`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	rows, err := r.pool.Query(ctx,
		`SELECT f.id, f.user_id, COALESCE(u.name, u.phone), u.phone, f.content, f.reply, f.replied_at, f.created_at
		 FROM feedbacks f JOIN users u ON u.id = f.user_id
		 ORDER BY f.created_at DESC LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []*model.Feedback
	for rows.Next() {
		var f model.Feedback
		if err := rows.Scan(&f.ID, &f.UserID, &f.UserName, &f.UserPhone, &f.Content, &f.Reply, &f.RepliedAt, &f.CreatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, &f)
	}
	return items, total, nil
}

func (r *FeedbackRepo) Reply(ctx context.Context, id, reply string) (*model.Feedback, error) {
	var f model.Feedback
	err := r.pool.QueryRow(ctx,
		`UPDATE feedbacks SET reply = $2, replied_at = NOW() WHERE id = $1
		 RETURNING id, user_id, content, reply, replied_at, created_at`,
		id, reply,
	).Scan(&f.ID, &f.UserID, &f.Content, &f.Reply, &f.RepliedAt, &f.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *FeedbackRepo) CountUnreplied(ctx context.Context) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM feedbacks WHERE reply IS NULL`).Scan(&count)
	return count, err
}
