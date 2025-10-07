package router

import (
	"rear/internal/container"
	"rear/internal/handler"

	"github.com/gin-gonic/gin"
)

// CleanupFunc 清理函数类型
type CleanupFunc func()

// SetupRoutes 设置路由
// 返回一个清理函数，用于在服务器关闭时清理资源
func SetupRoutes(r *gin.Engine, contain *container.DbContainer, imgContain *container.TaskContainer) CleanupFunc {
	// 默认访问
	r.GET("/", handler.BasicResponse)

	// 健康检查
	r.GET("/health", handler.HealthCheck)

	// 资料库处理
	libraryHandler := handler.NewLibraryHandler(contain, imgContain)
	devImageHandler := handler.NewDevImageHandler(contain)
	exifHandler := handler.NewExifHandler(contain)
	photoHandler := handler.NewPhotoHandler(contain, imgContain)
	batchThumbnailHandler := handler.NewBatchThumbnailHandler()

	// SSE 实时通信处理
	sseHandler := handler.NewSSEHandler()

	// 文件系统处理 (需要 SSE Manager)
	fileSystemHandler := handler.NewFileSystemHandler(sseHandler.Manager)
	fileOperationsHandler := handler.NewFileOperationsHandler()

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
			photo.GET("/:hash", photoHandler.GetPhoto)               // 获取图像文件
			photo.GET("/preview", photoHandler.GetPhotoPreview)      // 预览图片（通过路径）
			photo.POST("/batch-preview", batchThumbnailHandler.GetBatchThumbnails) // 批量获取缩略图
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

		// SSE 实时通信路由
		sse := v1.Group("/sse")
		{
			sse.GET("/connect", sseHandler.HandleSSEConnection)            // SSE 连接
			sse.POST("/broadcast", sseHandler.BroadcastMessage)            // 广播消息
			sse.POST("/send/:clientId", sseHandler.SendToClient)           // 发送给特定客户端
			sse.GET("/clients", sseHandler.GetClients)                     // 获取所有客户端
			sse.GET("/clients/:clientId", sseHandler.GetClientInfo)        // 获取客户端信息
			sse.DELETE("/clients/:clientId", sseHandler.DisconnectClient)  // 断开客户端
			sse.GET("/stats", sseHandler.GetStats)                         // 获取统计信息
			sse.GET("/test", sseHandler.TestEvent)                         // 测试事件

			// 订阅管理路由
			sse.POST("/subscribe", sseHandler.Subscribe)                          // 订阅路径
			sse.POST("/unsubscribe", sseHandler.Unsubscribe)                      // 取消订阅
			sse.GET("/subscriptions/:clientId", sseHandler.GetClientSubscriptions) // 获取订阅列表
		}

		// 文件系统相关路由
		filesystem := v1.Group("/filesystem")
		{
			filesystem.GET("/browse", fileSystemHandler.BrowseFileSystem)       // 浏览文件系统
			filesystem.GET("/disk-usage", fileSystemHandler.GetDiskUsage)      // 获取磁盘使用情况
			filesystem.GET("/item", fileSystemHandler.GetFileSystemItem)       // 获取文件系统项目信息
			filesystem.GET("/search", fileSystemHandler.SearchFiles)           // 搜索文件

			// 文件操作相关路由
			filesystem.POST("/directory", fileOperationsHandler.CreateDirectory)   // 创建目录
			filesystem.DELETE("/item", fileOperationsHandler.DeleteItem)           // 删除文件或目录
			filesystem.PUT("/item/move", fileOperationsHandler.MoveItem)           // 移动/重命名
			filesystem.POST("/item/copy", fileOperationsHandler.CopyItem)          // 复制文件或目录

			// 文件系统监听相关路由
			filesystem.POST("/watch", fileSystemHandler.WatchPath)                 // 订阅文件夹监听
			filesystem.POST("/unwatch", fileSystemHandler.UnwatchPath)             // 取消文件夹监听
			filesystem.GET("/watched", fileSystemHandler.GetWatchedPaths)          // 获取监听路径列表
		}
	}
	// 开发组
	dev := r.Group("/dev")
	{
		// 处理图片格式
		dev.GET("/tool/exiftool/get_exif", devImageHandler.GetExif)
	}

	// 返回清理函数
	return func() {
		// 关闭 SSE Manager
		if sseHandler != nil && sseHandler.Manager != nil {
			sseHandler.Manager.Close()
		}

		// 关闭文件系统监听器
		if fileSystemHandler != nil && fileSystemHandler.GetFileWatcher() != nil {
			fileSystemHandler.GetFileWatcher().Close()
		}
	}
}
