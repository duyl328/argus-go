//go:build darwin
// +build darwin

package utils

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// getCPUInfoPlatform 获取macOS下的CPU信息
func (s *SysUtils) getCPUInfoPlatform(info *CPUInfo) {
	cmd := exec.Command("sysctl", "-n", "machdep.cpu.brand_string")
	output, err := cmd.Output()
	if err != nil {
		return
	}
	info.ModelName = strings.TrimSpace(string(output))
}

// getMemoryInfoPlatform 获取macOS下的内存信息
func (s *SysUtils) getMemoryInfoPlatform(info *MemoryInfo) {
	cmd := exec.Command("sysctl", "-n", "hw.memsize")
	output, err := cmd.Output()
	if err != nil {
		return
	}

	total, err := strconv.ParseUint(strings.TrimSpace(string(output)), 10, 64)
	if err != nil {
		return
	}
	info.Total = total

	// 获取可用内存需要更复杂的计算，这里简化处理
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	info.Used = m.Sys
	info.Available = info.Total - info.Used
}

// getDiskUsagePlatform macOS平台获取磁盘信息
func (s *SysUtils) getDiskUsagePlatform(path string) (*DiskInfo, error) {
	var stat syscall.Statfs_t
	err := syscall.Statfs(path, &stat)
	if err != nil {
		return nil, fmt.Errorf("获取磁盘信息失败: %w", err)
	}

	total := stat.Blocks * uint64(stat.Bsize)
	free := stat.Bavail * uint64(stat.Bsize)
	used := total - free
	usedPct := float64(used) / float64(total) * 100

	// 获取文件系统类型
	fsType := "APFS" // macOS默认文件系统

	return &DiskInfo{
		Path:       path,
		Total:      total,
		Free:       free,
		Used:       used,
		UsedPct:    usedPct,
		FileSystem: fsType,
	}, nil
}

// getAllDisksInfoPlatform macOS平台获取所有磁盘信息
func (s *SysUtils) getAllDisksInfoPlatform() ([]*DiskInfo, error) {
	var disks []*DiskInfo

	// 获取挂载的卷
	cmd := exec.Command("df", "-h")
	output, err := cmd.Output()
	if err != nil {
		// 如果df命令失败，使用默认路径
		commonPaths := []string{"/", "/System/Volumes/Data"}
		for _, path := range commonPaths {
			if _, err := os.Stat(path); err == nil {
				diskInfo, err := s.GetDiskUsage(path)
				if err != nil {
					continue
				}
				disks = append(disks, diskInfo)
			}
		}
		return disks, nil
	}

	lines := strings.Split(string(output), "\n")
	for i, line := range lines {
		if i == 0 || line == "" { // 跳过标题行和空行
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 6 {
			mountPoint := fields[5]
			// 过滤掉一些虚拟文件系统
			if strings.HasPrefix(mountPoint, "/dev") ||
				strings.HasPrefix(mountPoint, "/private/var/vm") {
				continue
			}

			diskInfo, err := s.GetDiskUsage(mountPoint)
			if err != nil {
				continue
			}
			disks = append(disks, diskInfo)
		}
	}

	return disks, nil
}

// getGPUInfoPlatform 获取macOS下的GPU信息
func (s *SysUtils) getGPUInfoPlatform() []string {
	var gpus []string

	// 使用system_profiler获取GPU信息
	cmd := exec.Command("system_profiler", "SPDisplaysDataType")
	output, err := cmd.Output()
	if err != nil {
		return gpus
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "Chipset Model:") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				gpu := strings.TrimSpace(parts[1])
				gpus = append(gpus, gpu)
			}
		}
	}

	return gpus
}

// getSystemUptime 获取macOS系统启动时间
func (s *SysUtils) getSystemUptime() time.Duration {
	cmd := exec.Command("sysctl", "-n", "kern.boottime")
	output, err := cmd.Output()
	if err != nil {
		return 0
	}

	// 解析输出格式: { sec = 1234567890, usec = 123456 }
	line := strings.TrimSpace(string(output))
	if strings.Contains(line, "sec =") {
		parts := strings.Split(line, "sec =")
		if len(parts) > 1 {
			secPart := strings.Split(parts[1], ",")[0]
			secPart = strings.TrimSpace(secPart)
			sec, err := strconv.ParseInt(secPart, 10, 64)
			if err == nil {
				bootTime := time.Unix(sec, 0)
				return time.Since(bootTime)
			}
		}
	}

	return 0
}
