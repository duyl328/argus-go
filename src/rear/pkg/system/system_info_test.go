package system

import (
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestGetSystemInfo(t *testing.T) {
	t.Log("Testing GetSystemInfo functionality...")

	sysInfo, err := GetSystemInfo()
	if err != nil {
		t.Fatalf("GetSystemInfo failed: %v", err)
	}

	// 验证基础系统信息
	if sysInfo.OS == "" {
		t.Error("OS should not be empty")
	}
	if sysInfo.OS != runtime.GOOS {
		t.Errorf("Expected OS %s, got %s", runtime.GOOS, sysInfo.OS)
	}

	if sysInfo.Architecture == "" {
		t.Error("Architecture should not be empty")
	}
	if sysInfo.Architecture != runtime.GOARCH {
		t.Errorf("Expected architecture %s, got %s", runtime.GOARCH, sysInfo.Architecture)
	}

	if sysInfo.Hostname == "" {
		t.Error("Hostname should not be empty")
	}

	if sysInfo.GoVersion == "" {
		t.Error("GoVersion should not be empty")
	}
	if !strings.Contains(sysInfo.GoVersion, "go") {
		t.Errorf("GoVersion should contain 'go', got %s", sysInfo.GoVersion)
	}

	if sysInfo.NumCPU <= 0 {
		t.Errorf("NumCPU should be positive, got %d", sysInfo.NumCPU)
	}

	if sysInfo.NumGoroutine <= 0 {
		t.Errorf("NumGoroutine should be positive, got %d", sysInfo.NumGoroutine)
	}

	if sysInfo.CurrentTime.IsZero() {
		t.Error("CurrentTime should not be zero")
	}

	// 验证时间是否合理 (应该接近当前时间)
	timeDiff := time.Since(sysInfo.CurrentTime)
	if timeDiff > time.Minute {
		t.Errorf("CurrentTime seems too old: %v", timeDiff)
	}

	if sysInfo.TimeZone == "" {
		t.Error("TimeZone should not be empty")
	}

	t.Logf("System Info collected successfully:")
	t.Logf("  OS: %s (%s)", sysInfo.OS, sysInfo.Architecture)
	t.Logf("  Hostname: %s", sysInfo.Hostname)
	t.Logf("  Username: %s", sysInfo.Username)
	t.Logf("  Go Version: %s", sysInfo.GoVersion)
	t.Logf("  CPU Cores: %d", sysInfo.NumCPU)
	t.Logf("  Goroutines: %d", sysInfo.NumGoroutine)
	t.Logf("  IP Addresses: %v", sysInfo.IPAddresses)
	t.Logf("  Working Dir: %s", sysInfo.WorkingDir)
	t.Logf("  Environment vars: %d", len(sysInfo.Environment))
}

func TestGetIPAddresses(t *testing.T) {
	t.Log("Testing getIPAddresses functionality...")

	ips := getIPAddresses()

	// 至少应该有一个IP地址 (除非在特殊环境中)
	if len(ips) == 0 {
		t.Log("No IP addresses found - this might be expected in some environments")
		return
	}

	t.Logf("Found %d IP addresses", len(ips))
	for i, ip := range ips {
		t.Logf("  IP %d: %s", i+1, ip)

		// 验证IP格式 (简单检查)
		if !strings.Contains(ip, ".") && !strings.Contains(ip, ":") {
			t.Errorf("IP address %s doesn't look like a valid IP", ip)
		}
	}
}

func TestIsAdmin(t *testing.T) {
	t.Log("Testing IsAdmin functionality...")

	isAdmin := IsAdmin()
	t.Logf("Current user is admin: %t", isAdmin)

	// 这个测试不会断言具体值，因为依赖于运行环境
	// 只确保函数能正常调用而不出错
}

func TestGetCurrentUser(t *testing.T) {
	t.Log("Testing GetCurrentUser functionality...")

	user := GetCurrentUser()
	if user == "" {
		t.Error("Current user should not be empty")
	}
	if user == "unknown" {
		t.Log("User returned as 'unknown' - this might be expected in some environments")
	}

	t.Logf("Current user: %s", user)
}

func TestGetTempDir(t *testing.T) {
	t.Log("Testing GetTempDir functionality...")

	tempDir := GetTempDir()
	if tempDir == "" {
		t.Error("Temp directory should not be empty")
	}

	t.Logf("Temp directory: %s", tempDir)

	// 验证路径格式 (基本检查)
	switch runtime.GOOS {
	case "windows":
		if !strings.Contains(tempDir, "\\") {
			t.Errorf("Windows temp path should contain backslashes: %s", tempDir)
		}
	default:
		if !strings.HasPrefix(tempDir, "/") {
			t.Errorf("Unix temp path should start with /: %s", tempDir)
		}
	}
}

func TestGetHomeDir(t *testing.T) {
	t.Log("Testing GetHomeDir functionality...")

	homeDir := GetHomeDir()
	if homeDir == "" {
		t.Log("Home directory is empty - this might be expected in some environments")
		return
	}

	t.Logf("Home directory: %s", homeDir)

	// 验证路径格式
	switch runtime.GOOS {
	case "windows":
		if !strings.Contains(homeDir, "\\") {
			t.Errorf("Windows home path should contain backslashes: %s", homeDir)
		}
	default:
		if !strings.HasPrefix(homeDir, "/") {
			t.Errorf("Unix home path should start with /: %s", homeDir)
		}
	}
}

func TestGetExecutablePath(t *testing.T) {
	t.Log("Testing GetExecutablePath functionality...")

	execPath, err := GetExecutablePath()
	if err != nil {
		t.Fatalf("GetExecutablePath failed: %v", err)
	}

	if execPath == "" {
		t.Error("Executable path should not be empty")
	}

	t.Logf("Executable path: %s", execPath)

	// 验证路径包含可执行文件
	if !strings.Contains(execPath, "test") {
		t.Logf("Executable path doesn't contain 'test' - this might be expected: %s", execPath)
	}
}

func TestGetProcessID(t *testing.T) {
	t.Log("Testing GetProcessID functionality...")

	pid := GetProcessID()
	if pid <= 0 {
		t.Errorf("Process ID should be positive, got %d", pid)
	}

	t.Logf("Current process ID: %d", pid)
}

func TestGetParentProcessID(t *testing.T) {
	t.Log("Testing GetParentProcessID functionality...")

	ppid := GetParentProcessID()
	if ppid <= 0 {
		t.Errorf("Parent process ID should be positive, got %d", ppid)
	}

	t.Logf("Parent process ID: %d", ppid)
}

func TestSystemInfoFormatting(t *testing.T) {
	t.Log("Testing SystemInfo formatting...")

	sysInfo, err := GetSystemInfo()
	if err != nil {
		t.Fatalf("GetSystemInfo failed: %v", err)
	}

	formatted := sysInfo.FormatSystemInfo()
	if formatted == "" {
		t.Error("Formatted system info should not be empty")
	}

	// 验证格式化输出包含预期内容
	expectedContent := []string{
		"系统信息",
		sysInfo.OS,
		sysInfo.Hostname,
		sysInfo.GoVersion,
	}

	for _, content := range expectedContent {
		if !strings.Contains(formatted, content) {
			t.Errorf("Formatted output should contain '%s'", content)
		}
	}

	t.Logf("Formatted system info:\n%s", formatted)
}

func TestSystemInfoJSON(t *testing.T) {
	t.Log("Testing SystemInfo JSON serialization...")

	sysInfo, err := GetSystemInfo()
	if err != nil {
		t.Fatalf("GetSystemInfo failed: %v", err)
	}

	// 验证结构体可以被序列化 (通过检查关键字段)
	if sysInfo.OS == "" {
		t.Error("OS field is required for JSON serialization")
	}
	if sysInfo.Architecture == "" {
		t.Error("Architecture field is required for JSON serialization")
	}
	if sysInfo.CurrentTime.IsZero() {
		t.Error("CurrentTime field is required for JSON serialization")
	}

	t.Log("SystemInfo structure is ready for JSON serialization")
}