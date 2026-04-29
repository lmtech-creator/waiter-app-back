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
		&entity.Restaurant{},
		&entity.Table{},
		&entity.Request{},
		&entity.Feedback{},
		&entity.AdminUser{},
	); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	// Ensure restaurant_id is NOT NULL on admin_users (enforce constraint after AutoMigrate).
	if err := db.Exec(`ALTER TABLE admin_users ALTER COLUMN restaurant_id SET NOT NULL`).Error; err != nil {
		slog.Warn("could not set NOT NULL on admin_users.restaurant_id", "error", err)
	}

	slog.Info("database migrations completed")
	return db, nil
}
