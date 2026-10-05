package persistence

import (
	"fmt"
	"log/slog"

	"github.com/waiter/back/domain/entity"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// reportTimezone is the IANA zone used to bucket timestamps into hours for
// statistics queries. `created_at`/`completed_at` are `timestamptz`, so
// EXTRACT(HOUR FROM ts) yields the UTC hour; every hour bucket must therefore be
// shifted explicitly with `AT TIME ZONE` or the reports are off by the UTC offset.
// America/Argentina/Buenos_Aires has been a fixed UTC-3 with no DST since 2009.
const reportTimezone = "America/Argentina/Buenos_Aires"

func NewDatabase(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	slog.Info("connected to database")

	if err := db.AutoMigrate(
		&entity.Restaurant{},
		&entity.Table{},
		&entity.Request{},
		&entity.Feedback{},
		&entity.AdminUser{},
	); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	slog.Info("database migrations completed")
	return db, nil
}
