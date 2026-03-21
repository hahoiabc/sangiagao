package model

import "time"

type Report struct {
	ID          string     `json:"id"`
	ReporterID  string     `json:"reporter_id"`
	TargetType  string     `json:"target_type"`
	TargetID    string     `json:"target_id"`
	Reason      string     `json:"reason"`
	Description *string    `json:"description,omitempty"`
	Status      string     `json:"status"`
	AdminAction *string    `json:"admin_action,omitempty"`
	AdminNote   *string    `json:"admin_note,omitempty"`
	ResolvedBy  *string    `json:"resolved_by,omitempty"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

type CreateReportRequest struct {
	TargetType  string  `json:"target_type" binding:"required,oneof=listing user rating"`
	TargetID    string  `json:"target_id" binding:"required"`
	Reason      string  `json:"reason" binding:"required"`
	Description *string `json:"description"`
}

type ResolveReportRequest struct {
	AdminAction string  `json:"admin_action" binding:"required"`
	AdminNote   *string `json:"admin_note"`
}
