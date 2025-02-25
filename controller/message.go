package controller

import (
	"net/http"
	"one-api/model"

	"github.com/gin-gonic/gin"
)

// GetMessages 获取会话的所有消息
func GetMessages(c *gin.Context) {
	conversationId := c.Param("conversation_id")
	messages, err := model.GetMessagesByConversationID(conversationId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    messages,
	})
}
