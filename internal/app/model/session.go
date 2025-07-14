package model

import "gorm.io/gorm"

type Session struct {
	Model            Model
	Title            string     `gorm:"size:128;comment:会话标题（可自动生成）" json:"title"`
	UserID           string     `gorm:"type:char(36);not null;comment:用户ID" json:"userId"`
	MessageCount     int        `gorm:"default:0;comment:总消息数" json:"messageCount"`
	FreeMessageCount int        `gorm:"default:0;comment:剩余免费消息额度" json:"freeMessageCount"`
	Data             JSON       `gorm:"type:json;comment:扩展数据" json:"data"`
	IsDel            int8       `gorm:"default:0;comment:0-正常 1-删除" json:"isDel"`
	Messages         []Message  `gorm:"foreignKey:SessionID"`
	Workflows        []Workflow `gorm:"foreignKey:SessionID"`
	Resources        []Resource `gorm:"foreignKey:SessionID"`
}

func (s *Session) TableName() string {
	return "session"
}

type SessionData struct {
	InputTypes   []string `json:"inputType"`    // ["text","image"]
	LastPreview  string   `json:"lastPreview"`  // 缩略图URL
	ModelVersion string   `json:"modelVersion"` // 使用模型
}

func (s *Session) GetData() (*SessionData, error) {
	var data SessionData
	err := s.Data.ToStruct(&data)
	return &data, err
}

// SetData 设置SessionData
func (s *Session) SetData(data *SessionData) error {
	return s.Data.FromStruct(data)
}

type SessionRepository interface {
	Generic[Session]
	GetByUserID(db *gorm.DB, userID string, page, pageSize int) ([]Session, error)
	CountByUserID(db *gorm.DB, userID string) (int64, error)
	SoftDelete(db *gorm.DB, id uint) error
}

type SessionRepo struct {
	GenericImpl[Session]
}

func NewSessionRepository() SessionRepository {
	return &SessionRepo{}
}

func (r *SessionRepo) GetByUserID(db *gorm.DB, userID string, page, pageSize int) ([]Session, error) {
	var sessions []Session
	err := db.Where("user_id = ? AND is_del = 0", userID).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&sessions).Error
	return sessions, err
}

func (r *SessionRepo) CountByUserID(db *gorm.DB, userID string) (int64, error) {
	var count int64
	err := db.Model(&Session{}).Where("user_id = ? AND is_del = 0", userID).Count(&count).Error
	return count, err
}

func (r *SessionRepo) SoftDelete(db *gorm.DB, id uint) error {
	return db.Model(&Session{}).Where("id = ?", id).Update("is_del", 1).Error
}
