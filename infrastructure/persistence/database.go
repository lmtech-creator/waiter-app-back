package persistence

import (
	"fmt"
	"log/slog"

	"github.com/waiter/back/domain/entity"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewDatabase(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	slog.Info("connected to database")

	if err := db.AutoMigrate(
		&entity.Plan{},
		&entity.Restaurant{},
		&entity.Table{},
		&entity.Request{},
		&entity.Feedback{},
		&entity.AdminUser{},
	); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	warnIfTimestampsAreNotZoned(db)

	slog.Info("database migrations completed")
	return db, nil
}

// warnIfTimestampsAreNotZoned surfaces a deployment problem that silently
// shifts hourly statistics: statistics assume `created_at`/`completed_at` are
// `timestamptz`. AutoMigrate never changes an existing column type, so a table
// created as `timestamp without time zone` stays that way and must be migrated
// by hand.
func warnIfTimestampsAreNotZoned(db *gorm.DB) {
	var columns []struct {
		TableName  string
		ColumnName string
		DataType   string
	}

	err := db.Raw(`
		SELECT table_name, column_name, data_type
		FROM information_schema.columns
		WHERE table_schema = 'public'
		  AND (
		    (table_name = 'requests' AND column_name IN ('created_at', 'completed_at'))
		    OR (table_name = 'feedbacks' AND column_name = 'created_at')
		  )
	`).Scan(&columns).Error
	if err != nil {
		slog.Warn("could not verify timestamp column types", "error", err)
		return
	}

	for _, column := range columns {
		if column.DataType != "timestamp with time zone" {
			slog.Warn("timestamp column is not timezone-aware; hourly statistics will be wrong",
				"table", column.TableName,
				"column", column.ColumnName,
				"type", column.DataType,
			)
		}
	}
}
