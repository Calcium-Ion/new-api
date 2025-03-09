package model

import (
	"fmt"
	"gorm.io/gorm"
	"one-api/common"
	"one-api/setting/operation_setting"
	"sync"
	"time"
	"github.com/xuri/excelize/v2"
)

// QuotaData 柱状图数据
type QuotaData struct {
	Id        int    `json:"id"`
	UserID    int    `json:"user_id" gorm:"index"`
	Username  string `json:"username" gorm:"index:idx_qdt_model_user_name,priority:2;size:64;default:''"`
	ModelName string `json:"model_name" gorm:"index:idx_qdt_model_user_name,priority:1;size:64;default:''"`
	CreatedAt int64  `json:"created_at" gorm:"bigint;index:idx_qdt_created_at,priority:2"`
	TokenUsed int    `json:"token_used" gorm:"default:0"`
	Count     int    `json:"count" gorm:"default:0"`
	Quota     int    `json:"quota" gorm:"default:0"`
}

type BillingData struct {
	ChannelId         int    `json:"chanel_id"`
	ChannelName       string `json:"channel_name"`
	Count             int    `json:"count"`
	ModelName         string `json:"model_name"`
	PromptTokens      int    `json:"prompt_tokens"`
	CompletionsTokens int    `json:"completions_tokens"`
}

type BillingJsonData struct {
	ChannelId          int     `json:"chanel_id"`
	CurrentDate        string  `json:"current_date"`
	ChannelName        string  `json:"channel_name"`
	Count              int     `json:"count"`
	ModelName          string  `json:"model_name"`
	PromptTokens       float32 `json:"prompt_tokens"`
	CompletionsTokens  float32 `json:"completions_tokens"`
	PromptPricing      float32 `json:"prompt_pricing"`
	CompletionsPricing float32 `json:"completions_pricing"`
	/**/ Cost          float32 `json:"cost"`
}

func UpdateQuotaData() {
	// recover
	defer func() {
		if r := recover(); r != nil {
			common.SysLog(fmt.Sprintf("UpdateQuotaData panic: %s", r))
		}
	}()
	for {
		if common.DataExportEnabled {
			common.SysLog("正在更新数据看板数据...")
			SaveQuotaDataCache()
		}
		time.Sleep(time.Duration(common.DataExportInterval) * time.Minute)
	}
}

var CacheQuotaData = make(map[string]*QuotaData)
var CacheQuotaDataLock = sync.Mutex{}

func logQuotaDataCache(userId int, username string, modelName string, quota int, createdAt int64, tokenUsed int) {
	key := fmt.Sprintf("%d-%s-%s-%d", userId, username, modelName, createdAt)
	quotaData, ok := CacheQuotaData[key]
	if ok {
		quotaData.Count += 1
		quotaData.Quota += quota
		quotaData.TokenUsed += tokenUsed
	} else {
		quotaData = &QuotaData{
			UserID:    userId,
			Username:  username,
			ModelName: modelName,
			CreatedAt: createdAt,
			Count:     1,
			Quota:     quota,
			TokenUsed: tokenUsed,
		}
	}
	CacheQuotaData[key] = quotaData
}

func LogQuotaData(userId int, username string, modelName string, quota int, createdAt int64, tokenUsed int) {
	// 只精确到小时
	createdAt = createdAt - (createdAt % 3600)

	CacheQuotaDataLock.Lock()
	defer CacheQuotaDataLock.Unlock()
	logQuotaDataCache(userId, username, modelName, quota, createdAt, tokenUsed)
}

func SaveQuotaDataCache() {
	CacheQuotaDataLock.Lock()
	defer CacheQuotaDataLock.Unlock()
	size := len(CacheQuotaData)
	// 如果缓存中有数据，就保存到数据库中
	// 1. 先查询数据库中是否有数据
	// 2. 如果有数据，就更新数据
	// 3. 如果没有数据，就插入数据
	for _, quotaData := range CacheQuotaData {
		quotaDataDB := &QuotaData{}
		DB.Table("quota_data").Where("user_id = ? and username = ? and model_name = ? and created_at = ?",
			quotaData.UserID, quotaData.Username, quotaData.ModelName, quotaData.CreatedAt).First(quotaDataDB)
		if quotaDataDB.Id > 0 {
			//quotaDataDB.Count += quotaData.Count
			//quotaDataDB.Quota += quotaData.Quota
			//DB.Table("quota_data").Save(quotaDataDB)
			increaseQuotaData(quotaData.UserID, quotaData.Username, quotaData.ModelName, quotaData.Count, quotaData.Quota, quotaData.CreatedAt, quotaData.TokenUsed)
		} else {
			DB.Table("quota_data").Create(quotaData)
		}
	}
	CacheQuotaData = make(map[string]*QuotaData)
	common.SysLog(fmt.Sprintf("保存数据看板数据成功，共保存%d条数据", size))
}

