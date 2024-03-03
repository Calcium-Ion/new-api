package service

import "one-api/common"

func GetCallbackAddress() string {
	if common.CustomCallbackAddress == "" {
		return common.ServerAddress
	}
	return common.CustomCallbackAddress
}
