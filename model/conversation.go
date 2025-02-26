package model

import (
	"one-api/dto"
	"time"

	"github.com/google/uuid"
)

type Conversation struct {
	ID        string    `gorm:"primaryKey;type:varchar(255);not null" json:"id"`
	UserID    int       `gorm:"index;not null" json:"user_id"`
	Title     string    `gorm:"type:varchar(255)" json:"title"`
	Model     string    `gorm:"type:varchar(50)" json:"model"`
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
		Order("created_at desc").
		Find(&conversations).Error

	// 提交事务
	if err = tx.Commit().Error; err != nil {
		return nil, err
	}

	return conversations, nil
}

// CreateConversation 创建新的会话
func CreateConversation(userID int, req dto.CreateConversationRequest) (string, error) {
	var title string
	if req.Title == "" {
		title = "新对话"
	} else {
		title = req.Title
	}
	conversation := &Conversation{
		ID:     uuid.New().String(),
		UserID: userID,
		Title:  title,
		Model:  req.Model,
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

func UpdateConversationTitle(conversationID string, title string) {
	if title == "" {
		return
	}
	conversation := &Conversation{
		ID: conversationID,
	}
	err := DB.First(conversation, "id = ?", conversation.ID).Error
	if err != nil {
		return
	}
	// 已更新过无需再更新
	if conversation.Title != "新对话" {
		return
	}
	// 更新指定会话的标题
	DB.Model(conversation).Update("title", title)
}
