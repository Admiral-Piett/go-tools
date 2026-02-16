package models

import "time"

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type TokenResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}
