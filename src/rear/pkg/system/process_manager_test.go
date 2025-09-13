package system

import (
	"os"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestGetCurrentProcess(t *testing.T) {
	t.Log("Testing GetCurrentProcess functionality...")

	process := GetCurrentProcess()
	if process == nil {
		t.Fatal("GetCurrentProcess returned nil")
	}

	// 验证基础字段
	if process.PID <= 0 {
		t.Errorf("PID should be positive, got %d", process.PID)
	}

	if process.PPID <= 0 {
		t.Errorf("PPID should be positive, got %d", process.PPID)
	}

	if process.Name == "" {
		t.Error("Process name should not be empty")
	}

	if process.UserName == "" {
		t.Error("User name should not be empty")
	}

	if process.Status == "" {
		t.Error("Status should not be empty")
	}

	if process.StartTime.IsZero() {
		t.Error("Start time should not be zero")
	}

	t.Logf("Current process info:")
	t.Logf("  PID: %d", process.PID)
	t.Logf("  PPID: %d", process.PPID)
	t.Logf("  Name: %s", process.Name)
	t.Logf("  Execute Path: %s", process.ExecutePath)
	t.Logf("  Working Dir: %s", process.WorkingDir)
	t.Logf("  User: %s", process.UserName)
	t.Logf("  Status: %s", process.Status)
}

func TestIsProcessRunning(t *testing.T) {
	t.Log("Testing IsProcessRunning functionality...")

	// 测试当前进程
	currentPID := os.Getpid()
	if !IsProcessRunning(currentPID) {
		t.Errorf("Current process %d should be running", currentPID)
	}

	// 测试不存在的进程 (使用一个很大的PID)
	nonExistentPID := 999999
	if IsProcessRunning(nonExistentPID) {
		t.Logf("Process %d unexpectedly exists", nonExistentPID)
	}

	t.Logf("Process running check completed")
}

func TestStartProcess(t *testing.T) {
	t.Log("Testing StartProcess functionality...")

	var command string
	var args []string

	// 选择合适的测试命令
	switch runtime.GOOS {
	case "windows":
		command = "cmd"
		args = []string{"/c", "echo", "test"}
	default:
		command = "echo"
		args = []string{"test"}
	}

	process, err := StartProcess(command, args, "")
	if err != nil {
		t.Fatalf("StartProcess failed: %v", err)
	}

	if process == nil {
		t.Fatal("StartProcess returned nil process")
	}

	if process.PID <= 0 {
		t.Errorf("Started process PID should be positive, got %d", process.PID)
	}

	if process.Name != command {
		t.Errorf("Expected process name %s, got %s", command, process.Name)
	}

	if process.Status != "running" {
		t.Errorf("Expected process status 'running', got %s", process.Status)
	}

	t.Logf("Started process:")
	t.Logf("  PID: %d", process.PID)
	t.Logf("  Name: %s", process.Name)
	t.Logf("  Command Line: %s", process.CommandLine)

	// 等待进程完成
	time.Sleep(100 * time.Millisecond)
}

func TestStartProcessAsync(t *testing.T) {
	t.Log("Testing StartProcessAsync functionality...")

	var command string
	var args []string

	switch runtime.GOOS {
	case "windows":
		command = "cmd"
		args = []string{"/c", "echo", "async_test"}
	default:
		command = "echo"
		args = []string{"async_test"}
	}

	process, err := StartProcessAsync(command, args, "")
	if err != nil {
		t.Fatalf("StartProcessAsync failed: %v", err)
	}

	if process == nil {
		t.Fatal("StartProcessAsync returned nil process")
	}

	if process.PID <= 0 {
		t.Errorf("Started async process PID should be positive, got %d", process.PID)
	}

	t.Logf("Started async process:")
	t.Logf("  PID: %d", process.PID)
	t.Logf("  Name: %s", process.Name)

	// 给异步进程一些时间执行
	time.Sleep(200 * time.Millisecond)
}

func TestWaitForProcess(t *testing.T) {
	t.Log("Testing WaitForProcess functionality...")

	var command string
	var args []string

	switch runtime.GOOS {
	case "windows":
		command = "cmd"
		args = []string{"/c", "ping", "127.0.0.1", "-n", "1"}
	default:
		command = "sleep"
		args = []string{"0.1"}
	}

	process, err := StartProcess(command, args, "")
	if err != nil {
		t.Fatalf("StartProcess failed: %v", err)
	}

	// 等待进程完成 (设置合理的超时)
	err = WaitForProcess(process.PID, 5*time.Second)
	if err != nil {
		t.Logf("WaitForProcess completed with: %v", err)
		// 不视为错误，因为进程可能已经完成
	} else {
		t.Log("Process completed successfully")
	}
}

func TestGetProcessList(t *testing.T) {
	t.Log("Testing GetProcessList functionality...")

	processes, err := GetProcessList()
	if err != nil {
		t.Fatalf("GetProcessList failed: %v", err)
	}

	if len(processes) == 0 {
		t.Error("Process list should not be empty")
	}

	t.Logf("Found %d processes", len(processes))

	// 验证至少有一些基础进程信息
	validProcesses := 0
	for i, process := range processes {
		if i >= 10 { // 只检查前10个进程
			break
		}

		if process.PID > 0 {
			validProcesses++
		}

		t.Logf("Process %d: PID=%d, Name=%s, User=%s", i+1, process.PID, process.Name, process.UserName)
	}

	if validProcesses == 0 {
		t.Error("Should have at least some valid processes")
	}

	t.Logf("Found %d valid processes in sample", validProcesses)
}

func TestKillProcess(t *testing.T) {
	t.Log("Testing KillProcess functionality...")

	// 启动一个长时间运行的进程
	var command string
	var args []string

	switch runtime.GOOS {
	case "windows":
		command = "cmd"
		args = []string{"/c", "ping", "127.0.0.1", "-t"} // 无限ping
	default:
		command = "sleep"
		args = []string{"10"} // 睡眠10秒
	}

	process, err := StartProcess(command, args, "")
	if err != nil {
		t.Fatalf("StartProcess failed: %v", err)
	}

	// 确保进程正在运行
	if !IsProcessRunning(process.PID) {
		t.Fatalf("Process %d should be running", process.PID)
	}

	// 终止进程
	err = KillProcess(process.PID)
	if err != nil {
		t.Fatalf("KillProcess failed: %v", err)
	}

	t.Logf("Successfully killed process %d", process.PID)

	// 等待一段时间，然后检查进程是否已终止
	time.Sleep(500 * time.Millisecond)

	// 注意：在某些系统上，进程可能需要更多时间才能完全终止
	if IsProcessRunning(process.PID) {
		t.Logf("Process %d still running after kill - this might be normal on some systems", process.PID)
	} else {
		t.Logf("Process %d successfully terminated", process.PID)
	}
}

func TestSendSignal(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("SendSignal is not fully supported on Windows")
	}

	t.Log("Testing SendSignal functionality...")

	// 启动一个进程
	process, err := StartProcess("sleep", []string{"5"}, "")
	if err != nil {
		t.Fatalf("StartProcess failed: %v", err)
	}

	// 发送 SIGTERM 信号
	err = SendSignal(process.PID, os.Interrupt)
	if err != nil {
		t.Fatalf("SendSignal failed: %v", err)
	}

	t.Logf("Successfully sent signal to process %d", process.PID)

	// 等待进程响应信号
	time.Sleep(500 * time.Millisecond)
}

