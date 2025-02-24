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

// CreateMessage 创建新消息
func CreateMessage(c *gin.Context) {
	var message model.Message
	err := c.ShouldBindJSON(&message)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	newMessage, err := model.CreateMessage(message.ConversationID, message.Role, message.Content)
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
		"data":    newMessage,
	})
}
