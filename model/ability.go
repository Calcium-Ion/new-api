package model

import (
	"errors"
	"fmt"
	"github.com/samber/lo"
	"gorm.io/gorm"
	"one-api/common"
	"strings"
)

type Ability struct {
	Group     string `json:"group" gorm:"type:varchar(64);primaryKey;autoIncrement:false"`
	Model     string `json:"model" gorm:"type:varchar(64);primaryKey;autoIncrement:false"`
	ChannelId int    `json:"channel_id" gorm:"primaryKey;autoIncrement:false;index"`
	Enabled   bool   `json:"enabled"`
	Priority  *int64 `json:"priority" gorm:"bigint;default:0;index"`
	Weight    uint   `json:"weight" gorm:"default:0;index"`
}

func GetGroupModels(group string) []string {
	var models []string
	// Find distinct models
	groupCol := "`group`"
	if common.UsingPostgreSQL {
		groupCol = `"group"`
	}
	DB.Table("abilities").Where(groupCol+" = ? and enabled = ?", group, true).Distinct("model").Pluck("model", &models)
	return models
}

func getPriority(group string, model string, retry int) (int, error) {
	groupCol := "`group`"
	trueVal := "1"
	if common.UsingPostgreSQL {
		groupCol = `"group"`
		trueVal = "true"
	}

	var priorities []int
	err := DB.Model(&Ability{}).
		Select("DISTINCT(priority)").
		Where(groupCol+" = ? and model = ? and enabled = "+trueVal, group, model).
		Order("priority DESC").              // 按优先级降序排序
		Pluck("priority", &priorities).Error // Pluck用于将查询的结果直接扫描到一个切片中

	if err != nil {
		// 处理错误
		return 0, err
	}

	// 确定要使用的优先级
	var priorityToUse int
	if retry >= len(priorities) {
		// 如果重试次数大于优先级数，则使用最小的优先级
		priorityToUse = priorities[len(priorities)-1]
	} else {
		priorityToUse = priorities[retry]
	}
	return priorityToUse, nil
}

func getChannelQuery(group string, model string, retry int) *gorm.DB {
	groupCol := "`group`"
	trueVal := "1"
	if common.UsingPostgreSQL {
		groupCol = `"group"`
		trueVal = "true"
	}
	maxPrioritySubQuery := DB.Model(&Ability{}).Select("MAX(priority)").Where(groupCol+" = ? and model = ? and enabled = "+trueVal, group, model)
	channelQuery := DB.Where(groupCol+" = ? and model = ? and enabled = "+trueVal+" and priority = (?)", group, model, maxPrioritySubQuery)
	if retry != 0 {
		priority, err := getPriority(group, model, retry)
		if err != nil {
			common.SysError(fmt.Sprintf("Get priority failed: %s", err.Error()))
		} else {
			channelQuery = DB.Where(groupCol+" = ? and model = ? and enabled = "+trueVal+" and priority = ?", group, model, priority)
		}
	}

	return channelQuery
}

func GetRandomSatisfiedChannel(group string, model string, retry int) (*Channel, error) {
	var abilities []Ability

	var err error = nil
	channelQuery := getChannelQuery(group, model, retry)
	if common.UsingSQLite || common.UsingPostgreSQL {
		err = channelQuery.Order("weight DESC").Find(&abilities).Error
	} else {
		err = channelQuery.Order("weight DESC").Find(&abilities).Error
	}
	if err != nil {
		return nil, err
	}
	channel := Channel{}
	if len(abilities) > 0 {
		// Randomly choose one
		weightSum := uint(0)
		for _, ability_ := range abilities {
			weightSum += ability_.Weight + 10
		}
		// Randomly choose one
		weight := common.GetRandomInt(int(weightSum))
		for _, ability_ := range abilities {
			weight -= int(ability_.Weight) + 10
			//log.Printf("weight: %d, ability weight: %d", weight, *ability_.Weight)
			if weight <= 0 {
				channel.Id = ability_.ChannelId
				break
			}
		}
	} else {
		return nil, errors.New("channel not found")
	}
	err = DB.First(&channel, "id = ?", channel.Id).Error
	return &channel, err
}

func (channel *Channel) AddAbilities() error {
	models_ := strings.Split(channel.Models, ",")
	groups_ := strings.Split(channel.Group, ",")
	abilities := make([]Ability, 0, len(models_))
	for _, model := range models_ {
		for _, group := range groups_ {
			ability := Ability{
				Group:     group,
				Model:     model,
				ChannelId: channel.Id,
				Enabled:   channel.Status == common.ChannelStatusEnabled,
				Priority:  channel.Priority,
				Weight:    uint(channel.GetWeight()),
			}
			abilities = append(abilities, ability)
		}
	}
	if len(abilities) == 0 {
		return nil
	}
	for _, chunk := range lo.Chunk(abilities, 50) {
		err := DB.Create(&chunk).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func (channel *Channel) DeleteAbilities() error {
	return DB.Where("channel_id = ?", channel.Id).Delete(&Ability{}).Error
}

// UpdateAbilities updates abilities of this channel.
// Make sure the channel is completed before calling this function.
func (channel *Channel) UpdateAbilities() error {
	// A quick and dirty way to update abilities
	// First delete all abilities of this channel
	err := channel.DeleteAbilities()
	if err != nil {
		return err
	}
	// Then add new abilities
	err = channel.AddAbilities()
	if err != nil {
		return err
	}
	return nil
}

func UpdateAbilityStatus(channelId int, status bool) error {
	return DB.Model(&Ability{}).Where("channel_id = ?", channelId).Select("enabled").Update("enabled", status).Error
}

func FixAbility() (int, error) {
	var channelIds []int
	count := 0
	// Find all channel ids from channel table
	err := DB.Model(&Channel{}).Pluck("id", &channelIds).Error
	if err != nil {
		common.SysError(fmt.Sprintf("Get channel ids from channel table failed: %s", err.Error()))
		return 0, err
	}
	// Delete abilities of channels that are not in channel table
	err = DB.Where("channel_id NOT IN (?)", channelIds).Delete(&Ability{}).Error
	if err != nil {
		common.SysError(fmt.Sprintf("Delete abilities of channels that are not in channel table failed: %s", err.Error()))
		return 0, err
	}
	common.SysLog(fmt.Sprintf("Delete abilities of channels that are not in channel table successfully, ids: %v", channelIds))
	count += len(channelIds)

	// Use channelIds to find channel not in abilities table
	var abilityChannelIds []int
	err = DB.Model(&Ability{}).Pluck("channel_id", &abilityChannelIds).Error
	if err != nil {
		common.SysError(fmt.Sprintf("Get channel ids from abilities table failed: %s", err.Error()))
		return 0, err
	}
	var channels []Channel

	if len(abilityChannelIds) == 0 {
		err = DB.Find(&channels).Error
	} else {
		err = DB.Where("id NOT IN (?)", abilityChannelIds).Find(&channels).Error
	}
	if err != nil {
		return 0, err
	}
	for _, channel := range channels {
		err := channel.UpdateAbilities()
		if err != nil {
			common.SysError(fmt.Sprintf("Update abilities of channel %d failed: %s", channel.Id, err.Error()))
		} else {
			common.SysLog(fmt.Sprintf("Update abilities of channel %d successfully", channel.Id))
			count++
		}
	}
	InitChannelCache()
	return count, nil
}
