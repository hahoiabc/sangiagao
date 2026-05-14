package model

import "time"

type Subscription struct {
	ID             string    `json:"id"`
	UserID         string    `json:"user_id"`
	Plan           string    `json:"plan"`
	DurationMonths int       `json:"duration_months"`
	Amount         int64     `json:"amount"`
	StartedAt      time.Time `json:"started_at"`
	ExpiresAt      time.Time `json:"expires_at"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
}

// SubscriptionPlan defines a pricing tier stored in the database.
// Amount is what the user pays. ListAmount is an optional "giá niêm yết"
// shown crossed-out next to Amount on the pricing page; when ListAmount > 0
// the frontend derives a discount % from the ratio.
type SubscriptionPlan struct {
	ID         string `json:"id"`
	Months     int    `json:"months"`
	Amount     int64  `json:"amount"`      // VND — actual charge
	ListAmount int64  `json:"list_amount"` // VND — strike-through reference, 0 = hide
	Label      string `json:"label"`
	IsActive   bool   `json:"is_active"`
	SortOrder  int    `json:"sort_order"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

type UpdatePlanRequest struct {
	Months     *int    `json:"months"`
	Amount     *int64  `json:"amount"`
	ListAmount *int64  `json:"list_amount"`
	Label      *string `json:"label"`
	IsActive   *bool   `json:"is_active"`
}

type CreatePlanRequest struct {
	Months     int    `json:"months"`
	Amount     int64  `json:"amount"`
	ListAmount int64  `json:"list_amount"`
	Label      string `json:"label"`
}

// SubscriptionPlans is the fallback list of available plans (used if DB is empty).
var SubscriptionPlans = []SubscriptionPlan{
	{Months: 1, Amount: 35000, Label: "1 tháng", IsActive: true},
	{Months: 3, Amount: 96000, Label: "3 tháng", IsActive: true},
	{Months: 6, Amount: 180000, Label: "6 tháng", IsActive: true},
	{Months: 12, Amount: 300000, Label: "12 tháng", IsActive: true},
}

// FindPlan returns the plan for the given months from a plan list.
func FindPlan(months int) *SubscriptionPlan {
	for _, p := range SubscriptionPlans {
		if p.Months == months {
			return &p
		}
	}
	return nil
}

// FindPlanInList searches a plan list by months.
func FindPlanInList(plans []SubscriptionPlan, months int) *SubscriptionPlan {
	for _, p := range plans {
		if p.Months == months && p.IsActive {
			return &p
		}
	}
	return nil
}
