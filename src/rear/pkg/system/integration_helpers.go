package system

import (
	"fmt"
	"path/filepath"
	"rear/pkg/utils"
)

// StorageAnalyzer 存储分析器 - 连接设备管理和文件管理
type StorageAnalyzer struct {
	deviceManager *DeviceManager
}

// NewStorageAnalyzer 创建存储分析器
func NewStorageAnalyzer() *StorageAnalyzer {
	return &StorageAnalyzer{
		deviceManager: NewDeviceManager(),
	}
}

// AnalyzeDirectoryStorage 分析目录存储使用情况
func (sa *StorageAnalyzer) AnalyzeDirectoryStorage(dirPath string) (*DirectoryStorageInfo, error) {
	// 扫描设备
	err := sa.deviceManager.ScanDevices()
	if err != nil {
		return nil, fmt.Errorf("扫描设备失败: %w", err)
	}

	// 检查目录是否存在
	if !utils.FileUtils.Exists(dirPath) {
		return nil, fmt.Errorf("目录不存在: %s", dirPath)
	}

	if !utils.FileUtils.IsDir(dirPath) {
		return nil, fmt.Errorf("路径不是目录: %s", dirPath)
	}

	// 获取目录所在的设备
	device := sa.findDeviceForPath(dirPath)

	// 获取目录大小
	dirSize, err := utils.FileUtils.GetDirSize(dirPath)
	if err != nil {
		return nil, fmt.Errorf("获取目录大小失败: %w", err)
	}

	// 获取文件统计
	files, err := utils.FileUtils.GetAllFiles(dirPath, true)
	if err != nil {
		return nil, fmt.Errorf("获取文件列表失败: %w", err)
	}

	info := &DirectoryStorageInfo{
		Path:        dirPath,
		Size:        dirSize,
		FileCount:   len(files),
		Device:      device,
		SizePercent: 0,
	}

	// 计算在设备中的占用百分比
	if device != nil && device.TotalSpace > 0 {
		info.SizePercent = float64(dirSize) * 100 / float64(device.TotalSpace)
	}

	return info, nil
}

// findDeviceForPath 查找路径所在的设备
func (sa *StorageAnalyzer) findDeviceForPath(path string) *StorageDevice {
	path = filepath.Clean(path)
	devices := sa.deviceManager.GetDevices()

	var bestMatch *StorageDevice
	var bestMatchLen int

	for i := range devices {
		device := &devices[i]
		if device.IsReady {
			mountPoint := filepath.Clean(device.MountPoint)
			// 检查路径是否在这个挂载点下
			if len(mountPoint) <= len(path) && path[:len(mountPoint)] == mountPoint {
				if len(mountPoint) > bestMatchLen {
					bestMatch = device
					bestMatchLen = len(mountPoint)
				}
			}
		}
	}

	return bestMatch
}

// DirectoryStorageInfo 目录存储信息
type DirectoryStorageInfo struct {
	Path        string         `json:"path"`         // 目录路径
	Size        int64          `json:"size"`         // 目录大小
	FileCount   int            `json:"file_count"`   // 文件数量
	Device      *StorageDevice `json:"device"`       // 所在设备
	SizePercent float64        `json:"size_percent"` // 在设备中的占用百分比
}

// RecommendStorageLocation 推荐存储位置
func (sa *StorageAnalyzer) RecommendStorageLocation(requiredSpace int64) (*StorageRecommendation, error) {
	err := sa.deviceManager.ScanDevices()
	if err != nil {
		return nil, fmt.Errorf("扫描设备失败: %w", err)
	}

	devices := sa.deviceManager.GetReadyDevices()
	if len(devices) == 0 {
		return nil, fmt.Errorf("没有可用的存储设备")
	}

	var recommendations []DeviceRecommendation

	for _, device := range devices {
		if device.FreeSpace >= requiredSpace {
			score := calculateStorageScore(&device, requiredSpace)
			recommendation := DeviceRecommendation{
				Device:     &device,
				Score:      score,
				Reason:     generateRecommendationReason(&device, requiredSpace),
				Suitable:   true,
				FreeSpace:  device.FreeSpace,
				UsageRate:  float64(device.UsedSpace) * 100 / float64(device.TotalSpace),
			}
			recommendations = append(recommendations, recommendation)
		}
	}

	if len(recommendations) == 0 {
		return nil, fmt.Errorf("没有足够空间的存储设备")
	}

	// 按评分排序，找到最佳推荐
	bestRecommendation := recommendations[0]
	for _, rec := range recommendations[1:] {
		if rec.Score > bestRecommendation.Score {
			bestRecommendation = rec
		}
	}

	return &StorageRecommendation{
		BestChoice:      bestRecommendation,
		AllOptions:      recommendations,
		RequiredSpace:   requiredSpace,
		RequiredSpaceStr: utils.FileUtils.FormatFileSize(requiredSpace),
	}, nil
}

