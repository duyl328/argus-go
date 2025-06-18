package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io"
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
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
)

//go:embed tools/windows_amd64/exiftool/*
var exiftoolFS embed.FS

func main() {
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

	// 创建运行目录
	execPath, err := os.Executable()
	execDir := filepath.Dir(execPath)

	// 将 CLI 放到到运行目录
	join := filepath.Join(execDir, "tools", "exiftool")
	srcDir := `.\tools\windows_amd64\exiftool\exiftool` // 源目录
	//destDir := `.\build\tools\exiftool`                 // 目标目录

	// 复制整个目录
	if err := copyDirectory(srcDir, join); err != nil {
		fmt.Printf("复制失败: %v\n", err)
		return
	}

	// 初始化外部（CLI）工具管理
	// utils.EnsureInitialized()

	//if err := initDatabase(); err != nil {
	//	logger.Fatalf("Failed to initialize database: %v", err)
	//}
	// 依赖注入

	// 启动 http
	startHttp(newContainer)
}

func startHttp(con *container.Container) {
	// 配置加载
	netConfig := config.InitConfig()
	// 设置Gin模式
	gin.SetMode(netConfig.Mode)

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
	if netConfig.Mode == "debug" {
		pprof.Register(r)
	}

	// 设置路由
	router.SetupRoutes(r, con)

	// 创建HTTP服务器
	srv := &http.Server{
		Addr:         ":" + netConfig.Port,
		Handler:      r,
		ReadTimeout:  netConfig.ReadTimeout,
		WriteTimeout: netConfig.WriteTimeout,
		IdleTimeout:  netConfig.IdleTimeout,
	}

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 启动服务器
	go func() {
		logger.Infof("Server starting on port 127.0.0.1:%s", netConfig.Port)
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

func copyDirectory(src, dest string) error {

	// 获取源目录信息
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("源目录不存在: %s, 错误: %v", src, err)
	}

	// 创建目标目录
	if err := os.MkdirAll(dest, srcInfo.Mode()); err != nil {
		return fmt.Errorf("创建目标目录失败: %v", err)
	}

	// 读取源目录内容
	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("读取源目录失败: %v", err)
	}

	// 遍历并复制每个文件/目录
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		destPath := filepath.Join(dest, entry.Name())

		if entry.IsDir() {
			// 递归复制子目录
			if err := copyDirectory(srcPath, destPath); err != nil {
				return err
			}
		} else {
			// 复制文件
			if err := copyFile(srcPath, destPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func copyFile(src, dest string) error {
	// 打开源文件
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// 获取源文件信息
	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	// 确保目标目录存在
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}

	// 创建目标文件
	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destFile.Close()

	// 复制文件内容
	if _, err := io.Copy(destFile, srcFile); err != nil {
		return err
	}

	// 设置文件权限
	return os.Chmod(dest, srcInfo.Mode())
}
