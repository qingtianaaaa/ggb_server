package repository

import (
	"ggb_server/internal/app/model"
	"gorm.io/gorm"
)

type SessionRepository[T any] interface {
	Generic[T]
	GetByUserID(db *gorm.DB, userID string, page, pageSize int) ([]model.Session, error)
	CountByUserID(db *gorm.DB, userID string) (int64, error)
	SoftDelete(db *gorm.DB, id uint) error
}

type SessionRepo[T any] struct {
	GenericImpl[T]
}

func NewSessionRepository[T any]() SessionRepository[T] {
	return &SessionRepo[T]{
		GenericImpl[T]{},
	}
}

func (r *SessionRepo[T]) GetByUserID(db *gorm.DB, userID string, page, pageSize int) ([]model.Session, error) {
	var sessions []model.Session
	err := db.Where("user_id = ? AND is_del = 0", userID).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&sessions).Error
	return sessions, err
}

func (r *SessionRepo[T]) CountByUserID(db *gorm.DB, userID string) (int64, error) {
	var count int64
	err := db.Model(&model.Session{}).Where("user_id = ? AND is_del = 0", userID).Count(&count).Error
	return count, err
}

func (r *SessionRepo[T]) SoftDelete(db *gorm.DB, id uint) error {
	return db.Model(&model.Session{}).Where("id = ?", id).Update("is_del", 1).Error
}
