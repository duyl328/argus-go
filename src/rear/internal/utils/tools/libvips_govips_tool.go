//go:build linux || darwin

package tools

import (
	"context"
	"fmt"
	"os"
	"rear/internal/consts"
	"sync"

	vips "github.com/davidbyttow/govips/v2/vips"
)

var (
	goVipsInitOnce sync.Once
	goVipsInitErr  error
)

// ensureGoVipsInitialized 确保 govips 已初始化（仅在 Linux/macOS 下编译）
func ensureGoVipsInitialized() error {
	goVipsInitOnce.Do(func() {
		// 使用默认配置启动 govips，具体配置可在以后按需扩展
		goVipsInitErr = vips.Startup(nil)
	})
	return goVipsInitErr
}

// IsGoVipsAvailable 检查 govips 是否可用（是否能够成功初始化）
func IsGoVipsAvailable() bool {
	return ensureGoVipsInitialized() == nil
}

// ConvertVipsImageWithGoVips 使用 govips 进行格式转换
// 签名与 ConvertVipsImage 基本一致，方便在调用侧切换实现。
func ConvertVipsImageWithGoVips(ctx context.Context, inputPath, outputPath string, format consts.ImageFormat, options *VipsConvertOptions) error {
	if err := ensureGoVipsInitialized(); err != nil {
		return fmt.Errorf("govips 初始化失败: %w", err)
	}

	if options == nil {
		options = &VipsConvertOptions{Quality: 85}
	}

	image, err := vips.NewImageFromFile(inputPath)
	if err != nil {
		return fmt.Errorf("govips 加载图片失败: %w", err)
	}
	defer image.Close()

	var (
		data []byte
	)

	switch format {
	case consts.FormatJPG:
		params := vips.JpegExportParams{
			Quality:       options.Quality,
			StripMetadata: options.StripMeta,
			Interlace:     options.Progressive,
		}
		data, _, err = image.ExportJpeg(&params)
	case consts.FormatPNG:
		compression := 0
		if options.Quality > 0 {
			// 与 CLI 方式保持一致，将 1-100 映射到 0-9 压缩级别
			compression = 9 - (options.Quality-1)*9/99
		}
		params := vips.PngExportParams{
			Compression:   compression,
			StripMetadata: options.StripMeta,
		}
		data, _, err = image.ExportPng(&params)
	case consts.FormatWEBP:
		params := vips.WebpExportParams{
			Quality:       options.Quality,
			Lossless:      options.Lossless,
			StripMetadata: options.StripMeta,
		}
		data, _, err = image.ExportWebp(&params)
	default:
		return fmt.Errorf("unsupported format for govips: %s", format)
	}

	if err != nil {
		return fmt.Errorf("govips 导出图片失败: %w", err)
	}

	if err := os.WriteFile(outputPath, data, 0o644); err != nil {
		return fmt.Errorf("写出图片失败: %w", err)
	}

	return nil
}

// CompressVipsImageWithGoVips 使用 govips 进行综合压缩（缩放 + 去元数据 + 质量控制 + 格式转换）
// 行为尽可能与 CompressVipsImage 保持一致，但采用 libvips C 接口而非 CLI。
func CompressVipsImageWithGoVips(ctx context.Context, inputPath, outputPath string,
	targetFormat consts.ImageFormat, targetWidth, targetHeight, quality int) error {

	if err := ensureGoVipsInitialized(); err != nil {
		return fmt.Errorf("govips 初始化失败: %w", err)
	}

	if targetWidth <= 0 || targetHeight <= 0 {
		return fmt.Errorf("invalid target dimensions: width and height must be positive")
	}
	if quality < 1 || quality > 100 {
		return fmt.Errorf("invalid quality: must be between 1 and 100")
	}

	image, err := vips.NewImageFromFile(inputPath)
	if err != nil {
		return fmt.Errorf("govips 加载图片失败: %w", err)
	}
	defer image.Close()

	// 计算缩放比例（保持宽高比，且不放大）
	width := image.Width()
	height := image.Height()
	if width <= 0 || height <= 0 {
		return fmt.Errorf("invalid source dimensions: %dx%d", width, height)
	}

	scaleX := float64(targetWidth) / float64(width)
	scaleY := float64(targetHeight) / float64(height)
	scale := scaleX
	if scaleY < scaleX {
		scale = scaleY
	}
	if scale <= 0 {
		return fmt.Errorf("invalid resize scale computed: %f", scale)
	}
	if scale > 1.0 {
		scale = 1.0
	}

	if err := image.Resize(scale, vips.KernelLanczos3); err != nil {
		return fmt.Errorf("govips 缩放图片失败: %w", err)
	}

	// 导出为目标格式
	var data []byte

	switch targetFormat {
	case consts.FormatJPG:
		params := vips.JpegExportParams{
			Quality:       quality,
			StripMetadata: true,
			Interlace:     false,
		}
		data, _, err = image.ExportJpeg(&params)
	case consts.FormatPNG:
		compression := 9 - (quality-1)*9/99
		params := vips.PngExportParams{
			Compression:   compression,
			StripMetadata: true,
		}
		data, _, err = image.ExportPng(&params)
	case consts.FormatWEBP:
		params := vips.WebpExportParams{
			Quality:       quality,
			Lossless:      false,
			StripMetadata: true,
		}
		data, _, err = image.ExportWebp(&params)
	default:
		return fmt.Errorf("unsupported target format for govips: %s", targetFormat)
	}

	if err != nil {
		return fmt.Errorf("govips 导出压缩图片失败: %w", err)
	}

	if err := os.WriteFile(outputPath, data, 0o644); err != nil {
		return fmt.Errorf("写出压缩图片失败: %w", err)
	}

	return nil
}
