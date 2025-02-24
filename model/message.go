package model

import "time"

type Message struct {
	ID             uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	ConversationID string    `gorm:"index;type:varchar(255);not null" json:"conversation_id"`
	Role           string    `gorm:"type:varchar(20);not null" json:"role"`
	Content        string    `gorm:"type:text;not null" json:"content"`
	CreatedTime    time.Time `gorm:"not null;index" json:"created_time"`
}

func GetMessagesByConversationID(conversationID string) ([]*Message, error) {
	var messages []*Message
	// 开始事务
	tx := DB.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 构建基础查询，添加 conversationID 条件并按时间戳排序
	err := tx.Where("conversation_id = ?", conversationID).
		Order("created_time asc").
		Find(&messages).Error

	// 提交事务
	if err = tx.Commit().Error; err != nil {
		return nil, err
	}

	return messages, nil
}

// CreateMessage 创建新的消息
func CreateMessage(conversationID string, role string, content string) (*Message, error) {
	message := &Message{
		ConversationID: conversationID,
		Role:           role,
		Content:        content,
		CreatedTime:    time.Now(),
	}

	// 开始事务
	tx := DB.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 创建消息记录
	if err := tx.Create(message).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return message, nil
}
