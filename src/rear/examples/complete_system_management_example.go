package main

import (
	"fmt"
	"log"
	"time"

	"rear/pkg/system"
)

func main() {
	fmt.Println("=== 完整系统管理示例 ===\n")

	// 1. 系统信息获取
	fmt.Println("1. 📊 系统信息:")
	systemInfo, err := system.GetSystemInfo()
	if err != nil {
		log.Printf("获取系统信息失败: %v", err)
	} else {
		fmt.Print(systemInfo.FormatSystemInfo())
	}
	fmt.Println()

	// 2. 设备管理
	fmt.Println("2. 💾 设备管理:")
	deviceManager := system.NewDeviceManager()
	err = deviceManager.ScanDevices()
	if err != nil {
		log.Printf("设备扫描失败: %v", err)
	} else {
		devices := deviceManager.GetDevices()
		fmt.Printf("发现 %d 个存储设备:\n", len(devices))
		for i, device := range devices {
			fmt.Printf("  [%d] %s: %s 可用空间, %s 文件系统\n",
				i+1, device.Name,
				formatFileSize(device.FreeSpace),
				device.FileSystem)
		}
	}
	fmt.Println()

	// 3. 进程管理
	fmt.Println("3. ⚙️ 进程管理:")
	currentProcess := system.GetCurrentProcess()
	fmt.Print(currentProcess.FormatProcessInfo())
	fmt.Println()

	// 4. 网络管理
	fmt.Println("4. 🌐 网络管理:")
	networkManager := system.NewNetworkManager()
	err = networkManager.ScanInterfaces()
	if err != nil {
		log.Printf("网络扫描失败: %v", err)
	} else {
		interfaces := networkManager.GetActiveInterfaces()
		fmt.Printf("发现 %d 个活动网络接口:\n", len(interfaces))
		for _, iface := range interfaces {
			fmt.Printf("  - %s: %v\n", iface.Name, iface.IPAddresses)
		}

		// 测试网络连通性
		fmt.Println("\n网络连通性测试:")
		testHosts := []string{"google.com:80", "github.com:443", "microsoft.com:80"}
		for _, host := range testHosts {
			err := system.PingHost(host, 3*time.Second)
			status := "❌ 失败"
			if err == nil {
				status = "✅ 成功"
			}
			fmt.Printf("  %s: %s\n", host, status)
		}
	}
	fmt.Println()

	// 5. 存储分析和推荐
	fmt.Println("5. 📈 存储分析:")
	storageAnalyzer := system.NewStorageAnalyzer()

	// 分析当前目录
	currentDir := "."
	dirInfo, err := storageAnalyzer.AnalyzeDirectoryStorage(currentDir)
	if err != nil {
		log.Printf("目录分析失败: %v", err)
	} else {
		fmt.Printf("当前目录分析:\n")
		fmt.Printf("  路径: %s\n", dirInfo.Path)
		fmt.Printf("  大小: %s\n", formatFileSize(dirInfo.Size))
		fmt.Printf("  文件数: %d\n", dirInfo.FileCount)
		if dirInfo.Device != nil {
			fmt.Printf("  所在设备: %s\n", dirInfo.Device.Name)
			fmt.Printf("  占设备空间: %.2f%%\n", dirInfo.SizePercent)
		}
	}
	fmt.Println()

	// 6. 存储推荐
	fmt.Println("6. 💡 存储推荐:")
	requiredSpace := int64(1024 * 1024 * 1024) // 1GB
	recommendation, err := storageAnalyzer.RecommendStorageLocation(requiredSpace)
	if err != nil {
		log.Printf("存储推荐失败: %v", err)
	} else {
		fmt.Print(recommendation.FormatStorageRecommendation())
	}
	fmt.Println()

	// 7. 清理建议
	fmt.Println("7. 🧹 清理建议:")
	if len(deviceManager.GetDevices()) > 0 {
		device := deviceManager.GetDevices()[0]
		cleanupInfo, err := storageAnalyzer.CleanupStorage(device.MountPoint)
		if err != nil {
			log.Printf("清理分析失败: %v", err)
		} else {
			fmt.Printf("设备 %s 清理建议:\n", device.Name)
			for i, suggestion := range cleanupInfo.Suggestions {
				fmt.Printf("  %d. %s\n", i+1, suggestion)
			}
			if cleanupInfo.CanFreeSpace > 0 {
				fmt.Printf("预计可释放空间: %s\n", formatFileSize(cleanupInfo.CanFreeSpace))
			}
		}
	}
	fmt.Println()

	// 8. 权限检查
	fmt.Println("8. 🔐 权限检查:")
	isAdmin := system.IsAdmin()
	fmt.Printf("当前用户是否为管理员: %s\n", map[bool]string{true: "是", false: "否"}[isAdmin])
	fmt.Printf("当前用户: %s\n", system.GetCurrentUser())
	fmt.Printf("进程ID: %d\n", system.GetProcessID())
	fmt.Printf("父进程ID: %d\n", system.GetParentProcessID())
	fmt.Println()

	// 9. 系统路径信息
	fmt.Println("9. 📁 系统路径:")
	if execPath, err := system.GetExecutablePath(); err == nil {
		fmt.Printf("可执行文件路径: %s\n", execPath)
	}
	fmt.Printf("临时目录: %s\n", system.GetTempDir())
	if homeDir := system.GetHomeDir(); homeDir != "" {
		fmt.Printf("用户主目录: %s\n", homeDir)
	}
	fmt.Println()

	// 10. 实用功能演示
	fmt.Println("10. 🛠️ 实用功能演示:")

	// MAC 地址获取
	if macAddr, err := system.GetMACAddress(); err == nil {
		fmt.Printf("主网卡MAC地址: %s\n", macAddr)
	}

	// 本地IP获取
	if localIP, err := system.GetLocalIP(); err == nil {
		fmt.Printf("本机IP地址: %s\n", localIP)
	}

	// 主机名解析
	if ips, err := system.ResolveHostname("google.com"); err == nil && len(ips) > 0 {
		fmt.Printf("google.com 解析结果: %s\n", ips[0])
	}

	// 端口扫描演示 (扫描本地常用端口)
	fmt.Println("\n本地端口扫描 (常用端口):")
	commonPorts := []int{22, 80, 443, 3306, 5432, 6379, 8080, 9200}
	openPorts := []int{}
	for _, port := range commonPorts {
		if system.IsPortOpen("localhost", port, 100*time.Millisecond) {
			openPorts = append(openPorts, port)
		}
	}
	if len(openPorts) > 0 {
		fmt.Printf("开放端口: %v\n", openPorts)
	} else {
		fmt.Println("未发现开放的常用端口")
	}

	fmt.Println("\n=== 系统管理演示完成 ===")
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