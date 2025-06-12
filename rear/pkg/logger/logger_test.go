package logger

import (
	"errors"
	"go.uber.org/zap"
	"testing"
	"time"
)

func TestLog(t *testing.T) {
	//err := InitDefaultLogger()
	//if err != nil {
	//	panic(err)
	//}
	Infof("App started with default logger")
	Sync()
}

func main() {
	// 示例1: 使用默认配置
	defaultLogger, err := NewLogger(1)
	if err != nil {
		panic(err)
	}
	defer defaultLogger.Close()

	defaultLogger.Info("Application started with default config")

	// 示例2: 使用自定义配置
	customConfig := &Config{
		Level:           InfoLevel,
		LogPath:         "./logs/custom.log",
		MaxSize:         2,    // 2MB
		MaxBackups:      10,   // 保留10个备份
		MaxAge:          3,    // 保留3天
		Compress:        true, // 压缩旧文件
		EnableConsole:   true, // 同时输出到控制台
		EnableCaller:    true, // 显示文件名和行号
		TimeFormat:      "2006-01-02 15:04:05",
		AsyncBufferSize: 500, // 异步缓冲区大小
	}

	customLogger, err := NewLogger(1, customConfig)
	if err != nil {
		panic(err)
	}
	defer customLogger.Close()

	// 示例3: 各种日志级别的使用
	customLogger.Debug("This is a debug message")
	customLogger.Info("This is an info message")
	customLogger.Warn("This is a warning message")
	customLogger.Error("This is an error message")

	// 示例4: 格式化日志
	customLogger.Infof("User %s logged in at %s", "john_doe", time.Now().Format("15:04:05"))
	customLogger.Warnf("Failed to process %d items", 5)

	// 示例5: 结构化日志（带字段）
	customLogger.Info("User operation",
		zap.String("user_id", "12345"),
		zap.String("action", "login"),
		zap.Duration("duration", time.Millisecond*150),
		zap.Bool("success", true),
	)

	customLogger.Error("Database connection failed",
		zap.String("database", "mysql"),
		zap.String("host", "localhost:3306"),
		zap.Error(errors.New("connection timeout")),
	)

	// 示例6: 使用With添加上下文字段
	userLogger := customLogger.With(
		zap.String("user_id", "67890"),
		zap.String("session_id", "abc123"),
	)

	userLogger.Info("User action performed")
	userLogger.Warn("Rate limit approached")

	// 示例7: 使用WithFields添加多个字段
	requestLogger := customLogger.WithFields(map[string]interface{}{
		"request_id": "req_001",
		"method":     "POST",
		"path":       "/api/users",
		"ip":         "192.168.1.1",
	})

	requestLogger.Info("Request started")
	requestLogger.Infof("Request completed in %dms", 250)

	// 示例8: 初始化全局默认日志器
	err = InitDefaultLogger(&Config{
		Level:         DebugLevel,
		LogPath:       "./logs/global.log",
		MaxSize:       1,
		EnableConsole: true,
		EnableCaller:  true,
	})
	if err != nil {
		panic(err)
	}

	// 使用全局日志函数
	Info("Using global logger")
	Debugf("Debug message with value: %d", 42)
	Error("Global error", zap.String("component", "main"))

	// 示例9: 模拟不同场景的日志
	simulateApplicationFlow(customLogger)

	// 示例10: 性能测试（异步日志的好处）
	performanceTest(customLogger)

	// 确保所有日志都被写入
	Sync()
}

// simulateApplicationFlow 模拟应用程序流程
func simulateApplicationFlow(l *Logger) {
	l.Info("Starting application flow simulation")

	// 模拟用户注册
	userID := "user_12345"
	userLogger := l.With(zap.String("user_id", userID))

	userLogger.Info("User registration started")
	userLogger.Debug("Validating user input")

	// 模拟数据库操作
	dbLogger := userLogger.With(zap.String("component", "database"))
	dbLogger.Info("Inserting user into database")

	// 模拟成功
	dbLogger.Info("User created successfully",
		zap.Duration("db_time", time.Millisecond*45),
		zap.String("table", "users"),
	)

	userLogger.Info("User registration completed successfully")

	// 模拟错误场景
	errorLogger := l.With(zap.String("component", "payment"))
	errorLogger.Error("Payment processing failed",
		zap.String("payment_id", "pay_67890"),
		zap.String("gateway", "stripe"),
		zap.Error(errors.New("insufficient funds")),
		zap.Float64("amount", 99.99),
	)

	l.Info("Application flow simulation completed")
}

// performanceTest 性能测试
func performanceTest(l *Logger) {
	l.Info("Starting performance test")

	start := time.Now()
	count := 1000

	for i := 0; i < count; i++ {
		l.Infof("Performance test log entry %d", i)
	}

	duration := time.Since(start)
	l.Info("Performance test completed",
		zap.Int("log_count", count),
		zap.Duration("total_time", duration),
		zap.Float64("logs_per_second", float64(count)/duration.Seconds()),
	)
}

// 业务函数示例
func processUserRequest(userID string, l *Logger) error {
	requestLogger := l.With(
		zap.String("function", "processUserRequest"),
		zap.String("user_id", userID),
	)

	requestLogger.Info("Processing user request")

	// 模拟一些处理逻辑
	time.Sleep(time.Millisecond * 10)

	if userID == "invalid" {
		err := errors.New("invalid user ID")
		requestLogger.Error("Request processing failed",
			zap.Error(err),
		)
		return err
	}

	requestLogger.Info("Request processed successfully")
	return nil
}
