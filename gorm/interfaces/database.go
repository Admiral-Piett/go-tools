package interfaces

import (
	"database/sql"
	"gorm.io/gorm"
)

type DatabaseInterface interface {
	Close() error
	DB() *gorm.DB
	AutoMigrate(dst ...interface{}) error
	Model(value interface{}) *gorm.DB
	Transaction(
		fc func(tx *gorm.DB) error,
		opts ...*sql.TxOptions,
	) error
}
