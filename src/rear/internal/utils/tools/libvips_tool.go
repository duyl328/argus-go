package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"rear/internal/utils"
	"strconv"
	"strings"
)

// ProcessOptions 图像处理选项
type ProcessOptions struct {
	// 基础选项
	MaxSize int    // 最长边的最大尺寸（保持比例）
	Quality int    // 压缩质量 (1-100)
	Format  string // 输出格式 (jpeg, png, webp, avif等)

	// 高级选项
	Strip      bool   // 是否移除EXIF等元数据
	Interlace  bool   // 是否启用渐进式JPEG
	Optimize   bool   // 是否启用优化
	Background string // 背景色（处理透明图像时）
	Sharpen    bool   // 是否启用锐化
	AutoRotate bool   // 是否根据EXIF自动旋转

	// WebP特定选项
	WebPLossless bool // WebP无损压缩
	WebPEffort   int  // WebP压缩努力程度 (0-6)

	// 尺寸控制
	Width  int  // 固定宽度（如果设置则忽略MaxSize）
	Height int  // 固定高度（如果设置则忽略MaxSize）
	Crop   bool // 是否裁剪以填充指定尺寸
}

// VipsImageInfo vips图片信息
type VipsImageInfo struct {
	Format     string
	Width      int
	Height     int
	Channels   int
	ColorSpace string
	FileSize   string
}

// DefaultOptions 返回默认处理选项
func DefaultOptions() *ProcessOptions {
	return &ProcessOptions{
		MaxSize:      512,
		Quality:      85,
		Format:       "jpeg",
		Strip:        true,
		Interlace:    false,
		Optimize:     true,
		Background:   "white",
		Sharpen:      false,
		AutoRotate:   true,
		WebPLossless: false,
		WebPEffort:   4,
		Width:        0,
		Height:       0,
		Crop:         false,
	}
}

// ProcessImageWithVips 使用 libvips 处理图像
// 示例: ProcessImageWithVips(ctx, "input.jpg", "output.webp", DefaultOptions())
func ProcessImageWithVips(ctx context.Context, inputPath, outputPath string, options *ProcessOptions) error {
	if err := utils.EnsureInitialized(); err != nil {
		return err
	}

	if options == nil {
		options = DefaultOptions()
	}

	// 检查输入文件是否存在
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return fmt.Errorf("input file not found: %s", inputPath)
	}

	// 确保输出目录存在
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// 构建vips命令参数
	args := buildVipsArgs(inputPath, outputPath, options)

	// 执行命令
	result, err := utils.ExecuteCommand(ctx, utils.VipsPath, args...)
	if err != nil {
		return fmt.Errorf("vips command failed: %w, stderr: %s", err, string(result.Stderr))
	}

	return nil
}

// ThumbnailImage 生成缩略图（使用vips thumbnail命令）
func ThumbnailImage(ctx context.Context, input, output string, size int) error {
	return ProcessImageWithVips(ctx, input, output, &ProcessOptions{
		MaxSize: size,
		Quality: 85,
		Strip:   true,
	})
}

// ResizeImageWithVips 使用vips调整图片大小
func ResizeImageWithVips(ctx context.Context, input, output string, width, height int) error {
	return ProcessImageWithVips(ctx, input, output, &ProcessOptions{
		Width:   width,
		Height:  height,
		Quality: 85,
		Strip:   true,
	})
}

// ResizeImageKeepAspectVips 使用vips按比例调整图片大小（保持宽高比）
func ResizeImageKeepAspectVips(ctx context.Context, input, output string, maxSize int) error {
	return ProcessImageWithVips(ctx, input, output, &ProcessOptions{
		MaxSize: maxSize,
		Quality: 85,
		Strip:   true,
	})
}

// ConvertImageFormat 转换图像格式
func ConvertImageFormat(ctx context.Context, input, output, format string) error {
	return ProcessImageWithVips(ctx, input, output, &ProcessOptions{
		Format:  format,
		Quality: 85,
		Strip:   true,
	})
}

// CompressImage 压缩图像
func CompressImage(ctx context.Context, input, output string, quality int) error {
	return ProcessImageWithVips(ctx, input, output, &ProcessOptions{
		Quality:  quality,
		Strip:    true,
		Optimize: true,
	})
}

