package db

import (
	"rear/internal/model"
)

func AutoMigrate() error {
	return DB.AutoMigrate(
		&model.User{},
		// 在这里添加其他模型
	)
}
