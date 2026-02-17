package handlers

import (
	"encoding/json"
	"github.com/Admiral-Piett/go-tools/gin/test_helpers"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHealthCheck_Success(t *testing.T) {
	os.Setenv("VERSION", "local.0")
	os.Setenv("GIT_SHA", "git-sha")
	defer func() {
		os.Unsetenv("VERSION")
		os.Unsetenv("GIT_SHA")
	}()
	w := test_helpers.ServeRequest("GET", "/health", HealthCheck, nil)

	assert.Equal(t, http.StatusOK, w.Code)

	// Parse response
	var response HealthCheckResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.Nil(t, err)

	assert.Equal(t, "healthy", response.Status)
	assert.Equal(t, "polytracker-backend", response.Service)
	assert.Equal(t, "local.0", response.Version)
	assert.Equal(t, "git-sha", response.GitSha)
}
