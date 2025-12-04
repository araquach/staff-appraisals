package repos

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/araquach/phorest-datahub/internal/models"
)

type PhProductStockRepo struct {
	db *gorm.DB
}

func NewPhProductStockRepo(db *gorm.DB) *PhProductStockRepo {
	return &PhProductStockRepo{db: db}
}

func (r *PhProductStockRepo) GetByProductAndBranch(ctx context.Context, productID, branchID string) (*models.PhProductStock, error) {
	var s models.PhProductStock
	err := r.db.WithContext(ctx).
		Where("product_id = ? AND branch_id = ?", productID, branchID).
		First(&s).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *PhProductStockRepo) Upsert(ctx context.Context, s *models.PhProductStock) error {
	// set LastSyncedAt on each upsert
	s.LastSyncedAt = time.Now()

	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "product_id"}, {Name: "branch_id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"price",
				"min_quantity",
				"max_quantity",
				"quantity_in_stock",
				"reorder_count",
				"reorder_cost",
				"archived",
				"created_at_ph",
				"updated_at_ph",
				"last_synced_at",
			}),
		}).
		Create(s).Error
}

func (r *PhProductStockRepo) InsertHistory(ctx context.Context, h *models.PhProductStockHistory) error {
	if h.SnapshotTime.IsZero() {
		h.SnapshotTime = time.Now()
	}
	return r.db.WithContext(ctx).Create(h).Error
}
