package service

import (
	"fmt"
	"one-api/common"
	"one-api/constant"
	"one-api/dto"
	"one-api/model"
	"strings"
)

func notifyRootUser(subject string, content string) {
	if common.RootUserEmail == "" {
		common.RootUserEmail = model.GetRootUserEmail()
	}
	err := common.SendEmail(subject, common.RootUserEmail, content)
	if err != nil {
		common.SysError(fmt.Sprintf("failed to send email: %s", err.Error()))
	}
}

func NotifyUser(user *model.UserCache, data dto.Notify) error {
	userSetting := user.GetSetting()
	notifyType, ok := userSetting[constant.UserSettingNotifyType]
	if !ok {
		notifyType = constant.NotifyTypeEmail
	}
	switch notifyType {
	case constant.NotifyTypeEmail:
		userEmail := user.Email
		// check setting email
		if settingEmail, ok := userSetting[constant.UserSettingNotificationEmail]; ok {
			userEmail = settingEmail.(string)
		}
		if userEmail == "" {
			common.SysLog(fmt.Sprintf("user %d has no email, skip sending email", user.Id))
			return nil
		}
		return sendEmailNotify(userEmail, data)
	case constant.NotifyTypeWebhook:
		webhookURL, ok := userSetting[constant.UserSettingWebhookUrl]
		if !ok {
			common.SysError(fmt.Sprintf("user %d has no webhook url, skip sending webhook", user.Id))
			return nil
		}
		// TODO: 实现webhook通知
		_ = webhookURL // 临时处理未使用警告，等待webhook实现
	}
	return nil // 添加缺失的return
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
