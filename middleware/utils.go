package middleware

import (
	"github.com/gin-gonic/gin"
	"one-api/common"
)

func abortWithMessage(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, gin.H{
		"error": gin.H{
			"message": common.MessageWithRequestId(message, c.GetString(common.RequestIdKey)),
			"type":    "new_api_error",
		},
	})
	c.Abort()
	common.LogError(c.Request.Context(), message)
}
