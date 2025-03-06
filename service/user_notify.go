package service

import (
	"fmt"
	"one-api/common"
	"one-api/constant"
	"one-api/dto"
	"one-api/model"
	"strings"
)

func NotifyRootUser(t string, subject string, content string) {
	user := model.GetRootUser().ToBaseUser()
	err := NotifyUser(user.Id, user.Email, user.GetSetting(), dto.NewNotify(t, subject, content, nil))
	if err != nil {
		common.SysError(fmt.Sprintf("failed to notify root user: %s", err.Error()))
	}
}

func NotifyUser(userId int, userEmail string, userSetting map[string]interface{}, data dto.Notify) error {
	notifyType, ok := userSetting[constant.UserSettingNotifyType]
	if !ok {
		notifyType = constant.NotifyTypeEmail
	}

	// Check notification limit
	canSend, err := CheckNotificationLimit(userId, data.Type)
	if err != nil {
		common.SysError(fmt.Sprintf("failed to check notification limit: %s", err.Error()))
		return err
	}
	if !canSend {
		return fmt.Errorf("notification limit exceeded for user %d with type %s", userId, notifyType)
	}

	switch notifyType {
	case constant.NotifyTypeEmail:
		// check setting email
		if settingEmail, ok := userSetting[constant.UserSettingNotificationEmail]; ok {
			userEmail = settingEmail.(string)
		}
		if userEmail == "" {
			common.SysLog(fmt.Sprintf("user %d has no email, skip sending email", userId))
			return nil
		}
		return sendEmailNotify(userEmail, data)
	case constant.NotifyTypeWebhook:
		webhookURL, ok := userSetting[constant.UserSettingWebhookUrl]
		if !ok {
			common.SysError(fmt.Sprintf("user %d has no webhook url, skip sending webhook", userId))
			return nil
		}
		webhookURLStr, ok := webhookURL.(string)
		if !ok {
			common.SysError(fmt.Sprintf("user %d webhook url is not string type", userId))
			return nil
		}

		// 获取 webhook secret
		var webhookSecret string
		if secret, ok := userSetting[constant.UserSettingWebhookSecret]; ok {
			webhookSecret, _ = secret.(string)
		}

		return SendWebhookNotify(webhookURLStr, webhookSecret, data)
	}
	return nil
}

func sendEmailNotify(userEmail string, data dto.Notify) error {
	// make email content
	content := data.Content
	// 处理占位符
	for _, value := range data.Values {
		content = strings.Replace(content, dto.ContentValueParam, fmt.Sprintf("%v", value), 1)
	}
	return common.SendEmail(data.Title, userEmail, content)
}
