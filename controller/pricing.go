package controller

import (
	"github.com/gin-gonic/gin"
	"one-api/common"
	"one-api/model"
)

func GetPricing(c *gin.Context) {
	pricing := model.GetPricing()
	c.JSON(200, gin.H{
		"success":     true,
		"data":        pricing,
		"group_ratio": common.GroupRatio,
	})
}

func ResetModelRatio(c *gin.Context) {
	defaultStr := common.DefaultModelRatio2JSONString()
	err := model.UpdateOption("ModelRatio", defaultStr)
	if err != nil {
		c.JSON(200, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	err = common.UpdateModelRatioByJSONString(defaultStr)
	if err != nil {
		c.JSON(200, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(200, gin.H{
		"success": true,
		"message": "重置模型倍率成功",
	})
}
