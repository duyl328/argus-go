//go:build !windows

package system

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
)

// scanLinuxDevices 扫描 Linux 设备
func (dm *DeviceManager) scanLinuxDevices() error {
	return dm.parseUnixMounts("/proc/mounts")
}

// scanMacOSDevices 扫描 macOS 设备
func (dm *DeviceManager) scanMacOSDevices() error {
	// 检查根目录
	if err := dm.addUnixMountPoint("/", "/dev/root", "rootfs"); err == nil {
		// 继续扫描
	}

	// 检查 /Volumes 下的挂载点
	volumesPath := "/Volumes"
	if isDir(volumesPath) {
		if dirs, err := getDirectories(volumesPath); err == nil {
			for _, dir := range dirs {
				mountPath := filepath.Join(volumesPath, dir)
				dm.addUnixMountPoint(mountPath, fmt.Sprintf("/dev/disk_%s", dir), "external")
			}
		}
	}

	return nil
}

// scanUnixDevices 扫描通用 Unix 设备
func (dm *DeviceManager) scanUnixDevices() error {
	// 尝试不同的挂载信息文件
	mountFiles := []string{"/proc/mounts", "/etc/mtab", "/etc/fstab"}

	for _, mountFile := range mountFiles {
		if _, err := os.Stat(mountFile); err == nil {
			return dm.parseUnixMounts(mountFile)
		}
	}

	// 如果都不存在，至少添加根目录
	return dm.addUnixMountPoint("/", "/dev/root", "rootfs")
}

// parseUnixMounts 解析 Unix 系统的挂载文件
func (dm *DeviceManager) parseUnixMounts(mountFile string) error {
	file, err := os.Open(mountFile)
	if err != nil {
		return fmt.Errorf("打开挂载文件失败: %w", err)
	}
	defer file.Close()

	seen := make(map[string]bool) // 避免重复挂载点
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}

		devicePath := fields[0]
		mountPoint := fields[1]
		fsType := fields[2]

		// 跳过特殊的文件系统类型
		if isSpecialFileSystem(fsType) {
			continue
		}

		// 避免重复
		if seen[mountPoint] {
			continue
		}
		seen[mountPoint] = true

		// 检查挂载点是否可访问
		if !isDir(mountPoint) {
			continue
		}

		if err := dm.addUnixMountPoint(mountPoint, devicePath, fsType); err != nil {
			// 记录错误但继续处理其他挂载点
			continue
		}
	}

	return scanner.Err()
}

// addUnixMountPoint 添加 Unix 挂载点
func (dm *DeviceManager) addUnixMountPoint(mountPoint, devicePath, fsType string) error {
	device := StorageDevice{
		ID:         generateDeviceID(devicePath, mountPoint),
		Name:       getDeviceName(mountPoint, devicePath),
		MountPoint: mountPoint,
		DevicePath: devicePath,
		Type:       classifyUnixFileSystemType(fsType),
		FileSystem: fsType,
		IsReady:    isDir(mountPoint),
		Label:      getUnixVolumeLabel(mountPoint),
	}

	// 判断是否为可移动设备
	device.IsRemovable = isRemovableDevice(devicePath)

	// 获取磁盘使用情况
	if total, free, err := getDiskUsage(mountPoint); err == nil {
		device.TotalSpace = total
		device.FreeSpace = free
		device.UsedSpace = total - free
	}

	dm.addDevice(device)
	return nil
}

// getDiskUsage Unix 版本的磁盘使用情况获取
func getDiskUsage(path string) (total, free int64, err error) {
	var stat syscall.Statfs_t
	err = syscall.Statfs(path, &stat)
	if err != nil {
		return 0, 0, err
	}

	// 计算总空间和可用空间
	total = int64(stat.Blocks) * int64(stat.Bsize)
	free = int64(stat.Bavail) * int64(stat.Bsize)

	return total, free, nil
}

// 辅助函数
func generateDeviceID(devicePath, mountPoint string) string {
	if devicePath != "" {
		return fmt.Sprintf("dev_%s", strings.ReplaceAll(strings.TrimPrefix(devicePath, "/dev/"), "/", "_"))
	}
	return fmt.Sprintf("mount_%s", strings.ReplaceAll(strings.TrimPrefix(mountPoint, "/"), "/", "_"))
}

func getDeviceName(mountPoint, devicePath string) string {
	if mountPoint == "/" {
		return "Root"
	}
	name := filepath.Base(mountPoint)
	if name == "." || name == "" {
		return filepath.Base(devicePath)
	}
	return name
}

func getUnixVolumeLabel(mountPoint string) string {
	// 尝试从不同位置获取卷标
	labelPaths := []string{
		filepath.Join("/media", filepath.Base(mountPoint)),
		filepath.Join("/mnt", filepath.Base(mountPoint)),
	}

	for _, labelPath := range labelPaths {
		if isDir(labelPath) {
			return filepath.Base(labelPath)
		}
	}

	return ""
}

func isRemovableDevice(devicePath string) bool {
	// 简单的启发式判断
	removablePatterns := []string{
		"/dev/sd", "/dev/mmcblk", "/dev/nvme",
		"usb", "removable", "external",
	}

	deviceLower := strings.ToLower(devicePath)
	for _, pattern := range removablePatterns {
		if strings.Contains(deviceLower, pattern) {
			return true
		}
	}

	return false
}

func isSpecialFileSystem(fsType string) bool {
	specialFS := []string{
		"proc", "sysfs", "devfs", "tmpfs", "devtmpfs", "cgroup", "cgroup2",
		"securityfs", "debugfs", "tracefs", "fusectl", "pstore", "bpf",
		"hugetlbfs", "mqueue", "configfs", "selinuxfs", "rpc_pipefs",
		"binfmt_misc", "autofs", "overlayfs", "squashfs",
	}

	for _, special := range specialFS {
		if fsType == special {
			return true
		}
	}
	return false
}

func classifyUnixFileSystemType(fsType string) DeviceType {
	switch strings.ToLower(fsType) {
	case "ext2", "ext3", "ext4", "xfs", "btrfs", "zfs":
		return DeviceTypeFixed
	case "apfs", "hfs", "hfs+":
		return DeviceTypeFixed
	case "iso9660", "udf":
		return DeviceTypeCDROM
	case "tmpfs", "ramfs":
		return DeviceTypeRAM
	case "nfs", "nfs4", "cifs", "smb", "ftp":
		return DeviceTypeNetwork
	case "vfat", "msdos", "exfat":
		return DeviceTypeRemovable
	default:
		return DeviceTypeFixed
	}
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func getDirectories(dirPath string) ([]string, error) {
	var dirs []string
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry.Name())
		}
	}

	return dirs, nil
}

// 在 Unix 系统上的占位符方法
func (dm *DeviceManager) scanWindowsDevices() error {
	return fmt.Errorf("Windows device scanning not supported on Unix systems")
}