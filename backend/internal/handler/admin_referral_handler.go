package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sangiagao/rice-marketplace/internal/model"
	"github.com/sangiagao/rice-marketplace/internal/repository"
	"github.com/sangiagao/rice-marketplace/internal/service"
)

// AdminReferralHandler exposes commission/payout management gated by
// permission keys (not hardcoded roles). Owner/admin typically have:
//   - referrals.view_all     (see all partners)
//   - referrals.manage_rules (edit commission rule defaults / overrides)
//   - referrals.create_payout (create + mark payout sent)
// Aff role typically has:
//   - referrals.view_own (filtered to own referrer_user_id)
// Admin can re-assign these per role via /users → "Vai trò & Quyền hạn".
type AdminReferralHandler struct {
	repo            *repository.AffiliateRepo
	referralService *service.ReferralService
	perm            permissionChecker
}

// permissionChecker is the slice of PermissionService that we need.
type permissionChecker interface {
	HasPermission(role, key string) bool
}

func NewAdminReferralHandler(repo *repository.AffiliateRepo) *AdminReferralHandler {
	return &AdminReferralHandler{repo: repo}
}

func (h *AdminReferralHandler) SetReferralService(s *service.ReferralService) {
	h.referralService = s
}

func (h *AdminReferralHandler) SetPermissionChecker(p permissionChecker) {
	h.perm = p
}

// can reports whether the caller has the given permission key. Defaults to
// the conservative role check (owner/admin) when no permission service is
// wired (e.g. unit tests).
func (h *AdminReferralHandler) can(role, key string) bool {
	if h.perm != nil {
		return h.perm.HasPermission(role, key)
	}
	return role == "owner" || role == "admin"
}

// Permission keys (must match keys configured in /users → Vai trò & Quyền hạn).
const (
	permViewAll     = "referrals.view_all"
	permManageRules = "referrals.manage_rules"
	permCreatePayout = "referrals.create_payout"
)

// ListRules — view_all sees all; otherwise sees only their own per-partner rule.
func (h *AdminReferralHandler) ListRules(c *gin.Context) {
	role := c.GetString("user_role")
	userID := c.GetString("user_id")
	rules, err := h.repo.ListRules(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list rules"})
		return
	}
	if !h.can(role, permViewAll) {
		// aff: filter to their own rule (need to look up their referral_code_id)
		myCode, err := h.repo.GetCodeByUser(c.Request.Context(), userID)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"data": []any{}})
			return
		}
		filtered := []*model.CommissionRule{}
		for _, r := range rules {
			if r.ReferralCodeID != nil && *r.ReferralCodeID == myCode.ID {
				filtered = append(filtered, r)
			}
		}
		c.JSON(http.StatusOK, gin.H{"data": filtered})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rules})
}

type upsertRuleRequest struct {
	ReferralCodeID *string `json:"referral_code_id"` // nil = default
	Stage1Days     int     `json:"stage1_days" binding:"required,min=1"`
	Stage1Pct      float64 `json:"stage1_pct" binding:"required,min=0,max=1"`
	Stage2Days     int     `json:"stage2_days" binding:"required,min=1"`
	Stage2Pct      float64 `json:"stage2_pct" binding:"min=0,max=1"`
	Stage3Pct      float64 `json:"stage3_pct" binding:"min=0,max=1"`
	BaseType       string  `json:"base_type" binding:"oneof=gross net"`
	MinimumPayout  int64   `json:"minimum_payout" binding:"min=0"`
}

// UpsertRule — gated by referrals.manage_rules. Creates new active version, expires previous.
func (h *AdminReferralHandler) UpsertRule(c *gin.Context) {
	role := c.GetString("user_role")
	if !h.can(role, permManageRules) {
		c.JSON(http.StatusForbidden, gin.H{"error": "permission denied"})
		return
	}
	var req upsertRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	rule := &model.CommissionRule{
		ReferralCodeID: req.ReferralCodeID,
		Stage1Days:     req.Stage1Days,
		Stage1Pct:      req.Stage1Pct,
		Stage2Days:     req.Stage2Days,
		Stage2Pct:      req.Stage2Pct,
		Stage3Pct:      req.Stage3Pct,
		BaseType:       req.BaseType,
		MinimumPayout:  req.MinimumPayout,
	}
	if err := h.repo.UpsertRule(c.Request.Context(), rule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save rule: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, rule)
}

