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
}

// InitConfig 初始化配置
func InitConfig() *Config {
	return &Config{
		Port:         utils.GetEnv("PORT", "8080"),
		Mode:         utils.GetEnv("GIN_MODE", "debug"),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}
