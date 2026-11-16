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

	// Upgrade the restaurants.plan_id column before AutoMigrate runs: GORM would
	// otherwise emit `ADD COLUMN plan_id uuid NOT NULL`, which fails on a table
	// that already has rows.
	if err := migrateRestaurantPlanID(db); err != nil {
		return nil, fmt.Errorf("failed to migrate restaurants.plan_id: %w", err)
	}

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

// migrateRestaurantPlanID upgrades databases created before the plans module.
//
// GORM's AutoMigrate emits `ALTER TABLE restaurants ADD COLUMN plan_id uuid
// NOT NULL`, which Postgres rejects when the table already has rows. This adds
// the column as nullable, backfills it from the legacy `plan` text column
// (falling back to the default plan) and only then enforces NOT NULL.
//
// It is idempotent and a no-op on a fresh database, where AutoMigrate creates
// the table with the column already in place.
func migrateRestaurantPlanID(db *gorm.DB) error {
	var restaurantsExist bool
	if err := db.Raw(`SELECT EXISTS (
		SELECT 1 FROM information_schema.tables
		WHERE table_schema = CURRENT_SCHEMA() AND table_name = 'restaurants'
	)`).Scan(&restaurantsExist).Error; err != nil {
		return fmt.Errorf("check restaurants table: %w", err)
	}
	if !restaurantsExist {
		return nil
	}

	steps := []string{
		// plan_id references plans, so the table must exist first.
		`CREATE TABLE IF NOT EXISTS plans (
			id UUID PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT ''
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_plans_name ON plans(name)`,
		// Default plan — the fallback target for restaurants without one.
		`INSERT INTO plans (id, name, description)
			VALUES ('00000000-0000-0000-0000-000000000001', 'free', 'Plan gratuito')
			ON CONFLICT (name) DO NOTHING`,
		// Nullable first so the ADD succeeds on a populated table.
		`ALTER TABLE restaurants ADD COLUMN IF NOT EXISTS plan_id UUID`,
	}
	for _, step := range steps {
		if err := db.Exec(step).Error; err != nil {
			return err
		}
	}

	// Backfill from the legacy `plan` name when that column still exists. The
	// check lives in Go because Postgres resolves column references at parse
	// time — a plain UPDATE would fail once `plan` has been dropped.
	var legacyPlanExists bool
	if err := db.Raw(`SELECT EXISTS (
		SELECT 1 FROM information_schema.columns
		WHERE table_schema = CURRENT_SCHEMA()
		  AND table_name = 'restaurants' AND column_name = 'plan'
	)`).Scan(&legacyPlanExists).Error; err != nil {
		return fmt.Errorf("check restaurants.plan column: %w", err)
	}
	if legacyPlanExists {
		// Preserve the distinct legacy plan names as real plans.
		if err := db.Exec(`INSERT INTO plans (id, name, description)
			SELECT gen_random_uuid(), trim(r.plan), ''
			FROM (
				SELECT DISTINCT plan FROM restaurants
				WHERE plan IS NOT NULL AND trim(plan) <> ''
			) r
			WHERE NOT EXISTS (
				SELECT 1 FROM plans p WHERE lower(p.name) = lower(trim(r.plan))
			)`).Error; err != nil {
			return err
		}
		// Point each restaurant at the plan matching its legacy name.
		if err := db.Exec(`UPDATE restaurants r
			SET plan_id = p.id
			FROM plans p
			WHERE r.plan_id IS NULL AND lower(p.name) = lower(trim(r.plan))`).Error; err != nil {
			return err
		}
	}

	// Any remaining or dangling value falls back to the default plan.
	if err := db.Exec(`UPDATE restaurants r
		SET plan_id = (SELECT id FROM plans WHERE name = 'free' LIMIT 1)
		WHERE r.plan_id IS NULL
		   OR NOT EXISTS (SELECT 1 FROM plans p WHERE p.id = r.plan_id)`).Error; err != nil {
		return err
	}

	// Enforce the constraints AutoMigrate will not add to an existing column.
	if err := db.Exec(`ALTER TABLE restaurants ALTER COLUMN plan_id SET NOT NULL`).Error; err != nil {
		return err
	}
	if err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_restaurants_plan_id ON restaurants(plan_id)`).Error; err != nil {
		return err
	}

	// Drop the legacy column once the data has been migrated.
	if err := db.Exec(`ALTER TABLE restaurants DROP COLUMN IF EXISTS plan`).Error; err != nil {
		return err
	}

	return nil
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
