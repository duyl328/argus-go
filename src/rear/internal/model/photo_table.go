package model

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

	// 人物数量
	//

	IsAlgorithm    bool   `json:"isAlgorithm"`
	AlgorithmScore *int   `json:"algorithmScore,omitempty"`
	LastViewedTime *int64 `json:"lastViewedTime,omitempty"`
}