// Leaderboard lists all aff-role users + anyone with at least 1 commission,
// with aggregate totals (zeros if no commission yet). Admin sees all rows.
// Aff sees only their own row.
func (h *AdminReferralHandler) Leaderboard(c *gin.Context) {
	role := c.GetString("user_role")
	userID := c.GetString("user_id")

	q := `SELECT u.id AS referrer_user_id,
	             u.phone, COALESCE(u.name, '') AS name,
	             COALESCE(rc.code, '') AS code,
	             COUNT(DISTINCT cr.referee_user_id) AS total_referrals,
	             COALESCE(SUM(cr.commission_amount), 0) AS total_earned,
	             COALESCE(SUM(CASE WHEN cr.status='payable' THEN cr.commission_amount END), 0) AS payable_amount,
	             COALESCE(SUM(CASE WHEN cr.status='pending' THEN cr.commission_amount END), 0) AS pending_amount,
	             COALESCE(SUM(CASE WHEN cr.status='paid' THEN cr.commission_amount END), 0) AS paid_amount
	        FROM users u
	        LEFT JOIN referral_codes rc ON rc.user_id = u.id
	        LEFT JOIN commission_records cr ON cr.referrer_user_id = u.id`
	args := []any{}
	if !h.can(role, permViewAll) {
		q += " WHERE u.id = $1"
		args = append(args, userID)
	} else {
		q += ` WHERE u.role = 'aff'
		         OR EXISTS (SELECT 1 FROM commission_records WHERE referrer_user_id = u.id)
		         OR rc.id IS NOT NULL`
	}
	q += ` GROUP BY u.id, u.phone, u.name, rc.code
	       ORDER BY total_earned DESC, u.created_at DESC`

	rows, err := h.repo.Pool().Query(c.Request.Context(), q, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed: " + err.Error()})
		return
	}
	defer rows.Close()

	type row struct {
		ReferrerUserID string `json:"referrer_user_id"`
		Phone          string `json:"phone"`
		Name           string `json:"name"`
		Code           string `json:"code"`
		TotalReferrals int    `json:"total_referrals"`
		TotalEarned    int64  `json:"total_earned"`
		PayableAmount  int64  `json:"payable_amount"`
		PendingAmount  int64  `json:"pending_amount"`
		PaidAmount     int64  `json:"paid_amount"`
	}
	out := []row{}
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.ReferrerUserID, &r.Phone, &r.Name, &r.Code,
			&r.TotalReferrals, &r.TotalEarned, &r.PayableAmount, &r.PendingAmount, &r.PaidAmount); err != nil {
			continue
		}
		out = append(out, r)
	}

	// Backfill: any row with empty code → lazy-create. One-shot per user.
	if h.referralService != nil {
		for i := range out {
			if out[i].Code == "" {
				if rc, err := h.referralService.GetOrCreateCode(c.Request.Context(), out[i].ReferrerUserID); err == nil && rc != nil {
					out[i].Code = rc.Code
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"data": out})
}

// ListPayablePerReferrer — gated by referrals.create_payout. List payable
// (status='payable') records for a given referrer. Used to preview before
// creating a payout.
func (h *AdminReferralHandler) ListPayablePerReferrer(c *gin.Context) {
	role := c.GetString("user_role")
	if !h.can(role, permCreatePayout) {
		c.JSON(http.StatusForbidden, gin.H{"error": "permission denied"})
		return
	}
	referrerID := c.Query("referrer_user_id")
	if referrerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "referrer_user_id required"})
		return
	}
	rows, err := h.repo.Pool().Query(c.Request.Context(),
		`SELECT id, referee_user_id, commission_amount, stage, rate,
		        to_char(payable_after AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS payable_after,
		        to_char(created_at AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS created_at
		   FROM commission_records
		  WHERE referrer_user_id = $1 AND status = 'payable'
		  ORDER BY created_at ASC`, referrerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed: " + err.Error()})
		return
	}
	defer rows.Close()
	type item struct {
		ID               string  `json:"id"`
		RefereeUserID    string  `json:"referee_user_id"`
		CommissionAmount int64   `json:"commission_amount"`
		Stage            int     `json:"stage"`
		Rate             float64 `json:"rate"`
		PayableAfter     string  `json:"payable_after"`
		CreatedAt        string  `json:"created_at"`
	}
	out := []item{}
	var totalAmount int64
	for rows.Next() {
		var it item
		if err := rows.Scan(&it.ID, &it.RefereeUserID, &it.CommissionAmount,
			&it.Stage, &it.Rate, &it.PayableAfter, &it.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "scan failed: " + err.Error()})
			return
		}
		totalAmount += it.CommissionAmount
		out = append(out, it)
	}
	c.JSON(http.StatusOK, gin.H{"data": out, "total_amount": totalAmount, "count": len(out)})
}

