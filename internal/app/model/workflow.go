package model

type Workflow struct {
	Model
	SessionID  uint    `gorm:"not null;comment:关联会话ID" json:"sessionId"`
	MessageID  uint    `gorm:"not null;comment:关联消息ID" json:"messageId"`
	RootID     uint    `gorm:"default:0;comment:根工作流ID" json:"rootId"`
	ParentID   uint    `gorm:"default:0;comment:上一步工作流ID" json:"parentId"`
	Type       int     `gorm:"default:0;comment:步骤类型：1-提示词生成 2-GGB指令生成 3-HTML生成 4-HTML优化" json:"type"`
	Input      string  `gorm:"type:text;comment:步骤输入内容" json:"input"`
	Output     string  `gorm:"type:text;comment:步骤输出内容" json:"output"`
	ExecTimeMs int     `gorm:"default:0;comment:执行耗时(毫秒)" json:"execTimeMs"`
	Session    Session `gorm:"foreignKey:SessionID"`
	Message    Message `gorm:"foreignKey:MessageID"`
}

func (w *Workflow) TableName() string {
	return "workflow"
}
