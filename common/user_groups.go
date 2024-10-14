package common

import (
	"encoding/json"
)

var UserUsableGroups = map[string]string{
	"default": "默认分组",
	"vip":     "vip分组",
}

func UserUsableGroups2JSONString() string {
	jsonBytes, err := json.Marshal(UserUsableGroups)
	if err != nil {
		SysError("error marshalling user groups: " + err.Error())
	}
	return string(jsonBytes)
}

func UpdateUserUsableGroupsByJSONString(jsonStr string) error {
	UserUsableGroups = make(map[string]string)
	return json.Unmarshal([]byte(jsonStr), &UserUsableGroups)
}

func GetUserUsableGroups(userGroup string) map[string]string {
	if userGroup == "" {
		// 如果userGroup为空，返回UserUsableGroups
		return UserUsableGroups
	}
	// 如果userGroup不在UserUsableGroups中，返回UserUsableGroups + userGroup
	if _, ok := UserUsableGroups[userGroup]; !ok {
		appendUserUsableGroups := make(map[string]string)
		for k, v := range UserUsableGroups {
			appendUserUsableGroups[k] = v
		}
		appendUserUsableGroups[userGroup] = "用户分组"
		return appendUserUsableGroups
	}
	// 如果userGroup在UserUsableGroups中，返回UserUsableGroups
	return UserUsableGroups
}

func GroupInUserUsableGroups(groupName string) bool {
	_, ok := UserUsableGroups[groupName]
	return ok
}
