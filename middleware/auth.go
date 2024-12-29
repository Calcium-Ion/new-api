package middleware

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
	"one-api/common"
	"one-api/model"
	"strconv"
	"strings"
)

func validUserInfo(username string, role int) bool {
	// check username is empty
	if strings.TrimSpace(username) == "" {
		return false
	}
	if !common.IsValidateRole(role) {
		return false
	}
	return true
}

func authHelper(c *gin.Context, minRole int) {
	session := sessions.Default(c)
	username := session.Get("username")
	role := session.Get("role")
	id := session.Get("id")
	status := session.Get("status")
	useAccessToken := false
	if username == nil {
		// Check access token
		accessToken := c.Request.Header.Get("Authorization")
		if accessToken == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "无权进行此操作，未登录且未提供 access token",
			})
			c.Abort()
			return
		}
		user := model.ValidateAccessToken(accessToken)
		if user != nil && user.Username != "" {
			if !validUserInfo(user.Username, user.Role) {
				c.JSON(http.StatusOK, gin.H{
					"success": false,
					"message": "无权进行此操作，用户信息无效",
				})
				c.Abort()
				return
			}
			// Token is valid
			username = user.Username
			role = user.Role
			id = user.Id
			status = user.Status
			useAccessToken = true
		} else {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "无权进行此操作，access token 无效",
			})
			c.Abort()
			return
		}
	}
	if !useAccessToken {
		// get header New-Api-User
		apiUserIdStr := c.Request.Header.Get("New-Api-User")
		if apiUserIdStr == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "无权进行此操作，请刷新页面或清空缓存后重试",
			})
			c.Abort()
			return
		}
		apiUserId, err := strconv.Atoi(apiUserIdStr)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "无权进行此操作，登录信息无效，请重新登录",
			})
			c.Abort()
			return

		}
		if id != apiUserId {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "无权进行此操作，与登录用户不匹配，请重新登录",
			})
			c.Abort()
			return
		}
	}
	if status.(int) == common.UserStatusDisabled {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "用户已被封禁",
		})
		c.Abort()
		return
	}
	if role.(int) < minRole {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无权进行此操作，权限不足",
		})
		c.Abort()
		return
	}
	if !validUserInfo(username.(string), role.(int)) {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无权进行此操作，用户信息无效",
		})
		c.Abort()
		return
	}
	c.Set("username", username)
	c.Set("role", role)
	c.Set("id", id)
	c.Set("group", session.Get("group"))
	c.Set("use_access_token", useAccessToken)
	c.Next()
}

func TryUserAuth() func(c *gin.Context) {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		id := session.Get("id")
		if id != nil {
			c.Set("id", id)
		}
		c.Next()
	}
}

func UserAuth() func(c *gin.Context) {
	return func(c *gin.Context) {
		authHelper(c, common.RoleCommonUser)
	}
}

func AdminAuth() func(c *gin.Context) {
	return func(c *gin.Context) {
		authHelper(c, common.RoleAdminUser)
	}
}

func RootAuth() func(c *gin.Context) {
	return func(c *gin.Context) {
		authHelper(c, common.RoleRootUser)
	}
}

func WssAuth(c *gin.Context) {

}

func TokenAuth() func(c *gin.Context) {
	return func(c *gin.Context) {
		// 先检测是否为ws
		if c.Request.Header.Get("Sec-WebSocket-Protocol") != "" {
			// Sec-WebSocket-Protocol: realtime, openai-insecure-api-key.sk-xxx, openai-beta.realtime-v1
			// read sk from Sec-WebSocket-Protocol
			key := c.Request.Header.Get("Sec-WebSocket-Protocol")
			parts := strings.Split(key, ",")
			for _, part := range parts {
				part = strings.TrimSpace(part)
				if strings.HasPrefix(part, "openai-insecure-api-key") {
					key = strings.TrimPrefix(part, "openai-insecure-api-key.")
					break
				}
			}
			c.Request.Header.Set("Authorization", "Bearer "+key)
		}
		key := c.Request.Header.Get("Authorization")
		parts := make([]string, 0)
		key = strings.TrimPrefix(key, "Bearer ")
		if key == "" || key == "midjourney-proxy" {
			key = c.Request.Header.Get("mj-api-secret")
			key = strings.TrimPrefix(key, "Bearer ")
			key = strings.TrimPrefix(key, "sk-")
			parts = strings.Split(key, "-")
			key = parts[0]
		} else {
			key = strings.TrimPrefix(key, "sk-")
			parts = strings.Split(key, "-")
			key = parts[0]
		}
		token, err := model.ValidateUserToken(key)
		if token != nil {
			id := c.GetInt("id")
			if id == 0 {
				c.Set("id", token.UserId)
			}
		}
		if err != nil {
			abortWithOpenAiMessage(c, http.StatusUnauthorized, err.Error())
			return
		}
		userEnabled, err := model.IsUserEnabled(token.UserId, false)
		if err != nil {
			abortWithOpenAiMessage(c, http.StatusInternalServerError, err.Error())
			return
		}
		if !userEnabled {
			abortWithOpenAiMessage(c, http.StatusForbidden, "用户已被封禁")
			return
		}
		c.Set("id", token.UserId)
		c.Set("token_id", token.Id)
		c.Set("token_key", token.Key)
		c.Set("token_name", token.Name)
		c.Set("token_unlimited_quota", token.UnlimitedQuota)
		if !token.UnlimitedQuota {
			c.Set("token_quota", token.RemainQuota)
		}
		if token.ModelLimitsEnabled {
			c.Set("token_model_limit_enabled", true)
			c.Set("token_model_limit", token.GetModelLimitsMap())
		} else {
			c.Set("token_model_limit_enabled", false)
		}
		c.Set("allow_ips", token.GetIpLimitsMap())
		c.Set("token_group", token.Group)
		if len(parts) > 1 {
			if model.IsAdmin(token.UserId) {
				c.Set("specific_channel_id", parts[1])
			} else {
				abortWithOpenAiMessage(c, http.StatusForbidden, "普通用户不支持指定渠道")
				return
			}
		}
		c.Next()
	}
}
