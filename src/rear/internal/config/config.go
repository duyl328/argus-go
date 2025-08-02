package config

import (
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"rear/internal/consts"
	"rear/pkg/logger"
	"time"
)

// ImageCompressionOptions 压缩图像配置
type ImageCompressionOptions struct {
	// 缩略图格式
	ThumbnailFormat consts.ImageFormat
	// 缩略图大小
	ThumbnailSize []int
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
		ThumbnailFormat:  convertToImageFormat(fileConfig.Image.ThumbnailFormat),
		ThumbnailSize:    fileConfig.Image.ThumbnailSize,
		ThumbnailQuality: fileConfig.Image.ThumbnailQuality,
		ScreenSize:       fileConfig.Image.ScreenSize,
	}

	pathConfig := PathConfig{
		CachePath:     fileConfig.Paths.CachePath,
		ThumbnailPath: fileConfig.Paths.ThumbnailPath,
		LogPath:       fileConfig.Paths.LogPath,
		TempPath:      fileConfig.Paths.TempPath,
		PngTempPath:   fileConfig.Paths.PngTempPath,
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
		return CONFIG.FileConfig.Database
	}
	// 返回默认配置
	return DatabaseConfig{
		Type:         SQLite,
		Database:     "test.db",
		MaxIdleConns: 1,
		MaxOpenConns: 1,
		MaxLifetime:  0,
	}
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
