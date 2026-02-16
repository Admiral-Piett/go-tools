package utils

import (
	"crypto/rand"
	"encoding/base64"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (hash, salt string, err error) {
	// Generate random salt
	saltBytes := make([]byte, 16)
	if _, err := rand.Read(saltBytes); err != nil {
		return "", "", err
	}
	salt = base64.StdEncoding.EncodeToString(saltBytes)

	// Hash password with salt
	saltedPassword := password + salt
	hashBytes, err := bcrypt.GenerateFromPassword(
		[]byte(saltedPassword),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return "", "", err
	}

	hash = string(hashBytes)
	return hash, salt, nil
}

func ValidatePassword(password, hash, salt string) bool {
	saltedPassword := password + salt
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(saltedPassword))
	return err == nil
}
