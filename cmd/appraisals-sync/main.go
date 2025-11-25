package main

import (
	"staff-appraisals/internal/phorest"
	"time"

	"github.com/joho/godotenv"
	"staff-appraisals/internal/config"
	"staff-appraisals/internal/db"
)

func main() {
	_ = godotenv.Load()

	cfg := config.Load()
	logger := cfg.Logger

	gdb, err := db.Open(cfg.DatabaseURL)
	if err != nil {
		logger.Fatalf("DB connection failed: %v", err)
	}
	defer db.Close(gdb)

	if err := db.HealthCheck(gdb, 3*time.Second); err != nil {
		logger.Fatalf("DB health check failed: %v", err)
	}
	logger.Println("✅ Database connection healthy.")

	// Real Migrator
	if cfg.AutoMigrate {
		logger.Println("Running SQL migrations...")
		if err := db.RunMigrations(cfg.DatabaseURL, "migrations", logger); err != nil {
			logger.Fatalf("Database migration failed: %v", err)
		}
		logger.Println("✅ Database migrated successfully.")
	}

	// GORM Migrator
	//if cfg.AutoMigrate {
	//	logger.Println("Running AutoMigrate...")
	//	if err := db.AutoMigrate(gdb); err != nil {
	//		logger.Fatalf("AutoMigrate failed: %v", err)
	//	}
	//	logger.Println("✅ Database migrated successfully.")
	//}

	for _, b := range cfg.Branches {
		logger.Printf("Branch: %s (ID: %s)\n", b.Name, b.BranchID)
	}

	logger.Println("✅ Startup complete. Ready to sync Phorest data.")

	runner := phorest.NewRunner(gdb, cfg, logger)
	if err := runner.ImportAllTransactionsCSVs("data/transactions"); err != nil {
		logger.Fatalf("Import failed: %v", err)
	}

	if err := runner.ImportAllClientCSVs("data/clients"); err != nil {
		logger.Fatalf("clients import failed: %v", err)
	}

	if err := runner.SyncStaffFromAPI(); err != nil {
		logger.Fatalf("staff sync failed: %v", err)
	}

	if err := runner.SyncBranchesFromAPI(); err != nil {
		logger.Printf("branch sync ended with errors: %v", err)
	}

	//if err := runner.SyncReviewsFromAPI(); err != nil {
	//	logger.Printf("reviews sync ended with errors: %v", err)
	//}

	if err := runner.SyncLatestReviewsFromAPI(10); err != nil {
		logger.Printf("reviews latest-N sync ended with errors: %v", err)
	}

	if err := runner.BootstrapWatermarks(); err != nil {
		logger.Printf("⚠️ Bootstrap watermarks failed: %v", err)
	} else {
		logger.Printf("✅ Bootstrap watermarks completed successfully.")
	}
}