// GetVipsImageInfo 获取图像信息
func GetVipsImageInfo(ctx context.Context, imagePath string) (*VipsImageInfo, error) {
	if err := utils.EnsureInitialized(); err != nil {
		return nil, err
	}

	result, err := utils.ExecuteCommand(ctx, utils.VipsPath, "identify", imagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get image info: %w", err)
	}

	// 解析vips identify输出
	info := &VipsImageInfo{}
	lines := strings.Split(string(result.Stdout), "\n")
	for _, line := range lines {
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				switch key {
				case "format":
					info.Format = value
				case "width":
					if w, err := strconv.Atoi(value); err == nil {
						info.Width = w
					}
				case "height":
					if h, err := strconv.Atoi(value); err == nil {
						info.Height = h
					}
				case "channels":
					if c, err := strconv.Atoi(value); err == nil {
						info.Channels = c
					}
				case "interpretation":
					info.ColorSpace = value
				}
			}
		}
	}

	// 获取文件大小
	if stat, err := os.Stat(imagePath); err == nil {
		info.FileSize = fmt.Sprintf("%d bytes", stat.Size())
	}

	return info, nil
}

// BatchProcessImages 批量处理图像
func BatchProcessImages(ctx context.Context, inputDir, outputDir string, options *ProcessOptions) error {
	if options == nil {
		options = DefaultOptions()
	}

	// 支持的图像格式
	supportedExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
		".tiff": true,
		".bmp":  true,
		".tif":  true,
	}

	return filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if !supportedExts[ext] {
			return nil
		}

		// 计算相对路径
		relPath, err := filepath.Rel(inputDir, path)
		if err != nil {
			return err
		}

		// 构建输出路径
		outputPath := filepath.Join(outputDir, relPath)

		// 根据输出格式修改扩展名
		if options.Format != "" {
			outputPath = strings.TrimSuffix(outputPath, filepath.Ext(outputPath)) + "." + options.Format
		}

		return ProcessImageWithVips(ctx, path, outputPath, options)
	})
}

// IsVipsAvailable 检查 libvips 是否可用
func IsVipsAvailable() bool {
	if err := utils.EnsureInitialized(); err != nil {
		return false
	}
	return utils.VipsPath != ""
}

// buildVipsArgs 构建vips命令参数
func buildVipsArgs(inputPath, outputPath string, options *ProcessOptions) []string {
	args := []string{"thumbnail"}

	// 输入文件
	args = append(args, inputPath)

	// 输出文件
	args = append(args, outputPath)

	// 尺寸参数
	if options.Width > 0 && options.Height > 0 {
		if options.Crop {
			// 裁剪模式：填充指定尺寸
			args = append(args, fmt.Sprintf("%dx%d^", options.Width, options.Height))
			args = append(args, "--crop", "centre")
		} else {
			// 适应模式：保持比例
			args = append(args, fmt.Sprintf("%dx%d", options.Width, options.Height))
		}
	} else if options.Width > 0 {
		args = append(args, strconv.Itoa(options.Width))
	} else if options.Height > 0 {
		args = append(args, fmt.Sprintf("x%d", options.Height))
	} else {
		// 使用MaxSize（最长边）
		args = append(args, strconv.Itoa(options.MaxSize))
	}

	// 质量设置
	if options.Quality > 0 && options.Quality <= 100 {
		args = append(args, "--quality", strconv.Itoa(options.Quality))
	}

	// 格式转换
	if options.Format != "" {
		args = append(args, "--format", options.Format)
	}

	// 移除元数据
	if options.Strip {
		args = append(args, "--strip")
	}

	// 渐进式JPEG
	if options.Interlace {
		args = append(args, "--interlace")
	}

	// 优化
	if options.Optimize {
		args = append(args, "--optimize")
	}

	// 背景色
	if options.Background != "" {
		args = append(args, "--background", options.Background)
	}

	// 锐化
	if options.Sharpen {
		args = append(args, "--sharpen", "1")
	}

	// 自动旋转
	if options.AutoRotate {
		args = append(args, "--auto-rotate")
	}

	// WebP特定选项
	if strings.ToLower(options.Format) == "webp" {
		if options.WebPLossless {
			args = append(args, "--webp-lossless")
		}
		if options.WebPEffort >= 0 && options.WebPEffort <= 6 {
			args = append(args, "--webp-effort", strconv.Itoa(options.WebPEffort))
		}
	}

	return args
}
