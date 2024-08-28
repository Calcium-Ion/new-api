package vertex

import "one-api/common"

func GetModelRegion(other string, localModelName string) string {
	// if other is json string
	if common.IsJsonStr(other) {
		m := common.StrToMap(other)
		if m[localModelName] != nil {
			return m[localModelName].(string)
		} else {
			return m["default"].(string)
		}
	}
	return other
}
