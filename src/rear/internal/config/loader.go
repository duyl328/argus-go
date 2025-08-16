package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// AppConfig 应用程序配置
type AppConfig struct {
	Port         string        `yaml:"port" json:"port"`
	Mode         string        `yaml:"mode" json:"mode"`
	ReadTimeout  time.Duration `yaml:"read_timeout" json:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout" json:"write_timeout"`
	IdleTimeout  time.Duration `yaml:"idle_timeout" json:"idle_timeout"`
	Development  *bool         `yaml:"development" json:"development"` // 开发阶段标识，nil表示未设置
}

// ToolsConfig 工具路径配置
type ToolsConfig struct {
	ExifToolPath    string `yaml:"exiftool_path" json:"exiftool_path"`
	ImageMagickPath string `yaml:"imagemagick_path" json:"imagemagick_path"`
	LibVipsPath     string `yaml:"libvips_path" json:"libvips_path"`
}

// PathsConfig 路径配置
type PathsConfig struct {
	CachePath     string `yaml:"cache_path" json:"cache_path"`
	ThumbnailPath string `yaml:"thumbnail_path" json:"thumbnail_path"`
	LogPath       string `yaml:"log_path" json:"log_path"`
	TempPath      string `yaml:"temp_path" json:"temp_path"`
	PngTempPath   string `yaml:"png_temp_path" json:"png_temp_path"`
	DatabasePath  string `yaml:"database_path" json:"database_path"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level    string `yaml:"level" json:"level"`
	FilePath string `yaml:"file_path" json:"file_path"`
}

// ImageConfig 图像处理配置
type ImageConfig struct {
	ThumbnailFormat      string `yaml:"thumbnail_format" json:"thumbnail_format"`
	ThumbnailSize        []int  `yaml:"thumbnail_size" json:"thumbnail_size"`
	DefaultThumbnailSize int    `yaml:"default_thumbnail_size" json:"default_thumbnail_size"`
	ThumbnailQuality     int    `yaml:"thumbnail_quality" json:"thumbnail_quality"`
	ScreenSize           int    `yaml:"screen_size" json:"screen_size"`
	SupportedFormats     struct {
		Base      []string `yaml:"base" json:"base"`
		Special   []string `yaml:"special" json:"special"`
		Thumbnail []string `yaml:"thumbnail" json:"thumbnail"`
	} `yaml:"supported_formats" json:"supported_formats"`
}

// TaskConfig 任务处理配置
type TaskConfig struct {
	// 并发工作协程数量，0表示自动检测（CPU核心数）
	Concurrency int `yaml:"concurrency" json:"concurrency"`
	// 任务队列最大容量
	QueueCapacity int `yaml:"queue_capacity" json:"queue_capacity"`
	// 是否启用自动调整并发数
	AutoAdjust bool `yaml:"auto_adjust" json:"auto_adjust"`
	// 系统监控间隔（秒）
	MonitorInterval int `yaml:"monitor_interval" json:"monitor_interval"`
	// 最大并发数限制（防止系统过载）
	MaxConcurrency int `yaml:"max_concurrency" json:"max_concurrency"`
	// 最小并发数限制
	MinConcurrency int `yaml:"min_concurrency" json:"min_concurrency"`
	// 是否在启动时自动开始任务管理器
	AutoStart bool `yaml:"auto_start" json:"auto_start"`
}

// FileConfig 完整的配置文件结构
type FileConfig struct {
	App      AppConfig      `yaml:"app" json:"app"`
	Tools    ToolsConfig    `yaml:"tools" json:"tools"`
	Paths    PathsConfig    `yaml:"paths" json:"paths"`
	Logging  LoggingConfig  `yaml:"logging" json:"logging"`
	Database DatabaseConfig `yaml:"database" json:"database"`
	Image    ImageConfig    `yaml:"image" json:"image"`
	Task     TaskConfig     `yaml:"task" json:"task"`
}

// LoadConfigFromFile 从文件加载配置
func LoadConfigFromFile(configPath string) (*FileConfig, error) {
	// 检查配置文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("配置文件不存在: %s", configPath)
	}

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 解析YAML
	var config FileConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	return &config, nil
}

// GetDefaultFileConfig 获取默认配置
func GetDefaultFileConfig() *FileConfig {
	return &FileConfig{
		App: AppConfig{
			Port:         getEnv("PORT", "8080"),
			Mode:         getEnv("GIN_MODE", "debug"),
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  60 * time.Second,
			Development:  nil, // 默认不设置，表示生产阶段
		},
		Tools: ToolsConfig{
			ExifToolPath:    "",
			ImageMagickPath: "",
			LibVipsPath:     "",
		},
		Paths: PathsConfig{
			CachePath:     "cache",
			ThumbnailPath: "thumbnail",
			LogPath:       "app-logs",
			TempPath:      "app-tmp",
			PngTempPath:   "png-tmp",
			DatabasePath:  "db",
		},
		Logging: LoggingConfig{
			Level:    "info",
			FilePath: "",
		},
		Database: DatabaseConfig{
			Type:         SQLite,
			Database:     "argus",
			Host:         "localhost",
			Port:         "3306",
			Username:     "",
			Password:     "",
			Path:         "",
			MaxIdleConns: 1,
			MaxOpenConns: 1,
			MaxLifetime:  0,
		},
		Image: ImageConfig{
			ThumbnailFormat:      "jpg",
			ThumbnailSize:        []int{512, 720},
			DefaultThumbnailSize: 720,
			ThumbnailQuality:     80,
			ScreenSize:           1080,
			SupportedFormats: struct {
				Base      []string `yaml:"base" json:"base"`
				Special   []string `yaml:"special" json:"special"`
				Thumbnail []string `yaml:"thumbnail" json:"thumbnail"`
			}{
				Base:      []string{".jpg", ".jpeg", ".png", ".tif", ".tiff", ".bmp"},
				Special:   []string{".gif", ".heic", ".heif", ".webp", ".avif", ".jxl"},
				Thumbnail: []string{".jpg", ".webp"},
			},
		},
		Task: TaskConfig{
			Concurrency:     0,    // 0表示自动检测CPU核心数
			QueueCapacity:   1000, // 任务队列容量
			AutoAdjust:      true, // 启用自动调整
			MonitorInterval: 10,   // 10秒监控间隔
			MaxConcurrency:  0,    // 0表示CPU核心数*2
			MinConcurrency:  1,    // 最少1个并发
			AutoStart:       true, // 启动时自动开始
		},
	}
}

// SaveConfigToFile 保存配置到文件
func SaveConfigToFile(config *FileConfig, configPath string) error {
	// 确保目录存在
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}

	// 序列化为YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	return nil
}

// MergeWithDefaults 将文件配置与默认配置合并
func MergeWithDefaults(fileConfig *FileConfig) *FileConfig {
	defaultConfig := GetDefaultFileConfig()

	// 如果fileConfig为nil，直接返回默认配置
	if fileConfig == nil {
		return defaultConfig
	}

	// App配置合并
	if fileConfig.App.Port == "" {
		fileConfig.App.Port = defaultConfig.App.Port
	}
	if fileConfig.App.Mode == "" {
		fileConfig.App.Mode = defaultConfig.App.Mode
	}
	if fileConfig.App.ReadTimeout == 0 {
		fileConfig.App.ReadTimeout = defaultConfig.App.ReadTimeout
	}
	if fileConfig.App.WriteTimeout == 0 {
		fileConfig.App.WriteTimeout = defaultConfig.App.WriteTimeout
	}
	if fileConfig.App.IdleTimeout == 0 {
		fileConfig.App.IdleTimeout = defaultConfig.App.IdleTimeout
	}

	// Paths配置合并
	if fileConfig.Paths.CachePath == "" {
		fileConfig.Paths.CachePath = defaultConfig.Paths.CachePath
	}
	if fileConfig.Paths.ThumbnailPath == "" {
		fileConfig.Paths.ThumbnailPath = defaultConfig.Paths.ThumbnailPath
	}
	if fileConfig.Paths.LogPath == "" {
		fileConfig.Paths.LogPath = defaultConfig.Paths.LogPath
	}
	if fileConfig.Paths.TempPath == "" {
		fileConfig.Paths.TempPath = defaultConfig.Paths.TempPath
	}
	if fileConfig.Paths.PngTempPath == "" {
		fileConfig.Paths.PngTempPath = defaultConfig.Paths.PngTempPath
	}
	if fileConfig.Paths.DatabasePath == "" {
		fileConfig.Paths.DatabasePath = defaultConfig.Paths.DatabasePath
	}

	// Logging配置合并
	if fileConfig.Logging.Level == "" {
		fileConfig.Logging.Level = defaultConfig.Logging.Level
	}

	// Database配置合并
	if fileConfig.Database.Type == "" {
		fileConfig.Database.Type = defaultConfig.Database.Type
	}
	if fileConfig.Database.Database == "" {
		fileConfig.Database.Database = defaultConfig.Database.Database
	}
	if fileConfig.Database.Host == "" {
		fileConfig.Database.Host = defaultConfig.Database.Host
	}
	if fileConfig.Database.Port == "" {
		fileConfig.Database.Port = defaultConfig.Database.Port
	}
	if fileConfig.Database.MaxIdleConns == 0 {
		fileConfig.Database.MaxIdleConns = defaultConfig.Database.MaxIdleConns
	}
	if fileConfig.Database.MaxOpenConns == 0 {
		fileConfig.Database.MaxOpenConns = defaultConfig.Database.MaxOpenConns
	}

	// Image配置合并
	if fileConfig.Image.ThumbnailFormat == "" {
		fileConfig.Image.ThumbnailFormat = defaultConfig.Image.ThumbnailFormat
	}
	if len(fileConfig.Image.ThumbnailSize) == 0 {
		fileConfig.Image.ThumbnailSize = defaultConfig.Image.ThumbnailSize
	}
	if fileConfig.Image.ThumbnailQuality == 0 {
		fileConfig.Image.ThumbnailQuality = defaultConfig.Image.ThumbnailQuality
	}
	if fileConfig.Image.ScreenSize == 0 {
		fileConfig.Image.ScreenSize = defaultConfig.Image.ScreenSize
	}
	if fileConfig.Image.DefaultThumbnailSize == 0 {
		fileConfig.Image.DefaultThumbnailSize = defaultConfig.Image.DefaultThumbnailSize
	}
	if len(fileConfig.Image.SupportedFormats.Base) == 0 {
		fileConfig.Image.SupportedFormats.Base = defaultConfig.Image.SupportedFormats.Base
	}
	if len(fileConfig.Image.SupportedFormats.Special) == 0 {
		fileConfig.Image.SupportedFormats.Special = defaultConfig.Image.SupportedFormats.Special
	}
	if len(fileConfig.Image.SupportedFormats.Thumbnail) == 0 {
		fileConfig.Image.SupportedFormats.Thumbnail = defaultConfig.Image.SupportedFormats.Thumbnail
	}

	// Task配置合并
	if fileConfig.Task.QueueCapacity == 0 {
		fileConfig.Task.QueueCapacity = defaultConfig.Task.QueueCapacity
	}
	if fileConfig.Task.MonitorInterval == 0 {
		fileConfig.Task.MonitorInterval = defaultConfig.Task.MonitorInterval
	}
	if fileConfig.Task.MinConcurrency == 0 {
		fileConfig.Task.MinConcurrency = defaultConfig.Task.MinConcurrency
	}
	// 注意：Concurrency、MaxConcurrency、AutoAdjust、AutoStart 为0或false时可能是用户有意设置的
	// 所以这里不做默认值合并，而是在使用时进行处理

	return fileConfig
}

// GetConfigPaths 获取可能的配置文件路径
func GetConfigPaths() []string {
	execPath, _ := os.Executable()
	execDir := filepath.Dir(execPath)

	return []string{
		filepath.Join(execDir, "config.yaml"),
		filepath.Join(execDir, "config.yml"),
		filepath.Join(execDir, "configs", "config.yaml"),
		filepath.Join(execDir, "configs", "config.yml"),
		"config.yaml",
		"config.yml",
	}
}
