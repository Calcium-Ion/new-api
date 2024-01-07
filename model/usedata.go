package model

import (
	"fmt"
	"one-api/common"
	"time"
)

// QuotaData 柱状图数据
type QuotaData struct {
	Id        int    `json:"id"`
	UserID    int    `json:"user_id" gorm:"index"`
	Username  string `json:"username" gorm:"index:index_quota_data_model_user_name,priority:2;default:''"`
	ModelName string `json:"model_name" gorm:"index;index:index_quota_data_model_user_name,priority:1;default:''"`
	CreatedAt int64  `json:"created_at" gorm:"bigint;index:index_quota_data_created_at,priority:2"`
	Count     int    `json:"count" gorm:"default:0"`
	Quota     int    `json:"quota" gorm:"default:0"`
}

func UpdateQuotaData() {
	for {
		if common.DataExportEnabled {
			common.SysLog("正在更新数据看板数据...")
			SaveQuotaDataCache()
		}
		time.Sleep(time.Duration(common.DataExportInterval) * time.Minute)
	}
}

var CacheQuotaData = make(map[string]*QuotaData)

func LogQuotaDataCache(userId int, username string, modelName string, quota int, createdAt int64) {
	// 只精确到小时
	createdAt = createdAt - (createdAt % 3600)
	key := fmt.Sprintf("%d-%s-%s-%d", userId, username, modelName, createdAt)
	quotaData, ok := CacheQuotaData[key]
	if ok {
		quotaData.Count += 1
		quotaData.Quota += quota
	} else {
		quotaData = &QuotaData{
			UserID:    userId,
			Username:  username,
			ModelName: modelName,
			CreatedAt: createdAt,
			Count:     1,
			Quota:     quota,
		}
	}
	CacheQuotaData[key] = quotaData
}

func LogQuotaData(userId int, username string, modelName string, quota int, createdAt int64) {
	LogQuotaDataCache(userId, username, modelName, quota, createdAt)
}

func SaveQuotaDataCache() {
	// 如果缓存中有数据，就保存到数据库中
	// 1. 先查询数据库中是否有数据
	// 2. 如果有数据，就更新数据
	// 3. 如果没有数据，就插入数据
	for _, quotaData := range CacheQuotaData {
		quotaDataDB := &QuotaData{}
		DB.Table("quota_data").Where("user_id = ? and username = ? and model_name = ? and created_at = ?",
			quotaData.UserID, quotaData.Username, quotaData.ModelName, quotaData.CreatedAt).First(quotaDataDB)
		if quotaDataDB.Id > 0 {
			quotaDataDB.Count += quotaData.Count
			quotaDataDB.Quota += quotaData.Quota
			DB.Table("quota_data").Save(quotaDataDB)
		} else {
			DB.Table("quota_data").Create(quotaData)
		}
	}
	CacheQuotaData = make(map[string]*QuotaData)
}

func GetQuotaDataByUsername(username string) (quotaData []*QuotaData, err error) {
	var quotaDatas []*QuotaData
	// 从quota_data表中查询数据
	err = DB.Table("quota_data").Where("username = ?", username).Find(&quotaDatas).Error
	return quotaDatas, err
}

func GetAllQuotaDates() (quotaData []*QuotaData, err error) {
	var quotaDatas []*QuotaData
	// 从quota_data表中查询数据
	err = DB.Table("quota_data").Find(&quotaDatas).Error
	return quotaDatas, err
}
