package workflow

import (
	"rear/internal/repositories"
)

// isExifRepoNil 检查EXIF仓库是否为nil
func isExifRepoNil(repo *repositories.ExifRepository) bool {
	return repo == nil
}

// isPhotoRepoNil 检查Photo仓库是否为nil
func isPhotoRepoNil(repo *repositories.PhotoRepository) bool {
	return repo == nil
}