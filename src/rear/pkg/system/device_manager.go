package system

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"runtime"
	"time"
)

// DeviceType 设备类型枚举
type DeviceType string

const (
	DeviceTypeFixed     DeviceType = "fixed"     // 固定硬盘
	DeviceTypeRemovable DeviceType = "removable" // 可移动设备
	DeviceTypeNetwork   DeviceType = "network"   // 网络驱动器
	DeviceTypeCDROM     DeviceType = "cd-rom"    // 光驱
	DeviceTypeRAM       DeviceType = "ram"       // 内存盘
	DeviceTypeUnknown   DeviceType = "unknown"   // 未知类型
)

// StorageDevice 存储设备信息
type StorageDevice struct {
	ID          string     `json:"id"`           // 设备唯一标识
	Name        string     `json:"name"`         // 设备名称
	MountPoint  string     `json:"mount_point"`  // 挂载点/驱动器路径
	DevicePath  string     `json:"device_path"`  // 设备路径 (如 /dev/sda1, \\.\PhysicalDrive0)
	Type        DeviceType `json:"type"`         // 设备类型
	FileSystem  string     `json:"file_system"`  // 文件系统类型
	TotalSpace  int64      `json:"total_space"`  // 总空间 (字节)
	FreeSpace   int64      `json:"free_space"`   // 可用空间 (字节)
	UsedSpace   int64      `json:"used_space"`   // 已用空间 (字节)
	IsReady     bool       `json:"is_ready"`     // 是否可访问
	IsRemovable bool       `json:"is_removable"` // 是否可移动
	Label       string     `json:"label"`        // 卷标
	SerialNo    string     `json:"serial_no"`    // 序列号
	CreatedAt   time.Time  `json:"created_at"`   // 发现时间
	Metadata    map[string]interface{} `json:"metadata"` // 扩展元数据
}

// DeviceManager 设备管理器
type DeviceManager struct {
	devices []StorageDevice
	cache   map[string]*StorageDevice
}

// NewDeviceManager 创建设备管理器实例
func NewDeviceManager() *DeviceManager {
	return &DeviceManager{
		devices: make([]StorageDevice, 0),
		cache:   make(map[string]*StorageDevice),
	}
}

// ScanDevices 扫描所有存储设备
func (dm *DeviceManager) ScanDevices() error {
	dm.devices = dm.devices[:0] // 清空现有设备列表
	dm.cache = make(map[string]*StorageDevice)

	switch runtime.GOOS {
	case "windows":
		return dm.scanWindowsDevices()
	case "linux":
		return dm.scanLinuxDevices()
	case "darwin":
		return dm.scanMacOSDevices()
	default:
		return dm.scanUnixDevices()
	}
}

// 平台特定方法在各自的平台文件中实现：
// - device_manager_windows.go: scanWindowsDevices
// - device_manager_unix.go: scanLinuxDevices, scanMacOSDevices, scanUnixDevices

// GetDevices 获取所有设备列表
func (dm *DeviceManager) GetDevices() []StorageDevice {
	return dm.devices
}

// GetDevice 根据挂载点获取设备信息
func (dm *DeviceManager) GetDevice(mountPoint string) *StorageDevice {
	mountPoint = filepath.Clean(mountPoint)
	return dm.cache[mountPoint]
}

// GetDeviceByID 根据设备ID获取设备信息
func (dm *DeviceManager) GetDeviceByID(id string) *StorageDevice {
	for i := range dm.devices {
		if dm.devices[i].ID == id {
			return &dm.devices[i]
		}
	}
	return nil
}

// GetReadyDevices 获取所有可访问的设备
func (dm *DeviceManager) GetReadyDevices() []StorageDevice {
	var ready []StorageDevice
	for _, device := range dm.devices {
		if device.IsReady {
			ready = append(ready, device)
		}
	}
	return ready
}

// GetDevicesByType 根据类型筛选设备
func (dm *DeviceManager) GetDevicesByType(deviceType DeviceType) []StorageDevice {
	var filtered []StorageDevice
	for _, device := range dm.devices {
		if device.Type == deviceType {
			filtered = append(filtered, device)
		}
	}
	return filtered
}

// GetLargestFreeSpaceDevice 获取可用空间最大的设备
func (dm *DeviceManager) GetLargestFreeSpaceDevice() *StorageDevice {
	var largest *StorageDevice
	for i := range dm.devices {
		if dm.devices[i].IsReady && (largest == nil || dm.devices[i].FreeSpace > largest.FreeSpace) {
			largest = &dm.devices[i]
		}
	}
	return largest
}

// IsMountPoint 检查路径是否为挂载点
func (dm *DeviceManager) IsMountPoint(path string) bool {
	path = filepath.Clean(path)
	return dm.cache[path] != nil
}

// RefreshDevice 刷新指定设备的空间信息
func (dm *DeviceManager) RefreshDevice(mountPoint string) error {
	device := dm.GetDevice(mountPoint)
	if device == nil {
		return fmt.Errorf("设备不存在: %s", mountPoint)
	}

	total, free, err := getDiskUsage(mountPoint)
	if err != nil {
		return fmt.Errorf("获取磁盘使用情况失败: %w", err)
	}

	device.TotalSpace = total
	device.FreeSpace = free
	device.UsedSpace = total - free

	return nil
}

// ToJSON 将设备列表转换为JSON
func (dm *DeviceManager) ToJSON() ([]byte, error) {
	return json.MarshalIndent(dm.devices, "", "  ")
}

// addDevice 添加设备到管理器
func (dm *DeviceManager) addDevice(device StorageDevice) {
	device.CreatedAt = time.Now()
	dm.devices = append(dm.devices, device)
	dm.cache[device.MountPoint] = &dm.devices[len(dm.devices)-1]
}

// 默认的全局设备管理器实例
var DefaultDeviceManager = NewDeviceManager()

// 便捷函数，使用默认管理器
func ScanStorageDevices() error {
	return DefaultDeviceManager.ScanDevices()
}

func GetStorageDevices() []StorageDevice {
	return DefaultDeviceManager.GetDevices()
}

func GetStorageDevice(mountPoint string) *StorageDevice {
	return DefaultDeviceManager.GetDevice(mountPoint)
}

func GetReadyStorageDevices() []StorageDevice {
	return DefaultDeviceManager.GetReadyDevices()
}

func IsStorageMountPoint(path string) bool {
	return DefaultDeviceManager.IsMountPoint(path)
}