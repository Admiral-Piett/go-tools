package interfaces

import (
    "github.com/Admiral-Piett/go-tools/gin/models"
)

type TokenServiceInterface interface {
    GenerateTokenResponse(user UserModelInterface) (*models.TokenResponse, error)
    ValidateAccessToken(tokenString string) (*models.AuthClaims, error)
    ValidateRefreshToken(tokenString string) (string, error)
    DecryptUserID(encryptedUserID string) (int, error)
}
