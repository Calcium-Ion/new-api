package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"one-api/model"
)

func GetAllQuotaDates(c *gin.Context) {
	dates, err := model.GetAllQuotaDates()
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
		"data":    dates,
	})
	return
}
