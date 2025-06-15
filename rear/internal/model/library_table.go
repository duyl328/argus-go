package model

type LibraryTable struct {
	BaseModel
	// 照片存储库路径
	ImgPath string `gorm:"not null;size:255" json:"img_path"`
	// 是否开启
	IsEnable bool `gorm:"default:false;" json:"is_enable"`
}
