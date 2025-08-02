package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"os"
	"os/exec"
	"path/filepath"
	"rear/internal/config"
	"rear/pkg/logger"
	"runtime"
	"strings"
	"sync"
	"time"
)

// 全局变量存储工具路径
var (
	ImageMagickPath string
	VipsPath        string
	ExifToolPath    string
	toolsInitOnce   sync.Once
	toolsInitErr    error
)

// ToolConfig 初始化配置（已废弃，使用config.ToolsConfig）
type Config struct {
	// 工具路径（如果为空，会自动检测）
	ImageMagickPath string
	ExifToolPath    string
	VipsPath        string
}

// Initialize 使用配置文件初始化工具路径
func InitializeFromConfig(baseDir *string) error {
	toolsInitOnce.Do(func() {
		// 从全局配置获取工具路径
		toolsConfig := config.CONFIG.ToolsConfig

		ImageMagickPath = toolsConfig.ImageMagickPath
		ExifToolPath = toolsConfig.ExifToolPath
		VipsPath = toolsConfig.LibVipsPath

		toolsInitErr = detectTools(baseDir)
	})

	return toolsInitErr
}

// Initialize 初始化工具路径（保持向后兼容，可选调用，如果不调用会在第一次使用时自动初始化）
func Initialize(config *Config, baseDir *string) error {
	if config == nil {
		config = &Config{}
	}

	toolsInitOnce.Do(func() {
		ImageMagickPath = config.ImageMagickPath
		ExifToolPath = config.ExifToolPath
		VipsPath = config.VipsPath
		toolsInitErr = detectTools(baseDir)
	})

	return toolsInitErr
}

// detectTools 检测工具路径
func detectTools(baseDir *string) error {
	// 获取可执行文件所在目录
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	execDir := filepath.Dir(execPath)
	if baseDir != nil {
		execDir = *baseDir
	}

	// 检测 ImageMagick
	if ImageMagickPath == "" {
		ImageMagickPath = findTool("magick", execDir, "imagemagick")
		if ImageMagickPath == "" {
			ImageMagickPath = findTool("convert", execDir, "imagemagick")
		}
	} else {
		// 验证配置的路径是否有效
		if !validateToolPath(ImageMagickPath) {
			logger.Warn("配置的ImageMagick路径无效，尝试自动检测", zap.String("path", ImageMagickPath))
			ImageMagickPath = findTool("magick", execDir, "imagemagick")
			if ImageMagickPath == "" {
				ImageMagickPath = findTool("convert", execDir, "imagemagick")
			}
		}
	}
	if ImageMagickPath == "" {
		return fmt.Errorf("ImageMagick not found")
	}

	// 检测 ExifTool
	if ExifToolPath == "" {
		ExifToolPath = findTool("exiftool", execDir, "exiftool")
		// 特殊情况：检查 exiftool(-k).exe
		if ExifToolPath == "" {
			ExifToolPath = findTool("exiftool(-k)", execDir, "exiftool")
		}
	} else {
		// 验证配置的路径是否有效
		if !validateToolPath(ExifToolPath) {
			logger.Warn("配置的ExifTool路径无效，尝试自动检测", zap.String("path", ExifToolPath))
			ExifToolPath = findTool("exiftool", execDir, "exiftool")
			// 特殊情况：检查 exiftool(-k).exe
			if ExifToolPath == "" {
				ExifToolPath = findTool("exiftool(-k)", execDir, "exiftool")
			}
		}
	}
	if ExifToolPath == "" {
		return fmt.Errorf("ExifTool not found")
	}

	// 检测 libvips
	if VipsPath == "" {
		VipsPath = findTool("vips", execDir, "vips")
	} else {
		// 验证配置的路径是否有效
		if !validateToolPath(VipsPath) {
			logger.Warn("配置的LibVips路径无效，尝试自动检测", zap.String("path", VipsPath))
			VipsPath = findTool("vips", execDir, "vips")
		}
	}
	if VipsPath == "" {
		return fmt.Errorf("libvips not found")
	}

	logger.Info("工具路径检测完成",
		zap.String("ImageMagick", ImageMagickPath),
		zap.String("ExifTool", ExifToolPath),
		zap.String("LibVips", VipsPath))

	return nil
}

// validateToolPath 验证工具路径是否有效
func validateToolPath(path string) bool {
	if path == "" {
		return false
	}

	// 检查文件是否存在
	if _, err := os.Stat(path); err != nil {
		return false
	}

	return true
}

// findTool 查找工具
func findTool(name string, execDir string, identification string) string {
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
		filepath.Join(execDir, "tools", name, "bin", exeName),
		filepath.Join(execDir, "tools", identification, exeName),
		filepath.Join(execDir, "tools", identification, name, exeName),
		filepath.Join(execDir, "tools", identification, name, "bin", exeName),
		filepath.Join(execDir, identification, exeName),
		filepath.Join(execDir, identification, "bin", exeName),
		filepath.Join(execDir, identification, "tools", exeName),
		filepath.Join(execDir, identification, "tools", name, exeName),
		filepath.Join(execDir, identification, "tools", name, "bin", exeName),
	}
	jsonBytes, _ := json.Marshal(searchPaths)
	msg := string(jsonBytes)
	logger.Debug("查找工具路径", zap.String("tool", name), zap.String("paths", msg))

	for _, path := range searchPaths {
		if _, err := os.Stat(path); err == nil {
			logger.Info("找到工具", zap.String("tool", name), zap.String("path", path))
			return path
		}
	}

	// 尝试从 PATH 中查找
	if path, err := exec.LookPath(exeName); err == nil {
		logger.Info("从系统PATH找到工具", zap.String("tool", name), zap.String("path", path))
		return path
	}

	logger.Warn("未找到工具", zap.String("tool", name))
	return ""
}

