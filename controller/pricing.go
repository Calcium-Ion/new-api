package controller

import (
	"github.com/gin-gonic/gin"
	"one-api/common"
	"one-api/model"
)

func GetPricing(c *gin.Context) {
	pricing := model.GetPricing()
	userId, exists := c.Get("id")
	usableGroup := map[string]string{}
	groupRatio := common.GroupRatio
	var group string
	if exists {
		user, err := model.GetChannelById(userId.(int), false)
		if err != nil {
			c.JSON(200, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}
		group = user.Group
	}

	usableGroup = common.GetUserUsableGroups(group)
	// check groupRatio contains usableGroup
	for group := range common.GroupRatio {
		if _, ok := usableGroup[group]; !ok {
			delete(groupRatio, group)
		}
	}

	c.JSON(200, gin.H{
		"success":      true,
		"data":         pricing,
		"group_ratio":  groupRatio,
		"usable_group": usableGroup,
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
