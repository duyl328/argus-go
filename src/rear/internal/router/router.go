package router

import (
	"rear/internal/container"
	"rear/internal/handler"

	"github.com/gin-gonic/gin"
)

// SetupRoutes 设置路由
func SetupRoutes(r *gin.Engine, contain *container.DbContainer, imgContain *container.TaskContainer) {
	// 默认访问
	r.GET("/", handler.BasicResponse)

	// 健康检查
	r.GET("/health", handler.HealthCheck)

	// 资料库处理
	libraryHandler := handler.NewLibraryHandler(contain, imgContain)
	devImageHandler := handler.NewDevImageHandler(contain)
	exifHandler := handler.NewExifHandler(contain)
	photoHandler := handler.NewPhotoHandler(contain, imgContain)

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
			// 执行检索任务
			library.POST("indexed", libraryHandler.LibraryIndex)
		}
		// EXIF信息相关路由
		exif := v1.Group("/exif")
		{
			exif.GET("", exifHandler.GetAllExifs)                       // 获取所有EXIF记录
			exif.GET("/:hash", exifHandler.GetExifByHash)               // 根据照片哈希获取EXIF
			exif.GET("/statistics", exifHandler.GetExifStatistics)      // 获取统计信息
			exif.GET("/cameras", exifHandler.GetUniqueCameras)          // 获取相机列表
			exif.GET("/cameras/stats", exifHandler.GetCameraStatistics) // 获取相机统计
			exif.GET("/search", exifHandler.SearchExifs)                // 搜索EXIF
			exif.GET("/gps", exifHandler.GetExifsWithGPS)               // 获取GPS信息
			exif.GET("/iso", exifHandler.GetExifsByISO)                 // 根据ISO筛选
			exif.GET("/aperture", exifHandler.GetExifsByAperture)       // 根据光圈筛选
			exif.GET("/camera", exifHandler.GetExifsByCamera)           // 根据相机筛选
		}

		// 照片相关路由
		photo := v1.Group("/photo")
		{
			photo.GET("/:hash", photoHandler.GetPhoto) // 获取图像文件
		}

		// 照片列表相关路由
		photos := v1.Group("/photos")
		{
			photos.GET("/timeline", photoHandler.GetTimeline) // 获取照片时间线统计
			photos.GET("", photoHandler.GetPhotos)            // 获取照片列表（支持范围查询）
		}

		// 资产信息相关路由
		assets := v1.Group("/assets")
		{
			assets.GET("/:hash", photoHandler.GetPhotoAssets) // 获取图像详细信息
		}
	}
	// 开发组
	dev := r.Group("/dev")
	{
		// 处理图片格式
		dev.GET("/tool/exiftool/get_exif", devImageHandler.GetExif)
	}
}
