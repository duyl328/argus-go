package db

import "rear/internal/config"

// OptimizedDatabaseManager 优化版数据库管理器
type OptimizedDatabaseManager struct {
	*DatabaseManager
	config config.DatabaseConfig
}
