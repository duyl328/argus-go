package main

import (
	"fmt"
	"log"
	"rear/internal/config"
	"rear/internal/container"
	"rear/internal/db"
	"rear/internal/workflow"
	"time"
)

// ImgTaskDatabaseExample 展示img_task如何保存数据到数据库
func ImgTaskDatabaseExample() {
	fmt.Println("=== 图像任务数据库保存示例 ===")

	// 1. 初始化配置
	config.InitConfig()

	// 2. 初始化数据库
	if err := db.InitDatabase(); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	// 3. 自动迁移
	if err := db.AutoMigrate(); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// 4. 创建数据库容器
	dbContainer := container.NewContainer()

	// 5. 创建任务管理器
	taskManager := workflow.NewImgTaskManager(2, dbContainer)

	// 6. 添加图像任务进行测试
	// 注意：这里需要实际存在的图像文件路径
	testImagePaths := []string{
		// 替换为你实际的图像文件路径
		// "D:/test/sample.jpg",
		// "D:/test/photo.png",
	}

	// 如果没有测试图像，输出使用说明
	if len(testImagePaths) == 0 {
		fmt.Println("\n使用说明:")
		fmt.Println("1. 请在 testImagePaths 中添加实际的图像文件路径")
		fmt.Println("2. 确保图像文件存在且可访问")
		fmt.Println("3. 重新运行示例")
		fmt.Println("\n示例路径格式:")
		fmt.Println(`testImagePaths := []string{`)
		fmt.Println(`    "D:/photos/IMG_1234.jpg",`)
		fmt.Println(`    "D:/photos/DSC_5678.png",`)
		fmt.Println(`}`)
		return
	}

	fmt.Printf("\n准备处理 %d 个图像文件...\n", len(testImagePaths))

	// 添加任务
	var taskIDs []string
	for i, imagePath := range testImagePaths {
		taskID := taskManager.AddTask(imagePath)
		taskIDs = append(taskIDs, taskID)
		fmt.Printf("%d. 添加任务: %s (ID: %s)\n", i+1, imagePath, taskID)
	}

	// 7. 监控任务进度
	fmt.Println("\n开始监控任务进度...")
	monitorTasks(taskManager, taskIDs)

	fmt.Println("\n=== 示例完成 ===")
}

// monitorTasks 监控任务进度
func monitorTasks(taskManager *workflow.ImgTaskManager, taskIDs []string) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	timeout := time.After(5 * time.Minute) // 5分钟超时

	for {
		select {
		case <-timeout:
			fmt.Println("监控超时，退出...")
			return
		case <-ticker.C:
			allCompleted := true
			
			fmt.Println("\n--- 任务状态更新 ---")
			
			for i, taskID := range taskIDs {
				detail, err := taskManager.GetTaskDetail(taskID)
				if err != nil {
					fmt.Printf("任务 %d: 获取详情失败 - %v\n", i+1, err)
					continue
				}

				status := detail.Status
				progress := detail.Progress
				currentStep := detail.StepInfo.StepName

				fmt.Printf("任务 %d: %s | 进度: %.1f%% | 当前步骤: %s\n", 
					i+1, status, progress*100, currentStep)

				if detail.Error != "" {
					fmt.Printf("        错误: %s\n", detail.Error)
				}

				if status != workflow.StatusDone && status != workflow.StatusFailed {
					allCompleted = false
				}
			}

			// 显示整体进度
			overallProgress := taskManager.GetOverallProgress()
			fmt.Printf("整体进度: %d/%d 完成 (%.1f%%)\n",
				overallProgress.CompletedTasks,
				overallProgress.TotalTasks,
				overallProgress.OverallProgress*100)

			if allCompleted {
				fmt.Println("\n所有任务已完成！")
				
				// 显示数据库保存结果
				showDatabaseResults(taskManager, taskIDs)
				return
			}
		}
	}
}

// showDatabaseResults 显示数据库保存结果
func showDatabaseResults(taskManager *workflow.ImgTaskManager, taskIDs []string) {
	fmt.Println("\n--- 数据库保存结果 ---")
	
	// 这里可以查询数据库验证数据是否保存成功
	for i, taskID := range taskIDs {
		detail, err := taskManager.GetTaskDetail(taskID)
		if err != nil {
			continue
		}

		fmt.Printf("任务 %d (%s):\n", i+1, taskID)
		fmt.Printf("  文件路径: %s\n", detail.Path)
		fmt.Printf("  文件哈希: %s\n", detail.Hash)
		fmt.Printf("  最终状态: %s\n", detail.Status)
		fmt.Printf("  处理时长: %s\n", detail.Duration)
		
		if detail.Status == workflow.StatusDone {
			fmt.Printf("  ✓ 数据库保存: 成功\n")
		} else if detail.Status == workflow.StatusFailed {
			fmt.Printf("  ✗ 数据库保存: 失败 - %s\n", detail.Error)
		}
		fmt.Println()
	}
}

// 使用说明示例
func UsageExample() {
	fmt.Println("=== 使用说明 ===")
	fmt.Println()
	fmt.Println("这个示例展示如何使用img_task工作流保存图像数据到数据库。")
	fmt.Println()
	fmt.Println("主要步骤:")
	fmt.Println("1. 初始化配置和数据库")
	fmt.Println("2. 创建数据库容器(包含所有Repository)")
	fmt.Println("3. 创建任务管理器(传入数据库容器)")
	fmt.Println("4. 添加图像任务")
	fmt.Println("5. 监控任务进度")
	fmt.Println("6. 验证数据库保存结果")
	fmt.Println()
	fmt.Println("数据库表:")
	fmt.Println("- photos: 保存图像基础信息(路径、哈希、尺寸等)")
	fmt.Println("- photo_exif: 保存EXIF元数据信息")
	fmt.Println()
	fmt.Println("处理流程:")
	fmt.Println("1. 验证格式 → 2. 读取文件 → 3. 计算哈希")
	fmt.Println("4. 提取EXIF → 5. 格式转换 → 6. 生成缩略图")
	fmt.Println("7. 保存数据库 → 8. 智能分析 → 9. 完成")
}

func main() {
	UsageExample()
	fmt.Println()
	ImgTaskDatabaseExample()
}