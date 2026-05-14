package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sangiagao/rice-marketplace/internal/model"
	"github.com/sangiagao/rice-marketplace/internal/repository"
)

// ReferralService handles code generation + attribution. Commission calc is
// in CommissionEngine.
type ReferralService struct {
	pool    *pgxpool.Pool
	affRepo *repository.AffiliateRepo
}

func NewReferralService(pool *pgxpool.Pool, affRepo *repository.AffiliateRepo) *ReferralService {
	return &ReferralService{pool: pool, affRepo: affRepo}
}

// Pool exposes the underlying pgxpool for handlers needing ad-hoc SQL
// (e.g. /me/referees, /me/payouts).
func (s *ReferralService) Pool() *pgxpool.Pool { return s.pool }

var (
	ErrReferralSelfRefer    = errors.New("không thể tự dùng mã giới thiệu của mình")
	ErrReferralAlreadySet   = errors.New("tài khoản này đã có người giới thiệu")
	ErrReferralCodeNotFound = errors.New("mã giới thiệu không hợp lệ")
	ErrRoleNotEligible      = errors.New("vai trò hiện tại không thể tự nâng cấp")
)

// BecomeAffiliate self-activates the affiliate role for a 'member' user.
// Other roles (admin/editor/owner/aff) are no-ops or rejected.
// Idempotent: returns success if user is already 'aff'.
func (s *ReferralService) BecomeAffiliate(ctx context.Context, userID string) error {
	var currentRole string
	if err := s.pool.QueryRow(ctx, `SELECT role FROM users WHERE id = $1`, userID).Scan(&currentRole); err != nil {
		return err
	}
	if currentRole == "aff" {
		// idempotent — ensure code exists then return
		_, _ = s.GetOrCreateCode(ctx, userID)
		return nil
	}
	if currentRole != "member" {
		return ErrRoleNotEligible
	}
	_, err := s.pool.Exec(ctx,
		`UPDATE users SET role = 'aff', updated_at = NOW() WHERE id = $1 AND role = 'member'`, userID)
	if err != nil {
		return err
	}
	// Ensure referral_code exists right after upgrade
	_, _ = s.GetOrCreateCode(ctx, userID)
	return nil
}

// codeCharset excludes 0/O/1/I/L to avoid handwriting confusion.
const codeCharset = "ABCDEFGHJKMNPQRSTUVWXYZ23456789"
const codeLen = 6
const maxCodeGenAttempts = 10

// GetOrCreateCode returns the user's referral code, generating it lazily.
// Only creates a code for users currently in 'aff' role. Other roles get
// ErrReferralCodeNotFound (consumers should hide aff UI when this errors).
func (s *ReferralService) GetOrCreateCode(ctx context.Context, userID string) (*model.ReferralCode, error) {
	if rc, err := s.affRepo.GetCodeByUser(ctx, userID); err == nil {
		return rc, nil
	} else if !errors.Is(err, repository.ErrReferralCodeNotFound) {
		return nil, err
	}

	// Only create code for aff role (defense — caller should check too)
	var role string
	if err := s.pool.QueryRow(ctx, `SELECT role FROM users WHERE id = $1`, userID).Scan(&role); err != nil {
		return nil, err
	}
	if role != "aff" {
		return nil, repository.ErrReferralCodeNotFound
	}

	// Generate a unique 6-char code. Retry on collision (very unlikely).
	for attempt := 0; attempt < maxCodeGenAttempts; attempt++ {
		code, err := generateCode()
		if err != nil {
			return nil, err
		}
		rc, err := s.affRepo.CreateCode(ctx, userID, code)
		if err == nil {
			return rc, nil
		}
		// Unique violation on either user_id or code → retry
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			// If user_id collision: another request created it concurrently → re-read
			if rc, getErr := s.affRepo.GetCodeByUser(ctx, userID); getErr == nil {
				return rc, nil
			}
			continue // code collision → retry
		}
		return nil, err
	}
	return nil, fmt.Errorf("referral: could not generate unique code after %d attempts", maxCodeGenAttempts)
}

// AttributeReferral records that refereeID was referred by the owner of
// referralCode. Idempotent + defensive: ignores empty code, self-referral,
// already-attributed.
// CRITICAL: only attributes if the code owner is currently an active 'aff'
// role. Member users do not earn commission even if they somehow have a code.
func (s *ReferralService) AttributeReferral(ctx context.Context, referralCode, refereeID string) error {
	if referralCode == "" {
		return nil
	}
	rc, err := s.affRepo.GetCodeByCode(ctx, referralCode)
	if err != nil {
		if errors.Is(err, repository.ErrReferralCodeNotFound) {
			slog.Info("referral: unknown code on signup", "code", referralCode)
			return nil // don't fail signup over invalid code
		}
		return err
	}
	if rc.UserID == refereeID {
		return ErrReferralSelfRefer
	}

	// Guard: code owner must currently be 'aff' role to attribute.
	var ownerRole string
	if err := s.pool.QueryRow(ctx, `SELECT role FROM users WHERE id = $1`, rc.UserID).Scan(&ownerRole); err != nil {
		return err
	}
	if ownerRole != "aff" {
		slog.Info("referral: code owner not aff, skip attribution",
			"code", referralCode, "owner", rc.UserID, "owner_role", ownerRole)
		return nil
	}

	// Atomic update: only set if currently NULL (don't overwrite existing referrer)
	ct, err := s.pool.Exec(ctx,
		`UPDATE users
		    SET referrer_user_id = $1, referred_at = NOW()
		  WHERE id = $2 AND referrer_user_id IS NULL`,
		rc.UserID, refereeID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		// User already has a referrer or doesn't exist; silently skip
		return nil
	}
	slog.Info("referral: attributed", "referee", refereeID, "referrer", rc.UserID, "code", referralCode)
	return nil
}

// GetStats returns aggregated stats + the user's own code (lazy-created).
func (s *ReferralService) GetStats(ctx context.Context, userID string) (*model.ReferralStats, error) {
	rc, err := s.GetOrCreateCode(ctx, userID)
	if err != nil {
		return nil, err
	}
	stats, err := s.affRepo.StatsForReferrer(ctx, userID)
	if err != nil {
		return nil, err
	}
	stats.Code = rc.Code

	// Pull per-partner min_payout if exists, else default
	codeID := &rc.ID
	rule, err := s.affRepo.GetActiveRule(ctx, codeID)
	if err == nil && rule != nil {
		stats.MinimumPayout = rule.MinimumPayout
	}
	return stats, nil
}

// ListHistory returns the user's commission records (most recent first).
func (s *ReferralService) ListHistory(ctx context.Context, userID string, limit, offset int) ([]*model.CommissionRecord, error) {
	return s.affRepo.ListRecordsForReferrer(ctx, userID, limit, offset)
}

// ResolveReferrer returns the user_id of the referrer for a given code, or
// "" if the code is unknown. Used by web /r/{code} landing.
func (s *ReferralService) ResolveReferrer(ctx context.Context, code string) (string, error) {
	rc, err := s.affRepo.GetCodeByCode(ctx, code)
	if err != nil {
		if errors.Is(err, repository.ErrReferralCodeNotFound) {
			return "", nil
		}
		return "", err
	}
	return rc.UserID, nil
}

// generateCode picks codeLen random chars from codeCharset.
func generateCode() (string, error) {
	out := make([]byte, codeLen)
	buf := make([]byte, codeLen)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	for i, b := range buf {
		out[i] = codeCharset[int(b)%len(codeCharset)]
	}
	return string(out), nil
}

// Stub for compile-time check that pgx is imported (used in repo).
var _ = pgx.ErrNoRows
