package main

import (
	"fmt"
	"rear/internal/config"
)

// 示例：展示如何使用开发阶段标识

func exampleDevelopmentCheck() {
	// 方法1：使用 IsDevelopment() 函数
	if config.IsDevelopment() {
		fmt.Println("当前处于开发阶段")
		// 执行开发阶段特有的代码
		enableDebugFeatures()
		loadTestData()
	} else {
		fmt.Println("当前处于生产阶段")
		// 执行生产阶段的代码
		optimizeForProduction()
	}

	// 方法2：使用 IsProduction() 函数
	if config.IsProduction() {
		fmt.Println("生产环境配置启用")
		// 生产环境特有的配置
	}
}

func enableDebugFeatures() {
	fmt.Println("- 启用调试功能")
	fmt.Println("- 启用详细日志")
	fmt.Println("- 启用性能监控")
}

func loadTestData() {
	fmt.Println("- 加载测试数据")
	fmt.Println("- 启用模拟接口")
}

func optimizeForProduction() {
	fmt.Println("- 启用性能优化")
	fmt.Println("- 禁用调试信息")
	fmt.Println("- 启用缓存")
}
