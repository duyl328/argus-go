package container

import "rear/internal/repositories"

type DbContainer struct {
	LibraryRepo *repositories.LibraryRepository
	UserRepo    *repositories.UserService
	// 其他服务...
}

func NewContainer() *DbContainer {
	return &DbContainer{
		LibraryRepo: repositories.NewLibraryRepository(),
		UserRepo:    repositories.NewUserService(),
	}
}
