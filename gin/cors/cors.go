package cors

import (
	"github.com/Admiral-Piett/go-tools/settings"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func SetupCORS(cfg *settings.BaseSettings) gin.HandlerFunc {
	corsConfig := cors.DefaultConfig()

	corsConfig.AllowOrigins = cfg.AllowedOriginsSlice
	corsConfig.AllowMethods = []string{
		"GET",
		"POST",
		"PUT",
		"DELETE",
		"OPTIONS",
	}
	corsConfig.AllowHeaders = []string{
		"Origin",
		"Content-Type",
		"Accept",
		"Authorization",
	}
	corsConfig.AllowCredentials = true

	log.WithField("ALLOWED_ORIGINS", cfg.AllowedOriginsSlice).
		Debug("CORS configured")

	return cors.New(corsConfig)
}
