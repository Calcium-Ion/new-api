package common

import (
	"encoding/json"
)

var TopupGroupRatio = map[string]float64{
	"default": 1,
	"vip":     1,
	"svip":    1,
}

func TopupGroupRatio2JSONString() string {
	jsonBytes, err := json.Marshal(TopupGroupRatio)
	if err != nil {
		SysError("error marshalling model ratio: " + err.Error())
	}
	return string(jsonBytes)
}

func UpdateTopupGroupRatioByJSONString(jsonStr string) error {
	TopupGroupRatio = make(map[string]float64)
	return json.Unmarshal([]byte(jsonStr), &TopupGroupRatio)
}

func GetTopupGroupRatio(name string) float64 {
	ratio, ok := TopupGroupRatio[name]
	if !ok {
		SysError("topup group ratio not found: " + name)
		return 1
	}
	return ratio
}
