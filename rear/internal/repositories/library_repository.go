package repositories

import (
	"rear/internal/db"
	"rear/internal/model"
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
	return ExecuteWrite(func() error {
		return db.GetDB().Where("img_path = ?", imgPath).Delete(&model.LibraryTable{}).Error
	})
}

func (s *LibraryRepository) UpdateLibrary(oldImgPath string, newImgPath string) error {
	return ExecuteWrite(func() error {
		return db.GetDB().Where("img_path = ?", oldImgPath).
			Updates(&model.LibraryTable{ImgPath: newImgPath}).Error
	})
}

func (s *LibraryRepository) GetAllLibrary() ([]model.LibraryTable, error) {
	var library []model.LibraryTable
	err := ExecuteRead(func() error {
		return db.GetDB().Find(&library).Error
	})

	return library, err
}