// calculateStorageScore 计算存储设备评分
func calculateStorageScore(device *StorageDevice, requiredSpace int64) float64 {
	score := 0.0

	// 可用空间越多，分数越高
	if device.TotalSpace > 0 {
		freeSpaceRatio := float64(device.FreeSpace) / float64(device.TotalSpace)
		score += freeSpaceRatio * 40 // 最多40分
	}

	// 设备类型评分
	switch device.Type {
	case DeviceTypeFixed:
		score += 30 // 固定硬盘30分
	case DeviceTypeRemovable:
		score += 20 // 可移动设备20分
	case DeviceTypeNetwork:
		score += 10 // 网络驱动器10分
	default:
		score += 15 // 其他15分
	}

	// 剩余空间够用程度
	if device.FreeSpace > requiredSpace*10 {
		score += 20 // 剩余空间很充足
	} else if device.FreeSpace > requiredSpace*3 {
		score += 15 // 剩余空间充足
	} else if device.FreeSpace > requiredSpace*2 {
		score += 10 // 剩余空间适中
	} else {
		score += 5 // 剩余空间紧张
	}

	// 文件系统类型评分
	switch device.FileSystem {
	case "NTFS", "ext4", "apfs":
		score += 10 // 现代文件系统
	case "FAT32", "exFAT":
		score += 5 // 兼容性文件系统
	default:
		score += 7 // 其他文件系统
	}

	return score
}

// generateRecommendationReason 生成推荐理由
func generateRecommendationReason(device *StorageDevice, requiredSpace int64) string {
	reasons := []string{}

	usageRate := float64(device.UsedSpace) * 100 / float64(device.TotalSpace)

	if device.FreeSpace > requiredSpace*10 {
		reasons = append(reasons, "剩余空间非常充足")
	} else if device.FreeSpace > requiredSpace*3 {
		reasons = append(reasons, "剩余空间充足")
	}

	if usageRate < 50 {
		reasons = append(reasons, "使用率较低")
	} else if usageRate < 80 {
		reasons = append(reasons, "使用率适中")
	}

	switch device.Type {
	case DeviceTypeFixed:
		reasons = append(reasons, "固定硬盘，访问速度快")
	case DeviceTypeRemovable:
		reasons = append(reasons, "可移动设备，便于传输")
	case DeviceTypeNetwork:
		reasons = append(reasons, "网络驱动器，可远程访问")
	}

	if len(reasons) == 0 {
		return "可用的存储选项"
	}

	result := reasons[0]
	for i := 1; i < len(reasons); i++ {
		result += "，" + reasons[i]
	}

	return result
}

// StorageRecommendation 存储推荐结果
type StorageRecommendation struct {
	BestChoice       DeviceRecommendation   `json:"best_choice"`        // 最佳选择
	AllOptions       []DeviceRecommendation `json:"all_options"`        // 所有选项
	RequiredSpace    int64                  `json:"required_space"`     // 所需空间
	RequiredSpaceStr string                 `json:"required_space_str"` // 所需空间(格式化)
}

// DeviceRecommendation 设备推荐信息
type DeviceRecommendation struct {
	Device    *StorageDevice `json:"device"`     // 设备信息
	Score     float64        `json:"score"`      // 评分
	Reason    string         `json:"reason"`     // 推荐理由
	Suitable  bool           `json:"suitable"`   // 是否适合
	FreeSpace int64          `json:"free_space"` // 可用空间
	UsageRate float64        `json:"usage_rate"` // 使用率
}

