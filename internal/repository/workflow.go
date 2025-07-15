package repository

import (
	"ggb_server/internal/app/model"
	"gorm.io/gorm"
)

type WorkflowRepository interface {
	Generic[model.Workflow]
	GetByMessageID(db *gorm.DB, messageID int64) ([]model.Workflow, error)
	GetByRootID(db *gorm.DB, rootID int64) ([]model.Workflow, error)
	GetByType(db *gorm.DB, messageID int64, workflowType int8) (*model.Workflow, error)
}

type WorkflowRepo struct {
	GenericImpl[model.Workflow]
}

func NewWorkflowRepository() WorkflowRepository {
	return &WorkflowRepo{}
}

func (r *WorkflowRepo) GetByMessageID(db *gorm.DB, messageID int64) ([]model.Workflow, error) {
	var workflows []model.Workflow
	err := db.Where("message_id = ?", messageID).Find(&workflows).Error
	return workflows, err
}

func (r *WorkflowRepo) GetByRootID(db *gorm.DB, rootID int64) ([]model.Workflow, error) {
	var workflows []model.Workflow
	err := db.Where("root_id = ?", rootID).Find(&workflows).Error
	return workflows, err
}

func (r *WorkflowRepo) GetByType(db *gorm.DB, messageID int64, workflowType int8) (*model.Workflow, error) {
	var workflow model.Workflow
	err := db.Where("message_id = ? AND type = ?", messageID, workflowType).First(&workflow).Error
	return &workflow, err
}
