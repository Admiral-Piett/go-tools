package utils

import (
	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
)

func GetUserId(c *gin.Context) (int, bool) {
	v := c.Request.Context().Value("userId")
	if v == nil {
		return 0, false
	}
	userId, ok := v.(int)
	if !ok {
		log.Warning("userId invalid type")
		return 0, false
	}
	return userId, true
}
