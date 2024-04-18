package service

import (
	"one-api/common"
	"one-api/constant"
)

func GetCallbackAddress() string {
	if constant.CustomCallbackAddress == "" {
		return common.ServerAddress
	}
	return constant.CustomCallbackAddress
}
