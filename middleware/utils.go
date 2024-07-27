package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"one-api/common"
)

func abortWithOpenAiMessage(c *gin.Context, statusCode int, message string) {
	userId := c.GetInt("id")
	c.JSON(statusCode, gin.H{
		"error": gin.H{
			"message": common.MessageWithRequestId(message, c.GetString(common.RequestIdKey)),
			"type":    "new_api_error",
		},
	})
	c.Abort()
	common.LogError(c.Request.Context(), fmt.Sprintf("user %d | %s", userId, message))
}

func abortWithMidjourneyMessage(c *gin.Context, statusCode int, code int, description string) {
	c.JSON(statusCode, gin.H{
		"description": description,
		"type":        "new_api_error",
		"code":        code,
	})
	c.Abort()
	common.LogError(c.Request.Context(), description)
}
