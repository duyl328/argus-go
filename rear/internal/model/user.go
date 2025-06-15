package model

import "gorm.io/gorm"

type User struct {
	BaseModel
	Username string `gorm:"uniqueIndex;not null;size:50" json:"username"`
	Email    string `gorm:"uniqueIndex;not null;size:100" json:"email"`
	Password string `gorm:"not null;size:255" json:"-"`
	Age      int    `gorm:"default:0" json:"age"`
	Status   int    `gorm:"default:1;comment:1-active,0-inactive" json:"status"`
}

// TableName 指定表名
func (User) GetTableName() string {
	return "users"
}

// BeforeCreate GORM钩子，创建前执行
func (u *User) BeforeCreate(tx *gorm.DB) error {
	// 这里可以添加创建前的逻辑，比如密码加密
	return nil
}

// AfterCreate GORM钩子，创建后执行
func (u *User) AfterCreate(tx *gorm.DB) error {
	// 这里可以添加创建后的逻辑
	return nil
}
