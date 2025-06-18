package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"rear/pkg/logger"
	"runtime"
	"sync"
	"time"
)

// 全局变量存储工具路径
var (
	ImageMagickPath string
	ExifToolPath    string
	toolsInitOnce   sync.Once
	toolsInitErr    error
)

// Config 初始化配置
type Config struct {
	// 工具路径（如果为空，会自动检测）
	ImageMagickPath string
	ExifToolPath    string
}

// Initialize 初始化工具路径（可选调用，如果不调用会在第一次使用时自动初始化）
func Initialize(config *Config) error {
	if config == nil {
		config = &Config{}
	}

	toolsInitOnce.Do(func() {
		ImageMagickPath = config.ImageMagickPath
		ExifToolPath = config.ExifToolPath
		toolsInitErr = detectTools()
	})

	return toolsInitErr
}

// detectTools 检测工具路径
func detectTools() error {
	// 获取可执行文件所在目录
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	execDir := filepath.Dir(execPath)

	// 检测 ImageMagick
	if ImageMagickPath == "" {
		ImageMagickPath = findTool("convert", execDir)
		if ImageMagickPath == "" {
			ImageMagickPath = findTool("magick", execDir)
		}
	}
	if ImageMagickPath == "" {
		return fmt.Errorf("ImageMagick not found")
	}

	// 检测 ExifTool
	if ExifToolPath == "" {
		ExifToolPath = findTool("exiftool", execDir)
	}
	if ExifToolPath == "" {
		return fmt.Errorf("ExifTool not found")
	}

	return nil
}

// findTool 查找工具
func findTool(name string, execDir string) string {
	// Windows 下添加 .exe 后缀
	exeName := name
	if runtime.GOOS == "windows" {
		exeName = name + ".exe"
	}

	// 查找顺序：
	// 1. 执行文件同目录
	// 2. 执行文件同目录的 bin 子目录
	// 3. 执行文件同目录的 tools 子目录
	// 4. 系统 PATH

	searchPaths := []string{
		filepath.Join(execDir, exeName),
		filepath.Join(execDir, "bin", exeName),
		filepath.Join(execDir, "tools", exeName),
		filepath.Join(execDir, "tools", name, exeName),
	}
	jsonBytes, _ := json.Marshal(searchPaths)
	msg := string(jsonBytes)
	logger.Warn(msg)

	for _, path := range searchPaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// 尝试从 PATH 中查找
	if path, err := exec.LookPath(exeName); err == nil {
		return path
	}

	return ""
}

// EnsureInitialized 确保工具已初始化
func EnsureInitialized() error {
	if toolsInitErr == nil && ImageMagickPath == "" && ExifToolPath == "" {
		return Initialize(nil)
	}
	return toolsInitErr
}

// CommandResult 命令执行结果
type CommandResult struct {
	Stdout   []byte
	Stderr   []byte
	ExitCode int
	Duration time.Duration
}

// ExecuteCommand 执行命令的通用函数
func ExecuteCommand(ctx context.Context, program string, args ...string) (*CommandResult, error) {
	// 创建命令
	cmd := exec.CommandContext(ctx, program, args...)

	// 准备输出缓冲区
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// 记录开始时间
	start := time.Now()

	// 执行命令
	err := cmd.Run()
	duration := time.Since(start)

	result := &CommandResult{
		Stdout:   stdout.Bytes(),
		Stderr:   stderr.Bytes(),
		Duration: duration,
	}

	// 获取退出码
	if exitErr, ok := err.(*exec.ExitError); ok {
		result.ExitCode = exitErr.ExitCode()
	} else if err == nil {
		result.ExitCode = 0
	} else {
		result.ExitCode = -1
	}

	return result, err
}
