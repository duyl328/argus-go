//go:build windows

package system

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

// scanWindowsDevices 扫描 Windows 驱动器
func (dm *DeviceManager) scanWindowsDevices() error {
	// 检查从 A 到 Z 的所有可能驱动器字母
	for drive := 'A'; drive <= 'Z'; drive++ {
		drivePath := fmt.Sprintf("%c:\\", drive)

		// 检查驱动器是否存在
		if _, err := os.Stat(drivePath); err == nil {
			device := StorageDevice{
				ID:         fmt.Sprintf("drive_%c", drive),
				Name:       fmt.Sprintf("%c:", drive),
				MountPoint: drivePath,
				DevicePath: fmt.Sprintf("\\\\.\\%c:", drive),
				Type:       DeviceTypeUnknown,
				IsReady:    true,
			}

			// 获取驱动器类型
			device.Type = getWindowsDriveType(drivePath)

			// 获取卷标
			device.Label = getWindowsVolumeLabel(drivePath)

			// 获取文件系统类型
			device.FileSystem = getWindowsFileSystem(drivePath)

			// 获取空间信息
			if total, free, err := getDiskUsage(drivePath); err == nil {
				device.TotalSpace = total
				device.FreeSpace = free
				device.UsedSpace = total - free
			}

			// 判断是否为可移动设备
			device.IsRemovable = (device.Type == DeviceTypeRemovable)

			dm.addDevice(device)
		}
	}

	return nil
}

// getWindowsDriveType 获取 Windows 驱动器类型
func getWindowsDriveType(drivePath string) DeviceType {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getDriveType := kernel32.NewProc("GetDriveTypeW")

	pathPtr, err := syscall.UTF16PtrFromString(drivePath)
	if err != nil {
		return DeviceTypeUnknown
	}

	ret, _, _ := getDriveType.Call(uintptr(unsafe.Pointer(pathPtr)))

	switch ret {
	case 2: // DRIVE_REMOVABLE
		return DeviceTypeRemovable
	case 3: // DRIVE_FIXED
		return DeviceTypeFixed
	case 4: // DRIVE_REMOTE
		return DeviceTypeNetwork
	case 5: // DRIVE_CDROM
		return DeviceTypeCDROM
	case 6: // DRIVE_RAMDISK
		return DeviceTypeRAM
	default:
		return DeviceTypeUnknown
	}
}

// getWindowsVolumeLabel 获取 Windows 卷标
func getWindowsVolumeLabel(drivePath string) string {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getVolumeInfo := kernel32.NewProc("GetVolumeInformationW")

	pathPtr, err := syscall.UTF16PtrFromString(drivePath)
	if err != nil {
		return ""
	}

	volumeNameBuffer := make([]uint16, 256)
	var volumeSerialNumber uint32
	var maximumComponentLength uint32
	var fileSystemFlags uint32
	fileSystemNameBuffer := make([]uint16, 256)

	ret, _, _ := getVolumeInfo.Call(
		uintptr(unsafe.Pointer(pathPtr)),
		uintptr(unsafe.Pointer(&volumeNameBuffer[0])),
		uintptr(len(volumeNameBuffer)),
		uintptr(unsafe.Pointer(&volumeSerialNumber)),
		uintptr(unsafe.Pointer(&maximumComponentLength)),
		uintptr(unsafe.Pointer(&fileSystemFlags)),
		uintptr(unsafe.Pointer(&fileSystemNameBuffer[0])),
		uintptr(len(fileSystemNameBuffer)),
	)

	if ret != 0 {
		return syscall.UTF16ToString(volumeNameBuffer)
	}

	return ""
}

// getWindowsFileSystem 获取 Windows 文件系统类型
func getWindowsFileSystem(drivePath string) string {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getVolumeInfo := kernel32.NewProc("GetVolumeInformationW")

	pathPtr, err := syscall.UTF16PtrFromString(drivePath)
	if err != nil {
		return ""
	}

	volumeNameBuffer := make([]uint16, 256)
	var volumeSerialNumber uint32
	var maximumComponentLength uint32
	var fileSystemFlags uint32
	fileSystemNameBuffer := make([]uint16, 256)

	ret, _, _ := getVolumeInfo.Call(
		uintptr(unsafe.Pointer(pathPtr)),
		uintptr(unsafe.Pointer(&volumeNameBuffer[0])),
		uintptr(len(volumeNameBuffer)),
		uintptr(unsafe.Pointer(&volumeSerialNumber)),
		uintptr(unsafe.Pointer(&maximumComponentLength)),
		uintptr(unsafe.Pointer(&fileSystemFlags)),
		uintptr(unsafe.Pointer(&fileSystemNameBuffer[0])),
		uintptr(len(fileSystemNameBuffer)),
	)

	if ret != 0 {
		return syscall.UTF16ToString(fileSystemNameBuffer)
	}

	return ""
}

// getDiskUsage Windows 版本的磁盘使用情况获取
func getDiskUsage(path string) (total, free int64, err error) {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getDiskFreeSpaceEx := kernel32.NewProc("GetDiskFreeSpaceExW")

	var freeBytesToCaller, totalBytes, freeBytes int64

	pathPtr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return 0, 0, err
	}

	ret, _, errno := getDiskFreeSpaceEx.Call(
		uintptr(unsafe.Pointer(pathPtr)),
		uintptr(unsafe.Pointer(&freeBytesToCaller)),
		uintptr(unsafe.Pointer(&totalBytes)),
		uintptr(unsafe.Pointer(&freeBytes)),
	)

	if ret == 0 {
		return 0, 0, errno
	}

	return totalBytes, freeBytesToCaller, nil
}

// 在 Windows 上的占位符方法
func (dm *DeviceManager) scanLinuxDevices() error {
	return fmt.Errorf("Linux device scanning not supported on Windows")
}

func (dm *DeviceManager) scanMacOSDevices() error {
	return fmt.Errorf("macOS device scanning not supported on Windows")
}

func (dm *DeviceManager) scanUnixDevices() error {
	return fmt.Errorf("Unix device scanning not supported on Windows")
}