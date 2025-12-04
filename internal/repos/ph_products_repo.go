package repos

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/araquach/phorest-datahub/internal/models"
)

type PhProductRepo struct {
	db *gorm.DB
}

func NewPhProductRepo(db *gorm.DB) *PhProductRepo {
	return &PhProductRepo{db: db}
}

func (r *PhProductRepo) Upsert(ctx context.Context, p *models.PhProduct) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"parent_id",
				"name",
				"brand_id",
				"brand_name",
				"category_id",
				"category_name",
				"code",
				"type_raw",
				"measurement_qty",
				"measurement_unit",
				"archived",
				"created_at_ph",
				"updated_at_ph",
				"updated_at",
			}),
		}).
		Create(p).Error
}
