package db

import (
	"context"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Open returns a configured *gorm.DB. Prefer passing the handle around instead of globals.
func Open(dsn string) (*gorm.DB, error) {
	// Log level from env, default WARN in prod, INFO when GORM_DEBUG=1
	var lvl logger.LogLevel = logger.Warn
	if os.Getenv("GORM_DEBUG") == "1" {
		lvl = logger.Info // be careful: may log PII
	}

	// Structured-ish GORM logger with slow query threshold
	gormLogger := logger.New(
		log.New(os.Stdout, "", log.LstdFlags),
		logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  lvl,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	gdb, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:                                   gormLogger,
		PrepareStmt:                              true, // good default for APIs
		DisableForeignKeyConstraintWhenMigrating: false,
		NowFunc:                                  func() time.Time { return time.Now().UTC() }, // force UTC
	})
	if err != nil {
		return nil, err
	}

	// Configure the underlying sql.DB pool
	sqlDB, err := gdb.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(20) // tune as needed
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(10 * time.Minute)

	// Fast failure if DB is unreachable
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, err
	}

	log.Println("🧪 Effective GORM log level:", lvl)

	return gdb, nil
}

// Close is handy for CLIs like appraisals-sync
func Close(gdb *gorm.DB) error {
	sqlDB, err := gdb.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
