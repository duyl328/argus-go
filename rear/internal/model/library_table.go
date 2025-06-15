package model

type LibraryTable struct {
	BaseModel
	// 照片存储库路径
	ImgPath string `gorm:"not null;size:255" json:"imgPath"`
	// 是否开启
	IsEnable bool `gorm:"default:false;" json:"isEnable"`
}
