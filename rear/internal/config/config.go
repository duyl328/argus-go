package config

import (
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"rear/internal/consts"
	"rear/internal/utils"
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
}

var CONFIG Config

// InitConfig 初始化配置
func InitConfig() *Config {
	i := ImageCompressionOptions{
		ThumbnailFormat:  consts.FormatJPG,
		ThumbnailSize:    []int{256, 512, 720},
		ThumbnailQuality: 80,
	}
	pathConfig := PathConfig{
		CachePath:     "cache",
		ThumbnailPath: "thumbnail",
		LogPath:       "app-logs",
		TempPath:      "app-tmp",
		PngTempPath:   "png-tmp",
	}

	execPath, err := os.Executable()
	if err != nil {
		logger.Fatal("无法获取程序路径: %v", zap.Error(err))
	}

	CONFIG = Config{
		Port:                      utils.GetEnv("PORT", "8080"),
		Mode:                      utils.GetEnv("GIN_MODE", "debug"),
		ReadTimeout:               30 * time.Second,
		WriteTimeout:              30 * time.Second,
		IdleTimeout:               60 * time.Second,
		BaseSupportedFileTypes:    []string{".jpg", ".jpeg", ".png", ".tif", ".tiff", ".bmp"},
		SpecialSupportedFileTypes: []string{".gif", ".heic", ".heif", ".webp", ".avif", ".jxl"},
		SupportedThumbnailFormat:  []string{".jpg", ".webp"},
		ImageCompressionOption:    i,
		PathConfig:                pathConfig,
		AppPath:                   execPath,
		AppDir:                    filepath.Dir(execPath),
	}
	return &CONFIG
}
