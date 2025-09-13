package system

import (
	"net"
	"strings"
	"testing"
	"time"
)

func TestNetworkManagerScan(t *testing.T) {
	t.Log("Testing NetworkManager scanning functionality...")

	nm := NewNetworkManager()
	err := nm.ScanInterfaces()
	if err != nil {
		t.Fatalf("ScanInterfaces failed: %v", err)
	}

	interfaces := nm.GetInterfaces()
	if len(interfaces) == 0 {
		t.Error("Should find at least one network interface")
	}

	t.Logf("Found %d network interfaces:", len(interfaces))
	for i, iface := range interfaces {
		t.Logf("Interface %d:", i+1)
		t.Logf("  Name: %s", iface.Name)
		t.Logf("  Hardware Address: %s", iface.HardwareAddr)
		t.Logf("  Is Up: %t", iface.IsUp)
		t.Logf("  Is Loopback: %t", iface.IsLoopback)
		t.Logf("  MTU: %d", iface.MTU)
		t.Logf("  IP Addresses: %v", iface.IPAddresses)
		t.Log("  ---")
	}

	// 验证接口数据的合理性
	for _, iface := range interfaces {
		if iface.Name == "" {
			t.Error("Interface name should not be empty")
		}

		// MTU validation - loopback interfaces may have special MTU values like -1
		if iface.MTU <= 0 && !iface.IsLoopback {
			t.Errorf("Interface MTU should be positive for non-loopback interface, got %d for %s", iface.MTU, iface.Name)
		}

		// 验证IP地址格式
		for _, ip := range iface.IPAddresses {
			if net.ParseIP(ip) == nil {
				t.Errorf("Invalid IP address format: %s", ip)
			}
		}
	}
}

func TestGetActiveInterfaces(t *testing.T) {
	t.Log("Testing GetActiveInterfaces functionality...")

	nm := NewNetworkManager()
	err := nm.ScanInterfaces()
	if err != nil {
		t.Fatalf("ScanInterfaces failed: %v", err)
	}

	activeInterfaces := nm.GetActiveInterfaces()
	allInterfaces := nm.GetInterfaces()

	t.Logf("Found %d active interfaces out of %d total", len(activeInterfaces), len(allInterfaces))

	// 验证活动接口的条件
	for _, iface := range activeInterfaces {
		if !iface.IsUp {
			t.Errorf("Active interface %s should be up", iface.Name)
		}
		if iface.IsLoopback {
			t.Errorf("Active interface %s should not be loopback", iface.Name)
		}
	}

	for i, iface := range activeInterfaces {
		t.Logf("Active Interface %d: %s (%v)", i+1, iface.Name, iface.IPAddresses)
	}
}

func TestPingHost(t *testing.T) {
	t.Log("Testing PingHost functionality...")

	// 测试本地连接
	err := PingHost("127.0.0.1:80", 2*time.Second)
	if err != nil {
		t.Logf("Ping localhost:80 failed (expected): %v", err)
	}

	// 测试常见的可访问主机
	testHosts := []string{
		"google.com:80",
		"microsoft.com:80",
		"github.com:443",
	}

	for _, host := range testHosts {
		err := PingHost(host, 3*time.Second)
		if err != nil {
			t.Logf("Ping %s failed: %v", host, err)
		} else {
			t.Logf("Ping %s succeeded", host)
		}
	}

	// 测试无效主机
	err = PingHost("nonexistent.invalid:80", 1*time.Second)
	if err == nil {
		t.Error("Ping to nonexistent host should fail")
	}
}

func TestResolveHostname(t *testing.T) {
	t.Log("Testing ResolveHostname functionality...")

	// 测试本地主机
	ips, err := ResolveHostname("localhost")
	if err != nil {
		t.Logf("Resolve localhost failed: %v", err)
	} else {
		t.Logf("localhost resolves to: %v", ips)
		if len(ips) == 0 {
			t.Error("localhost should resolve to at least one IP")
		}
	}

	// 测试知名域名
	testDomains := []string{
		"google.com",
		"github.com",
		"microsoft.com",
	}

	for _, domain := range testDomains {
		ips, err := ResolveHostname(domain)
		if err != nil {
			t.Logf("Resolve %s failed: %v", domain, err)
		} else {
			t.Logf("%s resolves to: %v", domain, ips)
			if len(ips) == 0 {
				t.Errorf("%s should resolve to at least one IP", domain)
			}
		}
	}

	// 测试无效域名
	_, err = ResolveHostname("nonexistent.invalid.domain")
	if err == nil {
		t.Error("Resolving nonexistent domain should fail")
	}
}

func TestGetLocalIP(t *testing.T) {
	t.Log("Testing GetLocalIP functionality...")

	localIP, err := GetLocalIP()
	if err != nil {
		t.Fatalf("GetLocalIP failed: %v", err)
	}

	if localIP == "" {
		t.Error("Local IP should not be empty")
	}

	// 验证IP格式
	if net.ParseIP(localIP) == nil {
		t.Errorf("Invalid local IP format: %s", localIP)
	}

	t.Logf("Local IP: %s", localIP)

	// 验证不是回环地址
	if strings.HasPrefix(localIP, "127.") {
		t.Error("Local IP should not be loopback address")
	}
}

func TestIsPortOpen(t *testing.T) {
	t.Log("Testing IsPortOpen functionality...")

	// 测试不太可能开放的端口
	if IsPortOpen("127.0.0.1", 99999, 100*time.Millisecond) {
		t.Error("Port 99999 should not be open")
	}

	// 测试常见开放端口（如果有的话）
	commonPorts := []int{22, 80, 443, 3306, 5432, 6379}
	openPorts := []int{}

	for _, port := range commonPorts {
		if IsPortOpen("127.0.0.1", port, 100*time.Millisecond) {
			openPorts = append(openPorts, port)
		}
	}

	t.Logf("Open ports on localhost: %v", openPorts)
}

