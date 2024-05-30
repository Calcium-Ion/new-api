package service

import (
	"one-api/constant"
)

func GetCallbackAddress() string {
	if constant.CustomCallbackAddress == "" {
		return constant.ServerAddress
	}
	return constant.CustomCallbackAddress
}
