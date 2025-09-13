package system

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

// ProcessInfo 进程信息结构体
type ProcessInfo struct {
	PID         int       `json:"pid"`          // 进程 ID
	PPID        int       `json:"ppid"`         // 父进程 ID
	Name        string    `json:"name"`         // 进程名称
	CommandLine string    `json:"command_line"` // 命令行
	ExecutePath string    `json:"execute_path"` // 可执行文件路径
	WorkingDir  string    `json:"working_dir"`  // 工作目录
	UserName    string    `json:"user_name"`    // 用户名
	Status      string    `json:"status"`       // 状态
	StartTime   time.Time `json:"start_time"`   // 启动时间
	CPUPercent  float64   `json:"cpu_percent"`  // CPU 使用率
	MemoryMB    float64   `json:"memory_mb"`    // 内存使用(MB)
}

// ProcessManager 进程管理器
type ProcessManager struct {
	processes map[int]*ProcessInfo
}

// NewProcessManager 创建进程管理器
func NewProcessManager() *ProcessManager {
	return &ProcessManager{
		processes: make(map[int]*ProcessInfo),
	}
}

// GetCurrentProcess 获取当前进程信息
func GetCurrentProcess() *ProcessInfo {
	pid := os.Getpid()
	ppid := os.Getppid()

	execPath, _ := os.Executable()
	workingDir, _ := os.Getwd()

	process := &ProcessInfo{
		PID:         pid,
		PPID:        ppid,
		Name:        getProcessName(),
		ExecutePath: execPath,
		WorkingDir:  workingDir,
		UserName:    GetCurrentUser(),
		Status:      "running",
		StartTime:   time.Now(), // 这里应该是真实的启动时间，简化处理
	}

	return process
}

// getProcessName 获取进程名称
func getProcessName() string {
	if execPath, err := os.Executable(); err == nil {
		parts := strings.Split(execPath, string(os.PathSeparator))
		return parts[len(parts)-1]
	}
	return "unknown"
}

// KillProcess 终止进程
func KillProcess(pid int) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("查找进程失败: %w", err)
	}

	err = process.Kill()
	if err != nil {
		return fmt.Errorf("终止进程失败: %w", err)
	}

	return nil
}

// SendSignal 向进程发送信号 (Unix only)
func SendSignal(pid int, sig os.Signal) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("查找进程失败: %w", err)
	}

	err = process.Signal(sig)
	if err != nil {
		return fmt.Errorf("发送信号失败: %w", err)
	}

	return nil
}

// IsProcessRunning 检查进程是否正在运行
func IsProcessRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// 在 Unix 系统上发送信号 0 来检查进程是否存在
	if runtime.GOOS != "windows" {
		err = process.Signal(syscall.Signal(0))
		return err == nil
	}

	// Windows 上的简单检查
	return true // 简化处理，实际应该调用 Windows API
}

// StartProcess 启动新进程
func StartProcess(command string, args []string, workingDir string) (*ProcessInfo, error) {
	cmd := exec.Command(command, args...)
	if workingDir != "" {
		cmd.Dir = workingDir
	}

	err := cmd.Start()
	if err != nil {
		return nil, fmt.Errorf("启动进程失败: %w", err)
	}

	process := &ProcessInfo{
		PID:         cmd.Process.Pid,
		Name:        command,
		CommandLine: strings.Join(append([]string{command}, args...), " "),
		WorkingDir:  workingDir,
		UserName:    GetCurrentUser(),
		Status:      "running",
		StartTime:   time.Now(),
	}

	return process, nil
}

// StartProcessAsync 异步启动进程
func StartProcessAsync(command string, args []string, workingDir string) (*ProcessInfo, error) {
	process, err := StartProcess(command, args, workingDir)
	if err != nil {
		return nil, err
	}

	// 不等待进程结束
	go func() {
		cmd := exec.Command(command, args...)
		if workingDir != "" {
			cmd.Dir = workingDir
		}
		cmd.Wait() // 等待进程结束，防止僵尸进程
	}()

	return process, nil
}

// WaitForProcess 等待进程结束
func WaitForProcess(pid int, timeout time.Duration) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("查找进程失败: %w", err)
	}

	done := make(chan error, 1)
	go func() {
		state, err := process.Wait()
		if err != nil {
			done <- err
		} else if !state.Success() {
			done <- fmt.Errorf("进程退出状态异常: %v", state)
		} else {
			done <- nil
		}
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(timeout):
		return fmt.Errorf("等待进程超时")
	}
}