func increaseQuotaData(userId int, username string, modelName string, count int, quota int, createdAt int64, tokenUsed int) {
	err := DB.Table("quota_data").Where("user_id = ? and username = ? and model_name = ? and created_at = ?",
		userId, username, modelName, createdAt).Updates(map[string]interface{}{
		"count":      gorm.Expr("count + ?", count),
		"quota":      gorm.Expr("quota + ?", quota),
		"token_used": gorm.Expr("token_used + ?", tokenUsed),
	}).Error
	if err != nil {
		common.SysLog(fmt.Sprintf("increaseQuotaData error: %s", err))
	}
}

func GetQuotaDataByUsername(username string, startTime int64, endTime int64) (quotaData []*QuotaData, err error) {
	var quotaDatas []*QuotaData
	// 从quota_data表中查询数据
	err = DB.Table("quota_data").Where("username = ? and created_at >= ? and created_at <= ?", username, startTime, endTime).Find(&quotaDatas).Error
	return quotaDatas, err
}

func GetQuotaDataByUserId(userId int, startTime int64, endTime int64) (quotaData []*QuotaData, err error) {
	var quotaDatas []*QuotaData
	// 从quota_data表中查询数据
	err = DB.Table("quota_data").Where("user_id = ? and created_at >= ? and created_at <= ?", userId, startTime, endTime).Find(&quotaDatas).Error
	return quotaDatas, err
}

func GetAllQuotaDates(startTime int64, endTime int64, username string) (quotaData []*QuotaData, err error) {
	if username != "" {
		return GetQuotaDataByUsername(username, startTime, endTime)
	}
	var quotaDatas []*QuotaData
	// 从quota_data表中查询数据
	// only select model_name, sum(count) as count, sum(quota) as quota, model_name, created_at from quota_data group by model_name, created_at;
	//err = DB.Table("quota_data").Where("created_at >= ? and created_at <= ?", startTime, endTime).Find(&quotaDatas).Error
	err = DB.Table("quota_data").Select("model_name, sum(count) as count, sum(quota) as quota, sum(token_used) as token_used, created_at").Where("created_at >= ? and created_at <= ?", startTime, endTime).Group("model_name, created_at").Find(&quotaDatas).Error
	return quotaDatas, err
}

func GetBilling(startTime int64, endTime int64) (billingJsonData []*BillingJsonData, err error) {
	var billingData []*BillingData
	err = DB.Table("logs").
		Select("logs.channel_id,  channels.name as channel_name, logs.model_name, "+
			"SUM(logs.prompt_tokens) as prompt_tokens, "+
			"SUM(logs.completion_tokens) as completions_tokens, ", startTime).
		Joins("JOIN channels ON logs.channel_id = channels.id").
		Where("logs.created_at BETWEEN ? AND ?", startTime, endTime).
		Group("logs.channel_id, channels.name, logs.model_name").
		Order("logs.channel_id").
		Find(&billingData).Error

	for _, data := range billingData {
		modelPrice := operation_setting.GetDefaultModelRatioMap()[data.ModelName]

		billingJsonData = append(billingJsonData, &BillingJsonData{
			ChannelId:          data.ChannelId,
			ChannelName:        data.ChannelName,
			CurrentDate:        time.Unix(startTime, 0).Format("2006-01-02"),
			Count:              data.Count,
			ModelName:          data.ModelName,
			PromptTokens:       float32(data.PromptTokens),
			CompletionsTokens:  float32(data.CompletionsTokens),
			PromptPricing:      float32(modelPrice * 2),
			CompletionsPricing: float32(modelPrice * 2 * operation_setting.GetCompletionRatio(data.ModelName)),
			Cost:               (float32(data.PromptTokens)*float32(modelPrice*2) + float32(data.CompletionsTokens)*float32(modelPrice*2*operation_setting.GetCompletionRatio(data.ModelName))) / 100_0000,
		})
	}
	return billingJsonData, err
}

