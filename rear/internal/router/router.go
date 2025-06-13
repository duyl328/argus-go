package router

import (
	"github.com/gin-gonic/gin"
	"rear/internal/handler"
)

// 设置路由
func SetupRoutes(r *gin.Engine) {
	// 默认访问
	r.GET("/", handler.BasicResponse)

	// 健康检查
	r.GET("/health", handler.HealthCheck)

	// API版本组
	v1 := r.Group("/api/v1")
	{
		// 用户相关路由
		users := v1.Group("/users")
		{
			users.GET("", handler.GetUsers)
			users.GET("/:id", handler.GetUserByID)
			users.POST("", handler.CreateUser)
			users.PUT("/:id", handler.UpdateUser)
			users.DELETE("/:id", handler.DeleteUser)
		}
	}
}
