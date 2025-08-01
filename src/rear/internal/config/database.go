package config

import (
	"time"
)

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Type DatabaseType
	// Mysql 属性
	Host string
	// Mysql 属性
	Port     string
	Database string
	// Mysql 属性
	Username string
	// Mysql 属性
	Password string
	// SQLite specific
	DBPath       string
	MaxIdleConns int
	MaxOpenConns int
	MaxLifetime  time.Duration
}

// DatabaseType 数据库类型
type DatabaseType string

const (
	SQLite DatabaseType = "sqlite"
	MySQL  DatabaseType = "mysql"
)
