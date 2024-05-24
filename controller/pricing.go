package controller

import (
	"github.com/gin-gonic/gin"
	"one-api/common"
	"one-api/model"
)

func GetPricing(c *gin.Context) {
	userId := c.GetInt("id")
	// if no login, get default group ratio
	groupRatio := common.GetGroupRatio("default")
	group, err := model.CacheGetUserGroup(userId)
	if err == nil {
		groupRatio = common.GetGroupRatio(group)
	}
	pricing := model.GetPricing(group)
	c.JSON(200, gin.H{
		"success":     true,
		"data":        pricing,
		"group_ratio": groupRatio,
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
