package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sangiagao/rice-marketplace/internal/service"
)

// ReferralHandler exposes endpoints for the affiliate / referral program.
type ReferralHandler struct {
	svc *service.ReferralService
}

func NewReferralHandler(svc *service.ReferralService) *ReferralHandler {
	return &ReferralHandler{svc: svc}
}

// GET /api/v1/me/referral
// Returns the caller's referral code + aggregated stats.
func (h *ReferralHandler) GetMyReferral(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	stats, err := h.svc.GetStats(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "không tải được dữ liệu giới thiệu"})
		return
	}
	c.JSON(http.StatusOK, stats)
}

// GET /api/v1/me/referral/history?limit=20&offset=0
func (h *ReferralHandler) GetMyHistory(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	records, err := h.svc.ListHistory(c.Request.Context(), userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "không tải được lịch sử"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": records})
}

// GET /api/v1/me/referees — list of users referred by the caller.
// Always masks identity (mobile/web member surface) regardless of role.
// Returns subscription status + commission stats per referee.
func (h *ReferralHandler) GetMyReferees(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	rows, err := h.svc.Pool().Query(c.Request.Context(),
		`SELECT u.phone, COALESCE(u.name, ''),
		        to_char(u.created_at AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS registered_at,
		        COALESCE(s.status, 'none') AS sub_status,
		        to_char(s.expires_at AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS sub_expires_at,
		        (SELECT COUNT(*)::int FROM commission_records cr WHERE cr.referee_user_id = u.id AND cr.referrer_user_id = $1) AS commission_count,
		        (SELECT COALESCE(SUM(commission_amount), 0) FROM commission_records cr WHERE cr.referee_user_id = u.id AND cr.referrer_user_id = $1) AS total_commission,
		        (SELECT COALESCE(SUM(commission_amount), 0) FROM commission_records cr WHERE cr.referee_user_id = u.id AND cr.referrer_user_id = $1 AND cr.status = 'paid') AS paid_commission
		   FROM users u
		   LEFT JOIN LATERAL (
		     SELECT status, expires_at FROM subscriptions
		      WHERE user_id = u.id
		      ORDER BY started_at DESC LIMIT 1
		   ) s ON true
		  WHERE u.referrer_user_id = $1
		  ORDER BY u.created_at DESC
		  LIMIT 200`, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
		return
	}
	defer rows.Close()

	type item struct {
		Phone           string `json:"phone"`
		Name            string `json:"name"`
		RegisteredAt    string `json:"registered_at"`
		SubStatus       string `json:"sub_status"`
		SubExpiresAt    string `json:"sub_expires_at"`
		CommissionCount int    `json:"commission_count"`
		TotalCommission int64  `json:"total_commission"`
		PaidCommission  int64  `json:"paid_commission"`
	}

	out := []item{}
	for rows.Next() {
		var it item
		var subExpires *string
		if err := rows.Scan(&it.Phone, &it.Name, &it.RegisteredAt,
			&it.SubStatus, &subExpires, &it.CommissionCount, &it.TotalCommission, &it.PaidCommission); err != nil {
			continue
		}
		if subExpires != nil {
			it.SubExpiresAt = *subExpires
		}
		// Always mask on /me/ surface — partner sees their own data abstracted
		it.Phone = maskPhone(it.Phone)
		it.Name = maskName(it.Name)
		out = append(out, it)
	}
	c.JSON(http.StatusOK, gin.H{"data": out})
}

// GET /api/v1/me/payouts — caller's own payouts.
func (h *ReferralHandler) GetMyPayouts(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	rows, err := h.svc.Pool().Query(c.Request.Context(),
		`SELECT id, total_amount, record_count, method, status,
		        to_char(created_at AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS created_at,
		        to_char(sent_at AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS sent_at
		   FROM payouts WHERE referrer_user_id = $1 ORDER BY created_at DESC LIMIT 100`, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
		return
	}
	defer rows.Close()
	type item struct {
		ID          string `json:"id"`
		TotalAmount int64  `json:"total_amount"`
		RecordCount int    `json:"record_count"`
		Method      string `json:"method"`
		Status      string `json:"status"`
		CreatedAt   string `json:"created_at"`
		SentAt      string `json:"sent_at"`
	}
	out := []item{}
	for rows.Next() {
		var it item
		var sentAt *string
		if err := rows.Scan(&it.ID, &it.TotalAmount, &it.RecordCount, &it.Method, &it.Status, &it.CreatedAt, &sentAt); err != nil {
			continue
		}
		if sentAt != nil {
			it.SentAt = *sentAt
		}
		out = append(out, it)
	}
	c.JSON(http.StatusOK, gin.H{"data": out})
}

// POST /api/v1/me/become-affiliate
// Self-service: member → aff. Idempotent. Other roles return 403.
func (h *ReferralHandler) BecomeAffiliate(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	if err := h.svc.BecomeAffiliate(c.Request.Context(), userID); err != nil {
		if errors.Is(err, service.ErrRoleNotEligible) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Tài khoản admin/editor không thể tự đổi vai trò"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể kích hoạt vai trò đối tác"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"role": "aff"})
}

// GET /api/v1/referral/resolve/:code
// Public endpoint used by web /r/{code} landing to validate a code before redirect.
func (h *ReferralHandler) Resolve(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "code required"})
		return
	}
	referrerID, err := h.svc.ResolveReferrer(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "lookup failed"})
		return
	}
	if referrerID == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "mã không tồn tại"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"valid": true})
}
