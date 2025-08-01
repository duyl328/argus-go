package repositories

import (
	"fmt"
	"rear/internal/db"
	"rear/internal/model"
	"rear/pkg/logger"
)

type LibraryRepository struct{}

func NewLibraryRepository() *LibraryRepository {
	return &LibraryRepository{}
}

func (s *LibraryRepository) AddLibrary(lib *model.LibraryTable) error {
	var result model.LibraryTable

	err := db.GetDB().Where("img_path = ?", lib.ImgPath).FirstOrCreate(&result, lib).Error
	if err != nil {
		return err
	}
	return ExecuteWrite(func() error {
		// 如果找到了已存在的记录，更新 IsEnable
		if result.ID != 0 && result.ImgPath == lib.ImgPath {
			return db.GetDB().Model(&result).Update("is_enable", true).Error
		}

		return nil
	})
	//return ExecuteWrite(func() error {
	//	return db.GetDB().Create(lib).Error
	//})
}

func (s *LibraryRepository) DeleteLibrary(imgPath string) error {
	result := fmt.Sprintf("DeleteLibrary called with oldImgPath: %s", imgPath)
	logger.Warn(result)
	return ExecuteWrite(func() error {
		return db.GetDB().Where("img_path = ?", imgPath).Delete(&model.LibraryTable{}).Error
	})
}

func (s *LibraryRepository) UpdateLibrary(oldImgPath string, isEnable bool) error {
	result := fmt.Sprintf("UpdateLibrary called with oldImgPath: %s, isEnable: %t", oldImgPath, isEnable)
	logger.Warn(result)
	/*
	   SQL 更新语句中只更新了 updated_at 字段，而没有更新 is_enable 字段。这是 GORM 的一个常见陷阱。
	   问题原因：
	   GORM 的 Updates 方法默认会忽略零值字段。如果 isEnable 参数是 false（布尔类型的零值），GORM 不会将其包含在更新语句中。
	   解决方案有以下几种：
	   方案1：使用 Select 明确指定字段（推荐）
	*/

	return ExecuteWrite(func() error {
		return db.GetDB().Where("img_path = ?", oldImgPath).
			Select("is_enable"). // 明确指定要更新的字段
			Updates(&model.LibraryTable{IsEnable: isEnable}).Error
	})
}

func (s *LibraryRepository) GetAllLibrary() ([]model.LibraryTable, error) {
	var library []model.LibraryTable
	err := ExecuteRead(func() error {
		return db.GetDB().Find(&library).Error
	})

	return library, err
}
