package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"one-api/common"
	"one-api/model"
	// "strconv"
	"time"
)

type wechatLoginResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

func getWeChatIdByCode(code string) (string, error) {
	if code == "" {
		return "", errors.New("无效的参数")
	}
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/wechat/user?code=%s", common.WeChatServerAddress, code), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", common.WeChatServerToken)
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	httpResponse, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer httpResponse.Body.Close()
	var res wechatLoginResponse
	err = json.NewDecoder(httpResponse.Body).Decode(&res)
	if err != nil {
		return "", err
	}
	if !res.Success {
		return "", errors.New(res.Message)
	}
	if res.Data == "" {
		return "", errors.New("验证码错误或已过期")
	}
	return res.Data, nil
}

func WeChatAuth(c *gin.Context) {
    if !common.WeChatAuthEnabled {
        c.JSON(http.StatusOK, gin.H{
            "message": "管理员未开启通过微信登录以及注册",
            "success": false,
        })
        return
    }
    code := c.Query("code")
    wechatId, err := getWeChatIdByCode(code)
    if err != nil {
        c.JSON(http.StatusOK, gin.H{
            "message": err.Error(),
            "success": false,
        })
        return
    }
    user := model.User{
        WeChatId: wechatId,
    }
    if model.IsWeChatIdAlreadyTaken(wechatId) {
        err := user.FillUserByWeChatId()
        if err != nil {
            c.JSON(http.StatusOK, gin.H{
                "success": false,
                "message": err.Error(),
            })
            return
        }
    } else {
        c.JSON(http.StatusOK, gin.H{
            "success": false,
            "message": "管理员已关闭微信的新用户注册功能。现在，只有已经注册的用户可以使用微信进行登录和绑定。如果您想要注册新的微信账号，请使用您学校的邮箱（以.edu.cn结尾）进行注册。",
        })
        return
    }

    if user.Status != common.UserStatusEnabled {
        c.JSON(http.StatusOK, gin.H{
            "message": "用户已被封禁",
            "success": false,
        })
        return
    }
    setupLogin(&user, c)
}

func WeChatBind(c *gin.Context) {
	if !common.WeChatAuthEnabled {
		c.JSON(http.StatusOK, gin.H{
			"message": "管理员未开启通过微信登录以及注册",
			"success": false,
		})
		return
	}
	code := c.Query("code")
	wechatId, err := getWeChatIdByCode(code)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": err.Error(),
			"success": false,
		})
		return
	}
	if model.IsWeChatIdAlreadyTaken(wechatId) {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "该微信账号已被绑定",
		})
		return
	}
	id := c.GetInt("id")
	user := model.User{
		Id: id,
	}
	err = user.FillUserById()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	user.WeChatId = wechatId
	err = user.Update(false)
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
	})
	return
}
