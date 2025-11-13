package repos

import (
	"log"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"staff-appraisals/internal/models"
)

type ReviewsRepo struct {
	db *gorm.DB
	lg *log.Logger
}

func NewReviewsRepo(db *gorm.DB, lg *log.Logger) *ReviewsRepo {
	return &ReviewsRepo{db: db, lg: lg}
}

// UpsertMany: reviews are immutable, so we can safely DO NOTHING on conflict.
func (r *ReviewsRepo) UpsertMany(rows []models.Review) error {
	if len(rows) == 0 {
		return nil
	}

	res := r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "review_id"}},
		DoNothing: true,
	}).Create(&rows)

	if res.Error != nil {
		return res.Error
	}
	r.lg.Printf("Upserted/inserted reviews: %d", len(rows))
	return nil
}

// Watermark helpers (for incremental fetches by branch)
func (r *ReviewsRepo) MaxReviewDate(branchID string) (*string, error) {
	var ts *string
	err := r.db.
		Table("reviews").
		Select("MAX(review_date)::text").
		Where("branch_id = ?", branchID).
		Scan(&ts).Error
	return ts, err
}
