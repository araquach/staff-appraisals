package repos

import (
	"log"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"staff-appraisals/internal/models"
)

type StaffRepo struct {
	db *gorm.DB
	lg *log.Logger
}

func NewStaffRepo(db *gorm.DB, lg *log.Logger) *StaffRepo {
	return &StaffRepo{db: db, lg: lg}
}

func (r *StaffRepo) UpsertMany(rows []models.Staff) error {
	if len(rows) == 0 {
		return nil
	}

	// normalise date-only to midnight UTC (harmless if already nil)
	for i := range rows {
		rows[i].BirthDate = dateOnly(rows[i].BirthDate)
		rows[i].StartDate = dateOnly(rows[i].StartDate)
	}

	res := r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "staff_id"},
			{Name: "branch_id"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"staff_category_id", "user_id", "staff_category_name",
			"first_name", "last_name", "birth_date", "start_date",
			"self_employed", "archived", "mobile", "email", "gender",
			"notes", "online_profile", "hide_from_online_bookings",
			"hide_from_appointment_screen", "image_url", "updated_at",
		}),
	}).Create(&rows)

	if res.Error != nil {
		return res.Error
	}
	r.lg.Printf("Upserted %d staff rows", len(rows))
	return nil
}

func dateOnly(t *time.Time) *time.Time {
	if t == nil || t.IsZero() {
		return nil
	}
	d := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	return &d
}
