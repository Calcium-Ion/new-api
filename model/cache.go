package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"one-api/common"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	TokenCacheSeconds         = common.SyncFrequency
	UserId2GroupCacheSeconds  = common.SyncFrequency
	UserId2QuotaCacheSeconds  = common.SyncFrequency
	UserId2StatusCacheSeconds = common.SyncFrequency
)

// 仅用于定时同步缓存
var token2UserId = make(map[string]int)
var token2UserIdLock sync.RWMutex

func cacheSetToken(token *Token) error {
	jsonBytes, err := json.Marshal(token)
	if err != nil {
		return err
	}
	err = common.RedisSet(fmt.Sprintf("token:%s", token.Key), string(jsonBytes), time.Duration(TokenCacheSeconds)*time.Second)
	if err != nil {
		common.SysError(fmt.Sprintf("failed to set token %s to redis: %s", token.Key, err.Error()))
		return err
	}
	token2UserIdLock.Lock()
	defer token2UserIdLock.Unlock()
	token2UserId[token.Key] = token.UserId
	return nil
}

// CacheGetTokenByKey 从缓存中获取 token 并续期时间，如果缓存中不存在，则从数据库中获取
func CacheGetTokenByKey(key string) (*Token, error) {
	if !common.RedisEnabled {
		return GetTokenByKey(key)
	}
	var token *Token
	tokenObjectString, err := common.RedisGet(fmt.Sprintf("token:%s", key))
	if err != nil {
		// 如果缓存中不存在，则从数据库中获取
		token, err = GetTokenByKey(key)
		if err != nil {
			return nil, err
		}
		err = cacheSetToken(token)
		return token, nil
	}
	// 如果缓存中存在，则续期时间
	err = common.RedisExpire(fmt.Sprintf("token:%s", key), time.Duration(TokenCacheSeconds)*time.Second)
	err = json.Unmarshal([]byte(tokenObjectString), &token)
	return token, err
}

func SyncTokenCache(frequency int) {
	for {
		time.Sleep(time.Duration(frequency) * time.Second)
		common.SysLog("syncing tokens from database")
		token2UserIdLock.Lock()
		// 从token2UserId中获取所有的key
		var copyToken2UserId = make(map[string]int)
		for s, i := range token2UserId {
			copyToken2UserId[s] = i
		}
		token2UserId = make(map[string]int)
		token2UserIdLock.Unlock()

		for key := range copyToken2UserId {
			token, err := GetTokenByKey(key)
			if err != nil {
				// 如果数据库中不存在，则删除缓存
				common.SysError(fmt.Sprintf("failed to get token %s from database: %s", key, err.Error()))
				//delete redis
				err := common.RedisDel(fmt.Sprintf("token:%s", key))
				if err != nil {
					common.SysError(fmt.Sprintf("failed to delete token %s from redis: %s", key, err.Error()))
				}
			} else {
				// 如果数据库中存在，先检查redis
				_, err = common.RedisGet(fmt.Sprintf("token:%s", key))
				if err != nil {
					// 如果redis中不存在，则跳过
					continue
				}
				err = cacheSetToken(token)
				if err != nil {
					common.SysError(fmt.Sprintf("failed to update token %s to redis: %s", key, err.Error()))
				}
			}
		}
	}
}

func CacheGetUserGroup(id int) (group string, err error) {
	if !common.RedisEnabled {
		return GetUserGroup(id)
	}
	group, err = common.RedisGet(fmt.Sprintf("user_group:%d", id))
	if err != nil {
		group, err = GetUserGroup(id)
		if err != nil {
			return "", err
		}
		err = common.RedisSet(fmt.Sprintf("user_group:%d", id), group, time.Duration(UserId2GroupCacheSeconds)*time.Second)
		if err != nil {
			common.SysError("Redis set user group error: " + err.Error())
		}
	}
	return group, err
}

func CacheGetUsername(id int) (username string, err error) {
	if !common.RedisEnabled {
		return GetUsernameById(id)
	}
	username, err = common.RedisGet(fmt.Sprintf("user_name:%d", id))
	if err != nil {
		username, err = GetUsernameById(id)
		if err != nil {
			return "", err
		}
		err = common.RedisSet(fmt.Sprintf("user_name:%d", id), username, time.Duration(UserId2GroupCacheSeconds)*time.Second)
		if err != nil {
			common.SysError("Redis set user group error: " + err.Error())
		}
	}
	return username, err
}

func CacheGetUserQuota(id int) (quota int, err error) {
	if !common.RedisEnabled {
		return GetUserQuota(id)
	}
	quotaString, err := common.RedisGet(fmt.Sprintf("user_quota:%d", id))
	if err != nil {
		quota, err = GetUserQuota(id)
		if err != nil {
			return 0, err
		}
		err = common.RedisSet(fmt.Sprintf("user_quota:%d", id), fmt.Sprintf("%d", quota), time.Duration(UserId2QuotaCacheSeconds)*time.Second)
		if err != nil {
			common.SysError("Redis set user quota error: " + err.Error())
		}
		return quota, err
	}
	quota, err = strconv.Atoi(quotaString)
	return quota, err
}

func CacheUpdateUserQuota(id int) error {
	if !common.RedisEnabled {
		return nil
	}
	quota, err := GetUserQuota(id)
	if err != nil {
		return err
	}
	return cacheSetUserQuota(id, quota)
}

func cacheSetUserQuota(id int, quota int) error {
	err := common.RedisSet(fmt.Sprintf("user_quota:%d", id), fmt.Sprintf("%d", quota), time.Duration(UserId2QuotaCacheSeconds)*time.Second)
	return err
}

func CacheDecreaseUserQuota(id int, quota int) error {
	if !common.RedisEnabled {
		return nil
	}
	err := common.RedisDecrease(fmt.Sprintf("user_quota:%d", id), int64(quota))
	return err
}

func CacheIsUserEnabled(userId int) (bool, error) {
	if !common.RedisEnabled {
		return IsUserEnabled(userId)
	}
	enabled, err := common.RedisGet(fmt.Sprintf("user_enabled:%d", userId))
	if err == nil {
		return enabled == "1", nil
	}

	userEnabled, err := IsUserEnabled(userId)
	if err != nil {
		return false, err
	}
	enabled = "0"
	if userEnabled {
		enabled = "1"
	}
	err = common.RedisSet(fmt.Sprintf("user_enabled:%d", userId), enabled, time.Duration(UserId2StatusCacheSeconds)*time.Second)
	if err != nil {
		common.SysError("Redis set user enabled error: " + err.Error())
	}
	return userEnabled, err
}

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
    if (!common.MemoryCacheEnabled) {
        return
    }
    channelSyncLock.Lock()
    defer channelSyncLock.Unlock()
    if channel, ok := channelsIDM[id]; ok {
        channel.Status = status
    }
}
