package constant

var (
	UserSettingNotifyType            = "notify_type"                    // QuotaWarningType 额度预警类型
	UserSettingQuotaWarningThreshold = "quota_warning_threshold"        // QuotaWarningThreshold 额度预警阈值
	UserSettingWebhookUrl            = "webhook_url"                    // WebhookUrl webhook地址
	UserSettingWebhookSecret         = "webhook_secret"                 // WebhookSecret webhook密钥
	UserSettingNotificationEmail     = "notification_email"             // NotificationEmail 通知邮箱地址
	UserAcceptUnsetRatioModel        = "accept_unset_model_ratio_model" // AcceptUnsetRatioModel 是否接受未设置价格的模型
)

var (
	NotifyTypeEmail   = "email"   // Email 邮件
	NotifyTypeWebhook = "webhook" // Webhook
)
