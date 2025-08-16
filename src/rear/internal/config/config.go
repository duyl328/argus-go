package config

import (
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"rear/internal/consts"
	"rear/pkg/logger"
	"runtime"
	"time"
)

// ImageCompressionOptions 压缩图像配置
type ImageCompressionOptions struct {
	// 缩略图格式
	ThumbnailFormat consts.ImageFormat
	// 缩略图大小
	ThumbnailSize []int
	// 默认缩略图大小
	DefaultThumbnailSize int
	// 缩略图质量
	ThumbnailQuality int
	// 用户屏幕尺寸大小
	ScreenSize int
}

// PathConfig 路径相关配置
type PathConfig struct {
	// 缓存内容存放
	CachePath string
	// 缩略图存放路径
	ThumbnailPath string
	// 日志路径
	LogPath string
	// 临时文件路径
	TempPath string
	// png 临时文件
	PngTempPath string
	// 数据库路径
	DatabasePath string
}

// Config 配置结构
type Config struct {
	Port         string
	Mode         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	// 基础支持的格式【ImageMagick, libvips】
	BaseSupportedFileTypes []string
	// 特殊支持的格式
	SpecialSupportedFileTypes []string
	// 支持的缩略图格式
	SupportedThumbnailFormat []string
	ImageCompressionOption   ImageCompressionOptions
	// Raw 单独支持【rawtherapee-cli, dcraw】
	// heic 单独支持【heif-convert, magick】

	PathConfig PathConfig

	// 软件运行目录
	AppPath string
	AppDir  string

	// 新增配置字段
	ToolsConfig   ToolsConfig
	LoggingConfig LoggingConfig
	TaskConfig    TaskConfig  // 任务处理配置
	FileConfig    *FileConfig // 原始配置文件内容
}

var CONFIG Config

// InitConfig 初始化配置
func InitConfig() *Config {
	execPath, err := os.Executable()
	if err != nil {
		logger.Fatal("无法获取程序路径: %v", zap.Error(err))
	}

	execDir := filepath.Dir(execPath)

	// 尝试加载配置文件
	var fileConfig *FileConfig
	configPaths := GetConfigPaths()
	for _, configPath := range configPaths {
		if config, err := LoadConfigFromFile(configPath); err == nil {
			fileConfig = config
			logger.Info("加载配置文件成功", zap.String("path", configPath))
			break
		}
	}

	// 如果没有找到配置文件，使用默认配置
	if fileConfig == nil {
		logger.Info("未找到配置文件，使用默认配置")
		fileConfig = GetDefaultFileConfig()
	} else {
		// 与默认配置合并
		fileConfig = MergeWithDefaults(fileConfig)
	}

	// 转换为原有的配置格式，保持向后兼容
	imageCompressionOptions := ImageCompressionOptions{
		ThumbnailFormat:      convertToImageFormat(fileConfig.Image.ThumbnailFormat),
		ThumbnailSize:        fileConfig.Image.ThumbnailSize,
		DefaultThumbnailSize: fileConfig.Image.DefaultThumbnailSize,
		ThumbnailQuality:     fileConfig.Image.ThumbnailQuality,
		ScreenSize:           fileConfig.Image.ScreenSize,
	}

	pathConfig := PathConfig{
		CachePath:     fileConfig.Paths.CachePath,
		ThumbnailPath: fileConfig.Paths.ThumbnailPath,
		LogPath:       fileConfig.Paths.LogPath,
		TempPath:      fileConfig.Paths.TempPath,
		PngTempPath:   fileConfig.Paths.PngTempPath,
		DatabasePath:  fileConfig.Paths.DatabasePath,
	}

	CONFIG = Config{
		Port:                      fileConfig.App.Port,
		Mode:                      fileConfig.App.Mode,
		ReadTimeout:               fileConfig.App.ReadTimeout,
		WriteTimeout:              fileConfig.App.WriteTimeout,
		IdleTimeout:               fileConfig.App.IdleTimeout,
		BaseSupportedFileTypes:    fileConfig.Image.SupportedFormats.Base,
		SpecialSupportedFileTypes: fileConfig.Image.SupportedFormats.Special,
		SupportedThumbnailFormat:  fileConfig.Image.SupportedFormats.Thumbnail,
		ImageCompressionOption:    imageCompressionOptions,
		PathConfig:                pathConfig,
		AppPath:                   execPath,
		AppDir:                    execDir,
		ToolsConfig:               fileConfig.Tools,
		LoggingConfig:             fileConfig.Logging,
		TaskConfig:                fileConfig.Task,
		FileConfig:                fileConfig,
	}
	return &CONFIG
}

