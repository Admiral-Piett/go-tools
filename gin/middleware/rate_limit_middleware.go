package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Admiral-Piett/go-tools/gin/models"

	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

type RateLimitMiddleware struct {
	limiter *limiter.Limiter
}

func NewRateLimitMiddleware(
	limit int,
	window time.Duration,
) *RateLimitMiddleware {
	store := memory.NewStore()
	rate := limiter.Rate{
		Period: window,
		Limit:  int64(limit),
	}

	return &RateLimitMiddleware{
		limiter: limiter.New(store, rate),
	}
}

func (rlm *RateLimitMiddleware) Limit() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Use client IP as the key for rate limiting
		// TODO - test w/ gcloud headers as we might not get the real client IP anymore?
		key := c.ClientIP()

		context, err := rlm.limiter.Get(c, key)
		if err != nil {
			c.AbortWithStatusJSON(
				http.StatusInternalServerError,
				models.ErrorResponses.GeneralError,
			)
			return
		}

		// Add rate limit headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", context.Limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", context.Remaining))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", context.Reset))

		if context.Reached {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "Too many requests",
				"retry_after": context.Reset - time.Now().Unix(),
			})
			return
		}

		c.Next()
	})
}
