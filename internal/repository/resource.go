package repository

import (
	"ggb_server/internal/app/model"
	"gorm.io/gorm"
)

type ResourceRepository[T any] interface {
	Generic[T]
	GetByMessageID(db *gorm.DB, messageID int64) ([]model.Resource, error)
	GetByType(db *gorm.DB, messageID int64, resourceType int8) ([]model.Resource, error)
	BatchDelete(db *gorm.DB, ids []uint) error
}

type ResourceRepo[T any] struct {
	GenericImpl[T]
}

func NewResourceRepository[T any]() ResourceRepository[T] {
	return &ResourceRepo[T]{GenericImpl[T]{}}
}

func (r *ResourceRepo[T]) GetByMessageID(db *gorm.DB, messageID int64) ([]model.Resource, error) {
	var resources []model.Resource
	err := db.Where("message_id = ? AND is_del = 0", messageID).Find(&resources).Error
	return resources, err
}

func (r *ResourceRepo[T]) GetByType(db *gorm.DB, messageID int64, resourceType int8) ([]model.Resource, error) {
	var resources []model.Resource
	err := db.Where("message_id = ? AND type = ? AND is_del = 0", messageID, resourceType).Find(&resources).Error
	return resources, err
}

func (r *ResourceRepo[T]) BatchDelete(db *gorm.DB, ids []uint) error {
	return db.Model(&model.Resource{}).Where("id IN ?", ids).Update("is_del", 1).Error
}
