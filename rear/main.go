package main

import (
	"log"
	"rear/pkg/logger"
)

func main() {
	// 配置加载
	// 数据库初始化
	// 依赖注入
	// 日志初始化
	err := logger.InitDefaultLogger()
	if err != nil {
		// log.Fatal 会输出错误信息并调用 os.Exit(1)
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	//	设置路由
	//	启动服务
}
