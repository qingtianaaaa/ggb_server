package model

import "gorm.io/gorm"

type Message struct {
	Model     Model
	ParentID  int64      `gorm:"default:0;comment:父消息ID" json:"parentId"`
	SessionID int64      `gorm:"not null;comment:关联的会话ID" json:"sessionId"`
	UserID    string     `gorm:"type:char(36);not null" json:"userId"`
	Message   string     `gorm:"type:text;comment:消息正文内容" json:"message"`
	Identity  int8       `gorm:"default:0;comment:消息身份：0-用户 1-模型" json:"identity"`
	ExecTime  int        `gorm:"default:0;comment:处理耗时(毫秒)" json:"execTime"`
	Status    int8       `gorm:"default:0;comment:状态：0-正常 1-撤回 2-删除" json:"status"`
	Data      JSON       `gorm:"type:json;comment:扩展数据" json:"data"`
	IsDel     int8       `gorm:"default:0" json:"isDel"`
	Session   Session    `gorm:"foreignKey:SessionID"`
	Workflows []Workflow `gorm:"foreignKey:MessageID"`
	Resources []Resource `gorm:"foreignKey:MessageID"`
}

func (m *Message) TableName() string {
	return "message"
}

type MessageData struct {
}

func (m *Message) GetData() (*MessageData, error) {
	var data MessageData
	err := m.Data.ToStruct(&data)
	return &data, err
}

func (m *Message) SetData(data *MessageData) error {
	return m.Data.FromStruct(data)
}

type MessageRepository interface {
	Generic[Message]
	GetBySessionID(db *gorm.DB, sessionID int64, page, pageSize int) ([]Message, error)
	GetLastMessage(db *gorm.DB, sessionID int64) (*Message, error)
	BatchCreate(db *gorm.DB, messages []Message) error
}

type MessageRepo struct {
	GenericImpl[Message]
}

func NewMessageRepository() MessageRepository {
	return &MessageRepo{}
}

func (r *MessageRepo) GetBySessionID(db *gorm.DB, sessionID int64, page, pageSize int) ([]Message, error) {
	var messages []Message
	err := db.Where("session_id = ? AND is_del = 0", sessionID).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Order("created_at ASC").
		Find(&messages).Error
	return messages, err
}

func (r *MessageRepo) GetLastMessage(db *gorm.DB, sessionID int64) (*Message, error) {
	var message Message
	err := db.Where("session_id = ? AND is_del = 0", sessionID).
		Order("created_at DESC").
		First(&message).Error
	return &message, err
}

func (r *MessageRepo) BatchCreate(db *gorm.DB, messages []Message) error {
	return db.CreateInBatches(messages, 100).Error
}
