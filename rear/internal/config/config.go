package config

import (
	"rear/internal/utils"
	"time"
)

// Config 配置结构
type Config struct {
	Port         string
	Mode         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	// 基础支持的格式
	BaseSupportedFileTypes []string
	// 特殊支持的格式
	SpecialSupportedFileTypes []string
}

var CONFIG Config

// InitConfig 初始化配置
func InitConfig() *Config {
	CONFIG = Config{
		Port:                      utils.GetEnv("PORT", "8080"),
		Mode:                      utils.GetEnv("GIN_MODE", "debug"),
		ReadTimeout:               30 * time.Second,
		WriteTimeout:              30 * time.Second,
		IdleTimeout:               60 * time.Second,
		BaseSupportedFileTypes:    []string{".jpg", ".jpeg", ".png", ".tif", ".tiff", ".bmp"},
		SpecialSupportedFileTypes: []string{".gif", ".heic", ".heif", ".webp", ".avif", ".jxl"},
	}
	return &CONFIG
}
