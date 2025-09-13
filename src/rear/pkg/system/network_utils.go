package system

import (
	"fmt"
	"net"
	"time"
)

// NetworkInterface 网络接口信息
type NetworkInterface struct {
	Name         string   `json:"name"`          // 接口名称
	DisplayName  string   `json:"display_name"`  // 显示名称
	HardwareAddr string   `json:"hardware_addr"` // MAC 地址
	IPAddresses  []string `json:"ip_addresses"`  // IP 地址列表
	IsUp         bool     `json:"is_up"`         // 是否启用
	IsLoopback   bool     `json:"is_loopback"`   // 是否回环接口
	MTU          int      `json:"mtu"`           // 最大传输单元
}

// NetworkManager 网络管理器
type NetworkManager struct {
	interfaces []NetworkInterface
}

// NewNetworkManager 创建网络管理器
func NewNetworkManager() *NetworkManager {
	return &NetworkManager{}
}

// ScanInterfaces 扫描网络接口
func (nm *NetworkManager) ScanInterfaces() error {
	interfaces, err := net.Interfaces()
	if err != nil {
		return fmt.Errorf("获取网络接口失败: %w", err)
	}

	nm.interfaces = make([]NetworkInterface, 0, len(interfaces))

	for _, iface := range interfaces {
		netIface := NetworkInterface{
			Name:         iface.Name,
			HardwareAddr: iface.HardwareAddr.String(),
			IsUp:         iface.Flags&net.FlagUp != 0,
			IsLoopback:   iface.Flags&net.FlagLoopback != 0,
			MTU:          iface.MTU,
			IPAddresses:  []string{},
		}

		// 获取接口的 IP 地址
		addrs, err := iface.Addrs()
		if err == nil {
			for _, addr := range addrs {
				if ipnet, ok := addr.(*net.IPNet); ok {
					netIface.IPAddresses = append(netIface.IPAddresses, ipnet.IP.String())
				}
			}
		}

		nm.interfaces = append(nm.interfaces, netIface)
	}

	return nil
}

// GetInterfaces 获取所有网络接口
func (nm *NetworkManager) GetInterfaces() []NetworkInterface {
	return nm.interfaces
}

// GetActiveInterfaces 获取活动的网络接口
func (nm *NetworkManager) GetActiveInterfaces() []NetworkInterface {
	var active []NetworkInterface
	for _, iface := range nm.interfaces {
		if iface.IsUp && !iface.IsLoopback {
			active = append(active, iface)
		}
	}
	return active
}

// PingHost 测试主机连通性
func PingHost(host string, timeout time.Duration) error {
	conn, err := net.DialTimeout("tcp", host, timeout)
	if err != nil {
		return fmt.Errorf("连接失败: %w", err)
	}
	defer conn.Close()
	return nil
}

// ResolveHostname 解析主机名
func ResolveHostname(hostname string) ([]string, error) {
	ips, err := net.LookupIP(hostname)
	if err != nil {
		return nil, fmt.Errorf("解析主机名失败: %w", err)
	}

	var addresses []string
	for _, ip := range ips {
		addresses = append(addresses, ip.String())
	}

	return addresses, nil
}

// GetLocalIP 获取本机主要 IP 地址
func GetLocalIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", fmt.Errorf("获取本机IP失败: %w", err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
}

// IsPortOpen 检查端口是否开放
func IsPortOpen(host string, port int, timeout time.Duration) bool {
	address := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}

// ScanPorts 扫描端口范围
func ScanPorts(host string, startPort, endPort int, timeout time.Duration) []int {
	var openPorts []int

	for port := startPort; port <= endPort; port++ {
		if IsPortOpen(host, port, timeout) {
			openPorts = append(openPorts, port)
		}
	}

	return openPorts
}

// GetMACAddress 获取主要网络接口的 MAC 地址
func GetMACAddress() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", fmt.Errorf("获取网络接口失败: %w", err)
	}

	for _, iface := range interfaces {
		// 跳过回环接口和无 MAC 地址的接口
		if iface.Flags&net.FlagLoopback == 0 && len(iface.HardwareAddr) > 0 {
			return iface.HardwareAddr.String(), nil
		}
	}

	return "", fmt.Errorf("未找到有效的 MAC 地址")
}

// TraceRoute 简单的路由跟踪 (仅支持 Unix 系统)
func TraceRoute(host string, maxHops int) ([]string, error) {
	// 这里应该实现真正的路由跟踪功能
	// 为了简化，这里只返回目标主机的 IP
	ips, err := ResolveHostname(host)
	if err != nil {
		return nil, err
	}

	if len(ips) > 0 {
		return []string{ips[0]}, nil
	}

	return []string{}, nil
}

// NetworkSpeed 网络速度测试结果
type NetworkSpeed struct {
	DownloadMbps float64 `json:"download_mbps"` // 下载速度 (Mbps)
	UploadMbps   float64 `json:"upload_mbps"`   // 上传速度 (Mbps)
	Latency      int     `json:"latency"`       // 延迟 (ms)
}

// TestNetworkSpeed 测试网络速度 (简化版本)
func TestNetworkSpeed(testHost string) (*NetworkSpeed, error) {
	// 测试延迟
	start := time.Now()
	err := PingHost(testHost, 5*time.Second)
	latency := time.Since(start).Milliseconds()

	if err != nil {
		return nil, fmt.Errorf("网络连接测试失败: %w", err)
	}

	// 这里应该实现真正的速度测试
	// 为了简化，返回模拟数据
	speed := &NetworkSpeed{
		DownloadMbps: 100.0, // 模拟数据
		UploadMbps:   50.0,  // 模拟数据
		Latency:      int(latency),
	}

	return speed, nil
}

// FormatNetworkInterface 格式化网络接口信息
func (ni *NetworkInterface) FormatNetworkInterface() string {
	result := fmt.Sprintf("网络接口: %s\n", ni.Name)
	if ni.DisplayName != "" && ni.DisplayName != ni.Name {
		result += fmt.Sprintf("  显示名称: %s\n", ni.DisplayName)
	}
	if ni.HardwareAddr != "" {
		result += fmt.Sprintf("  MAC 地址: %s\n", ni.HardwareAddr)
	}
	result += fmt.Sprintf("  状态: %s\n", map[bool]string{true: "启用", false: "禁用"}[ni.IsUp])
	result += fmt.Sprintf("  类型: %s\n", map[bool]string{true: "回环接口", false: "物理接口"}[ni.IsLoopback])
	result += fmt.Sprintf("  MTU: %d\n", ni.MTU)

	if len(ni.IPAddresses) > 0 {
		result += "  IP 地址:\n"
		for _, ip := range ni.IPAddresses {
			result += fmt.Sprintf("    - %s\n", ip)
		}
	}

	return result
}

// 全局便捷函数
var DefaultNetworkManager = NewNetworkManager()

// ScanNetworkInterfaces 扫描网络接口 (使用默认管理器)
func ScanNetworkInterfaces() error {
	return DefaultNetworkManager.ScanInterfaces()
}

// GetNetworkInterfaces 获取网络接口 (使用默认管理器)
func GetNetworkInterfaces() []NetworkInterface {
	return DefaultNetworkManager.GetInterfaces()
}

// GetActiveNetworkInterfaces 获取活动网络接口 (使用默认管理器)
func GetActiveNetworkInterfaces() []NetworkInterface {
	return DefaultNetworkManager.GetActiveInterfaces()
}