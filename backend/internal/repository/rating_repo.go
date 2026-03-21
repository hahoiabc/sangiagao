package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sangiagao/rice-marketplace/internal/model"
)

var ErrDuplicateRating = errors.New("you already rated this seller")

type RatingRepo struct {
	pool *pgxpool.Pool
}

func NewRatingRepo(pool *pgxpool.Pool) *RatingRepo {
	return &RatingRepo{pool: pool}
}

func (r *RatingRepo) Create(ctx context.Context, reviewerID string, req *model.CreateRatingRequest) (*model.Rating, error) {
	var rating model.Rating
	err := r.pool.QueryRow(ctx,
		`INSERT INTO ratings (reviewer_id, seller_id, stars, comment)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, reviewer_id, seller_id, stars, comment, created_at`,
		reviewerID, req.SellerID, req.Stars, req.Comment,
	).Scan(&rating.ID, &rating.ReviewerID, &rating.SellerID, &rating.Stars, &rating.Comment, &rating.CreatedAt)
	if err != nil {
		if err.Error() == `ERROR: duplicate key value violates unique constraint "ratings_reviewer_id_seller_id_key" (SQLSTATE 23505)` {
			return nil, ErrDuplicateRating
		}
		return nil, err
	}
	return &rating, nil
}

func (r *RatingRepo) ListBySeller(ctx context.Context, sellerID string, page, limit int) ([]*model.Rating, int, error) {
	offset := (page - 1) * limit

	var total int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM ratings WHERE seller_id = $1`, sellerID,
	).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.Query(ctx,
		`SELECT id, reviewer_id, seller_id, stars, comment, created_at
		 FROM ratings WHERE seller_id = $1
		 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		sellerID, limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var ratings []*model.Rating
	for rows.Next() {
		var rating model.Rating
		if err := rows.Scan(&rating.ID, &rating.ReviewerID, &rating.SellerID, &rating.Stars, &rating.Comment, &rating.CreatedAt); err != nil {
			return nil, 0, err
		}
		ratings = append(ratings, &rating)
	}
	if ratings == nil {
		ratings = []*model.Rating{}
	}
	return ratings, total, rows.Err()
}

func (r *RatingRepo) GetSummary(ctx context.Context, sellerID string) (*model.RatingSummary, error) {
	var summary model.RatingSummary
	err := r.pool.QueryRow(ctx,
		`SELECT COALESCE(AVG(stars), 0), COUNT(*) FROM ratings WHERE seller_id = $1`, sellerID,
	).Scan(&summary.Average, &summary.Count)
	if err != nil {
		return nil, err
	}
	return &summary, nil
}

func (r *RatingRepo) HasRated(ctx context.Context, reviewerID, sellerID string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM ratings WHERE reviewer_id = $1 AND seller_id = $2)`,
		reviewerID, sellerID,
	).Scan(&exists)
	return exists, err
}

func (r *RatingRepo) GetSellerRole(ctx context.Context, userID string) (string, error) {
	var role string
	err := r.pool.QueryRow(ctx,
		`SELECT role FROM users WHERE id = $1`, userID,
	).Scan(&role)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	return role, err
}
