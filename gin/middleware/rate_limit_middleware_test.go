package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"

	"github.com/gin-gonic/gin"
)

type MockLimiterStore struct {
	MockContext *limiter.Context
}

func (store *MockLimiterStore) Get(
	ctx context.Context,
	key string,
	rate limiter.Rate,
) (limiter.Context, error) {
	if store.MockContext != nil {
		return *store.MockContext, nil
	}
	return limiter.Context{}, errors.New("boom")
}

func (store *MockLimiterStore) Peek(
	ctx context.Context,
	key string,
	rate limiter.Rate,
) (limiter.Context, error) {
	return limiter.Context{}, errors.New("boom")
}

func (store *MockLimiterStore) Reset(
	ctx context.Context,
	key string,
	rate limiter.Rate,
) (limiter.Context, error) {
	return limiter.Context{}, errors.New("boom")
}

func (store *MockLimiterStore) Increment(
	ctx context.Context,
	key string,
	count int64,
	rate limiter.Rate,
) (limiter.Context, error) {
	return limiter.Context{}, errors.New("boom")
}

func TestNewRateLimitMiddleware(t *testing.T) {
	result := NewRateLimitMiddleware(3, time.Duration(5)*time.Minute)

	assert.Equal(t, time.Duration(5)*time.Minute, result.limiter.Rate.Period)
	assert.Equal(t, int64(3), result.limiter.Rate.Limit)
}

func TestRateLimitMiddleware_Limit_success(t *testing.T) {
	h := RateLimitMiddleware{
		limiter: limiter.New(memory.NewStore(), limiter.Rate{
			Period: time.Duration(5) * time.Minute,
			Limit:  3,
		}),
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/temp", h.Limit())
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/temp", nil)
	r.Header.Add("Authorization", "Bearer valid-token")
	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRateLimitMiddleware_Limit_fetchContextError_500(t *testing.T) {
	store := &MockLimiterStore{}
	h := RateLimitMiddleware{limiter: limiter.New(store, limiter.Rate{
		Period: time.Duration(5) * time.Minute,
		Limit:  3,
	})}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/temp", h.Limit())
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/temp", nil)
	r.Header.Add("Authorization", "Bearer valid-token")
	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestRateLimitMiddleware_Limit_limitReached_429(t *testing.T) {
	store := &MockLimiterStore{
		MockContext: &limiter.Context{
			Reached: true,
		},
	}
	h := RateLimitMiddleware{limiter: limiter.New(store, limiter.Rate{
		Period: time.Duration(5) * time.Minute,
		Limit:  3,
	})}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/temp", h.Limit())
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/temp", nil)
	r.Header.Add("Authorization", "Bearer valid-token")
	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
}
