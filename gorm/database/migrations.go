package database

import (
	"fmt"
	"sort"
	"time"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Local log fields
var (
	MIGRATION_ID              = "migration_id"
	ERROR                     = "error"
	MIGRATION_DESCRIPTION     = "migration_description"
	MIGRATION_IDS_TO_ROLLBACK = "migration_ids_to_rollback"
)

// MigrationRecord tracks which migrations have been applied
type MigrationRecord struct {
	Id          string    `gorm:"primaryKey"`
	Description string    `gorm:"not null"`
	AppliedAt   time.Time `gorm:"not null"`
}

// Migration represents a database migration
type Migration struct {
	Id          string
	Description string
	Up          func(*gorm.DB) error
	Down        func(*gorm.DB) error
}

// MigrationRegistry holds all registered migrations
var MigrationRegistry []Migration

// RegisterMigration adds a migration to the registry
func RegisterMigration(migration Migration) {
	MigrationRegistry = append(MigrationRegistry, migration)
}

// RunMigrations executes all pending migrations
func RunMigrations(db *Database) (err error) {
	// Create migrations table if it doesn't exist
	if err := db.AutoMigrate(&MigrationRecord{}); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Sort migrations by Id to ensure consistent order
	sort.Slice(MigrationRegistry, func(i, j int) bool {
		return MigrationRegistry[i].Id < MigrationRegistry[j].Id
	})

	appliedCount := 0
	currentMigrationId := "UNKNOWN"
	var appliedMigrations []string
	defer func() {
		if err != nil {
			log.WithFields(log.Fields{
				MIGRATION_ID: currentMigrationId,
				ERROR:        err,
			}).Error("Migration failed, rolling back")

			// Attempt to rollback all migrations applied in this session
			if rollbackErr := rollbackMigrations(db, appliedMigrations); rollbackErr != nil {
				log.WithError(rollbackErr).
					Error("Failed to rollback migrations - database may be in inconsistent state")
				err = fmt.Errorf(
					"migration %s failed and rollback failed: %w (rollback error: %v)",
					currentMigrationId,
					err,
					rollbackErr,
				)
				return
			}

			err = fmt.Errorf(
				"migration %s failed (successfully rolled back): %w",
				currentMigrationId,
				err,
			)
			return
		}
	}()

	for _, migration := range MigrationRegistry {
		currentMigrationId = migration.Id
		// Check if migration has been applied
		var count int64
		db.Model(&MigrationRecord{}).Where("id = ?", migration.Id).Count(&count)

		if count == 0 {
			log.Info(
				fmt.Sprintf(
					"Running database migration %s",
					currentMigrationId,
				),
			)
			log.WithFields(log.Fields{
				MIGRATION_ID:          migration.Id,
				MIGRATION_DESCRIPTION: migration.Description,
			}).Info("Running migration")

			// Execute migration in a transaction
			err = db.Transaction(func(tx *gorm.DB) error {
				if err := migration.Up(tx); err != nil {
					return err
				}

				// Record migration
				record := MigrationRecord{
					Id:          migration.Id,
					Description: migration.Description,
					AppliedAt:   time.Now().UTC(),
				}
				return tx.Create(&record).Error
			})
			if err != nil {
				return err
			}

			// Track this migration for potential rollback
			appliedMigrations = append(appliedMigrations, migration.Id)
			appliedCount++
		}
	}

	return nil
}

// rollbackMigrations rolls back a list of migrations that have already been preformed today in reverse order
func rollbackMigrations(db *Database, appliedMigrations []string) error {
	log.WithField(MIGRATION_IDS_TO_ROLLBACK, len(appliedMigrations)).
		Warn("Rolling back migrations")

	// Rollback in reverse order
	for i := len(appliedMigrations) - 1; i >= 0; i-- {
		migrationID := appliedMigrations[i]

		// Find the migration
		var migration *Migration
		for _, m := range MigrationRegistry {
			if m.Id == migrationID {
				migration = &m
				break
			}
		}

		if migration == nil {
			log.WithField(MIGRATION_ID, migrationID).
				Error("Migration not found in registry for rollback")
			continue
		}

		log.WithFields(log.Fields{
			MIGRATION_ID:          migration.Id,
			MIGRATION_DESCRIPTION: migration.Description,
		}).Info("Rolling back migration")

		// Execute rollback in a transaction
		err := db.Transaction(func(tx *gorm.DB) error {
			if err := migration.Down(tx); err != nil {
				return err
			}

			// Remove migration record
			return tx.Where("id = ?", migration.Id).
				Delete(&MigrationRecord{}).
				Error
		})
		if err != nil {
			log.WithFields(log.Fields{
				MIGRATION_ID: migration.Id,
				ERROR:        err,
			}).Error("Failed to rollback migration")
			return fmt.Errorf(
				"failed to rollback migration %s: %w",
				migration.Id,
				err,
			)
		}
	}

	log.Info("Migration rollback completed")
	return nil
}
