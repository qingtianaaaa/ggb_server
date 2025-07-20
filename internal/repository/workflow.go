package repository

import (
	"ggb_server/internal/app/model"
	"gorm.io/gorm"
)

type WorkflowRepository[T any] interface {
	Generic[T]
	GetByMessageID(db *gorm.DB, messageID int64) ([]model.Workflow, error)
	GetByRootID(db *gorm.DB, rootID int64) ([]model.Workflow, error)
	GetByType(db *gorm.DB, messageID int64, workflowType int8) (*model.Workflow, error)
}

type WorkflowRepo[T any] struct {
	GenericImpl[T]
}

func NewWorkflowRepository[T any]() WorkflowRepository[T] {
	return &WorkflowRepo[T]{
		GenericImpl[T]{},
	}
}

func (r *WorkflowRepo[T]) GetByMessageID(db *gorm.DB, messageID int64) ([]model.Workflow, error) {
	var workflows []model.Workflow
	err := db.Where("message_id = ?", messageID).Find(&workflows).Error
	return workflows, err
}

func (r *WorkflowRepo[T]) GetByRootID(db *gorm.DB, rootID int64) ([]model.Workflow, error) {
	var workflows []model.Workflow
	err := db.Where("root_id = ?", rootID).Find(&workflows).Error
	return workflows, err
}

func (r *WorkflowRepo[T]) GetByType(db *gorm.DB, messageID int64, workflowType int8) (*model.Workflow, error) {
	var workflow model.Workflow
	err := db.Where("message_id = ? AND type = ?", messageID, workflowType).First(&workflow).Error
	return &workflow, err
}
