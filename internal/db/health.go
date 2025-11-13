package db

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// HealthCheck tries to ping the DB within a timeout.
func HealthCheck(gdb *gorm.DB, timeout time.Duration) error {
	sqlDB, err := gdb.DB()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return sqlDB.PingContext(ctx)
}
