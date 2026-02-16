package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Admiral-Piett/go-tools/gin/models"

	"github.com/stretchr/testify/assert"

	"github.com/gin-gonic/gin"

	"github.com/Admiral-Piett/go-tools/gin/mocks"
)

func TestAuthMiddleware_RequireAuth_success(t *testing.T) {
	tok := &mocks.MockTokenService{}
	tok.MockValidateAccessToken = func(tokenString string) (*models.AuthClaims, error) {
		r := &models.AuthClaims{
			EncryptedUserID: "encrypted-user-id",
		}
		return r, nil
	}
	h := AuthMiddleware{tokenService: tok}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/temp", h.RequireAuth())
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/temp", nil)
	r.Header.Add("Authorization", "Bearer valid-token")
	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)

	assert.Equal(
		t,
		[]interface{}{"valid-token"},
		tok.ValidateAccessTokenCalledWith,
	)
	assert.Equal(
		t,
		[]interface{}{"encrypted-user-id"},
		tok.DecryptUserIDCalledWith,
	)
}

func TestAuthMiddleware_RequireAuth_missingAuthHeader_401(t *testing.T) {
	h := AuthMiddleware{tokenService: &mocks.MockTokenService{}}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/temp", h.RequireAuth())
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/temp", nil)
	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_RequireAuth_tooManyPieces_401(t *testing.T) {
	h := AuthMiddleware{tokenService: &mocks.MockTokenService{}}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/temp", h.RequireAuth())
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/temp", nil)
	r.Header.Add("Authorization", "Bearer valid-token extra-nonsense")
	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_RequireAuth_tooFewPieces_401(t *testing.T) {
	h := AuthMiddleware{tokenService: &mocks.MockTokenService{}}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/temp", h.RequireAuth())
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/temp", nil)
	r.Header.Add("Authorization", "Bearer")
	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_RequireAuth_missingBearer_401(t *testing.T) {
	h := AuthMiddleware{tokenService: &mocks.MockTokenService{}}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/temp", h.RequireAuth())
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/temp", nil)
	r.Header.Add("Authorization", "valid-token extra-nonsense")
	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_RequireAuth_invalidAccessToken_401(t *testing.T) {
	tok := &mocks.MockTokenService{}
	tok.MockValidateAccessToken = func(tokenString string) (*models.AuthClaims, error) {
		return &models.AuthClaims{}, errors.New("boom")
	}
	h := AuthMiddleware{tokenService: tok}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/temp", h.RequireAuth())
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/temp", nil)
	r.Header.Add("Authorization", "Bearer valid-token")
	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_RequireAuth_unableToDecryptUserId_401(t *testing.T) {
	tok := &mocks.MockTokenService{}
	tok.MockDecryptUserID = func(encryptedUserID string) (int, error) {
		return 0, errors.New("boom")
	}
	h := AuthMiddleware{tokenService: tok}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/temp", h.RequireAuth())
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/temp", nil)
	r.Header.Add("Authorization", "Bearer valid-token")
	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
