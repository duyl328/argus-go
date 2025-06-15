package model

import (
	"gorm.io/gorm"
	"time"
)

// BaseModel 基础模型，包含通用字段
type BaseModel struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	IsDeleted bool           `json:"is_deleted"`
}
