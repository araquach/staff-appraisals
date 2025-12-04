package repos

import (
	"log"

	"github.com/araquach/phorest-datahub/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type BranchRepo struct {
	db *gorm.DB
	lg *log.Logger
}

func NewBranchRepo(db *gorm.DB, lg *log.Logger) *BranchRepo {
	return &BranchRepo{db: db, lg: lg}
}

func (r *BranchRepo) UpsertMany(rows []models.Branch) error {
	if len(rows) == 0 {
		return nil
	}

	res := r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "branch_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"name", "time_zone", "latitude", "longitude",
			"street_address_1", "street_address_2", "city", "state",
			"postal_code", "country", "currency_code", "account_id", "updated_at",
		}),
	}).Create(&rows)

	if res.Error != nil {
		return res.Error
	}
	r.lg.Printf("Upserted %d branches", len(rows))
	return nil
}