func TestScanPorts(t *testing.T) {
	t.Log("Testing ScanPorts functionality...")

	// 扫描小范围端口
	openPorts := ScanPorts("127.0.0.1", 8080, 8090, 50*time.Millisecond)
	t.Logf("Open ports in range 8080-8090: %v", openPorts)

	// 验证返回的端口在指定范围内
	for _, port := range openPorts {
		if port < 8080 || port > 8090 {
			t.Errorf("Port %d is outside the specified range 8080-8090", port)
		}
	}
}

func TestGetMACAddress(t *testing.T) {
	t.Log("Testing GetMACAddress functionality...")

	macAddr, err := GetMACAddress()
	if err != nil {
		t.Fatalf("GetMACAddress failed: %v", err)
	}

	if macAddr == "" {
		t.Error("MAC address should not be empty")
	}

	// 验证MAC地址格式 (简单检查)
	if !strings.Contains(macAddr, ":") {
		t.Errorf("MAC address should contain colons: %s", macAddr)
	}

	// MAC地址应该是6组16进制数
	parts := strings.Split(macAddr, ":")
	if len(parts) != 6 {
		t.Errorf("MAC address should have 6 parts, got %d: %s", len(parts), macAddr)
	}

	t.Logf("MAC Address: %s", macAddr)
}

func TestTraceRoute(t *testing.T) {
	t.Log("Testing TraceRoute functionality...")

	// 测试到本地主机的追踪
	routes, err := TraceRoute("localhost", 10)
	if err != nil {
		t.Logf("TraceRoute to localhost failed: %v", err)
	} else {
		t.Logf("Route to localhost: %v", routes)
	}

	// 测试到外部主机的追踪
	routes, err = TraceRoute("google.com", 5)
	if err != nil {
		t.Logf("TraceRoute to google.com failed: %v", err)
	} else {
		t.Logf("Route to google.com: %v", routes)
	}
}

func TestTestNetworkSpeed(t *testing.T) {
	t.Log("Testing TestNetworkSpeed functionality...")

	// 测试到常见主机的网络速度
	speed, err := TestNetworkSpeed("google.com:80")
	if err != nil {
		t.Logf("Network speed test failed: %v", err)
	} else {
		t.Logf("Network speed test results:")
		t.Logf("  Download: %.1f Mbps", speed.DownloadMbps)
		t.Logf("  Upload: %.1f Mbps", speed.UploadMbps)
		t.Logf("  Latency: %d ms", speed.Latency)

		// 验证数据合理性
		if speed.Latency < 0 {
			t.Error("Latency should not be negative")
		}
		if speed.DownloadMbps < 0 || speed.UploadMbps < 0 {
			t.Error("Speed values should not be negative")
		}
	}
}

func TestNetworkInterfaceFormatting(t *testing.T) {
	t.Log("Testing NetworkInterface formatting...")

	nm := NewNetworkManager()
	err := nm.ScanInterfaces()
	if err != nil {
		t.Fatalf("ScanInterfaces failed: %v", err)
	}

	interfaces := nm.GetInterfaces()
	if len(interfaces) == 0 {
		t.Skip("No network interfaces found")
	}

	// 测试第一个接口的格式化
	firstInterface := interfaces[0]
	formatted := firstInterface.FormatNetworkInterface()

	if formatted == "" {
		t.Error("Formatted network interface should not be empty")
	}

	// 验证格式化输出包含关键信息
	expectedContent := []string{
		"网络接口",
		firstInterface.Name,
		"状态",
		"MTU",
	}

	for _, content := range expectedContent {
		if !strings.Contains(formatted, content) {
			t.Errorf("Formatted output should contain '%s'", content)
		}
	}

	t.Logf("Formatted network interface:\n%s", formatted)
}

func TestGlobalNetworkFunctions(t *testing.T) {
	t.Log("Testing global network functions...")

	// 测试全局函数
	err := ScanNetworkInterfaces()
	if err != nil {
		t.Fatalf("ScanNetworkInterfaces failed: %v", err)
	}

	interfaces := GetNetworkInterfaces()
	if len(interfaces) == 0 {
		t.Error("Should find at least one network interface")
	}

	activeInterfaces := GetActiveNetworkInterfaces()
	t.Logf("Global functions - Total: %d, Active: %d", len(interfaces), len(activeInterfaces))

	// 验证活动接口是总接口的子集
	if len(activeInterfaces) > len(interfaces) {
		t.Error("Active interfaces count should not exceed total interfaces count")
	}
}

func TestNetworkUtilsEdgeCases(t *testing.T) {
	t.Log("Testing NetworkUtils edge cases...")

	// 测试无效主机的ping
	err := PingHost("", 1*time.Second)
	if err == nil {
		t.Error("Ping to empty host should fail")
	}

	// 测试无效端口
	if IsPortOpen("127.0.0.1", 0, 100*time.Millisecond) {
		t.Error("Port 0 should not be considered open")
	}

	if IsPortOpen("127.0.0.1", 65536, 100*time.Millisecond) {
		t.Error("Port 65536 is invalid and should not be open")
	}

	// 测试空的域名解析
	_, err = ResolveHostname("")
	if err == nil {
		t.Error("Resolving empty hostname should fail")
	}

	t.Log("Edge case testing completed")
}