package service

import (
	"one-api/constant"
)

func GetCallbackAddress() string {
	if constant.EpayCallbackAddress == "" {
		return constant.ServerAddress
	}
	return constant.EpayCallbackAddress
}
