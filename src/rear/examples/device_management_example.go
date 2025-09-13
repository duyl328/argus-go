package main

import (
	"fmt"
	"log"

	"rear/pkg/system"
)

func main() {
	fmt.Println("=== 设备管理系统示例 ===\n")

	// 1. 创建设备管理器实例
	fmt.Println("1. 创建设备管理器并扫描设备...")
	deviceManager := system.NewDeviceManager()

	err := deviceManager.ScanDevices()
	if err != nil {
		log.Fatalf("设备扫描失败: %v", err)
	}

	devices := deviceManager.GetDevices()
	fmt.Printf("✅ 成功扫描到 %d 个存储设备\n\n", len(devices))

	// 2. 显示所有设备详细信息
	fmt.Println("2. 设备详细信息:")
	for i, device := range devices {
		fmt.Printf("设备 %d:\n", i+1)
		fmt.Printf("  📱 名称: %s\n", device.Name)
		fmt.Printf("  📂 挂载点: %s\n", device.MountPoint)
		fmt.Printf("  🔧 设备路径: %s\n", device.DevicePath)
		fmt.Printf("  📊 类型: %s\n", device.Type)
		fmt.Printf("  💾 文件系统: %s\n", device.FileSystem)

		if device.Label != "" {
			fmt.Printf("  🏷️  卷标: %s\n", device.Label)
		}

		fmt.Printf("  ✅ 状态: %s\n", map[bool]string{true: "可访问", false: "不可访问"}[device.IsReady])
		fmt.Printf("  🔄 可移动: %s\n", map[bool]string{true: "是", false: "否"}[device.IsRemovable])

		if device.TotalSpace > 0 {
			fmt.Printf("  💿 总空间: %s\n", formatFileSize(device.TotalSpace))
			fmt.Printf("  🆓 可用空间: %s\n", formatFileSize(device.FreeSpace))
			fmt.Printf("  📈 已用空间: %s\n", formatFileSize(device.UsedSpace))
			fmt.Printf("  📊 使用率: %.1f%%\n", float64(device.UsedSpace)*100/float64(device.TotalSpace))
		}

		fmt.Printf("  🕐 发现时间: %s\n", device.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Println()
	}

	// 3. 按类型分类显示设备
	fmt.Println("3. 按类型分类的设备:")
	deviceTypes := []system.DeviceType{
		system.DeviceTypeFixed,
		system.DeviceTypeRemovable,
		system.DeviceTypeNetwork,
		system.DeviceTypeCDROM,
		system.DeviceTypeRAM,
	}

	for _, deviceType := range deviceTypes {
		typeDevices := deviceManager.GetDevicesByType(deviceType)
		if len(typeDevices) > 0 {
			fmt.Printf("  %s 设备: %d 个\n", getDeviceTypeIcon(deviceType), len(typeDevices))
			for _, device := range typeDevices {
				fmt.Printf("    - %s (%s)\n", device.Name, device.MountPoint)
			}
		}
	}
	fmt.Println()

	// 4. 智能设备推荐
	fmt.Println("4. 智能设备推荐:")

	// 可用空间最大的设备
	largestDevice := deviceManager.GetLargestFreeSpaceDevice()
	if largestDevice != nil {
		fmt.Printf("  💾 存储推荐: %s (%s)\n", largestDevice.Name, largestDevice.MountPoint)
		fmt.Printf("     可用空间: %s\n", formatFileSize(largestDevice.FreeSpace))
		fmt.Printf("     建议用途: 大文件存储、备份\n")
	}

	// 固定硬盘设备（用于系统文件）
	fixedDevices := deviceManager.GetDevicesByType(system.DeviceTypeFixed)
	if len(fixedDevices) > 0 {
		fmt.Printf("  🔧 系统推荐: %s (%s)\n", fixedDevices[0].Name, fixedDevices[0].MountPoint)
		fmt.Printf("     建议用途: 系统文件、应用程序\n")
	}

	// 可移动设备（用于临时存储）
	removableDevices := deviceManager.GetDevicesByType(system.DeviceTypeRemovable)
	if len(removableDevices) > 0 {
		fmt.Printf("  📱 便携推荐: %s (%s)\n", removableDevices[0].Name, removableDevices[0].MountPoint)
		fmt.Printf("     建议用途: 文件传输、临时备份\n")
	}
	fmt.Println()

	// 5. 设备管理操作演示
	fmt.Println("5. 设备管理操作:")
	if len(devices) > 0 {
		testDevice := devices[0]

		// 检查是否为挂载点
		isMountPoint := deviceManager.IsMountPoint(testDevice.MountPoint)
		fmt.Printf("  🔍 %s 是挂载点: %t\n", testDevice.MountPoint, isMountPoint)

		// 刷新设备信息
		err := deviceManager.RefreshDevice(testDevice.MountPoint)
		if err != nil {
			fmt.Printf("  🔄 刷新设备信息失败: %v\n", err)
		} else {
			fmt.Printf("  🔄 成功刷新设备 %s 的信息\n", testDevice.Name)
		}

		// 根据ID查找设备
		foundDevice := deviceManager.GetDeviceByID(testDevice.ID)
		if foundDevice != nil {
			fmt.Printf("  🎯 根据ID查找设备成功: %s -> %s\n", testDevice.ID, foundDevice.Name)
		}
	}
	fmt.Println()

	// 6. 使用全局设备管理器
	fmt.Println("6. 全局设备管理器:")
	err = system.ScanStorageDevices()
	if err != nil {
		fmt.Printf("  ❌ 全局扫描失败: %v\n", err)
	} else {
		globalDevices := system.GetStorageDevices()
		readyDevices := system.GetReadyStorageDevices()
		fmt.Printf("  ✅ 全局管理器: %d 个设备，%d 个可用\n", len(globalDevices), len(readyDevices))
	}

	// 7. JSON 导出示例
	fmt.Println("\n7. JSON 导出:")
	jsonData, err := deviceManager.ToJSON()
	if err != nil {
		fmt.Printf("  ❌ JSON 导出失败: %v\n", err)
	} else {
		fmt.Printf("  ✅ JSON 导出成功，数据长度: %d 字节\n", len(jsonData))
		fmt.Println("  JSON 预览:")
		// 只显示前200个字符
		preview := string(jsonData)
		if len(preview) > 200 {
			preview = preview[:200] + "..."
		}
		fmt.Printf("    %s\n", preview)
	}
}

// 辅助函数
func formatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func getDeviceTypeIcon(deviceType system.DeviceType) string {
	switch deviceType {
	case system.DeviceTypeFixed:
		return "🔧 固定硬盘"
	case system.DeviceTypeRemovable:
		return "📱 可移动设备"
	case system.DeviceTypeNetwork:
		return "🌐 网络驱动器"
	case system.DeviceTypeCDROM:
		return "💿 光驱"
	case system.DeviceTypeRAM:
		return "⚡ 内存盘"
	default:
		return "❓ 未知设备"
	}
}