// CleanupStorage 存储清理建议
func (sa *StorageAnalyzer) CleanupStorage(devicePath string) (*StorageCleanupInfo, error) {
	device := sa.deviceManager.GetDevice(devicePath)
	if device == nil {
		return nil, fmt.Errorf("设备不存在: %s", devicePath)
	}

	info := &StorageCleanupInfo{
		Device:       device,
		Suggestions:  []string{},
		CanFreeSpace: 0,
	}

	usageRate := float64(device.UsedSpace) * 100 / float64(device.TotalSpace)

	if usageRate > 90 {
		info.Suggestions = append(info.Suggestions, "设备空间严重不足，建议立即清理")
		info.CanFreeSpace = device.TotalSpace / 10 // 估计可释放10%空间
	} else if usageRate > 80 {
		info.Suggestions = append(info.Suggestions, "设备空间不足，建议适当清理")
		info.CanFreeSpace = device.TotalSpace / 20 // 估计可释放5%空间
	} else if usageRate > 70 {
		info.Suggestions = append(info.Suggestions, "设备空间使用率较高，可考虑清理")
		info.CanFreeSpace = device.TotalSpace / 50 // 估计可释放2%空间
	} else {
		info.Suggestions = append(info.Suggestions, "设备空间充足，无需立即清理")
	}

	// 添加通用清理建议
	info.Suggestions = append(info.Suggestions, "清理临时文件和缓存")
	info.Suggestions = append(info.Suggestions, "删除不需要的大文件")
	info.Suggestions = append(info.Suggestions, "使用磁盘清理工具")

	return info, nil
}

// StorageCleanupInfo 存储清理信息
type StorageCleanupInfo struct {
	Device       *StorageDevice `json:"device"`         // 设备信息
	Suggestions  []string       `json:"suggestions"`    // 清理建议
	CanFreeSpace int64          `json:"can_free_space"` // 预计可释放空间
}

// FormatStorageRecommendation 格式化存储推荐
func (sr *StorageRecommendation) FormatStorageRecommendation() string {
	result := fmt.Sprintf("=== 存储推荐 ===\n")
	result += fmt.Sprintf("所需空间: %s\n\n", sr.RequiredSpaceStr)

	result += fmt.Sprintf("🏆 最佳选择: %s (%s)\n", sr.BestChoice.Device.Name, sr.BestChoice.Device.MountPoint)
	result += fmt.Sprintf("   评分: %.1f/100\n", sr.BestChoice.Score)
	result += fmt.Sprintf("   可用空间: %s\n", utils.FileUtils.FormatFileSize(sr.BestChoice.FreeSpace))
	result += fmt.Sprintf("   使用率: %.1f%%\n", sr.BestChoice.UsageRate)
	result += fmt.Sprintf("   推荐理由: %s\n\n", sr.BestChoice.Reason)

	if len(sr.AllOptions) > 1 {
		result += "其他可选项:\n"
		for i, option := range sr.AllOptions {
			if i == 0 {
				continue // 跳过最佳选择
			}
			result += fmt.Sprintf("  %d. %s - 评分: %.1f, 可用: %s\n",
				i+1, option.Device.Name, option.Score, utils.FileUtils.FormatFileSize(option.FreeSpace))
		}
	}

	return result
}

// 全局便捷函数
var DefaultStorageAnalyzer = NewStorageAnalyzer()

// AnalyzeDirectory 分析目录存储 (使用默认分析器)
func AnalyzeDirectory(dirPath string) (*DirectoryStorageInfo, error) {
	return DefaultStorageAnalyzer.AnalyzeDirectoryStorage(dirPath)
}

// RecommendStorage 推荐存储位置 (使用默认分析器)
func RecommendStorage(requiredSpace int64) (*StorageRecommendation, error) {
	return DefaultStorageAnalyzer.RecommendStorageLocation(requiredSpace)
}

// SuggestCleanup 建议清理 (使用默认分析器)
func SuggestCleanup(devicePath string) (*StorageCleanupInfo, error) {
	return DefaultStorageAnalyzer.CleanupStorage(devicePath)
}