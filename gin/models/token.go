package models

import "github.com/golang-jwt/jwt"

type AuthClaims struct {
	EncryptedUserID string `json:"uid"`
	DeviceToken     string `json:"device,omitempty"`
	jwt.StandardClaims
}
