package db

import (
	"rear/internal/model/tables"
)

func AutoMigrate() error {
	return DB.AutoMigrate(
		&tables.User{},
		&tables.LibraryTable{},
		&tables.Photo{},
		&tables.PhotoExif{},
		// 在这里添加其他模型
	)
}
