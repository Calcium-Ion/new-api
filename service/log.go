package service

import (
	"github.com/gin-gonic/gin"
	relaycommon "one-api/relay/common"
)

func GenerateTextOtherInfo(ctx *gin.Context, relayInfo *relaycommon.RelayInfo, modelRatio, groupRatio, completionRatio, modelPrice float64) map[string]interface{} {
	other := make(map[string]interface{})
	other["model_ratio"] = modelRatio
	other["group_ratio"] = groupRatio
	other["completion_ratio"] = completionRatio
	other["model_price"] = modelPrice
	other["frt"] = float64(relayInfo.FirstResponseTime.UnixMilli() - relayInfo.StartTime.UnixMilli())
	adminInfo := make(map[string]interface{})
	adminInfo["use_channel"] = ctx.GetStringSlice("use_channel")
	other["admin_info"] = adminInfo
	return other
}
