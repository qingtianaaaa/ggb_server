package model

import "gorm.io/gorm"

type Workflow struct {
	Model      Model
	SessionID  int64   `gorm:"not null;comment:关联会话ID" json:"sessionId"`
	MessageID  int64   `gorm:"not null;comment:关联消息ID" json:"messageId"`
	RootID     int64   `gorm:"default:0;comment:根工作流ID" json:"rootId"`
	ParentID   int64   `gorm:"default:0;comment:上一步工作流ID" json:"parentId"`
	Type       int8    `gorm:"default:0;comment:步骤类型：1-提示词生成 2-GGB指令生成 3-HTML生成 4-HTML优化" json:"type"`
	Input      string  `gorm:"type:text;comment:步骤输入内容" json:"input"`
	Output     string  `gorm:"type:text;comment:步骤输出内容" json:"output"`
	ExecTimeMs int     `gorm:"default:0;comment:执行耗时(毫秒)" json:"execTimeMs"`
	Session    Session `gorm:"foreignKey:SessionID"`
	Message    Message `gorm:"foreignKey:MessageID"`
}

func (w *Workflow) TableName() string {
	return "workflow"
}

type WorkflowRepository interface {
	Generic[Workflow]
	GetByMessageID(db *gorm.DB, messageID int64) ([]Workflow, error)
	GetByRootID(db *gorm.DB, rootID int64) ([]Workflow, error)
	GetByType(db *gorm.DB, messageID int64, workflowType int8) (*Workflow, error)
}

type WorkflowRepo struct {
	GenericImpl[Workflow]
}

func NewWorkflowRepository() WorkflowRepository {
	return &WorkflowRepo{}
}

func (r *WorkflowRepo) GetByMessageID(db *gorm.DB, messageID int64) ([]Workflow, error) {
	var workflows []Workflow
	err := db.Where("message_id = ?", messageID).Find(&workflows).Error
	return workflows, err
}

func (r *WorkflowRepo) GetByRootID(db *gorm.DB, rootID int64) ([]Workflow, error) {
	var workflows []Workflow
	err := db.Where("root_id = ?", rootID).Find(&workflows).Error
	return workflows, err
}

func (r *WorkflowRepo) GetByType(db *gorm.DB, messageID int64, workflowType int8) (*Workflow, error) {
	var workflow Workflow
	err := db.Where("message_id = ? AND type = ?", messageID, workflowType).First(&workflow).Error
	return &workflow, err
}
