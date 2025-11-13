package db

import (
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// RunMigrations applies SQL migrations from the given directory (e.g. "migrations").
func RunMigrations(dsn string, migrationsDir string, lg *log.Logger) error {
	lg.Println("🗂  Running SQL migrations...")

	m, err := migrate.New(
		"file://"+migrationsDir,
		dsn,
	)
	if err != nil {
		return err
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}

	if err == migrate.ErrNoChange {
		lg.Println("ℹ️  No migration changes to apply.")
	} else {
		lg.Println("✅ SQL migrations applied successfully.")
	}

	return nil
}
