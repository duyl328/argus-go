package main

import (
	"context"
	"errors"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"os/signal"
	"rear/internal/config"
	"rear/internal/db"
	"rear/internal/model"
	"rear/internal/repositories"
	"rear/internal/router"
	"rear/internal/service"
	"rear/pkg/logger"
	"syscall"
	"time"
)

func main() {
	// 日志初始化
	err := logger.InitDefaultLogger()
	if err != nil {
		// log.Fatal 会输出错误信息并调用 os.Exit(1)
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	// 数据库初始化
	if err := db.InitDatabase(); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	// 自动迁移
	if err := db.AutoMigrate(); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// 初始化基础服务（启动写操作处理协程）
	repositories.InitBaseService()

	// 使用示例
	userService := repositories.NewUserService()

	// 创建用户
	user := &model.User{
		Email:    "john@example.com",
		Username: "John Doe",
		Age:      25,
	}

	if err := userService.CreateUser(user); err != nil {
		log.Printf("Failed to create user: %v", err)
	} else {
		log.Printf("User created successfully: %+v", user)
	}

	// 更新用户名称
	if err := userService.UpdateUserName(user.ID, "John Smith"); err != nil {
		log.Printf("Failed to update user name: %v", err)
	}

	// 查询用户（可以并发执行）
	if foundUser, err := userService.GetUserByID(user.ID); err != nil {
		log.Printf("Failed to get user: %v", err)
	} else if foundUser != nil {
		log.Printf("Found user: %+v", foundUser)
	}

	// 获取所有用户
	if users, err := userService.GetAllUsers(); err != nil {
		log.Printf("Failed to get all users: %v", err)
	} else {
		log.Printf("Total users: %d", len(users))
	}

	//if err := initDatabase(); err != nil {
	//	logger.Fatalf("Failed to initialize database: %v", err)
	//}
	// 依赖注入

	// 启动 http
	startHttp()
}

func startHttp() {
	// 配置加载
	netConfig := config.InitConfig()
	// 设置Gin模式
	gin.SetMode(netConfig.Mode)

	// 创建Gin引擎
	r := gin.New()

	// CORS 处理
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization", "X-Request-ID"}

	// 添加中间件
	r.Use(service.RequestIDMiddleware())    // 最先生成请求ID
	r.Use(service.LoggerMiddleware())       // 记录日志
	r.Use(gin.Recovery())                   // 恢复panic
	r.Use(cors.New(corsConfig))             // CORS 处理
	r.Use(service.ErrorHandlerMiddleware()) // 最后处理错误

	// 性能分析 (仅在debug模式下)
	if netConfig.Mode == "debug" {
		pprof.Register(r)
	}

	// 设置路由
	router.SetupRoutes(r)

	// 创建HTTP服务器
	srv := &http.Server{
		Addr:         ":" + netConfig.Port,
		Handler:      r,
		ReadTimeout:  netConfig.ReadTimeout,
		WriteTimeout: netConfig.WriteTimeout,
		IdleTimeout:  netConfig.IdleTimeout,
	}

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 启动服务器
	go func() {
		logger.Infof("Server starting on port 127.0.0.1:%s", netConfig.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatalf("Failed to start server: %v", err)
			// 发送信号给主goroutine，让它知道启动失败
			quit <- syscall.SIGTERM
		}
	}()

	//  阻塞主goroutine，等待信号
	<-quit
	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Info("Server exited")
}
