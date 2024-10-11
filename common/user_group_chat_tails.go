package common

import (
	"encoding/json"
)

var UserUsableGroupChatTails = map[string]string{
	"default": "--我是default的小尾巴",
}

func UserUsableGroupChatTails2JSONString() string {
	jsonBytes, err := json.Marshal(UserUsableGroupChatTails)
	if err != nil {
		SysError("error marshalling user groups: " + err.Error())
	}
	return string(jsonBytes)
}

func UpdateUserUsableGroupChatTailsByJSONString(jsonStr string) error {
	UserUsableGroupChatTails = make(map[string]string)
	return json.Unmarshal([]byte(jsonStr), &UserUsableGroupChatTails)
}

func GroupInUserUsableGroupChatTails(groupName string) bool {
	_, ok := UserUsableGroupChatTails[groupName]
	return ok
}