func GetBillingAndExportExcel(startTime int64, endTime int64) ([]byte, error) {
	billingData, err := GetBilling(startTime, endTime)
	if err != nil {
		return nil, err
	}

	// 创建新的Excel文件
	f := excelize.NewFile()
	defer f.Close()

	// 设置表头
	headers := []string{"渠道ID", "渠道名称", "日期", "调用次数", "模型名字", 
		"提示Tokens", "补全Tokens", "提示价格", "补全价格", "金额"}
	for i, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue("Sheet1", cell, header)
	}

	row := 2
	currentChannelID := -1
	var channelTotal float32 = 0

	// 写入数据
	for _, data := range billingData {
		// 如果是新的渠道ID，且不是第一条数据
		if currentChannelID != -1 && currentChannelID != data.ChannelId && channelTotal > 0 {
			// 写入渠道总计行
			f.SetCellValue("Sheet1", fmt.Sprintf("A%d", row), currentChannelID)
			f.SetCellValue("Sheet1", fmt.Sprintf("B%d", row), "-")
			f.SetCellValue("Sheet1", fmt.Sprintf("C%d", row), "-")
			f.SetCellValue("Sheet1", fmt.Sprintf("D%d", row), "-")
			f.SetCellValue("Sheet1", fmt.Sprintf("E%d", row), "-")
			f.SetCellValue("Sheet1", fmt.Sprintf("F%d", row), "-")
			f.SetCellValue("Sheet1", fmt.Sprintf("G%d", row), "-")
			f.SetCellValue("Sheet1", fmt.Sprintf("H%d", row), "-")
			f.SetCellValue("Sheet1", fmt.Sprintf("I%d", row), "-")
			f.SetCellValue("Sheet1", fmt.Sprintf("J%d", row), channelTotal)
			row++
			channelTotal = 0
		}

		// 写入详细数据
		f.SetCellValue("Sheet1", fmt.Sprintf("A%d", row), data.ChannelId)
		f.SetCellValue("Sheet1", fmt.Sprintf("B%d", row), data.ChannelName)
		f.SetCellValue("Sheet1", fmt.Sprintf("C%d", row), data.CurrentDate)
		f.SetCellValue("Sheet1", fmt.Sprintf("D%d", row), data.Count)
		f.SetCellValue("Sheet1", fmt.Sprintf("E%d", row), data.ModelName)
		f.SetCellValue("Sheet1", fmt.Sprintf("F%d", row), data.PromptTokens)
		f.SetCellValue("Sheet1", fmt.Sprintf("G%d", row), data.CompletionsTokens)
		f.SetCellValue("Sheet1", fmt.Sprintf("H%d", row), data.PromptPricing)
		f.SetCellValue("Sheet1", fmt.Sprintf("I%d", row), data.CompletionsPricing)
		f.SetCellValue("Sheet1", fmt.Sprintf("J%d", row), data.Cost)

		channelTotal += data.Cost
		currentChannelID = data.ChannelId
		row++
	}

	// 写入最后一个渠道的总计行
	if channelTotal > 0 {
		f.SetCellValue("Sheet1", fmt.Sprintf("A%d", row), currentChannelID)
		f.SetCellValue("Sheet1", fmt.Sprintf("B%d", row), "-")
		f.SetCellValue("Sheet1", fmt.Sprintf("C%d", row), "-")
		f.SetCellValue("Sheet1", fmt.Sprintf("D%d", row), "-")
		f.SetCellValue("Sheet1", fmt.Sprintf("E%d", row), "-")
		f.SetCellValue("Sheet1", fmt.Sprintf("F%d", row), "-")
		f.SetCellValue("Sheet1", fmt.Sprintf("G%d", row), "-")
		f.SetCellValue("Sheet1", fmt.Sprintf("H%d", row), "-")
		f.SetCellValue("Sheet1", fmt.Sprintf("I%d", row), "-")
		f.SetCellValue("Sheet1", fmt.Sprintf("J%d", row), channelTotal)
	}

	// 删除保存文件的代码，改为返回字节流
	buffer, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}