type createPayoutRequest struct {
	ReferrerUserID string          `json:"referrer_user_id" binding:"required"`
	RecordIDs      []string        `json:"record_ids" binding:"required"`
	Method         string          `json:"method" binding:"oneof=bank momo cash other"`
	BankInfo       json.RawMessage `json:"bank_info"`
	Note           string          `json:"note"`
}

// CreatePayout — gated by referrals.create_payout. Marks the given record IDs
// as paid + creates payout row. Atomic in repo.
func (h *AdminReferralHandler) CreatePayout(c *gin.Context) {
	role := c.GetString("user_role")
	if !h.can(role, permCreatePayout) {
		c.JSON(http.StatusForbidden, gin.H{"error": "permission denied"})
		return
	}
	createdBy := c.GetString("user_id")

	var req createPayoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(req.RecordIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "record_ids cannot be empty"})
		return
	}

	// Calculate total + check minimum_payout threshold
	var total int64
	row := h.repo.Pool().QueryRow(c.Request.Context(),
		`SELECT COALESCE(SUM(commission_amount), 0) FROM commission_records
		  WHERE id = ANY($1::uuid[]) AND referrer_user_id = $2 AND status = 'payable'`,
		req.RecordIDs, req.ReferrerUserID)
	if err := row.Scan(&total); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "sum failed: " + err.Error()})
		return
	}
	if total == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no payable records found"})
		return
	}

	// Resolve referrer's referral_code_id to find their rule minimum_payout
	var codeID *string
	_ = h.repo.Pool().QueryRow(c.Request.Context(),
		`SELECT id FROM referral_codes WHERE user_id = $1`, req.ReferrerUserID).Scan(&codeID)
	rule, err := h.repo.GetActiveRule(c.Request.Context(), codeID)
	if err == nil && rule != nil && total < rule.MinimumPayout {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":          "below minimum payout threshold",
			"total":          total,
			"minimum_payout": rule.MinimumPayout,
		})
		return
	}

	p := &model.Payout{
		ReferrerUserID: req.ReferrerUserID,
		TotalAmount:    total,
		RecordCount:    len(req.RecordIDs),
		Method:         req.Method,
		BankInfo:       req.BankInfo,
		CreatedBy:      createdBy,
	}
	if req.Note != "" {
		p.Note = &req.Note
	}
	if err := h.repo.CreatePayout(c.Request.Context(), p, req.RecordIDs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, p)
}

// ListPayouts — view_all sees all, otherwise filtered to own.
func (h *AdminReferralHandler) ListPayouts(c *gin.Context) {
	role := c.GetString("user_role")
	userID := c.GetString("user_id")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	var ref *string
	if !h.can(role, permViewAll) {
		ref = &userID
	} else if q := c.Query("referrer_user_id"); q != "" {
		ref = &q
	}

	payouts, err := h.repo.ListPayouts(c.Request.Context(), ref, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": payouts})
}

// MarkPayoutSent — gated by referrals.create_payout. Transitions pending → sent.
func (h *AdminReferralHandler) MarkPayoutSent(c *gin.Context) {
	role := c.GetString("user_role")
	if !h.can(role, permCreatePayout) {
		c.JSON(http.StatusForbidden, gin.H{"error": "permission denied"})
		return
	}
	id := c.Param("id")
	if err := h.repo.MarkPayoutSent(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
