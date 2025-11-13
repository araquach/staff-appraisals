package db

import (
	"log"

	"gorm.io/gorm"
	"staff-appraisals/internal/models"
)

// AutoMigrate handles dev schema refresh + migration.
func AutoMigrate(g *gorm.DB) error {
	log.Println("🧹 Dropping and recreating tables (dev mode only)...")

	if err := g.Migrator().DropTable(&models.Branch{}, &models.Staff{}, &models.Client{}, &models.Transaction{}, &models.TransactionItem{}, &models.Review{}); err != nil {
		return err
	}

	// Now recreate the tables
	if err := g.AutoMigrate(
		&models.Branch{},
		&models.Staff{},
		&models.Client{},
		&models.Transaction{},
		&models.TransactionItem{},
		&models.Review{},
	); err != nil {
		return err
	}

	log.Println("✅ AutoMigrate completed (tables dropped and recreated).")
	return nil
}
