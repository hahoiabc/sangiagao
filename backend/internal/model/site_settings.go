package model

import "time"

type SiteSetting struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UpdateSiteSettingRequest struct {
	Value string `json:"value" binding:"required,max=500"`
}
