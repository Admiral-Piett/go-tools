package handlers

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// HealthCheckResponse represents the health check postLoginResponse
type HealthCheckResponse struct {
	Status    string    `json:"status"`
	Service   string    `json:"service"`
	Version   string    `json:"version"`
	GitSha    string    `json:"sha"`
	Timestamp time.Time `json:"timestamp"`
	Uptime    string    `json:"uptime"`
}

var startTime = time.Now()

// HealthCheck handles health check requests for Cloud Run
func HealthCheck(c *gin.Context) {
	uptime := time.Since(startTime).Round(time.Second)

	response := HealthCheckResponse{
		Status:    "healthy",
		Service:   "polytracker-backend",
		Version:   os.Getenv("VERSION"),
		GitSha:    os.Getenv("GIT_SHA"),
		Timestamp: time.Now().UTC(),
		Uptime:    uptime.String(),
	}

	log.Debug("Health check requested")
	c.JSON(http.StatusOK, response)
}
