package router

import (
	"rear/internal/container"
	"rear/internal/handler"

	"github.com/gin-gonic/gin"
)

// SetupRoutes 设置路由
func SetupRoutes(r *gin.Engine, contain *container.Container) {
	// 默认访问
	r.GET("/", handler.BasicResponse)

	// 健康检查
	r.GET("/health", handler.HealthCheck)

	// 资料库处理
	libraryHandler := handler.NewLibraryHandler(contain)
	devImageHandler := handler.NewDevImageHandler(contain)
	// API版本组
	v1 := r.Group("/api/v1")
	{
		v1.GET("", handler.BasicResponseV1)
		// 用户相关路由
		users := v1.Group("/users")
		{
			users.GET("", handler.GetUsers)
			users.GET("/:id", handler.GetUserByID)
			users.POST("", handler.CreateUser)
			users.PUT("/:id", handler.UpdateUser)
			users.DELETE("/:id", handler.DeleteUser)
		}
		// 存储库相关
		library := v1.Group("/library")
		{
			library.GET("", libraryHandler.GetLibrary)
			library.POST("", libraryHandler.AddLibrary)
			library.PUT("", libraryHandler.UpdateLibrary)
			library.DELETE("", libraryHandler.DeleteLibrary)
		}
	}
	// 开发组
	dev := r.Group("/dev")
	{
		// 处理图片格式
		dev.GET("/tool/exiftool/get_exif", devImageHandler.GetExif)
	}
}
