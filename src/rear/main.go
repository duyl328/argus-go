package main

import (
	"archive/zip"
	"context"
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
	"rear/internal/utils"
	"rear/pkg/logger"
	utilsPkg "rear/pkg/utils"
	"runtime"
	"strings"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
)

func main() {
	// 基础配置加载
	config.InitConfig()

	// 日志初始化
	err, logPath := getLogPath() // 获取日志路径
	if err != nil {
		log.Fatal("获取日志路径失败:", err)
		return
	}
	logConfig := logger.DefaultConfig()
	logConfig.LogPath = logPath
	// 使用配置文件中的日志级别
	if config.CONFIG.LoggingConfig.Level != "" {
		logConfig.Level = parseLogLevel(config.CONFIG.LoggingConfig.Level)
	}
	err = logger.InitDefaultLogger(logConfig)
	if err != nil {
		// log.Fatal 会输出错误信息并调用 os.Exit(1)
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	// 初始化工具依赖（仅在开发阶段）
	if config.IsDevelopment() {
		err = initializeToolDependencies()
		if err != nil {
			logger.Warnf("Failed to initialize tool dependencies: %v", err)
			logger.Warn("继续运行但某些功能可能受限")
		}
	}

	// 初始化工具路径
	err = initializeTools()
	if err != nil {
		logger.Warnf("Failed to initialize tools: %v", err)
		logger.Warn("继续运行但某些功能可能受限")
	}

	// 输出工具版本信息
	printToolsInfo()

	// 根据开发阶段执行不同的逻辑
	if config.IsDevelopment() {
		logger.Info("运行在开发模式")
		// 开发阶段特有的初始化代码
		initDevelopmentFeatures()
	} else {
		logger.Info("运行在生产模式")
		// 生产阶段的优化设置
		initProductionOptimizations()
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

	//if err := initDatabase(); err != nil {
	//	logger.Fatalf("Failed to initialize database: %v", err)
	//}
	// 依赖注入

	// 启动 http
	startHttp(newContainer, newTaskContainer)
}

// parseLogLevel 将字符串转换为LogLevel
func parseLogLevel(level string) logger.LogLevel {
	switch level {
	case "debug":
		return logger.DebugLevel
	case "info":
		return logger.InfoLevel
	case "warn", "warning":
		return logger.WarnLevel
	case "error":
		return logger.ErrorLevel
	case "fatal":
		return logger.FatalLevel
	default:
		return logger.InfoLevel
	}
}

// 获取日志路径
func getLogPath() (error, string) {
	var logPath string
	// 如果配置文件中指定了日志文件路径，使用配置的路径
	if config.CONFIG.LoggingConfig.FilePath != "" {
		logPath = config.CONFIG.LoggingConfig.FilePath
		// 如果是相对路径，转换为绝对路径
		if !filepath.IsAbs(logPath) {
			logPath = filepath.Join(config.CONFIG.AppDir, logPath)
		}
	} else {
		// 使用默认路径
		logPath = filepath.Join(config.CONFIG.AppDir, config.CONFIG.PathConfig.LogPath)
	}
	err := utilsPkg.FileUtils.CreateDir(logPath)
	if err != nil {
		logger.Error("日志文件夹创建失败！", zap.String("path", logPath), zap.Error(err))
		return err, ""
	}
	return nil, logPath
}

// initDevelopmentFeatures 初始化开发阶段功能
func initDevelopmentFeatures() {
	logger.Info("启用开发模式功能")
}

// initProductionOptimizations 初始化生产阶段优化
func initProductionOptimizations() {
	logger.Info("启用生产模式优化")
}

// initializeTools 初始化工具
func initializeTools() error {
	// 使用新的配置初始化方法
	return utils.InitializeFromConfig(&config.CONFIG.AppDir)
}

// printToolsInfo 输出工具版本信息
func printToolsInfo() {
	logger.Info("=== 工具信息 ===")

	tools := utils.GetToolsInfo()
	for _, tool := range tools {
		if tool.Error != nil {
			logger.Warnf("%-12s: %s (路径: %s)", tool.Name, tool.Error.Error(), tool.Path)
		} else {
			logger.Infof("%-12s: %s", tool.Name, tool.Version)
			logger.Infof("%-12s  路径: %s", "", tool.Path)
		}
	}

	logger.Info("================")
}

// 创建软件所需的缓存目录等内容
func createCachePath(dir string) {
	// 缩略图目录
	thumbnailPath := filepath.Join(dir, config.CONFIG.PathConfig.CachePath, config.CONFIG.PathConfig.ThumbnailPath)
	err := utilsPkg.FileUtils.CreateDir(thumbnailPath)
	if err != nil {
		logger.Error("缩略图路径创建失败！", zap.String("path", dir), zap.Error(err))
		return
	}
	// 临时文件路径
	tempPath := filepath.Join(dir, config.CONFIG.PathConfig.TempPath, config.CONFIG.PathConfig.PngTempPath)
	err = utilsPkg.FileUtils.CreateDir(tempPath)
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

// initializeToolDependencies 初始化工具依赖（仅在开发阶段）
func initializeToolDependencies() error {
	logger.Info("开始初始化工具依赖...")

	// 定义需要处理的工具
	tools := []string{"exiftool", "imagemagick", "libvips"}

	for _, tool := range tools {
		if err := extractToolFromZip(tool); err != nil {
			return fmt.Errorf("初始化工具 %s 失败: %v", tool, err)
		}
	}

	logger.Info("工具依赖初始化完成")
	return nil
}

// getPlatformDir 获取当前平台对应的目录名
func getPlatformDir() string {
	osName := runtime.GOOS
	arch := runtime.GOARCH

	switch osName {
	case "windows":
		if arch == "amd64" {
			return "windows_amd64"
		}
		return "windows_" + arch
	case "darwin":
		if arch == "amd64" {
			return "darwin_amd64"
		} else if arch == "arm64" {
			return "darwin_arm64"
		}
		return "darwin_" + arch
	case "linux":
		if arch == "amd64" {
			return "linux_amd64"
		}
		return "linux_" + arch
	default:
		return osName + "_" + arch
	}
}

// extractToolFromZip 从压缩包中解压工具到构建目录
func extractToolFromZip(toolName string) error {
	logger.Infof("处理工具: %s", toolName)

	// 获取当前平台目录
	platformDir := getPlatformDir()
	logger.Infof("当前平台: %s", platformDir)

	// 源压缩包目录（包含平台子目录）
	sourceDir := filepath.Join("tools", platformDir, toolName)

	// 检查平台特定目录是否存在
	if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
		return fmt.Errorf("平台目录不存在: %s，请确保为当前平台 %s 提供了 %s 工具包", sourceDir, platformDir, toolName)
	}

	// 目标解压目录
	targetDir := filepath.Join(config.CONFIG.AppDir, "tools", toolName)

	// 检查目标目录是否已存在且内容完整
	if isToolReady(targetDir, toolName) {
		logger.Infof("工具 %s 已存在且完整，跳过解压", toolName)
		return nil
	}

	// 选择合适的压缩包
	zipFile, err := selectZipFile(sourceDir, toolName)
	if err != nil {
		return fmt.Errorf("选择压缩包失败: %v", err)
	}

	logger.Infof("使用压缩包: %s", zipFile)

	// 解压压缩包
	if err := extractZip(zipFile, targetDir); err != nil {
		return fmt.Errorf("解压失败: %v", err)
	}

	// Windows平台特定的后处理
	if err := postProcessTool(toolName, targetDir); err != nil {
		logger.Warnf("工具 %s 后处理失败: %v", toolName, err)
	}

	logger.Infof("工具 %s 解压完成", toolName)
	return nil
}

// postProcessTool 对解压后的工具进行平台特定的后处理
func postProcessTool(toolName, targetDir string) error {
	// 只在Windows平台进行后处理
	if runtime.GOOS != "windows" {
		return nil
	}

	switch toolName {
	case "exiftool":
		return postProcessExiftool(targetDir)
	}

	return nil
}

// postProcessExiftool 对exiftool进行Windows平台的后处理
func postProcessExiftool(targetDir string) error {
	// 检查是否存在 exiftool(-k).exe
	oldPath := filepath.Join(targetDir, "exiftool(-k).exe")
	newPath := filepath.Join(targetDir, "exiftool.exe")

	// 如果 exiftool(-k).exe 存在
	if fileExists(oldPath) {
		// 如果 exiftool.exe 已经存在，先删除它
		if fileExists(newPath) {
			if err := os.Remove(newPath); err != nil {
				return fmt.Errorf("删除现有的 exiftool.exe 失败: %v", err)
			}
			logger.Infof("删除现有的 exiftool.exe")
		}

		// 重命名 exiftool(-k).exe 为 exiftool.exe
		if err := os.Rename(oldPath, newPath); err != nil {
			return fmt.Errorf("重命名 exiftool(-k).exe 为 exiftool.exe 失败: %v", err)
		}

		logger.Infof("已将 exiftool(-k).exe 重命名为 exiftool.exe")
	}

	return nil
}

// selectZipFile 选择合适的压缩包文件
func selectZipFile(sourceDir, toolName string) (string, error) {
	// 扫描目录中的zip文件
	zipFiles, err := filepath.Glob(filepath.Join(sourceDir, "*.zip"))
	if err != nil {
		return "", fmt.Errorf("扫描zip文件失败: %v", err)
	}

	if len(zipFiles) == 0 {
		return "", fmt.Errorf("在目录 %s 中未找到zip文件", sourceDir)
	}

	if len(zipFiles) == 1 {
		// 只有一个zip文件，直接使用
		return zipFiles[0], nil
	}

	// 多个zip文件的情况，查找默认名称的文件
	defaultZip := filepath.Join(sourceDir, toolName+".zip")
	for _, zipFile := range zipFiles {
		if zipFile == defaultZip {
			return zipFile, nil
		}
	}

	// 如果没有找到默认名称的文件，报错
	return "", fmt.Errorf("发现多个zip文件但未找到默认文件 %s.zip，请确保存在默认压缩包", toolName)
}

// isToolReady 检查工具是否已准备就绪（目录存在且有内容）
func isToolReady(targetDir, toolName string) bool {
	// 检查目录是否存在
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		return false
	}

	// 检查目录是否为空
	entries, err := os.ReadDir(targetDir)
	if err != nil || len(entries) == 0 {
		return false
	}

	// 根据不同工具检查关键文件是否存在
	switch toolName {
	case "exiftool":
		// Windows平台下经过后处理后应该只有 exiftool.exe
		if runtime.GOOS == "windows" {
			return fileExists(filepath.Join(targetDir, "exiftool.exe"))
		}
		// 其他平台检查两个可能的文件名
		return fileExists(filepath.Join(targetDir, "exiftool(-k).exe")) ||
			fileExists(filepath.Join(targetDir, "exiftool.exe"))
	case "imagemagick":
		return fileExists(filepath.Join(targetDir, "magick.exe"))
	case "libvips":
		return dirExists(filepath.Join(targetDir, "bin"))
	default:
		return true // 对于未知工具，只要有内容就认为准备就绪
	}
}

// fileExists 检查文件是否存在
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	return err == nil && !info.IsDir()
}

