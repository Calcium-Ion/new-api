package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
	epay "github.com/star-horizon/go-epay"
	"log"
	"net/url"
	"one-api/common"
	"one-api/model"
	"one-api/service"
	"strconv"
	"time"
)

type EpayRequest struct {
	Amount        int    `json:"amount"`
	PaymentMethod string `json:"payment_method"`
	TopUpCode     string `json:"top_up_code"`
}

type AmountRequest struct {
	Amount    int    `json:"amount"`
	TopUpCode string `json:"top_up_code"`
}

func GetEpayClient() *epay.Client {
	if common.PayAddress == "" || common.EpayId == "" || common.EpayKey == "" {
		return nil
	}
	withUrl, err := epay.NewClientWithUrl(&epay.Config{
		PartnerID: common.EpayId,
		Key:       common.EpayKey,
	}, common.PayAddress)
	if err != nil {
		return nil
	}
	return withUrl
}

func GetAmount(count float64, user model.User) float64 {
	// 别问为什么用float64，问就是这么点钱没必要
	topupGroupRatio := common.GetTopupGroupRatio(user.Group)
	if topupGroupRatio == 0 {
		topupGroupRatio = 1
	}
	amount := count * common.Price * topupGroupRatio
	return amount
}

func RequestEpay(c *gin.Context) {
	var req EpayRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(200, gin.H{"message": err.Error(), "data": 10})
		return
	}
	if req.Amount < common.MinTopUp {
		c.JSON(200, gin.H{"message": fmt.Sprintf("充值数量不能小于 %d", common.MinTopUp), "data": 10})
		return
	}

	id := c.GetInt("id")
	user, _ := model.GetUserById(id, false)
	payMoney := GetAmount(float64(req.Amount), *user)

	var payType epay.PurchaseType
	if req.PaymentMethod == "zfb" {
		payType = epay.Alipay
	}
	if req.PaymentMethod == "wx" {
		req.PaymentMethod = "wxpay"
		payType = epay.WechatPay
	}
	callBackAddress := service.GetCallbackAddress()
	returnUrl, _ := url.Parse(common.ServerAddress + "/log")
	notifyUrl, _ := url.Parse(callBackAddress + "/api/user/epay/notify")
	tradeNo := strconv.FormatInt(time.Now().Unix(), 10)
	client := GetEpayClient()
	if client == nil {
		c.JSON(200, gin.H{"message": "error", "data": "当前管理员未配置支付信息"})
		return
	}
	uri, params, err := client.Purchase(&epay.PurchaseArgs{
		Type:           payType,
		ServiceTradeNo: "A" + tradeNo,
		Name:           "B" + tradeNo,
		Money:          strconv.FormatFloat(payMoney, 'f', 2, 64),
		Device:         epay.PC,
		NotifyUrl:      notifyUrl,
		ReturnUrl:      returnUrl,
	})
	if err != nil {
		c.JSON(200, gin.H{"message": "error", "data": "拉起支付失败"})
		return
	}
	topUp := &model.TopUp{
		UserId:     id,
		Amount:     req.Amount,
		Money:      payMoney,
		TradeNo:    "A" + tradeNo,
		CreateTime: time.Now().Unix(),
		Status:     "pending",
	}
	err = topUp.Insert()
	if err != nil {
		c.JSON(200, gin.H{"message": "error", "data": "创建订单失败"})
		return
	}
	c.JSON(200, gin.H{"message": "success", "data": params, "url": uri})
}

func EpayNotify(c *gin.Context) {
	params := lo.Reduce(lo.Keys(c.Request.URL.Query()), func(r map[string]string, t string, i int) map[string]string {
		r[t] = c.Request.URL.Query().Get(t)
		return r
	}, map[string]string{})
	client := GetEpayClient()
	if client == nil {
		log.Println("易支付回调失败 未找到配置信息")
		_, err := c.Writer.Write([]byte("fail"))
		if err != nil {
			log.Println("易支付回调写入失败")
		}
	}
	verifyInfo, err := client.Verify(params)
	if err == nil && verifyInfo.VerifyStatus {
		_, err := c.Writer.Write([]byte("success"))
		if err != nil {
			log.Println("易支付回调写入失败")
		}
	} else {
		_, err := c.Writer.Write([]byte("fail"))
		if err != nil {
			log.Println("易支付回调写入失败")
		}
	}

	if verifyInfo.TradeStatus == epay.StatusTradeSuccess {
		log.Println(verifyInfo)
		topUp := model.GetTopUpByTradeNo(verifyInfo.ServiceTradeNo)
		if topUp.Status == "pending" {
			topUp.Status = "success"
			err := topUp.Update()
			if err != nil {
				log.Printf("易支付回调更新订单失败: %v", topUp)
				return
			}
			//user, _ := model.GetUserById(topUp.UserId, false)
			//user.Quota += topUp.Amount * 500000
			err = model.IncreaseUserQuota(topUp.UserId, topUp.Amount*500000)
			if err != nil {
				log.Printf("易支付回调更新用户失败: %v", topUp)
				return
			}
			log.Printf("易支付回调更新用户成功 %v", topUp)
			model.RecordLog(topUp.UserId, model.LogTypeTopup, fmt.Sprintf("使用在线充值成功，充值金额: %v，支付金额：%f", common.LogQuota(topUp.Amount*500000), topUp.Money))
		}
	} else {
		log.Printf("易支付异常回调: %v", verifyInfo)
	}
}

func RequestAmount(c *gin.Context) {
	var req AmountRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(200, gin.H{"message": "error", "data": "参数错误"})
		return
	}
	if req.Amount < common.MinTopUp {
		c.JSON(200, gin.H{"message": "error", "data": fmt.Sprintf("充值数量不能小于 %d", common.MinTopUp)})
		return
	}
	id := c.GetInt("id")
	user, _ := model.GetUserById(id, false)
	payMoney := GetAmount(float64(req.Amount), *user)
	c.JSON(200, gin.H{"message": "success", "data": strconv.FormatFloat(payMoney, 'f', 2, 64)})
}
