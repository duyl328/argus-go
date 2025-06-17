package utils

import (
	"fmt"
	"os"
	"runtime"
	"time"
)

// SystemInfo 系统信息结构体
type SystemInfo struct {
	OS           string `json:"os"`
	Arch         string `json:"arch"`
	CPUCount     int    `json:"cpu_count"`
	Is64Bit      bool   `json:"is_64_bit"`
	Hostname     string `json:"hostname"`
	Username     string `json:"username"`
	GoVersion    string `json:"go_version"`
	Compiler     string `json:"compiler"`
	NumGoroutine int    `json:"num_goroutine"`
}

// CPUInfo CPU信息结构体
type CPUInfo struct {
	ModelName string  `json:"model_name"`
	Cores     int     `json:"cores"`
	Threads   int     `json:"threads"`
	Frequency string  `json:"frequency"`
	Usage     float64 `json:"usage"`
}

// MemoryInfo 内存信息结构体
type MemoryInfo struct {
	Total     uint64  `json:"total"`
	Available uint64  `json:"available"`
	Used      uint64  `json:"used"`
	UsedPct   float64 `json:"used_percent"`
}

// DiskInfo 磁盘信息结构体
type DiskInfo struct {
	Path       string  `json:"path"`         // 磁盘路径
	Total      uint64  `json:"total"`        // 总空间 (字节)
	Free       uint64  `json:"free"`         // 可用空间 (字节)
	Used       uint64  `json:"used"`         // 已使用空间 (字节)
	UsedPct    float64 `json:"used_percent"` // 使用百分比
	FileSystem string  `json:"file_system"`  // 文件系统类型
}

// NetworkInfo 网络信息结构体
type NetworkInfo struct {
	InterfaceName string `json:"interface_name"`
	IPAddress     string `json:"ip_address"`
	MACAddress    string `json:"mac_address"`
	MTU           int    `json:"mtu"`
}

// ProcessInfo 进程信息结构体
type ProcessInfo struct {
	PID         int     `json:"pid"`
	PPID        int     `json:"ppid"`
	Name        string  `json:"name"`
	CPUPercent  float64 `json:"cpu_percent"`
	MemoryUsage uint64  `json:"memory_usage"`
}

// SysUtils 系统工具类
type SysUtils struct{}

// NewSysUtils 创建新的系统工具实例
func NewSysUtils() *SysUtils {
	return &SysUtils{}
}

// GetSystemInfo 获取基本系统信息
func (s *SysUtils) GetSystemInfo() SystemInfo {
	hostname, _ := os.Hostname()
	username := os.Getenv("USER")
	if username == "" {
		username = os.Getenv("USERNAME") // Windows
	}

	return SystemInfo{
		OS:           runtime.GOOS,
		Arch:         runtime.GOARCH,
		CPUCount:     runtime.NumCPU(),
		Is64Bit:      s.Is64Bit(),
		Hostname:     hostname,
		Username:     username,
		GoVersion:    runtime.Version(),
		Compiler:     runtime.Compiler,
		NumGoroutine: runtime.NumGoroutine(),
	}
}

// GetCurrentProcessInfo 获取当前进程信息
func (s *SysUtils) GetCurrentProcessInfo() ProcessInfo {
	pid := os.Getpid()
	ppid := os.Getppid()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return ProcessInfo{
		PID:         pid,
		PPID:        ppid,
		Name:        os.Args[0],
		MemoryUsage: m.Alloc,
	}
}

// GetSystemUptime 获取系统启动时间
func (s *SysUtils) GetSystemUptime() time.Duration {
	return s.getSystemUptime()
}

// FormatBytes 格式化字节数为人类可读格式
func (s *SysUtils) FormatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// PrintSystemInfo 打印完整的系统信息
func (s *SysUtils) PrintSystemInfo() {
	fmt.Println("========== 系统信息 ==========")
	sysInfo := s.GetSystemInfo()
	fmt.Printf("操作系统: %s\n", sysInfo.OS)
	fmt.Printf("架构: %s\n", sysInfo.Arch)
	fmt.Printf("CPU核心数: %d\n", sysInfo.CPUCount)
	fmt.Printf("64位系统: %v\n", sysInfo.Is64Bit)
	fmt.Printf("主机名: %s\n", sysInfo.Hostname)
	fmt.Printf("用户名: %s\n", sysInfo.Username)
	fmt.Printf("Go版本: %s\n", sysInfo.GoVersion)
	fmt.Printf("编译器: %s\n", sysInfo.Compiler)
	fmt.Printf("Goroutine数量: %d\n", sysInfo.NumGoroutine)

	fmt.Println("\n========== CPU信息 ==========")
	cpuInfo := s.GetCPUInfo()
	fmt.Printf("型号: %s\n", cpuInfo.ModelName)
	fmt.Printf("核心数: %d\n", cpuInfo.Cores)
	fmt.Printf("线程数: %d\n", cpuInfo.Threads)

	fmt.Println("\n========== 内存信息 ==========")
	memInfo := s.GetMemoryInfo()
	fmt.Printf("总内存: %s\n", s.FormatBytes(memInfo.Total))
	fmt.Printf("已用内存: %s\n", s.FormatBytes(memInfo.Used))
	fmt.Printf("可用内存: %s\n", s.FormatBytes(memInfo.Available))
	fmt.Printf("使用率: %.2f%%\n", memInfo.UsedPct)

	fmt.Println("\n========== 磁盘信息 ==========")
	diskInfo, _ := s.GetDiskUsage(".")
	if diskInfo != nil {
		fmt.Printf("路径: %s\n", diskInfo.Path)
		fmt.Printf("总空间: %s\n", s.FormatBytes(diskInfo.Total))
		fmt.Printf("已用空间: %s\n", s.FormatBytes(diskInfo.Used))
		fmt.Printf("可用空间: %s\n", s.FormatBytes(diskInfo.Free))
		fmt.Printf("使用率: %.2f%%\n", diskInfo.UsedPct)
	}

	fmt.Println("\n========== 网络接口 ==========")
	networkInfos := s.GetNetworkInterfaces()
	for _, netInfo := range networkInfos {
		fmt.Printf("接口: %s\n", netInfo.InterfaceName)
		fmt.Printf("IP地址: %s\n", netInfo.IPAddress)
		fmt.Printf("MAC地址: %s\n", netInfo.MACAddress)
		fmt.Printf("MTU: %d\n", netInfo.MTU)
		fmt.Println("---")
	}

	fmt.Println("\n========== GPU信息 ==========")
	gpus := s.GetGPUInfo()
	for i, gpu := range gpus {
		fmt.Printf("GPU %d: %s\n", i+1, gpu)
	}

	fmt.Println("\n========== 进程信息 ==========")
	processInfo := s.GetCurrentProcessInfo()
	fmt.Printf("进程ID: %d\n", processInfo.PID)
	fmt.Printf("父进程ID: %d\n", processInfo.PPID)
	fmt.Printf("进程名: %s\n", processInfo.Name)
	fmt.Printf("内存使用: %s\n", s.FormatBytes(processInfo.MemoryUsage))

	fmt.Println("\n========== 系统运行时间 ==========")
	uptime := s.GetSystemUptime()
	fmt.Printf("运行时间: %s\n", uptime.String())
}
