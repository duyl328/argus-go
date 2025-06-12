package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// LogLevel 日志级别
type LogLevel int

const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

// Config 日志配置
type Config struct {
	Level           LogLevel `json:"level"`             // 日志级别
	LogPath         string   `json:"log_path"`          // 日志文件路径
	MaxSize         int      `json:"max_size"`          // 单个日志文件最大大小(MB)
	MaxBackups      int      `json:"max_backups"`       // 保留的旧日志文件个数
	MaxAge          int      `json:"max_age"`           // 保留旧日志文件的最大天数
	Compress        bool     `json:"compress"`          // 是否压缩旧日志文件
	EnableConsole   bool     `json:"enable_console"`    // 是否同时输出到控制台
	EnableCaller    bool     `json:"enable_caller"`     // 是否显示调用者信息(文件名和行号)
	TimeFormat      string   `json:"time_format"`       // 时间格式
	AsyncBufferSize int      `json:"async_buffer_size"` // 异步缓冲区大小
}

// Logger 日志器结构体
type Logger struct {
	zap    *zap.Logger
	sugar  *zap.SugaredLogger
	config *Config
}

// defaultConfig 默认配置
func defaultConfig() *Config {
	return &Config{
		Level:           InfoLevel,
		LogPath:         "./logs/app.log",
		MaxSize:         1,    // 1MB
		MaxBackups:      30,   // 保留30个备份
		MaxAge:          7,    // 保留7天
		Compress:        true, // 压缩旧文件
		EnableConsole:   true,
		EnableCaller:    true,
		TimeFormat:      "2006-01-02 15:04:05.000",
		AsyncBufferSize: 1000,
	}
}

// NewLogger 创建新的日志器
func NewLogger(skipCaller int, config ...*Config) (*Logger, error) {
	var cfg *Config
	if len(config) > 0 && config[0] != nil {
		cfg = config[0]
	} else {
		cfg = defaultConfig()
	}

	// 确保日志目录存在
	logDir := filepath.Dir(cfg.LogPath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// 配置日志轮转
	hook := &lumberjack.Logger{
		Filename:   cfg.LogPath,
		MaxSize:    cfg.MaxSize,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAge,
		Compress:   cfg.Compress,
		LocalTime:  true,
	}

	// 创建编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     getTimeEncoder(cfg.TimeFormat),
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 创建文件写入器
	var fileWriter zapcore.WriteSyncer
	if cfg.AsyncBufferSize > 0 {
		// 使用异步缓冲写入器
		fileWriter = &zapcore.BufferedWriteSyncer{
			WS:            zapcore.AddSync(hook),
			Size:          cfg.AsyncBufferSize,
			FlushInterval: time.Second,
		}
	} else {
		fileWriter = zapcore.AddSync(hook)
	}

	// 创建控制台写入器
	consoleWriter := zapcore.AddSync(os.Stdout)

	// 设置日志级别
	var level zapcore.Level
	switch cfg.Level {
	case DebugLevel:
		level = zapcore.DebugLevel
	case InfoLevel:
		level = zapcore.InfoLevel
	case WarnLevel:
		level = zapcore.WarnLevel
	case ErrorLevel:
		level = zapcore.ErrorLevel
	case FatalLevel:
		level = zapcore.FatalLevel
	default:
		level = zapcore.InfoLevel
	}

	// 创建核心组件
	var cores []zapcore.Core

	// 文件输出核心
	fileCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		fileWriter,
		level,
	)
	cores = append(cores, fileCore)

	// 控制台输出核心
	if cfg.EnableConsole {
		consoleEncoderConfig := encoderConfig
		consoleEncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		// zapcore.NewJSONEncoder(...) → 输出 JSON 格式日志（机器友好、可解析）
		// zapcore.NewConsoleEncoder(...) → 输出彩色文本日志（人类友好）
		consoleCore := zapcore.NewCore(
			zapcore.NewConsoleEncoder(consoleEncoderConfig),
			consoleWriter,
			level,
		)
		cores = append(cores, consoleCore)
	}

	// 创建tee核心（同时写入多个目标）
	core := zapcore.NewTee(cores...)

	// 配置选项
	var options []zap.Option
	if cfg.EnableCaller {
		// 向上寻找调用者【1 或 2 即可】
		options = append(options, zap.AddCaller(), zap.AddCallerSkip(skipCaller))
	}
	options = append(options, zap.AddStacktrace(zapcore.ErrorLevel))

	// 创建zap logger
	zapLogger := zap.New(core, options...)

	logger := &Logger{
		zap:    zapLogger,
		sugar:  zapLogger.Sugar(),
		config: cfg,
	}

	return logger, nil
}

