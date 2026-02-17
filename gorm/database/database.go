package database

import (
	"database/sql"
	"fmt"
	"github.com/Admiral-Piett/go-tools/gorm/interfaces"
	"github.com/Admiral-Piett/go-tools/settings"
	"gorm.io/driver/sqlite"
	"strings"

	log "github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NOTE: You're not really going to be able to unit test this layer.  Leave it for smoke tests.

// Database wraps the GORM database connection
type Database struct {
	db *gorm.DB
}

// NewDatabase creates a new database connection
func NewDatabase(cfg *settings.BaseSettings) (interfaces.DatabaseInterface, error) {
	log.Info("Connecting to database")

	// Configure GORM logger based on environment
	var gormLogger logger.Interface
	switch strings.ToUpper(cfg.SqlLogLevel) {
	case "INFO":
		gormLogger = logger.Default.LogMode(logger.Info)
	case "WARNING":
		gormLogger = logger.Default.LogMode(logger.Warn)
	case "ERROR":
		gormLogger = logger.Default.LogMode(logger.Error)
	default:
		gormLogger = logger.Default.LogMode(logger.Silent)
	}

	var db *gorm.DB
	var err error
	if strings.ToLower(cfg.SqlType) == "postgres" {
		db, err = gorm.Open(sqlite.Open(cfg.SqlUri), &gorm.Config{
			Logger: gormLogger,
		})
	} else {
		db, err = gorm.Open(postgres.Open(cfg.SqlUri), &gorm.Config{
			Logger: gormLogger,
		})
	}
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Info("Database connection established successfully")
	return &Database{db: db}, nil
}

// Close closes the database connection
func (d *Database) Close() error {
	sqlDB, err := d.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// NOTE: honestly, I don't know if these are worth it...we could go with sqlmock the whole way.
// Close closes the database connection
func (d *Database) DB() *gorm.DB {
	return d.db
}

func (d *Database) AutoMigrate(dst ...interface{}) error {
	return d.db.AutoMigrate(dst...)
}

// This returns a database instance linked to a given model/table.  It's tricky to do a local unit test.
func (d *Database) Model(value interface{}) *gorm.DB {
	return d.db.Model(value)
}

func (d *Database) Transaction(
	fc func(tx *gorm.DB) error,
	opts ...*sql.TxOptions,
) error {
	return d.db.Transaction(fc, opts...)
}
