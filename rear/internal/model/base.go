package model

import (
	"gorm.io/gorm"
	"time"
)

// BaseModel 基础模型，包含通用字段
type BaseModel struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	// 软删除 【查询时会检测删除时间】
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
