package service

import (
	"one-api/setting"
)

func GetCallbackAddress() string {
	if setting.CustomCallbackAddress == "" {
		return setting.ServerAddress
	}
	return setting.CustomCallbackAddress
}
