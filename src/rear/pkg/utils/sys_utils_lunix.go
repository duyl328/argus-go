//go:build linux
// +build linux

package utils

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// getCPUInfoPlatform 获取Linux下的CPU信息
func (s *SysUtils) getCPUInfoPlatform(info *CPUInfo) {
	file, err := os.Open("/proc/cpuinfo")
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "model name") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				info.ModelName = strings.TrimSpace(parts[1])
				break
			}
		}
	}
}

// getMemoryInfoPlatform 获取Linux下的内存信息
func (s *SysUtils) getMemoryInfoPlatform(info *MemoryInfo) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		value, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			continue
		}
		value *= 1024 // Convert KB to bytes

		switch fields[0] {
		case "MemTotal:":
			info.Total = value
		case "MemAvailable:":
			info.Available = value
		}
	}

	if info.Total > 0 && info.Available > 0 {
		info.Used = info.Total - info.Available
	}
}

// getDiskUsagePlatform Linux平台获取磁盘信息
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

	return &DiskInfo{
		Path:       path,
		Total:      total,
		Free:       free,
		Used:       used,
		UsedPct:    usedPct,
		FileSystem: "ext4", // Linux通常是ext4，这里简化处理
	}, nil
}

// getAllMountPoints 获取所有挂载点
func (s *SysUtils) getAllMountPoints() ([]string, error) {
	// 读取/proc/mounts文件
	content, err := os.ReadFile("/proc/mounts")
	if err != nil {
		return nil, fmt.Errorf("读取挂载点失败: %w", err)
	}

	var mountPoints []string
	lines := strings.Split(string(content), "\n")

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			mountPoint := fields[1]
			// 过滤掉一些虚拟文件系统
			if !strings.HasPrefix(mountPoint, "/proc") &&
				!strings.HasPrefix(mountPoint, "/sys") &&
				!strings.HasPrefix(mountPoint, "/dev") &&
				mountPoint != "/" {
				mountPoints = append(mountPoints, mountPoint)
			}
		}
	}

	return mountPoints, nil
}

// getAllDisksInfoPlatform Linux平台获取所有磁盘信息
func (s *SysUtils) getAllDisksInfoPlatform() ([]*DiskInfo, error) {
	var disks []*DiskInfo

	// 通常检查根目录和常见挂载点
	commonPaths := []string{"/", "/home", "/var", "/tmp"}

	for _, path := range commonPaths {
		if _, err := os.Stat(path); err == nil {
			diskInfo, err := s.GetDiskUsage(path)
			if err != nil {
				continue
			}
			disks = append(disks, diskInfo)
		}
	}

	// 也可以尝试从挂载点获取
	mountPoints, err := s.getAllMountPoints()
	if err == nil {
		for _, mount := range mountPoints {
			diskInfo, err := s.GetDiskUsage(mount)
			if err != nil {
				continue
			}
			// 检查是否已经存在
			exists := false
			for _, d := range disks {
				if d.Path == diskInfo.Path {
					exists = true
					break
				}
			}
			if !exists {
				disks = append(disks, diskInfo)
			}
		}
	}

	return disks, nil
}

// getGPUInfoPlatform 获取Linux下的GPU信息
func (s *SysUtils) getGPUInfoPlatform() []string {
	var gpus []string

	// 尝试使用lspci获取GPU信息
	cmd := exec.Command("lspci")
	output, err := cmd.Output()
	if err != nil {
		return gpus
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(strings.ToLower(line), "vga") ||
			strings.Contains(strings.ToLower(line), "3d") ||
			strings.Contains(strings.ToLower(line), "display") {
			gpus = append(gpus, strings.TrimSpace(line))
		}
	}

	return gpus
}

// getSystemUptime 获取Linux系统启动时间
func (s *SysUtils) getSystemUptime() time.Duration {
	file, err := os.Open("/proc/uptime")
	if err != nil {
		return 0
	}
	defer file.Close()

	var uptime float64
	fmt.Fscanf(file, "%f", &uptime)
	return time.Duration(uptime) * time.Second
}
