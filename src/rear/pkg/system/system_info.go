package system

import (
	"fmt"
	"net"
	"os"
	"runtime"
	"time"
)

// SystemInfo 系统信息结构体
type SystemInfo struct {
	// 基础系统信息
	OS           string `json:"os"`            // 操作系统
	Architecture string `json:"architecture"`  // 架构
	Hostname     string `json:"hostname"`      // 主机名
	Username     string `json:"username"`      // 当前用户
	WorkingDir   string `json:"working_dir"`   // 当前工作目录

	// 运行时信息
	GoVersion    string `json:"go_version"`    // Go 版本
	NumCPU       int    `json:"num_cpu"`       // CPU 核心数
	NumGoroutine int    `json:"num_goroutine"` // 协程数

	// 网络信息
	IPAddresses []string `json:"ip_addresses"` // IP 地址列表

	// 时间信息
	Uptime      time.Duration `json:"uptime"`       // 系统启动时间
	CurrentTime time.Time     `json:"current_time"` // 当前时间
	TimeZone    string        `json:"time_zone"`    // 时区

	// 环境变量
	Environment map[string]string `json:"environment"` // 重要环境变量
}

// GetSystemInfo 获取系统信息
func GetSystemInfo() (*SystemInfo, error) {
	info := &SystemInfo{
		OS:           runtime.GOOS,
		Architecture: runtime.GOARCH,
		GoVersion:    runtime.Version(),
		NumCPU:       runtime.NumCPU(),
		NumGoroutine: runtime.NumGoroutine(),
		CurrentTime:  time.Now(),
		Environment:  make(map[string]string),
	}

	// 获取主机名
	if hostname, err := os.Hostname(); err == nil {
		info.Hostname = hostname
	}

	// 获取当前用户
	if user := os.Getenv("USER"); user != "" {
		info.Username = user
	} else if user := os.Getenv("USERNAME"); user != "" {
		info.Username = user
	}

	// 获取工作目录
	if wd, err := os.Getwd(); err == nil {
		info.WorkingDir = wd
	}

	// 获取时区
	info.TimeZone = info.CurrentTime.Format("-0700 MST")

	// 获取 IP 地址
	info.IPAddresses = getIPAddresses()

	// 获取重要环境变量
	importantEnvVars := []string{
		"PATH", "HOME", "USERPROFILE", "TEMP", "TMP",
		"GOPATH", "GOROOT", "NODE_ENV", "PYTHONPATH",
	}

	for _, envVar := range importantEnvVars {
		if value := os.Getenv(envVar); value != "" {
			info.Environment[envVar] = value
		}
	}

	return info, nil
}

// getIPAddresses 获取本机所有 IP 地址
func getIPAddresses() []string {
	var ips []string

	interfaces, err := net.Interfaces()
	if err != nil {
		return ips
	}

	for _, iface := range interfaces {
		// 跳过回环接口和非活动接口
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					ips = append(ips, ipnet.IP.String())
				}
			}
		}
	}

	return ips
}

// IsAdmin 检查当前进程是否以管理员权限运行
func IsAdmin() bool {
	switch runtime.GOOS {
	case "windows":
		return isWindowsAdmin()
	default:
		return isUnixRoot()
	}
}

// isWindowsAdmin 和 isUnixRoot 在对应的平台特定文件中实现

// GetCurrentUser 获取当前用户信息
func GetCurrentUser() string {
	if user := os.Getenv("USER"); user != "" {
		return user
	}
	if user := os.Getenv("USERNAME"); user != "" {
		return user
	}
	return "unknown"
}

// GetTempDir 获取系统临时目录
func GetTempDir() string {
	return os.TempDir()
}

// GetHomeDir 获取用户主目录
func GetHomeDir() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	if home := os.Getenv("USERPROFILE"); home != "" {
		return home
	}
	return ""
}

// GetExecutablePath 获取当前可执行文件路径
func GetExecutablePath() (string, error) {
	return os.Executable()
}

// GetProcessID 获取当前进程 ID
func GetProcessID() int {
	return os.Getpid()
}

// GetParentProcessID 获取父进程 ID
func GetParentProcessID() int {
	return os.Getppid()
}

// FormatSystemInfo 格式化系统信息为易读的字符串
func (si *SystemInfo) FormatSystemInfo() string {
	result := fmt.Sprintf("=== 系统信息 ===\n")
	result += fmt.Sprintf("操作系统: %s (%s)\n", si.OS, si.Architecture)
	result += fmt.Sprintf("主机名: %s\n", si.Hostname)
	result += fmt.Sprintf("用户: %s\n", si.Username)
	result += fmt.Sprintf("工作目录: %s\n", si.WorkingDir)
	result += fmt.Sprintf("Go 版本: %s\n", si.GoVersion)
	result += fmt.Sprintf("CPU 核心数: %d\n", si.NumCPU)
	result += fmt.Sprintf("协程数: %d\n", si.NumGoroutine)
	result += fmt.Sprintf("当前时间: %s (%s)\n", si.CurrentTime.Format("2006-01-02 15:04:05"), si.TimeZone)

	if len(si.IPAddresses) > 0 {
		result += fmt.Sprintf("IP 地址: %v\n", si.IPAddresses)
	}

	if len(si.Environment) > 0 {
		result += "重要环境变量:\n"
		for key, value := range si.Environment {
			if len(value) > 50 {
				value = value[:47] + "..."
			}
			result += fmt.Sprintf("  %s: %s\n", key, value)
		}
	}

	return result
}