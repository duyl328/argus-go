package container

import "rear/internal/repositories"

type DbContainer struct {
	LibraryRepo *repositories.LibraryRepository
	UserRepo    *repositories.UserService
	ExifRepo    *repositories.ExifRepository
	PhotoRepo   *repositories.PhotoRepository
	// 其他服务...
}

func NewContainer() *DbContainer {
	return &DbContainer{
		LibraryRepo: repositories.NewLibraryRepository(),
		UserRepo:    repositories.NewUserService(),
		ExifRepo:    repositories.NewExifRepository(),
		PhotoRepo:   repositories.NewPhotoRepository(),
	}
}
