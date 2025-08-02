package config

import (
	"time"
)

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Type DatabaseType `yaml:"type" json:"type"`
	// Mysql 属性
	Host string `yaml:"host" json:"host"`
	// Mysql 属性
	Port     string `yaml:"port" json:"port"`
	Database string `yaml:"database" json:"database"`
	// Mysql 属性
	Username string `yaml:"username" json:"username"`
	// Mysql 属性
	Password string `yaml:"password" json:"password"`
	// SQLite specific - 数据库文件路径
	Path         string        `yaml:"path" json:"path"`
	MaxIdleConns int           `yaml:"max_idle_conns" json:"max_idle_conns"`
	MaxOpenConns int           `yaml:"max_open_conns" json:"max_open_conns"`
	MaxLifetime  time.Duration `yaml:"max_lifetime" json:"max_lifetime"`
}

// DatabaseType 数据库类型
type DatabaseType string

const (
	SQLite DatabaseType = "sqlite"
	MySQL  DatabaseType = "mysql"
)
