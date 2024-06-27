package common

import (
	"encoding/json"
)

var GroupRatio = map[string]float64{
	"default": 1,
	"vip":     1,
	"svip":    1,
}

func GroupRatio2JSONString() string {
	jsonBytes, err := json.Marshal(GroupRatio)
	if err != nil {
		SysError("error marshalling model ratio: " + err.Error())
	}
	return string(jsonBytes)
}

func UpdateGroupRatioByJSONString(jsonStr string) error {
	GroupRatio = make(map[string]float64)
	return json.Unmarshal([]byte(jsonStr), &GroupRatio)
}

func GetGroupRatio(name string) float64 {
	ratio, ok := GroupRatio[name]
	if !ok {
		SysError("group ratio not found: " + name)
		return 1
	}
	return ratio
}
