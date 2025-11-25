package models

import "time"

// SyncWatermark stores the last-updated marker for a given entity/branch pair.
// examples:
//
//	entity = "transactions", branch_id = "eajUo19q0jF069dv8UN3rQ"
//	entity = "clients",      branch_id = NULL (global)
type SyncWatermark struct {
	ID       uint    `gorm:"primaryKey"`
	Entity   string  `gorm:"column:entity;not null;index:idx_watermark_entity_branch,unique"`
	BranchID *string `gorm:"column:branch_id;index:idx_watermark_entity_branch,unique"`
	// LastUpdatedPhorest is the last updated_at_phorest (or equivalent) we’ve fully processed
	LastUpdatedPhorest *time.Time `gorm:"column:last_updated_phorest"`
	CreatedAt          time.Time  `gorm:"column:created_at"`
	UpdatedAt          time.Time  `gorm:"column:updated_at"`
}

// TableName is optional, but keeps things explicit.
func (SyncWatermark) TableName() string {
	return "sync_watermarks"
}
