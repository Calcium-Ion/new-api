package model

import (
	"errors"
	"fmt"
	"math/rand"
	"one-api/common"
	"sort"
	"strings"
	"sync"
	"time"
)

//func CacheGetUserGroup(id int) (group string, err error) {
//	if !common.RedisEnabled {
//		return GetUserGroup(id)
//	}
//	group, err = common.RedisGet(fmt.Sprintf("user_group:%d", id))
//	if err != nil {
//		group, err = GetUserGroup(id)
//		if err != nil {
//			return "", err
//		}
//		err = common.RedisSet(fmt.Sprintf("user_group:%d", id), group, time.Duration(constant.UserId2GroupCacheSeconds)*time.Second)
//		if err != nil {
//			common.SysError("Redis set user group error: " + err.Error())
//		}
//	}
//	return group, err
//}
//
//func CacheGetUsername(id int) (username string, err error) {
//	if !common.RedisEnabled {
//		return GetUsernameById(id)
//	}
//	username, err = common.RedisGet(fmt.Sprintf("user_name:%d", id))
//	if err != nil {
//		username, err = GetUsernameById(id)
//		if err != nil {
//			return "", err
//		}
//		err = common.RedisSet(fmt.Sprintf("user_name:%d", id), username, time.Duration(constant.UserId2GroupCacheSeconds)*time.Second)
//		if err != nil {
//			common.SysError("Redis set user group error: " + err.Error())
//		}
//	}
//	return username, err
//}
//
//func CacheGetUserQuota(id int) (quota int, err error) {
//	if !common.RedisEnabled {
//		return GetUserQuota(id)
//	}
//	quotaString, err := common.RedisGet(fmt.Sprintf("user_quota:%d", id))
//	if err != nil {
//		quota, err = GetUserQuota(id)
//		if err != nil {
//			return 0, err
//		}
//		return quota, nil
//	}
//	quota, err = strconv.Atoi(quotaString)
//	return quota, nil
//}
//
//func CacheUpdateUserQuota(id int) error {
//	if !common.RedisEnabled {
//		return nil
//	}
//	quota, err := GetUserQuota(id)
//	if err != nil {
//		return err
//	}
//	return cacheSetUserQuota(id, quota)
//}
//
//func cacheSetUserQuota(id int, quota int) error {
//	err := common.RedisSet(fmt.Sprintf("user_quota:%d", id), fmt.Sprintf("%d", quota), time.Duration(constant.UserId2QuotaCacheSeconds)*time.Second)
//	return err
//}
//
//func CacheDecreaseUserQuota(id int, quota int) error {
//	if !common.RedisEnabled {
//		return nil
//	}
//	err := common.RedisDecrease(fmt.Sprintf("user_quota:%d", id), int64(quota))
//	return err
//}
//
//func CacheIsUserEnabled(userId int) (bool, error) {
//	if !common.RedisEnabled {
//		return IsUserEnabled(userId)
//	}
//	enabled, err := common.RedisGet(fmt.Sprintf("user_enabled:%d", userId))
//	if err == nil {
//		return enabled == "1", nil
//	}
//
//	userEnabled, err := IsUserEnabled(userId)
//	if err != nil {
//		return false, err
//	}
//	enabled = "0"
//	if userEnabled {
//		enabled = "1"
//	}
//	err = common.RedisSet(fmt.Sprintf("user_enabled:%d", userId), enabled, time.Duration(constant.UserId2StatusCacheSeconds)*time.Second)
//	if err != nil {
//		common.SysError("Redis set user enabled error: " + err.Error())
//	}
//	return userEnabled, err
//}

var group2model2channels map[string]map[string][]*Channel
var channelsIDM map[int]*Channel
var channelSyncLock sync.RWMutex