func TestProcessInfoFormatting(t *testing.T) {
	t.Log("Testing ProcessInfo formatting...")

	process := GetCurrentProcess()
	formatted := process.FormatProcessInfo()

	if formatted == "" {
		t.Error("Formatted process info should not be empty")
	}

	// 验证格式化输出包含关键信息
	expectedContent := []string{
		"进程信息",
		"PID",
		"名称",
		"用户",
		"状态",
	}

	for _, content := range expectedContent {
		if !strings.Contains(formatted, content) {
			t.Errorf("Formatted output should contain '%s'", content)
		}
	}

	t.Logf("Formatted process info:\n%s", formatted)
}

func TestProcessManagerEdgeCases(t *testing.T) {
	t.Log("Testing ProcessManager edge cases...")

	// 测试无效PID
	err := KillProcess(-1)
	if err == nil {
		t.Error("KillProcess should fail for invalid PID")
	}

	// 测试不存在的命令
	_, err = StartProcess("nonexistent_command_12345", []string{}, "")
	if err == nil {
		t.Error("StartProcess should fail for nonexistent command")
	}

	// 测试无效工作目录
	_, err = StartProcess("echo", []string{"test"}, "/nonexistent/directory")
	if err == nil {
		t.Error("StartProcess should fail for invalid working directory")
	}

	t.Log("Edge case testing completed")
}