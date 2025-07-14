package model

import "gorm.io/gorm"

type Resource struct {
	Model     Model
	SessionID int64   `gorm:"not null;comment:关联会话ID" json:"sessionId"`
	MessageID int64   `gorm:"not null;comment:关联消息ID" json:"messageId"`
	Type      int8    `gorm:"default:0;comment:资源类型：1-图片 2-视频 3-HTML文件 4-其他" json:"type"`
	URL       string  `gorm:"type:text;comment:资源存储路径" json:"url"`
	Data      JSON    `gorm:"type:json;comment:扩展数据" json:"data"`
	IsDel     int8    `gorm:"default:0" json:"isDel"`
	Session   Session `gorm:"foreignKey:SessionID"`
	Message   Message `gorm:"foreignKey:MessageID"`
}

func (r *Resource) TableName() string {
	return "resource"
}

type ResourceData struct {
	Size     int64  `json:"size"`     // 文件大小
	Width    int    `json:"width"`    // 图片宽度
	Height   int    `json:"height"`   // 图片高度
	Duration int    `json:"duration"` // 视频时长(秒)
	Format   string `json:"format"`   // 文件格式
}

func (r *Resource) GetData() (*ResourceData, error) {
	var data ResourceData
	err := r.Data.ToStruct(&data)
	return &data, err
}

func (r *Resource) SetData(data *ResourceData) error {
	return r.Data.FromStruct(data)
}

type ResourceRepository interface {
	Generic[Resource]
	GetByMessageID(db *gorm.DB, messageID int64) ([]Resource, error)
	GetByType(db *gorm.DB, messageID int64, resourceType int8) ([]Resource, error)
	BatchDelete(db *gorm.DB, ids []uint) error
}

type ResourceRepo struct {
	GenericImpl[Resource]
}

func NewResourceRepository() ResourceRepository {
	return &ResourceRepo{}
}

func (r *ResourceRepo) GetByMessageID(db *gorm.DB, messageID int64) ([]Resource, error) {
	var resources []Resource
	err := db.Where("message_id = ? AND is_del = 0", messageID).Find(&resources).Error
	return resources, err
}

func (r *ResourceRepo) GetByType(db *gorm.DB, messageID int64, resourceType int8) ([]Resource, error) {
	var resources []Resource
	err := db.Where("message_id = ? AND type = ? AND is_del = 0", messageID, resourceType).Find(&resources).Error
	return resources, err
}

func (r *ResourceRepo) BatchDelete(db *gorm.DB, ids []uint) error {
	return db.Model(&Resource{}).Where("id IN ?", ids).Update("is_del", 1).Error
}