func InitChannelCache() {
	newChannelId2channel := make(map[int]*Channel)
	var channels []*Channel
	DB.Where("status = ?", common.ChannelStatusEnabled).Find(&channels)
	for _, channel := range channels {
		newChannelId2channel[channel.Id] = channel
	}
	var abilities []*Ability
	DB.Find(&abilities)
	groups := make(map[string]bool)
	for _, ability := range abilities {
		groups[ability.Group] = true
	}
	newGroup2model2channels := make(map[string]map[string][]*Channel)
	newChannelsIDM := make(map[int]*Channel)
	for group := range groups {
		newGroup2model2channels[group] = make(map[string][]*Channel)
	}
	for _, channel := range channels {
		newChannelsIDM[channel.Id] = channel
		groups := strings.Split(channel.Group, ",")
		for _, group := range groups {
			models := strings.Split(channel.Models, ",")
			for _, model := range models {
				if _, ok := newGroup2model2channels[group][model]; !ok {
					newGroup2model2channels[group][model] = make([]*Channel, 0)
				}
				newGroup2model2channels[group][model] = append(newGroup2model2channels[group][model], channel)
			}
		}
	}

	// sort by priority
	for group, model2channels := range newGroup2model2channels {
		for model, channels := range model2channels {
			sort.Slice(channels, func(i, j int) bool {
				return channels[i].GetPriority() > channels[j].GetPriority()
			})
			newGroup2model2channels[group][model] = channels
		}
	}

	channelSyncLock.Lock()
	group2model2channels = newGroup2model2channels
	channelsIDM = newChannelsIDM
	channelSyncLock.Unlock()
	common.SysLog("channels synced from database")
}

func SyncChannelCache(frequency int) {
	for {
		time.Sleep(time.Duration(frequency) * time.Second)
		common.SysLog("syncing channels from database")
		InitChannelCache()
	}
}

func CacheGetRandomSatisfiedChannel(group string, model string, retry int) (*Channel, error) {
	if strings.HasPrefix(model, "gpt-4-gizmo") {
		model = "gpt-4-gizmo-*"
	}
	if strings.HasPrefix(model, "gpt-4o-gizmo") {
		model = "gpt-4o-gizmo-*"
	}

	// if memory cache is disabled, get channel directly from database
	if !common.MemoryCacheEnabled {
		return GetRandomSatisfiedChannel(group, model, retry)
	}
	channelSyncLock.RLock()
	defer channelSyncLock.RUnlock()
	channels := group2model2channels[group][model]
	if len(channels) == 0 {
		return nil, errors.New("channel not found")
	}

	uniquePriorities := make(map[int]bool)
	for _, channel := range channels {
		uniquePriorities[int(channel.GetPriority())] = true
	}
	var sortedUniquePriorities []int
	for priority := range uniquePriorities {
		sortedUniquePriorities = append(sortedUniquePriorities, priority)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(sortedUniquePriorities)))

	if retry >= len(uniquePriorities) {
		retry = len(uniquePriorities) - 1
	}
	targetPriority := int64(sortedUniquePriorities[retry])

	// get the priority for the given retry number
	var targetChannels []*Channel
	for _, channel := range channels {
		if channel.GetPriority() == targetPriority {
			targetChannels = append(targetChannels, channel)
		}
	}

	// 平滑系数
	smoothingFactor := 10
	// Calculate the total weight of all channels up to endIdx
	totalWeight := 0
	for _, channel := range targetChannels {
		totalWeight += channel.GetWeight() + smoothingFactor
	}
	// Generate a random value in the range [0, totalWeight)
	randomWeight := rand.Intn(totalWeight)

	// Find a channel based on its weight
	for _, channel := range targetChannels {
		randomWeight -= channel.GetWeight() + smoothingFactor
		if randomWeight < 0 {
			return channel, nil
		}
	}
	// return null if no channel is not found
	return nil, errors.New("channel not found")
}

func CacheGetChannel(id int) (*Channel, error) {
	if !common.MemoryCacheEnabled {
		return GetChannelById(id, true)
	}
	channelSyncLock.RLock()
	defer channelSyncLock.RUnlock()

	c, ok := channelsIDM[id]
	if !ok {
		return nil, errors.New(fmt.Sprintf("当前渠道# %d，已不存在", id))
	}
	return c, nil
}

func CacheUpdateChannelStatus(id int, status int) {
	if !common.MemoryCacheEnabled {
		return
	}
	channelSyncLock.Lock()
	defer channelSyncLock.Unlock()
	if channel, ok := channelsIDM[id]; ok {
		channel.Status = status
	}
}
