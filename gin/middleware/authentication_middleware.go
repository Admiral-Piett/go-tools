package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/Admiral-Piett/go-tools/gin/interfaces"
	"github.com/Admiral-Piett/go-tools/gin/models"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type AuthMiddleware struct {
	tokenService interfaces.TokenServiceInterface
}

func NewAuthMiddleware(tokenService interfaces.TokenServiceInterface) *AuthMiddleware {
	return &AuthMiddleware{
		tokenService: tokenService,
	}
}

func (am *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		ctx, err := am.validateAuthHeader(c.Request)
		if err != nil {
			log.WithError(err).Warning("Validate Auth Header Failure")

			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				models.ErrorResponses.UnauthorizedError,
			)
			return
		}

		// Update request context
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})
}

func (am *AuthMiddleware) validateAuthHeader(
	r *http.Request,
) (context.Context, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return r.Context(), errors.New("authorization header missing")
	}

	// Parse Bearer token
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return r.Context(), errors.New("invalid authorization header format")
	}

	tokenString := parts[1]

	// Validate access token
	claims, err := am.tokenService.ValidateAccessToken(tokenString)
	if err != nil {
		return r.Context(), err
	}

	// Decrypt user ID
	userId, err := am.tokenService.DecryptUserID(claims.EncryptedUserID)
	if err != nil {
		return r.Context(), err
	}

	// Add user context
	ctx := context.WithValue(r.Context(), "userId", userId)
	ctx = context.WithValue(ctx, "deviceToken", claims.DeviceToken)

	return ctx, nil
}
