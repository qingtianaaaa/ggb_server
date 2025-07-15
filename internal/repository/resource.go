package repository

import (
	"ggb_server/internal/app/model"
	"gorm.io/gorm"
)

type ResourceRepository interface {
	Generic[model.Resource]
	GetByMessageID(db *gorm.DB, messageID int64) ([]model.Resource, error)
	GetByType(db *gorm.DB, messageID int64, resourceType int8) ([]model.Resource, error)
	BatchDelete(db *gorm.DB, ids []uint) error
}

type ResourceRepo struct {
	GenericImpl[model.Resource]
}

func NewResourceRepository() ResourceRepository {
	return &ResourceRepo{}
}

func (r *ResourceRepo) GetByMessageID(db *gorm.DB, messageID int64) ([]model.Resource, error) {
	var resources []model.Resource
	err := db.Where("message_id = ? AND is_del = 0", messageID).Find(&resources).Error
	return resources, err
}

func (r *ResourceRepo) GetByType(db *gorm.DB, messageID int64, resourceType int8) ([]model.Resource, error) {
	var resources []model.Resource
	err := db.Where("message_id = ? AND type = ? AND is_del = 0", messageID, resourceType).Find(&resources).Error
	return resources, err
}

func (r *ResourceRepo) BatchDelete(db *gorm.DB, ids []uint) error {
	return db.Model(&model.Resource{}).Where("id IN ?", ids).Update("is_del", 1).Error
}
