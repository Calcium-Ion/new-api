package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"one-api/model"
	"one-api/setting"
)

func GetGroups(c *gin.Context) {
	groupNames := make([]string, 0)
	for groupName, _ := range setting.GetGroupRatioCopy() {
		groupNames = append(groupNames, groupName)
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    groupNames,
	})
}

func GetUserGroups(c *gin.Context) {
	usableGroups := make(map[string]string)
	userGroup := ""
	userId := c.GetInt("id")
	userGroup, _ = model.CacheGetUserGroup(userId)
	for groupName, _ := range setting.GetGroupRatioCopy() {
		// UserUsableGroups contains the groups that the user can use
		userUsableGroups := setting.GetUserUsableGroups(userGroup)
		if _, ok := userUsableGroups[groupName]; ok {
			usableGroups[groupName] = userUsableGroups[groupName]
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    usableGroups,
	})
}
