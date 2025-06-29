package tools

import (
	"context"
	"fmt"
	"rear/internal/utils"
)

// ImageInfo 图片信息
type ImageInfo struct {
	Format     string
	Width      int
	Height     int
	ColorSpace string
	FileSize   string
}

// ConvertImage 使用 ImageMagick 转换图片
// 示例: ConvertImage(ctx, "input.jpg", "output.png", "-quality", "90")
func ConvertImage(ctx context.Context, input, output string, options ...string) error {
	if err := utils.EnsureInitialized(); err != nil {
		return err
	}

	args := append([]string{input}, options...)
	args = append(args, output)

	result, err := utils.ExecuteCommand(ctx, utils.ImageMagickPath, args...)
	if err != nil {
		return fmt.Errorf("convert failed: %w, stderr: %s", err, string(result.Stderr))
	}

	return nil
}

// ResizeImage 调整图片大小
func ResizeImage(ctx context.Context, input, output string, width, height int) error {
	size := fmt.Sprintf("%dx%d", width, height)
	return ConvertImage(ctx, input, output, "-resize", size)
}

// ResizeImageKeepAspect 按比例调整图片大小（保持宽高比）
func ResizeImageKeepAspect(ctx context.Context, input, output string, maxWidth, maxHeight int) error {
	size := fmt.Sprintf("%dx%d>", maxWidth, maxHeight)
	return ConvertImage(ctx, input, output, "-resize", size)
}

// CropImage 裁剪图片
func CropImage(ctx context.Context, input, output string, width, height, x, y int) error {
	crop := fmt.Sprintf("%dx%d+%d+%d", width, height, x, y)
	return ConvertImage(ctx, input, output, "-crop", crop)
}

// RotateImage 旋转图片
func RotateImage(ctx context.Context, input, output string, degrees float64) error {
	return ConvertImage(ctx, input, output, "-rotate", fmt.Sprintf("%.2f", degrees))
}

// IsImageMagickAvailable 检查 ImageMagick 是否可用
func IsImageMagickAvailable() bool {
	if err := utils.EnsureInitialized(); err != nil {
		return false
	}
	return utils.ImageMagickPath != ""
}
