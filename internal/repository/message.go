package repository

import (
	"ggb_server/internal/app/model"
	"gorm.io/gorm"
)

type MessageRepository interface {
	Generic[model.Message]
	GetBySessionID(db *gorm.DB, sessionID int64, page, pageSize int) ([]model.Message, error)
	GetLastMessage(db *gorm.DB, sessionID int64) (*model.Message, error)
	BatchCreate(db *gorm.DB, messages []model.Message) error
}

type MessageRepo struct {
	GenericImpl[model.Message]
}

func NewMessageRepository() MessageRepository {
	return &MessageRepo{}
}

func (r *MessageRepo) GetBySessionID(db *gorm.DB, sessionID int64, page, pageSize int) ([]model.Message, error) {
	var messages []model.Message
	err := db.Where("session_id = ? AND is_del = 0", sessionID).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Order("created_at ASC").
		Find(&messages).Error
	return messages, err
}

func (r *MessageRepo) GetLastMessage(db *gorm.DB, sessionID int64) (*model.Message, error) {
	var message model.Message
	err := db.Where("session_id = ? AND is_del = 0", sessionID).
		Order("created_at DESC").
		First(&message).Error
	return &message, err
}

func (r *MessageRepo) BatchCreate(db *gorm.DB, messages []model.Message) error {
	return db.CreateInBatches(messages, 100).Error
}
