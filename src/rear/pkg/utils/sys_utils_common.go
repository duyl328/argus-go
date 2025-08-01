package utils

import (
	"net"
	"runtime"
	"unsafe"
)

// Is64Bit 检查系统是否为64位
func (s *SysUtils) Is64Bit() bool {
	return unsafe.Sizeof(uintptr(0)) == 8
}

// GetNetworkInterfaces 获取网络接口信息
func (s *SysUtils) GetNetworkInterfaces() []NetworkInfo {
	var networkInfos []NetworkInfo

	interfaces, err := net.Interfaces()
	if err != nil {
		return networkInfos
	}

	for _, iface := range interfaces {
		// 跳过回环接口和未启用的接口
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		var ipAddr string
		for _, addr := range addrs {
			if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
				if ipNet.IP.To4() != nil {
					ipAddr = ipNet.IP.String()
					break
				}
			}
		}

		if ipAddr != "" {
			networkInfo := NetworkInfo{
				InterfaceName: iface.Name,
				IPAddress:     ipAddr,
				MACAddress:    iface.HardwareAddr.String(),
				MTU:           iface.MTU,
			}
			networkInfos = append(networkInfos, networkInfo)
		}
	}

	return networkInfos
}

// GetCPUInfo 获取CPU信息
func (s *SysUtils) GetCPUInfo() CPUInfo {
	cpuInfo := CPUInfo{
		Cores:   runtime.NumCPU(),
		Threads: runtime.NumCPU(), // 在Go中通常相等
	}

	// 调用平台特定的方法来获取详细信息
	s.getCPUInfoPlatform(&cpuInfo)

	return cpuInfo
}

// GetMemoryInfo 获取内存信息
func (s *SysUtils) GetMemoryInfo() MemoryInfo {
	var memInfo MemoryInfo

	// 调用平台特定的方法
	s.getMemoryInfoPlatform(&memInfo)

	if memInfo.Total > 0 {
		memInfo.UsedPct = float64(memInfo.Used) / float64(memInfo.Total) * 100
	}

	return memInfo
}

// GetDiskUsage 获取磁盘使用情况
func (s *SysUtils) GetDiskUsage(path string) (*DiskInfo, error) {
	return s.getDiskUsagePlatform(path)
}

// GetAllDisksInfo 跨平台获取所有磁盘信息
func (s *SysUtils) GetAllDisksInfo() ([]*DiskInfo, error) {
	return s.getAllDisksInfoPlatform()
}

// GetGPUInfo 获取GPU信息
func (s *SysUtils) GetGPUInfo() []string {
	return s.getGPUInfoPlatform()
}