// dirExists 检查目录是否存在
func dirExists(dirname string) bool {
	info, err := os.Stat(dirname)
	return err == nil && info.IsDir()
}

// extractZip 解压zip文件到指定目录
func extractZip(src, dest string) error {
	// 创建目标目录
	if err := os.MkdirAll(dest, 0755); err != nil {
		return fmt.Errorf("创建目标目录失败: %v", err)
	}

	// 打开zip文件
	r, err := zip.OpenReader(src)
	if err != nil {
		return fmt.Errorf("打开zip文件失败: %v", err)
	}
	defer r.Close()

	// 解压每个文件
	for _, f := range r.File {
		err := extractFile(f, dest)
		if err != nil {
			return fmt.Errorf("解压文件 %s 失败: %v", f.Name, err)
		}
	}

	return nil
}

// extractFile 解压单个文件（扁平化解压，跳过顶层目录）
func extractFile(f *zip.File, destDir string) error {
	// 获取文件在压缩包中的路径
	zipPath := f.Name

	// 如果是以 / 或 \ 开头的路径，去掉开头的分隔符
	zipPath = strings.TrimPrefix(zipPath, "/")
	zipPath = strings.TrimPrefix(zipPath, "\\")

	// 将路径分割为目录组件
	pathParts := strings.Split(zipPath, "/")
	if len(pathParts) == 0 {
		return nil // 跳过空路径
	}

	// 如果只有一个组件且是目录，跳过（这通常是顶层目录）
	if len(pathParts) == 1 && f.FileInfo().IsDir() {
		return nil
	}

	// 构建目标路径：跳过第一个目录组件（顶层目录）
	var relativePath string
	if len(pathParts) > 1 {
		// 跳过顶层目录，保留子路径
		relativePath = filepath.Join(pathParts[1:]...)
	} else {
		// 如果只有一个组件且是文件，直接使用文件名
		relativePath = pathParts[0]
	}

	// 构建完整的目标路径
	targetPath := filepath.Join(destDir, relativePath)

	// 确保路径安全（防止zip slip攻击）
	cleanDestDir := filepath.Clean(destDir)
	cleanTargetPath := filepath.Clean(targetPath)

	// 检查目标路径是否在destDir范围内
	if !strings.HasPrefix(cleanTargetPath, cleanDestDir) ||
		(len(cleanTargetPath) > len(cleanDestDir) &&
			cleanTargetPath[len(cleanDestDir)] != filepath.Separator) {
		return fmt.Errorf("无效的文件路径: %s", f.Name)
	}

	if f.FileInfo().IsDir() {
		// 创建目录（如果需要的话）
		if len(pathParts) > 1 {
			return os.MkdirAll(targetPath, f.FileInfo().Mode())
		}
		return nil // 跳过顶层目录
	}

	// 创建父目录
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return err
	}

	// 打开zip中的文件
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	// 创建目标文件
	outFile, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.FileInfo().Mode())
	if err != nil {
		return err
	}
	defer outFile.Close()

	// 复制文件内容
	_, err = io.Copy(outFile, rc)
	return err
}
