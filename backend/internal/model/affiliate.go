package model

import (
	"encoding/json"
	"time"
)

// ReferralCode is 1-1 with a user, lazy-generated on first request.
type ReferralCode struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Code      string    `json:"code"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
}

// CommissionRule defines 3-stage rate by member age. referral_code_id = NULL
// is the global default. Per-partner override created when admin negotiates.
type CommissionRule struct {
	ID             string     `json:"id"`
	ReferralCodeID *string    `json:"referral_code_id"` // nil = default
	Stage1Days     int        `json:"stage1_days"`
	Stage1Pct      float64    `json:"stage1_pct"`
	Stage2Days     int        `json:"stage2_days"`
	Stage2Pct      float64    `json:"stage2_pct"`
	Stage3Pct      float64    `json:"stage3_pct"` // perpetual after stage1+stage2 days
	BaseType       string     `json:"base_type"`  // gross | net
	MinimumPayout  int64      `json:"minimum_payout"`
	ActiveFrom     time.Time  `json:"active_from"`
	ActiveTo       *time.Time `json:"active_to,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// CommissionRecord is 1 row per payment event (Apple SUBSCRIBED/DID_RENEW or
// SePay PAID). Idempotent by (payment_source, payment_event_id).
type CommissionRecord struct {
	ID               string     `json:"id"`
	ReferrerUserID   string     `json:"referrer_user_id"`
	RefereeUserID    string     `json:"referee_user_id"`
	SubscriptionID   string     `json:"subscription_id"`
	PaymentSource    string     `json:"payment_source"` // apple | sepay | admin
	PaymentEventID   string     `json:"payment_event_id"`
	GrossAmount      int64      `json:"gross_amount"`
	PlatformFee      int64      `json:"platform_fee"`
	NetAmount        int64      `json:"net_amount"`
	BaseAmount       int64      `json:"base_amount"` // gross or net per rule
	Stage            int        `json:"stage"`       // 1, 2, 3
	Rate             float64    `json:"rate"`
	CommissionAmount int64      `json:"commission_amount"`
	RefereeAgeDays   int        `json:"referee_age_days"`
	RuleID           string     `json:"rule_id"`
	Status           string     `json:"status"` // pending | payable | paid | cancelled
	PayableAfter     time.Time  `json:"payable_after"`
	PaidAt           *time.Time `json:"paid_at,omitempty"`
	PayoutID         *string    `json:"payout_id,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
}

// Payout aggregates one or more commission_records into a single transfer.
type Payout struct {
	ID             string          `json:"id"`
	ReferrerUserID string          `json:"referrer_user_id"`
	TotalAmount    int64           `json:"total_amount"`
	TransferFee    int64           `json:"transfer_fee"`   // paid by aff, deducted from total
	RecordCount    int             `json:"record_count"`
	Method         string          `json:"method"` // bank | momo | cash | other
	BankInfo       json.RawMessage `json:"bank_info,omitempty"`
	Note           *string         `json:"note,omitempty"`
	Status         string          `json:"status"` // pending | sent | failed
	CreatedBy      string          `json:"created_by"`
	SentAt         *time.Time      `json:"sent_at,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
}

// AffBankInfo is the bank account where commission payouts are sent.
type AffBankInfo struct {
	UserID      string    `json:"user_id"`
	AccountNo   string    `json:"account_no"`
	BankName    string    `json:"bank_name"`
	HolderName  string    `json:"holder_name"`
	Note        *string   `json:"note,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CurrentTermsVersion is the live affiliate T&C version. Bump when content changes.
// Aff must accept the new version before further payouts can be created.
const CurrentTermsVersion = "1.0"

// ReferralStats is the aggregated view shown to a referrer (mobile + web).
type ReferralStats struct {
	Code            string `json:"code"`
	TotalReferrals  int    `json:"total_referrals"`
	ActiveReferees  int    `json:"active_referees"` // referees with active subscription
	TotalEarned     int64  `json:"total_earned"`
	PayableAmount   int64  `json:"payable_amount"` // status='payable', not yet paid
	PendingAmount   int64  `json:"pending_amount"` // status='pending'
	PaidAmount      int64  `json:"paid_amount"`
	MinimumPayout   int64  `json:"minimum_payout"`
}
