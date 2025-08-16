package tables

import (
	"gorm.io/gorm"
	"time"
)

// BaseModel 基础模型，包含通用字段
type BaseModel struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	// 软删除 【查询时会检测删除时间】
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

/*
索引

给 Hash 建唯一索引（你已经有了）。

给 TakenAt 建索引（时间线展示时会用到）。

JSON 字段

SQLite 3.38+ 原生支持 JSON → 可以直接存 TEXT 并用 json_extract 查询。

GORM 层建议用 type:JSON 或 type:TEXT。

大文本/EXIF

如果 EXIF 很大，建议存 JSON 串在单独表里，不要污染主表。
*/