// convertToImageFormat 转换字符串到ImageFormat
func convertToImageFormat(format string) consts.ImageFormat {
	switch format {
	case "jpg", "jpeg":
		return consts.FormatJPG
	case "png":
		return consts.FormatPNG
	case "webp":
		return consts.FormatWEBP
	default:
		return consts.FormatJPG
	}
}

// GetDatabaseConfig 获取数据库配置
func GetDatabaseConfig() DatabaseConfig {
	if CONFIG.FileConfig != nil {
		dbConfig := CONFIG.FileConfig.Database

		// 如果是SQLite且需要处理路径
		if dbConfig.Type == SQLite {
			dbConfig.Path = processSQLitePath(dbConfig.Path)
		}

		return dbConfig
	}
	// 返回默认配置
	return DatabaseConfig{
		Type:         SQLite,
		Path:         processSQLitePath(""),
		Database:     "argus",
		MaxIdleConns: 1,
		MaxOpenConns: 1,
		MaxLifetime:  0,
	}
}

// processSQLitePath 处理SQLite数据库路径
// 如果用户指定了路径，使用用户指定的
// 如果没有指定，在运行文件根目录下创建db文件夹，并使用db.sqlite作为文件名
func processSQLitePath(configPath string) string {
	// 如果用户指定了路径，直接使用
	if configPath != "" {
		// 如果是相对路径，转换为基于AppDir的绝对路径
		if !filepath.IsAbs(configPath) {
			return filepath.Join(CONFIG.AppDir, configPath)
		}
		return configPath
	}

	// 用户没有指定路径，使用默认路径：AppDir/db/db.sqlite
	dbDir := filepath.Join(CONFIG.AppDir, "db")
	dbFile := filepath.Join(dbDir, "db.sqlite")

	// 确保db目录存在
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		// 如果创建目录失败，记录警告但继续使用该路径
		logger.Warn("创建数据库目录失败，但将继续使用该路径",
			zap.String("path", dbDir),
			zap.Error(err))
	}

	return dbFile
}

// IsDevelopment 判断是否为开发阶段
// 如果配置文件中设置了 development 字段且为 true，则认为是开发阶段
// 如果未设置或为 false，则认为是生产阶段
func IsDevelopment() bool {
	if CONFIG.FileConfig != nil && CONFIG.FileConfig.App.Development != nil {
		return *CONFIG.FileConfig.App.Development
	}
	return false // 默认为生产阶段
}

// IsProduction 判断是否为生产阶段
func IsProduction() bool {
	return !IsDevelopment()
}

// GetTaskConfig 获取优化后的任务配置
func GetTaskConfig() TaskConfig {
	taskConfig := CONFIG.TaskConfig

	// 如果并发数为0，自动设置为CPU核心数
	if taskConfig.Concurrency == 0 {
		taskConfig.Concurrency = runtime.NumCPU()
	}

	// 如果最大并发数为0，设置为CPU核心数*2
	if taskConfig.MaxConcurrency == 0 {
		taskConfig.MaxConcurrency = runtime.NumCPU() * 2
	}

	// 确保最小并发数至少为1
	if taskConfig.MinConcurrency == 0 {
		taskConfig.MinConcurrency = 1
	}

	// 确保并发数在合理范围内
	if taskConfig.Concurrency < taskConfig.MinConcurrency {
		taskConfig.Concurrency = taskConfig.MinConcurrency
	}
	if taskConfig.Concurrency > taskConfig.MaxConcurrency {
		taskConfig.Concurrency = taskConfig.MaxConcurrency
	}

	// 如果队列容量为0，设置默认值
	if taskConfig.QueueCapacity == 0 {
		taskConfig.QueueCapacity = 1000
	}

	// 如果监控间隔为0，设置默认值
	if taskConfig.MonitorInterval == 0 {
		taskConfig.MonitorInterval = 10
	}

	logger.Info("任务配置已优化",
		zap.Int("concurrency", taskConfig.Concurrency),
		zap.Int("max_concurrency", taskConfig.MaxConcurrency),
		zap.Int("min_concurrency", taskConfig.MinConcurrency),
		zap.Int("queue_capacity", taskConfig.QueueCapacity),
		zap.Int("monitor_interval", taskConfig.MonitorInterval),
		zap.Bool("auto_adjust", taskConfig.AutoAdjust),
		zap.Bool("auto_start", taskConfig.AutoStart))

	return taskConfig
}

// GetDefaultThumbnailSize 获取默认缩略图尺寸
func GetDefaultThumbnailSize() int {
	if CONFIG.ImageCompressionOption.DefaultThumbnailSize > 0 {
		return CONFIG.ImageCompressionOption.DefaultThumbnailSize
	}
	// 如果未设置或为0，返回默认值720
	return 720
}
