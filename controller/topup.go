package controller

import (
	"sync"

	"github.com/gin-gonic/gin"
)

type PayAdaptor interface {
	RequestAmount(c *gin.Context, req *PayRequest)
	RequestPay(c *gin.Context, req *PayRequest)
}

var (
	payNameAdaptorMap = map[string]PayAdaptor{}
)

type PayRequest struct {
	Amount        int64  `json:"amount"`
	PaymentMethod string `json:"payment_method"`
	TopUpCode     string `json:"top_up_code"`
}

func RequestPay(c *gin.Context) {
	var req PayRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(200, gin.H{"message": "error", "data": "参数错误"})
		return
	}

	//if !setting.PaymentEnabled {
	//	c.JSON(200, gin.H{"message": "error", "data": "管理员未开启在线支付"})
	//	return
	//}

	payAdaptor, ok := payNameAdaptorMap[req.PaymentMethod]
	if !ok {
		c.JSON(200, gin.H{"message": "error", "data": "不支持的支付方式"})
		return
	}
	payAdaptor.RequestPay(c, &req)
}

// tradeNo lock
var orderLocks sync.Map
var createLock sync.Mutex

// LockOrder 尝试对给定订单号加锁
func LockOrder(tradeNo string) {
	lock, ok := orderLocks.Load(tradeNo)
	if !ok {
		createLock.Lock()
		defer createLock.Unlock()
		lock, ok = orderLocks.Load(tradeNo)
		if !ok {
			lock = new(sync.Mutex)
			orderLocks.Store(tradeNo, lock)
		}
	}
	lock.(*sync.Mutex).Lock()
}

// UnlockOrder 释放给定订单号的锁
func UnlockOrder(tradeNo string) {
	lock, ok := orderLocks.Load(tradeNo)
	if ok {
		lock.(*sync.Mutex).Unlock()
	}
}

func RequestAmount(c *gin.Context) {
	var req PayRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(200, gin.H{"message": "error", "data": "参数错误"})
		return
	}

	//if !setting.PaymentEnabled {
	//	c.JSON(200, gin.H{"message": "error", "data": "管理员未开启在线支付"})
	//	return
	//}

	payAdaptor, ok := payNameAdaptorMap[req.PaymentMethod]
	if !ok {
		c.JSON(200, gin.H{"message": "error", "data": "不支持的支付方式"})
		return
	}

	payAdaptor.RequestAmount(c, &req)
}
