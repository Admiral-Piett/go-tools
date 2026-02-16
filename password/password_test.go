package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	password     = "password"
	passwordHash = "$2a$10$yBIDtRKQQM4uP0MlYLjHlO4wNvYlJBZ872drjRAzkLzTSobGZZZHK"
	passwordSalt = "dNSczLZ/bqPL5GHpyx+Y1w=="
)

func TestHashPassword_success(t *testing.T) {
	hash, salt, err := HashPassword(password)

	assert.Nil(t, err)
	assert.NotEqual(t, "", hash)
	assert.NotEqual(t, "", salt)
}

func TestValidatePassword_success(t *testing.T) {
	ok := ValidatePassword(
		password,
		passwordHash,
		passwordSalt,
	)

	assert.True(t, ok)
}

func TestValidatePassword_invalidHash_failure(t *testing.T) {
	ok := ValidatePassword(
		password,
		"invalid",
		passwordSalt,
	)

	assert.False(t, ok)
}

func TestValidatePassword_invalidSalt_failure(t *testing.T) {
	ok := ValidatePassword(
		password,
		passwordHash,
		"invalid",
	)

	assert.False(t, ok)
}
