package controller

import (
	"github.com/gin-gonic/gin"
	"one-api/common"
	"one-api/constant"
	"one-api/model"
	"one-api/setting/operation_setting"
	"time"
)

type Setup struct {
	Status       bool   `json:"status"`
	RootInit     bool   `json:"root_init"`
	DatabaseType string `json:"database_type"`
}

type SetupRequest struct {
	Username           string `json:"username"`
	Password           string `json:"password"`
	ConfirmPassword    string `json:"confirmPassword"`
	SelfUseModeEnabled bool   `json:"SelfUseModeEnabled"`
	DemoSiteEnabled    bool   `json:"DemoSiteEnabled"`
}

func GetSetup(c *gin.Context) {
	setup := Setup{
		Status: constant.Setup,
	}
	if constant.Setup {
		c.JSON(200, gin.H{
			"success": true,
			"data":    setup,
		})
		return
	}
	setup.RootInit = model.RootUserExists()
	if common.UsingMySQL {
		setup.DatabaseType = "mysql"
	}
	if common.UsingPostgreSQL {
		setup.DatabaseType = "postgres"
	}
	if common.UsingSQLite {
		setup.DatabaseType = "sqlite"
	}
	c.JSON(200, gin.H{
		"success": true,
		"data":    setup,
	})
}

func PostSetup(c *gin.Context) {
	// Check if setup is already completed
	if constant.Setup {
		c.JSON(400, gin.H{
			"success": false,
			"message": "系统已经初始化完成",
		})
		return
	}

	// Check if root user already exists
	rootExists := model.RootUserExists()

	var req SetupRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(400, gin.H{
			"success": false,
			"message": "请求参数有误",
		})
		return
	}

	// If root doesn't exist, validate and create admin account
	if !rootExists {
		// Validate password
		if req.Password != req.ConfirmPassword {
			c.JSON(400, gin.H{
				"success": false,
				"message": "两次输入的密码不一致",
			})
			return
		}

		if len(req.Password) < 8 {
			c.JSON(400, gin.H{
				"success": false,
				"message": "密码长度至少为8个字符",
			})
			return
		}

		// Create root user
		hashedPassword, err := common.Password2Hash(req.Password)
		if err != nil {
			c.JSON(500, gin.H{
				"success": false,
				"message": "系统错误: " + err.Error(),
			})
			return
		}
		rootUser := model.User{
			Username:    req.Username,
			Password:    hashedPassword,
			Role:        common.RoleRootUser,
			Status:      common.UserStatusEnabled,
			DisplayName: "Root User",
			AccessToken: nil,
			Quota:       100000000,
		}
		err = model.DB.Create(&rootUser).Error
		if err != nil {
			c.JSON(500, gin.H{
				"success": false,
				"message": "创建管理员账号失败: " + err.Error(),
			})
			return
		}
	}

	// Set operation modes
	operation_setting.SelfUseModeEnabled = req.SelfUseModeEnabled
	operation_setting.DemoSiteEnabled = req.DemoSiteEnabled

	// Save operation modes to database for persistence
	err = model.UpdateOption("SelfUseModeEnabled", boolToString(req.SelfUseModeEnabled))
	if err != nil {
		c.JSON(500, gin.H{
			"success": false,
			"message": "保存自用模式设置失败: " + err.Error(),
		})
		return
	}

	err = model.UpdateOption("DemoSiteEnabled", boolToString(req.DemoSiteEnabled))
	if err != nil {
		c.JSON(500, gin.H{
			"success": false,
			"message": "保存演示站点模式设置失败: " + err.Error(),
		})
		return
	}

	// Update setup status
	constant.Setup = true

	setup := model.Setup{
		Version:       common.Version,
		InitializedAt: time.Now().Unix(),
	}
	err = model.DB.Create(&setup).Error
	if err != nil {
		c.JSON(500, gin.H{
			"success": false,
			"message": "系统初始化失败: " + err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "系统初始化成功",
	})
}

func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
