package models

import "time"

var SQL_NOT_FOUND = "record not found"

type User struct {
	Id            int
	Email         string
	PasswordHash  string
	PasswordSalt  string
	DeviceToken   string
	EmailVerified bool
	LastLoginAt   time.Time
}
