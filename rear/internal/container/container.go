package container

import "rear/internal/repositories"

type Container struct {
	LibraryRepo *repositories.LibraryRepository
	UserRepo    *repositories.UserService
	// 其他服务...
}

func NewContainer() *Container {
	return &Container{
		LibraryRepo: repositories.NewLibraryRepository(),
		UserRepo:    repositories.NewUserService(),
	}
}