// getTimeEncoder 获取时间编码器
func getTimeEncoder(format string) zapcore.TimeEncoder {
	return func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format(format))
	}
}

// Debug 调试级别日志
func (l *Logger) Debug(msg string, fields ...zap.Field) {
	l.zap.Debug(msg, fields...)
}

// Debugf 格式化调试日志
func (l *Logger) Debugf(template string, args ...interface{}) {
	l.sugar.Debugf(template, args...)
}

// Info 信息级别日志
func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.zap.Info(msg, fields...)
}

// Infof 格式化信息日志
func (l *Logger) Infof(template string, args ...interface{}) {
	l.sugar.Infof(template, args...)
}

// Warn 警告级别日志
func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.zap.Warn(msg, fields...)
}

// Warnf 格式化警告日志
func (l *Logger) Warnf(template string, args ...interface{}) {
	l.sugar.Warnf(template, args...)
}

// Error 错误级别日志
func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.zap.Error(msg, fields...)
}

// Errorf 格式化错误日志
func (l *Logger) Errorf(template string, args ...interface{}) {
	l.sugar.Errorf(template, args...)
}

// Fatal 致命错误日志（会调用os.Exit(1)）
func (l *Logger) Fatal(msg string, fields ...zap.Field) {
	l.zap.Fatal(msg, fields...)
}

// Fatalf 格式化致命错误日志
func (l *Logger) Fatalf(template string, args ...interface{}) {
	l.sugar.Fatalf(template, args...)
}

// With 添加字段到日志上下文
func (l *Logger) With(fields ...zap.Field) *Logger {
	return &Logger{
		zap:    l.zap.With(fields...),
		sugar:  l.zap.With(fields...).Sugar(),
		config: l.config,
	}
}

// WithFields 使用映射添加字段
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	var zapFields []zap.Field
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}
	return l.With(zapFields...)
}

// Sync 同步缓冲区（确保所有日志都被写入）
func (l *Logger) Sync() error {
	return l.zap.Sync()
}

// Close 关闭日志器
func (l *Logger) Close() error {
	return l.Sync()
}

// GetZapLogger 获取原始的zap logger（用于高级用法）
func (l *Logger) GetZapLogger() *zap.Logger {
	return l.zap
}

// GetSugarLogger 获取sugar logger（用于更简单的API）
func (l *Logger) GetSugarLogger() *zap.SugaredLogger {
	return l.sugar
}

// 全局默认日志器
var defaultLogger *Logger

// InitDefaultLogger 初始化默认日志器
func InitDefaultLogger(config ...*Config) error {
	var err error
	defaultLogger, err = NewLogger(2, config...)
	return err
}

// 自动初始化
func ensureDefaultLogger() {
	if defaultLogger == nil {
		_ = InitDefaultLogger() // 忽略错误或打印到stderr
	}
}

// Debug 全局日志函数
func Debug(msg string, fields ...zap.Field) {
	ensureDefaultLogger()
	defaultLogger.Debug(msg, fields...)
}

func Debugf(template string, args ...interface{}) {
	ensureDefaultLogger()
	defaultLogger.Debugf(template, args...)
}

func Info(msg string, fields ...zap.Field) {
	ensureDefaultLogger()
	defaultLogger.Info(msg, fields...)
}

func Infof(template string, args ...interface{}) {
	ensureDefaultLogger()
	defaultLogger.Infof(template, args...)
}

func Warn(msg string, fields ...zap.Field) {
	ensureDefaultLogger()
	defaultLogger.Warn(msg, fields...)
}

func Warnf(template string, args ...interface{}) {
	ensureDefaultLogger()
	defaultLogger.Warnf(template, args...)
}

func Error(msg string, fields ...zap.Field) {
	ensureDefaultLogger()
	defaultLogger.Error(msg, fields...)
}

func Errorf(template string, args ...interface{}) {
	ensureDefaultLogger()
	defaultLogger.Errorf(template, args...)
}

func Fatal(msg string, fields ...zap.Field) {
	ensureDefaultLogger()
	defaultLogger.Fatal(msg, fields...)
}

func Fatalf(template string, args ...interface{}) {
	ensureDefaultLogger()
	defaultLogger.Fatalf(template, args...)
}

// Sync 确保所有日志都被写入
func Sync() error {
	if defaultLogger != nil {
		return defaultLogger.Sync()
	}
	return nil
}
