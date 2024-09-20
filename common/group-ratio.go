package common

import (
	"encoding/json"
	"errors"
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
	tempGroupRatio := make(map[string]float64)
	err := json.Unmarshal([]byte(jsonStr), &tempGroupRatio)
	err = checkGroupRatio(tempGroupRatio)
	if err != nil {
		return err
	}
	GroupRatio = tempGroupRatio
	return err
}

func GetGroupRatio(name string) float64 {
	ratio, ok := GroupRatio[name]
	if !ok {
		SysError("group ratio not found: " + name)
		return 1
	}
	return ratio
}

func checkGroupRatio(checkGroupRatio map[string]float64) error {
	for name, ratio := range checkGroupRatio {
		if ratio < 0 {
			return errors.New("group ratio must be greater than 0: " + name)
		}
	}
	return nil
}
