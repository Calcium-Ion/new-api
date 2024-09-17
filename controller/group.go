package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"one-api/common"
)

func GetGroups(c *gin.Context) {
	groupNames := make([]string, 0)
	for groupName, _ := range common.GroupRatio {
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
	for groupName, _ := range common.GroupRatio {
		// UserUsableGroups contains the groups that the user can use
		if _, ok := common.UserUsableGroups[groupName]; ok {
			usableGroups[groupName] = common.UserUsableGroups[groupName]
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    usableGroups,
	})
}
