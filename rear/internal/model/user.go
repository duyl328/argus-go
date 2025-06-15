package model

type User struct {
	BaseModel
	Username string `gorm:"uniqueIndex;not null;size:50" json:"username"`
	Email    string `gorm:"uniqueIndex;not null;size:100" json:"email"`
	// json:"-" - 序列化时忽略此字段（密码不会出现在JSON中）
	Password string `gorm:"not null;size:255" json:"-"`
	Age      int    `gorm:"default:0" json:"age"`
	Status   int    `gorm:"default:1;comment:1-active,0-inactive" json:"status"`
}
