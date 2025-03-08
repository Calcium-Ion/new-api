package model

type TopUp struct {
	Id         int     `json:"id"`
	UserId     int     `json:"user_id" gorm:"index"`
	Amount     int64   `json:"amount"`
	Money      float64 `json:"money"`
	TradeNo    string  `json:"trade_no"`
	CreateTime int64   `json:"create_time"`
	Status     string  `json:"status"`
}

func (topUp *TopUp) Insert() error {
	var err error
	err = DB.Create(topUp).Error
	return err
}

func (topUp *TopUp) Update() error {
	var err error
	err = DB.Save(topUp).Error
	return err
}

func GetTopUpById(id int) *TopUp {
	var topUp *TopUp
	var err error
	err = DB.Where("id = ?", id).First(&topUp).Error
	if err != nil {
		return nil
	}
	return topUp
}

func GetTopUpByTradeNo(tradeNo string) *TopUp {
	var topUp *TopUp
	var err error
	err = DB.Where("trade_no = ?", tradeNo).First(&topUp).Error
	if err != nil {
		return nil
	}
	return topUp
}
