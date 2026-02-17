package database

import (
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestDatabase_Close_success(t *testing.T) {
	gormDb, mock := newGormDBMock()
	d := &Database{
		db: gormDb,
	}
	mock.ExpectClose()

	err := d.Close()
	assert.Nil(t, err)

	err = mock.ExpectationsWereMet()
	assert.Nil(t, err)
}

func TestDatabase_DB_success(t *testing.T) {
	gormDb, _ := newGormDBMock()
	d := &Database{
		db: gormDb,
	}

	result := d.DB()
	assert.Equal(t, gormDb, result)
}

func TestDatabase_AutoMigrate_success(t *testing.T) {
	gormDb, mock := newGormDBMock()
	d := &Database{
		db: gormDb,
	}
	// For some reason in the underlying package this had to return a result even if that doesn't make sense,
	// and the regex bit escapes a bunch of characters.
	mock.ExpectExec(regexp.QuoteMeta(`CREATE TABLE "migration_records" ("id" text,"description" text NOT NULL,"applied_at" timestamptz NOT NULL,PRIMARY KEY ("id"))`)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := d.AutoMigrate(&MigrationRecord{})
	assert.Nil(t, err)
	err = mock.ExpectationsWereMet()
	assert.Nil(t, err)
}

func TestDatabase_Model_success(t *testing.T) {
	gormDb, mock := newGormDBMock()
	d := &Database{
		db: gormDb,
	}

	type basicTable struct {
		ID        int       `gorm:"primaryKey"`
		CreatedAt time.Time `gorm:"not null"`
	}

	now := time.Now()

	expectedModel := basicTable{
		ID:        1,
		CreatedAt: now,
	}

	rows := sqlmock.NewRows([]string{"id", "created_at"}).
		AddRow(1, now)

	mock.ExpectExec(regexp.QuoteMeta(`CREATE TABLE "basic_tables" ("id" bigserial,"created_at" timestamptz NOT NULL,PRIMARY KEY ("id"))`)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("SELECT \\* FROM \"basic_tables\" WHERE id = (.+) ORDER BY \"basic_tables\".\"id\" LIMIT (.+)").
		WillReturnRows(rows)

	// Create the table we can fetch against with a model
	err := d.AutoMigrate(&basicTable{})
	assert.Nil(t, err)

	result := basicTable{}
	d.Model(&basicTable{}).Where("id = 1").First(&result)
	assert.Equal(t, expectedModel, result)

	err = mock.ExpectationsWereMet()
	assert.Nil(t, err)
}

func TestDatabase_Transaction_success(t *testing.T) {
	gormDb, mock := newGormDBMock()
	d := &Database{
		db: gormDb,
	}

	mock.ExpectBegin()
	mock.ExpectCommit()

	inputFunc := func(tx *gorm.DB) error { return nil }

	err := d.Transaction(inputFunc)
	assert.Nil(t, err)
}
