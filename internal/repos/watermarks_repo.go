package repos

import (
	"log"
	"time"

	"gorm.io/gorm"
)

// WatermarksRepo provides access to the sync_watermarks table.
type WatermarksRepo struct {
	db *gorm.DB
	lg *log.Logger
}

func NewWatermarksRepo(db *gorm.DB, lg *log.Logger) *WatermarksRepo {
	return &WatermarksRepo{db: db, lg: lg}
}

// SyncWatermark matches the sync_watermarks table schema.
// (We only use it for reads; writes use raw SQL upsert.)
type SyncWatermark struct {
	SourceName    string     `gorm:"column:source_name;primaryKey"`
	BranchID      string     `gorm:"column:branch_id;primaryKey"`
	LastUpdatedAt *time.Time `gorm:"column:last_updated_at"`
	LastRunAt     time.Time  `gorm:"column:last_run_at"`
}

func (SyncWatermark) TableName() string { return "sync_watermarks" }

// Get returns the last_updated_at for a given (source, branchID).
// If no row exists, it returns (nil, nil).
func (r *WatermarksRepo) Get(source, branchID string) (*time.Time, error) {
	var wm SyncWatermark
	err := r.db.
		Where("source_name = ? AND branch_id = ?", source, branchID).
		First(&wm).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return wm.LastUpdatedAt, nil
}

// UpsertMax advances the watermark if candidate is newer.
// If candidate is nil or zero, it does nothing.
func (r *WatermarksRepo) UpsertMax(source, branchID string, candidate *time.Time) error {
	if candidate == nil || candidate.IsZero() {
		return nil
	}

	r.lg.Printf("💾 Updating watermark for %s/%s → %s",
		source, branchID, candidate.UTC().Format(time.RFC3339))

	return r.db.Exec(`
INSERT INTO sync_watermarks (source_name, branch_id, last_updated_at, last_run_at)
VALUES (?, ?, ?, now())
ON CONFLICT (source_name, branch_id) DO UPDATE
SET last_updated_at = GREATEST(sync_watermarks.last_updated_at, EXCLUDED.last_updated_at),
    last_run_at     = now();
`, source, branchID, candidate.UTC()).Error
}
