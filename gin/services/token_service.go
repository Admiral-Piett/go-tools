package services

import (
	"encoding/hex"
	"errors"
	"strconv"
	"time"

	"github.com/Admiral-Piett/go-tools/gin/interfaces"

	"github.com/Admiral-Piett/go-tools/settings"

	"github.com/Admiral-Piett/go-tools/encryption"
	"github.com/Admiral-Piett/go-tools/gin/models"
	"github.com/golang-jwt/jwt"
)

type TokenService struct {
	jwtSecret     []byte
	encryptionKey []byte
	accessTTL     time.Duration
	refreshTTL    time.Duration
	appName       string
}

func NewTokenService(
	cfg *settings.BaseSettings,
) interfaces.TokenServiceInterface {
	decodedJwtHmacKey, _ := hex.DecodeString(cfg.JwtHmacKey)
	decodedEncryptionKey, _ := hex.DecodeString(cfg.EncryptionKey)
	return &TokenService{
		jwtSecret:     decodedJwtHmacKey,
		encryptionKey: decodedEncryptionKey,
		accessTTL:     time.Duration(cfg.JwtAccessTokenTTL) * time.Minute,
		refreshTTL:    time.Duration(cfg.JwtRefreshTokenTTL) * time.Minute,
		appName:       cfg.AppName,
	}
}

func (ts *TokenService) GenerateTokenResponse(
	user interfaces.UserModelInterface,
) (*models.TokenResponse, error) {
	// Encrypt user ID
	encryptedID, err := encryption.EncryptAES(
		strconv.Itoa(user.GetUserId()),
		ts.encryptionKey,
	)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	accessExp := now.Add(ts.accessTTL)
	refreshExp := now.Add(ts.refreshTTL)

	// Access token claims
	accessClaims := &models.AuthClaims{
		EncryptedUserID: encryptedID,
		DeviceToken:     user.GetDeviceToken(),
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: accessExp.Unix(),
			IssuedAt:  now.Unix(),
			NotBefore: now.Unix(),
			Issuer:    ts.appName,
			Subject:   "access",
		},
	}

	// Generate access token
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(ts.jwtSecret)
	if err != nil {
		return nil, err
	}

	// Refresh token (simpler claims)
	refreshClaims := &jwt.StandardClaims{
		ExpiresAt: refreshExp.Unix(),
		IssuedAt:  now.Unix(),
		Issuer:    "polytracker",
		Subject:   "refresh",
		Id:        encryptedID,
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(ts.jwtSecret)
	if err != nil {
		return nil, err
	}

	return &models.TokenResponse{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresAt:    accessExp,
	}, nil
}

func (ts *TokenService) ValidateAccessToken(
	tokenString string,
) (*models.AuthClaims, error) {
	claims := &models.AuthClaims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			// Verify signing method to prevent algorithm confusion attacks
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.NewValidationError(
					"unexpected signing method",
					jwt.ValidationErrorSignatureInvalid,
				)
			}
			return ts.jwtSecret, nil
		},
	)
	if err != nil {
		return nil, errors.New(
			err.Error(),
		) // repackage the internal errors so their easily accessible for logging
	}
	if !token.Valid || claims.Subject != "access" {
		return nil, errors.New("token invalid")
	}

	return claims, nil
}

func (ts *TokenService) ValidateRefreshToken(
	tokenString string,
) (string, error) {
	claims := &jwt.StandardClaims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			// Verify signing method to prevent algorithm confusion attacks
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.NewValidationError(
					"unexpected signing method",
					jwt.ValidationErrorSignatureInvalid,
				)
			}
			return ts.jwtSecret, nil
		},
	)
	if err != nil {
		return "", errors.New(
			err.Error(),
		) // repackage the internal errors so their easily accessible for logging
	}
	if !token.Valid || claims.Subject != "refresh" {
		return "", errors.New("token invalid")
	}

	// Return encrypted user ID from the refresh token
	return claims.Id, nil
}

func (ts *TokenService) DecryptUserID(encryptedUserID string) (int, error) {
	stringValue, err := encryption.DecryptAES(encryptedUserID, ts.encryptionKey)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(stringValue)
}
