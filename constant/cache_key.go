package constant

import "one-api/common"

var (
	TokenCacheSeconds         = common.SyncFrequency
	UserId2GroupCacheSeconds  = common.SyncFrequency
	UserId2QuotaCacheSeconds  = common.SyncFrequency
	UserId2StatusCacheSeconds = common.SyncFrequency
)

// Cache keys
const (
	UserGroupKeyFmt    = "user_group:%d"
	UserQuotaKeyFmt    = "user_quota:%d"
	UserEnabledKeyFmt  = "user_enabled:%d"
	UserUsernameKeyFmt = "user_name:%d"
)

const (
	TokenFiledRemainQuota = "RemainQuota"
	TokenFieldGroup       = "Group"
)
