package operation_setting

import (
	"encoding/json"
	"one-api/common"
	"sync"
)

var defaultCacheRatio = map[string]float64{
	"gpt-4":                        0.5,
	"o1":                           0.5,
	"o1-2024-12-17":                0.5,
	"o1-preview-2024-09-12":        0.5,
	"o1-preview":                   0.5,
	"o1-mini-2024-09-12":           0.5,
	"o1-mini":                      0.5,
	"gpt-4o-2024-11-20":            0.5,
	"gpt-4o-2024-08-06":            0.5,
	"gpt-4o":                       0.5,
	"gpt-4o-mini-2024-07-18":       0.5,
	"gpt-4o-mini":                  0.5,
	"gpt-4o-realtime-preview":      0.5,
	"gpt-4o-mini-realtime-preview": 0.5,
	"deepseek-chat":                0.1,
	"deepseek-reasoner":            0.1,
	"deepseek-coder":               0.1,
}

var defaultCreateCacheRatio = map[string]float64{}

var cacheRatioMap map[string]float64
var cacheRatioMapMutex sync.RWMutex

// GetCacheRatioMap returns the cache ratio map
func GetCacheRatioMap() map[string]float64 {
	cacheRatioMapMutex.Lock()
	defer cacheRatioMapMutex.Unlock()
	if cacheRatioMap == nil {
		cacheRatioMap = defaultCacheRatio
	}
	return cacheRatioMap
}

// CacheRatio2JSONString converts the cache ratio map to a JSON string
func CacheRatio2JSONString() string {
	GetCacheRatioMap()
	jsonBytes, err := json.Marshal(cacheRatioMap)
	if err != nil {
		common.SysError("error marshalling cache ratio: " + err.Error())
	}
	return string(jsonBytes)
}

// UpdateCacheRatioByJSONString updates the cache ratio map from a JSON string
func UpdateCacheRatioByJSONString(jsonStr string) error {
	cacheRatioMapMutex.Lock()
	defer cacheRatioMapMutex.Unlock()
	cacheRatioMap = make(map[string]float64)
	return json.Unmarshal([]byte(jsonStr), &cacheRatioMap)
}

// GetCacheRatio returns the cache ratio for a model
func GetCacheRatio(name string) (float64, bool) {
	GetCacheRatioMap()
	ratio, ok := cacheRatioMap[name]
	if !ok {
		return 1, false // Default to 0.5 if not found
	}
	return ratio, true
}

// DefaultCacheRatio2JSONString converts the default cache ratio map to a JSON string
func DefaultCacheRatio2JSONString() string {
	jsonBytes, err := json.Marshal(defaultCacheRatio)
	if err != nil {
		common.SysError("error marshalling default cache ratio: " + err.Error())
	}
	return string(jsonBytes)
}

// GetDefaultCacheRatioMap returns the default cache ratio map
func GetDefaultCacheRatioMap() map[string]float64 {
	return defaultCacheRatio
}
