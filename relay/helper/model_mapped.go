package helper

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"one-api/relay/common"
)

func ModelMappedHelper(c *gin.Context, info *common.RelayInfo) error {
	// map model name
	modelMapping := c.GetString("model_mapping")
	if modelMapping != "" && modelMapping != "{}" {
		modelMap := make(map[string]string)
		err := json.Unmarshal([]byte(modelMapping), &modelMap)
		if err != nil {
			return fmt.Errorf("unmarshal_model_mapping_failed")
		}
		if modelMap[info.OriginModelName] != "" {
			info.UpstreamModelName = modelMap[info.OriginModelName]
			info.IsModelMapped = true
		}
	}
	return nil
}
