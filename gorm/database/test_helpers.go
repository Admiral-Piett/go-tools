package database

import (
	"log"
	"os"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NOTE: Have to use this date
var DATABASE_DATETIME_FORMAT = "2006-01-02 15:04:05.000"

func NewTestableDatabase() (*Database, sqlmock.Sqlmock) {
	gormDb, mock := newGormDBMock()
	d := &Database{
		db: gormDb,
	}
	return d, mock
}

func newGormDBMock() (*gorm.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		panic(err)
	}

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold:             time.Second,
				LogLevel:                  logger.Info,
				IgnoreRecordNotFoundError: false,
				ParameterizedQueries:      false, // Show actual values
				Colorful:                  false,
			},
		),
	})
	if err != nil {
		panic(err)
	}

	return gormDB, mock
}
