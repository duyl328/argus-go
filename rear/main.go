package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"rear/internal/config"
	"rear/internal/container"
	"rear/internal/db"
	"rear/internal/repositories"
	"rear/internal/router"
	"rear/internal/service"
	"rear/pkg/logger"
	"rear/pkg/utils"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
)

//go:embed tools/windows_amd64/exiftool/*
var exiftoolFS embed.FS

func main() {
	// 基础配置加载
	config.InitConfig()

	// 日志初始化
	err := logger.InitDefaultLogger()
	if err != nil {
		// log.Fatal 会输出错误信息并调用 os.Exit(1)
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	// 数据库初始化
	if err := db.InitDatabase(); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	// 自动迁移
	if err := db.AutoMigrate(); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}
	// 初始化容器【数据库存储容器】
	newContainer := container.NewContainer()

	// 初始化基础服务（启动写操作处理协程）
	repositories.InitBaseService()

	// 初始化照片管理任务
	newTaskContainer := container.NewTaskContainer(newContainer)

	// 创建软件所需的缓存目录等内容
	createCachePath(config.CONFIG.AppDir)

	// 将 CLI 放到到运行目录
	join := filepath.Join(config.CONFIG.AppDir, "tools", "exiftool")
	srcDir := `.\tools\windows_amd64\exiftool\exiftool` // 源目录

	// 复制整个目录
	if err := utils.FileUtils.CopyDir(srcDir, join); err != nil {
		fmt.Printf("复制失败: %v\n", err)
		return
	}

	join1 := filepath.Join(config.CONFIG.AppDir, "tools", "vips")
	srcDir1 := `.\tools\windows_amd64\libvips\vips` // 源目录
	if err := utils.FileUtils.CopyDir(srcDir1, join1); err != nil {
		fmt.Printf("复制失败: %v\n", err)
		return
	}

	//if err := initDatabase(); err != nil {
	//	logger.Fatalf("Failed to initialize database: %v", err)
	//}
	// 依赖注入

	// 启动 http
	startHttp(newContainer, newTaskContainer)
}

// 创建软件所需的缓存目录等内容
func createCachePath(dir string) {
	// 缩略图目录
	thumbnailPath := filepath.Join(dir, config.CONFIG.PathConfig.CachePath, config.CONFIG.PathConfig.ThumbnailPath)
	err := utils.FileUtils.CreateDir(thumbnailPath)
	if err != nil {
		logger.Error("缩略图路径创建失败！", zap.String("path", dir), zap.Error(err))
		return
	}
	// 临时文件路径
	tempPath := filepath.Join(dir, config.CONFIG.PathConfig.TempPath, config.CONFIG.PathConfig.PngTempPath)
	err = utils.FileUtils.CreateDir(tempPath)
	if err != nil {
		logger.Error("临时文件夹创建失败！", zap.String("path", dir), zap.Error(err))
		return
	}
}

func startHttp(con *container.DbContainer, imgContain *container.TaskContainer) {
	// 设置Gin模式
	gin.SetMode(config.CONFIG.Mode)

	// 创建Gin引擎
	r := gin.New()

	// CORS 处理
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization", "X-Request-ID"}

	// 添加中间件
	r.Use(service.RequestIDMiddleware())    // 最先生成请求ID
	r.Use(service.LoggerMiddleware())       // 记录日志
	r.Use(gin.Recovery())                   // 恢复panic
	r.Use(cors.New(corsConfig))             // CORS 处理
	r.Use(service.ErrorHandlerMiddleware()) // 最后处理错误

	// 性能分析 (仅在debug模式下)
	if config.CONFIG.Mode == "debug" {
		pprof.Register(r)
	}

	// 设置路由
	router.SetupRoutes(r, con, imgContain)

	// 创建HTTP服务器
	srv := &http.Server{
		Addr:         ":" + config.CONFIG.Port,
		Handler:      r,
		ReadTimeout:  config.CONFIG.ReadTimeout,
		WriteTimeout: config.CONFIG.WriteTimeout,
		IdleTimeout:  config.CONFIG.IdleTimeout,
	}

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 启动服务器
	go func() {
		logger.Infof("Server starting on port 127.0.0.1:%s", config.CONFIG.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatalf("Failed to start server: %v", err)
			// 发送信号给主goroutine，让它知道启动失败
			quit <- syscall.SIGTERM
		}
	}()

	//  阻塞主goroutine，等待信号
	<-quit
	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Info("Server exited")
}
