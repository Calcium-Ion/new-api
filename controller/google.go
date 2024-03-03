package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"io"
	"net/http"
	"net/url"
	"one-api/common"
	"one-api/model"
	"strconv"
)

const (
	GoogleOAuthURL = "https://accounts.google.com/o/oauth2/auth"
	Scope          = "https://www.googleapis.com/auth/userinfo.email"
	GetTokenUrl    = "https://accounts.google.com/o/oauth2/token"
	GetUserUrl     = "https://www.googleapis.com/oauth2/v1/userinfo"
)

func GoogleOAuth(c *gin.Context) {
	oAuthUrl := fmt.Sprintf("%s?client_id=%s&redirect_uri=%s&scope=%s&response_type=code", GoogleOAuthURL, common.GoogleClientId, common.GoogleClientSecret, Scope)
	c.Redirect(302, oAuthUrl)
}

func GoogleOAuthCallback(c *gin.Context) {
	code := c.Query("code")

	if !common.GoogleOAuthEnabled {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "管理员未开启通过 Google 登录以及注册",
		})
		return
	}

	tokenResult, err := getTokenByCode(code)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	googleUser, err := getUserInfoByToken(tokenResult.AccessToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	user, err := model.GetUserByEmail(googleUser.Email)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if common.RegisterEnabled {
			user.Username = "google_" + strconv.Itoa(model.GetMaxUserId()+1)
			user.DisplayName = "Google User"

			user.Email = googleUser.Email
			user.Role = common.RoleCommonUser
			user.Status = common.UserStatusEnabled

			if err := user.Insert(0); err != nil {
				c.JSON(http.StatusOK, gin.H{
					"success": false,
					"message": err.Error(),
				})
				return
			}
		} else {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "管理员关闭了新用户注册",
			})
			return
		}
	}
	if user.Status != common.UserStatusEnabled {
		c.JSON(http.StatusOK, gin.H{
			"message": "用户已被封禁",
			"success": false,
		})
		return
	}
	setupLogin(user, c)
}

func getTokenByCode(code string) (*GoogleTokenResult, error) {
	redirect_url := fmt.Sprintf("%s/api/oauth/google/callback", common.ServerAddress)
	data := url.Values{}
	data.Set("client_id", common.GoogleClientId)
	data.Set("client_secret", common.GoogleClientSecret)
	data.Set("code", code)
	data.Set("grant_type", "authorization_code")
	data.Set("redirect_uri", redirect_url)
	response, err := http.PostForm(GetTokenUrl, data)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return nil, fmt.Errorf("failed to get token: %d", response.StatusCode)
	}
	getTokenResult, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	var tokenResult GoogleTokenResult
	err = json.Unmarshal(getTokenResult, &tokenResult)
	if err != nil {
		return nil, err
	}
	return &tokenResult, nil
}

func getUserInfoByToken(token string) (*GoogleUser, error) {
	req, err := http.NewRequest("GET", GetUserUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	client := http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return nil, fmt.Errorf("failed to get user info: %d", response.StatusCode)
	}
	userInfo, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	var user GoogleUser
	err = json.Unmarshal(userInfo, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

type GoogleTokenResult struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
	IdToken     string `json:"id_token"`
}

type GoogleUser struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Picture       string `json:"picture"`
}
