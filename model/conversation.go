package model

import (
	"time"

	"github.com/google/uuid"
)

type Conversation struct {
	ID        string    `gorm:"primaryKey;type:varchar(255);not null" json:"id"`
	UserID    int       `gorm:"index;not null" json:"user_id"`
	Title     string    `gorm:"type:varchar(255)" json:"title"`
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null" json:"updated_at"`
}

func GetConversationsByUserID(userID int) ([]*Conversation, error) {
	var conversations []*Conversation
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

	// 构建基础查询
	err := tx.Where("user_id = ?", userID).
		Order("created_time asc").
		Find(&conversations).Error

	// 提交事务
	if err = tx.Commit().Error; err != nil {
		return nil, err
	}

	return conversations, nil
}

// CreateConversation 创建新的会话
func CreateConversation(userID int) (string, error) {
	conversation := &Conversation{
		ID:     uuid.New().String(),
		UserID: userID,
		Title:  "新会话",
	}

	// 开始事务
	tx := DB.Begin()
	if tx.Error != nil {
		return "", tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 创建会话记录
	if err := tx.Create(conversation).Error; err != nil {
		tx.Rollback()
		return "", err
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return "", err
	}

	return conversation.ID, nil
}
