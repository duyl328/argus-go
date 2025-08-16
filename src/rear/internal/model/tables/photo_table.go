package tables

import "time"

// Photo 表结构 - 保存图片基础信息
type Photo struct {
	BaseModel
	ImgPath string `gorm:"column:img_path;not null" json:"imgPath"`
	ImgName string `gorm:"column:img_name;not null" json:"imgName"`
	Hash    string `gorm:"uniqueIndex;not null" json:"hash"`
	Width   int    `json:"width"`
	Height  int    `json:"height"`
	// 图片的宽高比
	AspectRatio float32 `json:"aspectRatio"`
	FileSize    int64   `json:"fileSize"`
	Format      string  `json:"format"`
	Notes       *string `json:"notes,omitempty"`

	// 文件创建时间
	FileCreatedAt *time.Time `json:"fileCreatedAt,omitempty"`

	// ========================= 用户行为数据 ==============================

	// 评分、分级
	Rating       int        `json:"rating"`
	LastViewedAt *time.Time `json:"lastViewedAt,omitempty"`
	// 访问次数
	ViewCount int `json:"viewCount"`
}
