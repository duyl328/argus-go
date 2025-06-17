//go:build windows
// +build windows

package utils

import (
	"fmt"
	"os/exec"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

// getCPUInfoPlatform 获取Windows下的CPU信息
func (s *SysUtils) getCPUInfoPlatform(info *CPUInfo) {
	cmd := exec.Command("wmic", "cpu", "get", "name", "/value")
	output, err := cmd.Output()
	if err != nil {
		return
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Name=") {
			info.ModelName = strings.TrimSpace(strings.TrimPrefix(line, "Name="))
			break
		}
	}
}

// getMemoryInfoPlatform 获取Windows下的内存信息
func (s *SysUtils) getMemoryInfoPlatform(info *MemoryInfo) {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	globalMemoryStatusEx := kernel32.NewProc("GlobalMemoryStatusEx")

	type memoryStatusEx struct {
		dwLength                uint32
		dwMemoryLoad            uint32
		ullTotalPhys            uint64
		ullAvailPhys            uint64
		ullTotalPageFile        uint64
		ullAvailPageFile        uint64
		ullTotalVirtual         uint64
		ullAvailVirtual         uint64
		ullAvailExtendedVirtual uint64
	}

	var memStatus memoryStatusEx
	memStatus.dwLength = uint32(unsafe.Sizeof(memStatus))

	ret, _, _ := globalMemoryStatusEx.Call(uintptr(unsafe.Pointer(&memStatus)))
	if ret != 0 {
		info.Total = memStatus.ullTotalPhys
		info.Available = memStatus.ullAvailPhys
		info.Used = info.Total - info.Available
	}
}

// getDiskUsagePlatform Windows平台获取磁盘信息
func (s *SysUtils) getDiskUsagePlatform(path string) (*DiskInfo, error) {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getDiskFreeSpaceEx := kernel32.NewProc("GetDiskFreeSpaceExW")

	var freeBytesAvailable uint64
	var totalNumberOfBytes uint64
	var totalNumberOfFreeBytes uint64

	pathPtr, _ := syscall.UTF16PtrFromString(path)
	ret, _, _ := getDiskFreeSpaceEx.Call(
		uintptr(unsafe.Pointer(pathPtr)),
		uintptr(unsafe.Pointer(&freeBytesAvailable)),
		uintptr(unsafe.Pointer(&totalNumberOfBytes)),
		uintptr(unsafe.Pointer(&totalNumberOfFreeBytes)),
	)

	if ret == 0 {
		return nil, fmt.Errorf("获取磁盘信息失败")
	}

	used := totalNumberOfBytes - totalNumberOfFreeBytes
	usedPct := float64(used) / float64(totalNumberOfBytes) * 100

	return &DiskInfo{
		Path:       path,
		Total:      totalNumberOfBytes,
		Free:       totalNumberOfFreeBytes,
		Used:       used,
		UsedPct:    usedPct,
		FileSystem: "NTFS", // Windows通常是NTFS，这里简化处理
	}, nil
}

// getAllDrivesWindows 获取所有磁盘驱动器列表
func (s *SysUtils) getAllDrivesWindows() ([]string, error) {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getLogicalDrives := kernel32.NewProc("GetLogicalDrives")

	ret, _, _ := getLogicalDrives.Call()
	if ret == 0 {
		return nil, fmt.Errorf("获取驱动器列表失败")
	}

	var drives []string
	for i := 0; i < 26; i++ {
		if ret&(1<<uint(i)) != 0 {
			drive := string(rune('A'+i)) + ":\\"
			drives = append(drives, drive)
		}
	}

	return drives, nil
}

// getAllDisksInfoPlatform Windows平台获取所有磁盘信息
func (s *SysUtils) getAllDisksInfoPlatform() ([]*DiskInfo, error) {
	var disks []*DiskInfo

	drives, err := s.getAllDrivesWindows()
	if err != nil {
		return nil, err
	}

	for _, drive := range drives {
		diskInfo, err := s.GetDiskUsage(drive)
		if err != nil {
			continue // 跳过无法访问的驱动器
		}
		disks = append(disks, diskInfo)
	}

	return disks, nil
}

// getGPUInfoPlatform 获取Windows下的GPU信息
func (s *SysUtils) getGPUInfoPlatform() []string {
	var gpus []string

	// 使用wmic获取GPU信息
	cmd := exec.Command("wmic", "path", "win32_VideoController", "get", "name", "/value")
	output, err := cmd.Output()
	if err != nil {
		return gpus
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Name=") && len(line) > 5 {
			gpu := strings.TrimSpace(strings.TrimPrefix(line, "Name="))
			if gpu != "" {
				gpus = append(gpus, gpu)
			}
		}
	}

	return gpus
}

// getSystemUptime 获取Windows系统启动时间
func (s *SysUtils) getSystemUptime() time.Duration {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getTickCount64 := kernel32.NewProc("GetTickCount64")

	ret, _, _ := getTickCount64.Call()
	return time.Duration(ret) * time.Millisecond
}
