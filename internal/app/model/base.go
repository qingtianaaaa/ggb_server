package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"gorm.io/gorm"
	"time"
)

type Model struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

type JSON json.RawMessage

func (j *JSON) Scan(value interface{}) error {
	if value == nil {
		*j = nil
	}
	s, ok := value.([]byte)
	if !ok {
		return errors.New("invalid JSON")
	}
	*j = append((*j)[0:0], s...)
	return nil
}

func (j *JSON) Value() (driver.Value, error) {
	if len(*j) == 0 {
		return nil, nil
	}
	return []byte(*j), nil
}

func (j *JSON) MarshalJSON() ([]byte, error) {
	if j == nil {
		return []byte("null"), nil
	}
	return *j, nil
}

func (j *JSON) UnmarshalJSON(data []byte) error {
	if j == nil {
		return errors.New("null input")
	}
	*j = append((*j)[0:0], data...)
	return nil
}

func (j *JSON) ToStruct(v interface{}) error {
	return json.Unmarshal(*j, v)
}

func (j *JSON) FromStruct(v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	*j = data
	return nil
}

type Generic[T any] interface {
	Create(db *gorm.DB, entity *T) error
	GetById(db *gorm.DB, id int64) (*T, error)
	Update(db *gorm.DB, entity *T) error
	List(db *gorm.DB, page, pageSize int) ([]T, error)
}

type GenericImpl[T any] struct{}

func (g GenericImpl[T]) Create(db *gorm.DB, entity *T) error {
	return db.Create(entity).Error
}

func (g GenericImpl[T]) GetById(db *gorm.DB, id int64) (*T, error) {
	var entity T
	err := db.First(&entity, id).Error
	return &entity, err
}

func (g GenericImpl[T]) Update(db *gorm.DB, entity *T) error {
	return db.Save(entity).Error
}

func (g GenericImpl[T]) List(db *gorm.DB, page, pageSize int) ([]T, error) {
	var entities []T
	err := db.Limit(pageSize).Offset((page - 1) * pageSize).Find(&entities).Error
	return entities, err
}
