package phorest

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/gorm"
	"staff-appraisals/internal/config"
	"staff-appraisals/internal/repos"
)

type Runner struct {
	DB     *gorm.DB
	Cfg    *config.Config
	Logger *log.Logger
}

// Accept cfg and store it so r.Cfg is valid everywhere
func NewRunner(db *gorm.DB, cfg *config.Config, lg *log.Logger) *Runner {
	return &Runner{
		DB:     db,
		Cfg:    cfg,
		Logger: lg,
	}
}

// ImportAllTransactionsCSVs loops through all .csv files in a directory and imports them.
func (r *Runner) ImportAllTransactionsCSVs(dir string) error {
	lg := r.Logger
	lg.Printf("🔍 Scanning directory: %s", dir)

	paths, err := filepath.Glob(filepath.Join(dir, "*.csv"))
	if err != nil {
		return fmt.Errorf("read directory: %w", err)
	}
	if len(paths) == 0 {
		lg.Printf("⚠️  No CSV files found in %s", dir)
		return nil
	}

	lg.Printf("📂 Found %d CSV files", len(paths))
	for _, path := range paths {
		name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		lg.Printf("──────────────────────────────────────────────")
		lg.Printf("🏁 Starting import for file: %s", name)

		if err := r.importSingleTransactionsCSV(path); err != nil {
			lg.Printf("❌ Failed import for %s: %v", name, err)
			continue
		}
		lg.Printf("✅ Completed import for %s", name)
	}

	lg.Printf("🎉 All CSV imports complete.")
	return nil
}

func (r *Runner) importSingleTransactionsCSV(csvPath string) error {
	lg := r.Logger

	batch, err := ParseTransactionsCSV(csvPath, lg)
	if err != nil {
		return err
	}
	lg.Printf("Importing CSV %s: %d transactions, %d items", csvPath, len(batch.Transactions), len(batch.Items))

	tx := r.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	tr := repos.NewTransactionsRepo(tx, lg)
	ir := repos.NewItemsRepo(tx, lg)

	if err := tr.UpsertBatch(batch.Transactions, 500); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := ir.UpsertBatch(batch.Items, 500); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}
	lg.Printf("✅ CSV %s committed.", filepath.Base(csvPath))
	return nil
}

// ImportAllClientCSVs scans a dir and imports every .csv as clients
func (r *Runner) ImportAllClientCSVs(dir string) error {
	r.Logger.Printf("🔍 Scanning clients dir: %s", dir)
	paths, err := filepath.Glob(filepath.Join(dir, "*.csv"))
	if err != nil {
		return fmt.Errorf("read directory: %w", err)
	}
	if len(paths) == 0 {
		r.Logger.Printf("⚠️  No client CSV files found in %s", dir)
		return nil
	}
	for _, p := range paths {
		if err := r.importSingleClientsCSV(p); err != nil {
			r.Logger.Printf("❌ Client import failed: %s: %v", p, err)
			continue
		}
	}
	r.Logger.Printf("✅ All client CSV imports complete.")
	return nil
}

func (r *Runner) importSingleClientsCSV(csvPath string) error {
	lg := r.Logger
	batch, err := ParseClientsCSV(csvPath, lg)
	if err != nil {
		return err
	}
	lg.Printf("Importing Clients CSV %s: %d clients", csvPath, len(batch.Clients))

	var maxTS *time.Time
	for i := range batch.Clients {
		if ts := batch.Clients[i].UpdatedAtPhorest; ts != nil {
			if maxTS == nil || ts.After(*maxTS) {
				maxTS = ts
			}
		}
	}
	if maxTS == nil {
		lg.Printf("⚠️  No UpdatedAtPhorest values in %s; skipping watermark update", csvPath)
	}

	tx := r.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	cr := repos.NewClientsRepo(tx, lg)
	if err := cr.UpsertBatch(batch.Clients, 1000); err != nil {
		_ = tx.Rollback()
		return err
	}

	if maxTS != nil {
		wr := repos.NewWatermarksRepo(tx, lg)
		// NOTE: branch = "ALL" for global clients CSV
		if err := wr.UpsertLastUpdated("clients_csv", "ALL", *maxTS); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("update clients_csv watermark: %w", err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}
	lg.Printf("✅ Clients CSV %s committed.", csvPath)
	return nil
}
