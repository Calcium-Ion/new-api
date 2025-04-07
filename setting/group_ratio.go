package setting

import (
	"encoding/json"
	"errors"
	"one-api/common"
	"sync"
)

var groupRatio = map[string]float64{
	"default": 1,
	"vip":     1,
	"svip":    1,
}
var groupRatioMutex sync.RWMutex

func GetGroupRatioCopy() map[string]float64 {
	groupRatioMutex.RLock()
	defer groupRatioMutex.RUnlock()
	
	groupRatioCopy := make(map[string]float64)
	for k, v := range groupRatio {
		groupRatioCopy[k] = v
	}
	return groupRatioCopy
}

func ContainsGroupRatio(name string) bool {
	groupRatioMutex.RLock()
	defer groupRatioMutex.RUnlock()
	
	_, ok := groupRatio[name]
	return ok
}

func GroupRatio2JSONString() string {
	groupRatioMutex.RLock()
	defer groupRatioMutex.RUnlock()
	
	jsonBytes, err := json.Marshal(groupRatio)
	if err != nil {
		common.SysError("error marshalling model ratio: " + err.Error())
	}
	return string(jsonBytes)
}

func UpdateGroupRatioByJSONString(jsonStr string) error {
	groupRatioMutex.Lock()
	defer groupRatioMutex.Unlock()
	
	groupRatio = make(map[string]float64)
	return json.Unmarshal([]byte(jsonStr), &groupRatio)
}

func GetGroupRatio(name string) float64 {
	groupRatioMutex.RLock()
	defer groupRatioMutex.RUnlock()
	
	ratio, ok := groupRatio[name]
	if !ok {
		common.SysError("group ratio not found: " + name)
		return 1
	}
	return ratio
}

func CheckGroupRatio(jsonStr string) error {
	checkGroupRatio := make(map[string]float64)
	err := json.Unmarshal([]byte(jsonStr), &checkGroupRatio)
	if err != nil {
		return err
	}
	for name, ratio := range checkGroupRatio {
		if ratio < 0 {
			return errors.New("group ratio must be not less than 0: " + name)
		}
	}
	return nil
}
