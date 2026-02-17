package database

import (
	"fmt"
	"regexp"
	"testing"

	"gorm.io/gorm"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestRegisterMigration(t *testing.T) {
	MigrationRegistry = []Migration{}
	defer func() {
		MigrationRegistry = []Migration{}
	}()

	inputMigration := Migration{
		Id:          "001_test_migration",
		Description: "Test Migration 1",
	}
	RegisterMigration(inputMigration)

	assert.Equal(t, []Migration{inputMigration}, MigrationRegistry)
}

func TestRunMigrations(t *testing.T) {
	defer func() {
		MigrationRegistry = []Migration{}
	}()

	upCalled := false
	upFunc := func(*gorm.DB) error {
		upCalled = true
		return nil
	}
	downCalled := false
	downFunc := func(*gorm.DB) error {
		downCalled = true
		return nil
	}
	inputMigration := Migration{
		Id:          "001_test_migration",
		Description: "Test Migration 1",
		Up:          upFunc,
		Down:        downFunc,
	}
	RegisterMigration(inputMigration)

	gormDb, mock := newGormDBMock()
	d := &Database{
		db: gormDb,
	}

	mock.ExpectExec(regexp.QuoteMeta(`CREATE TABLE "migration_records" ("id" text,"description" text NOT NULL,"applied_at" timestamptz NOT NULL,PRIMARY KEY ("id"))`)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "migration_records" WHERE id = $1`)).
		WithArgs(inputMigration.Id)
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "migration_records" ("id","description","applied_at") VALUES ($1,$2,$3)`)).
		WithArgs(
			inputMigration.Id,
			inputMigration.Description,
			sqlmock.AnyArg(), // time.Now()
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := RunMigrations(d)
	assert.Nil(t, err)

	assert.True(t, upCalled)
	assert.False(t, downCalled)

	err = mock.ExpectationsWereMet()
	assert.Nil(t, err)
}

func TestRunMigrations_already_applied(t *testing.T) {
	defer func() {
		MigrationRegistry = []Migration{}
	}()

	upCalled := false
	upFunc := func(*gorm.DB) error {
		upCalled = true
		return nil
	}
	downCalled := false
	downFunc := func(*gorm.DB) error {
		downCalled = true
		return nil
	}
	inputMigration := Migration{
		Id:          "001_test_migration",
		Description: "Test Migration 1",
		Up:          upFunc,
		Down:        downFunc,
	}

	RegisterMigration(inputMigration)

	gormDb, mock := newGormDBMock()
	d := &Database{
		db: gormDb,
	}

	rows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	mock.ExpectExec(regexp.QuoteMeta(`CREATE TABLE "migration_records" ("id" text,"description" text NOT NULL,"applied_at" timestamptz NOT NULL,PRIMARY KEY ("id"))`)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "migration_records" WHERE id = $1`)).
		WithArgs(inputMigration.Id).
		WillReturnRows(rows)

	err := RunMigrations(d)
	assert.Nil(t, err)

	assert.False(t, upCalled)
	assert.False(t, downCalled)

	err = mock.ExpectationsWereMet()
	assert.Nil(t, err)
}

func TestRunMigrations_AutoMigrations_failure(t *testing.T) {
	defer func() {
		MigrationRegistry = []Migration{}
	}()

	upCalled := false
	upFunc := func(*gorm.DB) error {
		upCalled = true
		return nil
	}
	downCalled := false
	downFunc := func(*gorm.DB) error {
		downCalled = true
		return nil
	}
	inputMigration := Migration{
		Id:          "001_test_migration",
		Description: "Test Migration 1",
		Up:          upFunc,
		Down:        downFunc,
	}
	RegisterMigration(inputMigration)

	gormDb, mock := newGormDBMock()
	d := &Database{
		db: gormDb,
	}

	mock.ExpectExec(regexp.QuoteMeta(`CREATE TABLE "migration_records" ("id" text,"description" text NOT NULL,"applied_at" timestamptz NOT NULL,PRIMARY KEY ("id"))`)).
		WillReturnError(fmt.Errorf("boom"))

	err := RunMigrations(d)
	assert.Error(t, err)

	assert.False(t, upCalled)
	assert.False(t, downCalled)

	err = mock.ExpectationsWereMet()
	assert.Nil(t, err)
}

func TestRunMigrations_Transaction_failure(t *testing.T) {
	defer func() {
		MigrationRegistry = []Migration{}
	}()

	upCalled := false
	upFunc := func(*gorm.DB) error {
		upCalled = true
		return fmt.Errorf("boom")
	}
	downCalled := false
	downFunc := func(*gorm.DB) error {
		downCalled = true
		return nil
	}
	inputMigration := Migration{
		Id:          "001_test_migration",
		Description: "Test Migration 1",
		Up:          upFunc,
		Down:        downFunc,
	}

	RegisterMigration(inputMigration)

	gormDb, mock := newGormDBMock()
	d := &Database{
		db: gormDb,
	}

	mock.ExpectExec(regexp.QuoteMeta(`CREATE TABLE "migration_records" ("id" text,"description" text NOT NULL,"applied_at" timestamptz NOT NULL,PRIMARY KEY ("id"))`)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "migration_records" WHERE id = $1`)).
		WithArgs(inputMigration.Id)
	mock.ExpectBegin()
	mock.ExpectRollback()

	err := RunMigrations(d)
	assert.Error(t, err)

	assert.True(t, upCalled)
	assert.False(t, downCalled)

	err = mock.ExpectationsWereMet()
	assert.Nil(t, err)
}

func TestRunMigrations_Transaction_Multi_failure_WithInCodeRollback(
	t *testing.T,
) {
	defer func() {
		MigrationRegistry = []Migration{}
	}()

	upCalled := false
	upFunc := func(*gorm.DB) error {
		// Fail on the second one
		if upCalled == true {
			return fmt.Errorf("boom")
		}
		upCalled = true
		return nil
	}
	downCalled := false
	downFunc := func(*gorm.DB) error {
		downCalled = true
		return nil
	}
	inputMigration := Migration{
		Id:          "001_test_migration",
		Description: "Test Migration 1",
		Up:          upFunc,
		Down:        downFunc,
	}
	inputMigration2 := Migration{
		Id:          "002_test_migration",
		Description: "Test Migration 2",
		Up:          upFunc,
		Down:        downFunc,
	}

	RegisterMigration(inputMigration)
	RegisterMigration(inputMigration2)

	gormDb, mock := newGormDBMock()
	d := &Database{
		db: gormDb,
	}

	// Baseline set up
	mock.ExpectExec(regexp.QuoteMeta(`CREATE TABLE "migration_records" ("id" text,"description" text NOT NULL,"applied_at" timestamptz NOT NULL,PRIMARY KEY ("id"))`)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	// First Migration
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "migration_records" WHERE id = $1`)).
		WithArgs(inputMigration.Id)
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "migration_records" ("id","description","applied_at") VALUES ($1,$2,$3)`)).
		WithArgs(
			inputMigration.Id,
			inputMigration.Description,
			sqlmock.AnyArg(), // time.Now()
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// Second Migration w/ Failure
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "migration_records" WHERE id = $1`)).
		WithArgs(inputMigration2.Id)
	mock.ExpectBegin()
	mock.ExpectRollback()

	// First Migration Rollback
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "migration_records" WHERE id = $1`)).
		WithArgs(
			inputMigration.Id,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := RunMigrations(d)
	assert.Error(t, err)

	assert.True(t, upCalled)
	assert.True(t, downCalled)

	err = mock.ExpectationsWereMet()
	assert.Nil(t, err)
}