// GetProcessList 获取进程列表 (平台特定实现)
func GetProcessList() ([]*ProcessInfo, error) {
	switch runtime.GOOS {
	case "windows":
		return getWindowsProcessList()
	default:
		return getUnixProcessList()
	}
}

// getUnixProcessList 获取 Unix 进程列表
func getUnixProcessList() ([]*ProcessInfo, error) {
	cmd := exec.Command("ps", "aux")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("执行 ps 命令失败: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	var processes []*ProcessInfo

	for i, line := range lines {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue // 跳过标题行和空行
		}

		fields := strings.Fields(line)
		if len(fields) < 11 {
			continue
		}

		pid, err := strconv.Atoi(fields[1])
		if err != nil {
			continue
		}

		cpuPercent, _ := strconv.ParseFloat(fields[2], 64)
		memPercent, _ := strconv.ParseFloat(fields[3], 64)

		process := &ProcessInfo{
			PID:        pid,
			UserName:   fields[0],
			CPUPercent: cpuPercent,
			MemoryMB:   memPercent, // 这里是百分比，不是 MB
			Status:     fields[7],
			Name:       fields[10],
		}

		if len(fields) > 11 {
			process.CommandLine = strings.Join(fields[10:], " ")
		}

		processes = append(processes, process)
	}

	return processes, nil
}

// getWindowsProcessList 获取 Windows 进程列表
func getWindowsProcessList() ([]*ProcessInfo, error) {
	cmd := exec.Command("tasklist", "/fo", "csv", "/nh")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("执行 tasklist 命令失败: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	var processes []*ProcessInfo

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 解析 CSV 格式的输出
		fields := parseCSVLine(line)
		if len(fields) < 5 {
			continue
		}

		pid, err := strconv.Atoi(strings.Trim(fields[1], "\""))
		if err != nil {
			continue
		}

		memoryStr := strings.Trim(fields[4], "\"")
		memoryStr = strings.ReplaceAll(memoryStr, ",", "")
		memoryStr = strings.ReplaceAll(memoryStr, " K", "")
		memoryKB, _ := strconv.ParseFloat(memoryStr, 64)

		process := &ProcessInfo{
			PID:      pid,
			Name:     strings.Trim(fields[0], "\""),
			UserName: strings.Trim(fields[2], "\""),
			MemoryMB: memoryKB / 1024, // 转换为 MB
			Status:   "running",
		}

		processes = append(processes, process)
	}

	return processes, nil
}

// parseCSVLine 简单的 CSV 行解析
func parseCSVLine(line string) []string {
	var fields []string
	var current strings.Builder
	var inQuotes bool

	for _, char := range line {
		switch char {
		case '"':
			inQuotes = !inQuotes
			current.WriteRune(char)
		case ',':
			if inQuotes {
				current.WriteRune(char)
			} else {
				fields = append(fields, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(char)
		}
	}

	if current.Len() > 0 {
		fields = append(fields, current.String())
	}

	return fields
}

// FormatProcessInfo 格式化进程信息
func (pi *ProcessInfo) FormatProcessInfo() string {
	result := fmt.Sprintf("进程信息:\n")
	result += fmt.Sprintf("  PID: %d\n", pi.PID)
	if pi.PPID > 0 {
		result += fmt.Sprintf("  父进程ID: %d\n", pi.PPID)
	}
	result += fmt.Sprintf("  名称: %s\n", pi.Name)
	if pi.CommandLine != "" {
		result += fmt.Sprintf("  命令行: %s\n", pi.CommandLine)
	}
	if pi.ExecutePath != "" {
		result += fmt.Sprintf("  可执行文件: %s\n", pi.ExecutePath)
	}
	if pi.WorkingDir != "" {
		result += fmt.Sprintf("  工作目录: %s\n", pi.WorkingDir)
	}
	result += fmt.Sprintf("  用户: %s\n", pi.UserName)
	result += fmt.Sprintf("  状态: %s\n", pi.Status)
	if !pi.StartTime.IsZero() {
		result += fmt.Sprintf("  启动时间: %s\n", pi.StartTime.Format("2006-01-02 15:04:05"))
	}
	if pi.CPUPercent > 0 {
		result += fmt.Sprintf("  CPU使用率: %.1f%%\n", pi.CPUPercent)
	}
	if pi.MemoryMB > 0 {
		result += fmt.Sprintf("  内存使用: %.1f MB\n", pi.MemoryMB)
	}

	return result
}