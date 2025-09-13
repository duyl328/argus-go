package main

import (
	"encoding/json"
	"fmt"
	"log"

	"rear/pkg/utils"
)

func main() {
	fmt.Println("=== 磁盘驱动器/挂载点发现示例 ===\n")

	// 1. 获取所有驱动器/挂载点
	fmt.Println("1. 获取所有可用的驱动器/挂载点:")
	drives, err := utils.FileUtils.GetDrives()
	if err != nil {
		log.Fatalf("获取驱动器失败: %v", err)
	}

	fmt.Printf("找到 %d 个驱动器/挂载点:\n", len(drives))
	for i, drive := range drives {
		fmt.Printf("  [%d] %s (%s)\n", i+1, drive.Name, drive.Path)
		fmt.Printf("      类型: %s, 文件系统: %s\n", drive.Type, drive.FileSystem)
		fmt.Printf("      总空间: %s, 可用空间: %s, 已用空间: %s\n",
			utils.FileUtils.FormatFileSize(drive.TotalSpace),
			utils.FileUtils.FormatFileSize(drive.FreeSpace),
			utils.FileUtils.FormatFileSize(drive.UsedSpace))
		if drive.Description != "" {
			fmt.Printf("      描述: %s\n", drive.Description)
		}
		fmt.Printf("      状态: 可访问=%t\n", drive.IsReady)
		fmt.Println()
	}

	// 2. 获取指定驱动器的详细信息
	if len(drives) > 0 {
		testDrive := drives[0].Path
		fmt.Printf("2. 获取驱动器 %s 的详细信息:\n", testDrive)

		driveInfo, err := utils.FileUtils.GetDriveInfo(testDrive)
		if err != nil {
			fmt.Printf("   错误: %v\n", err)
		} else {
			fmt.Printf("   名称: %s\n", driveInfo.Name)
			fmt.Printf("   路径: %s\n", driveInfo.Path)
			fmt.Printf("   总空间: %s\n", utils.FileUtils.FormatFileSize(driveInfo.TotalSpace))
			fmt.Printf("   可用空间: %s\n", utils.FileUtils.FormatFileSize(driveInfo.FreeSpace))
			fmt.Printf("   已用空间: %s\n", utils.FileUtils.FormatFileSize(driveInfo.UsedSpace))
			fmt.Printf("   使用率: %.1f%%\n", float64(driveInfo.UsedSpace)*100/float64(driveInfo.TotalSpace))
		}
		fmt.Println()
	}

	// 3. 检查路径是否为驱动器根目录
	fmt.Println("3. 检查路径是否为驱动器根目录:")
	testPaths := []string{
		"C:\\",
		"D:\\",
		"C:\\Windows",
		"D:\\temp",
		"/",
		"/home",
		"/tmp",
	}

	for _, path := range testPaths {
		if utils.FileUtils.Exists(path) {
			isDriveRoot := utils.FileUtils.IsDriveRoot(path)
			fmt.Printf("   %s -> %s\n", path, map[bool]string{true: "是驱动器根目录", false: "不是驱动器根目录"}[isDriveRoot])
		}
	}
	fmt.Println()

	// 4. JSON 格式输出
	fmt.Println("4. JSON 格式的驱动器信息:")
	jsonData, err := json.MarshalIndent(drives, "", "  ")
	if err != nil {
		fmt.Printf("   JSON 序列化失败: %v\n", err)
	} else {
		fmt.Printf("%s\n", string(jsonData))
	}

	// 5. 使用场景示例：选择可用空间最大的驱动器
	fmt.Println("5. 实用功能：找到可用空间最大的驱动器")
	var maxFreeDrive *utils.DriveInfo
	for i := range drives {
		if drives[i].IsReady && (maxFreeDrive == nil || drives[i].FreeSpace > maxFreeDrive.FreeSpace) {
			maxFreeDrive = &drives[i]
		}
	}

	if maxFreeDrive != nil {
		fmt.Printf("   可用空间最大的驱动器: %s\n", maxFreeDrive.Path)
		fmt.Printf("   可用空间: %s\n", utils.FileUtils.FormatFileSize(maxFreeDrive.FreeSpace))
		fmt.Printf("   建议用于存储大文件或备份\n")
	}
}