// EnsureInitialized 确保工具已初始化
func EnsureInitialized(config *Config, baseDir *string) error {
	if toolsInitErr == nil && ImageMagickPath == "" && ExifToolPath == "" && VipsPath == "" {
		return Initialize(config, baseDir)
	}
	return toolsInitErr
}

// EnsureInitializedFromConfig 从配置文件确保工具已初始化
func EnsureInitializedFromConfig(baseDir *string) error {
	if toolsInitErr == nil && ImageMagickPath == "" && ExifToolPath == "" && VipsPath == "" {
		return InitializeFromConfig(baseDir)
	}
	return toolsInitErr
}

// GetToolPaths 获取当前工具路径
func GetToolPaths() (string, string, string) {
	return ImageMagickPath, VipsPath, ExifToolPath
}

// ToolInfo 工具信息结构
type ToolInfo struct {
	Name    string
	Path    string
	Version string
	Error   error
}

// GetToolsInfo 获取所有工具的详细信息
func GetToolsInfo() []ToolInfo {
	var tools []ToolInfo

	// ImageMagick 信息
	if ImageMagickPath != "" {
		version, err := getImageMagickVersion()
		tools = append(tools, ToolInfo{
			Name:    "ImageMagick",
			Path:    ImageMagickPath,
			Version: version,
			Error:   err,
		})
	} else {
		tools = append(tools, ToolInfo{
			Name:    "ImageMagick",
			Path:    "未找到",
			Version: "N/A",
			Error:   fmt.Errorf("工具未找到"),
		})
	}

	// ExifTool 信息
	if ExifToolPath != "" {
		version, err := getExifToolVersion()
		tools = append(tools, ToolInfo{
			Name:    "ExifTool",
			Path:    ExifToolPath,
			Version: version,
			Error:   err,
		})
	} else {
		tools = append(tools, ToolInfo{
			Name:    "ExifTool",
			Path:    "未找到",
			Version: "N/A",
			Error:   fmt.Errorf("工具未找到"),
		})
	}

	// LibVips 信息
	if VipsPath != "" {
		version, err := getVipsVersion()
		tools = append(tools, ToolInfo{
			Name:    "LibVips",
			Path:    VipsPath,
			Version: version,
			Error:   err,
		})
	} else {
		tools = append(tools, ToolInfo{
			Name:    "LibVips",
			Path:    "未找到",
			Version: "N/A",
			Error:   fmt.Errorf("工具未找到"),
		})
	}

	return tools
}

// getImageMagickVersion 获取ImageMagick版本
func getImageMagickVersion() (string, error) {
	if ImageMagickPath == "" {
		return "", fmt.Errorf("ImageMagick path not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := ExecuteCommand(ctx, ImageMagickPath, "-version")
	if err != nil {
		return "", fmt.Errorf("failed to get ImageMagick version: %w", err)
	}

	// 解析版本信息
	output := string(result.Stdout)
	lines := bytes.Split(result.Stdout, []byte("\n"))
	if len(lines) > 0 {
		firstLine := string(lines[0])
		// 提取版本号 (例如: "Version: ImageMagick 7.1.1-47 Q16-HDRI x64")
		if strings.Contains(firstLine, "Version:") || strings.Contains(firstLine, "ImageMagick") {
			return strings.TrimSpace(firstLine), nil
		}
	}

	return output, nil
}

// getExifToolVersion 获取ExifTool版本
func getExifToolVersion() (string, error) {
	if ExifToolPath == "" {
		return "", fmt.Errorf("ExifTool path not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := ExecuteCommand(ctx, ExifToolPath, "-ver")
	if err != nil {
		return "", fmt.Errorf("failed to get ExifTool version: %w", err)
	}

	version := strings.TrimSpace(string(result.Stdout))
	if version != "" {
		return fmt.Sprintf("ExifTool %s", version), nil
	}

	return string(result.Stdout), nil
}

// getVipsVersion 获取LibVips版本
func getVipsVersion() (string, error) {
	if VipsPath == "" {
		return "", fmt.Errorf("LibVips path not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := ExecuteCommand(ctx, VipsPath, "--version")
	if err != nil {
		return "", fmt.Errorf("failed to get LibVips version: %w", err)
	}

	// 解析版本信息
	output := string(result.Stdout)
	lines := bytes.Split(result.Stdout, []byte("\n"))
	if len(lines) > 0 {
		firstLine := string(lines[0])
		// 提取版本号 (例如: "vips-8.17.0")
		if strings.Contains(firstLine, "vips") {
			return strings.TrimSpace(firstLine), nil
		}
	}

	return output, nil
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
