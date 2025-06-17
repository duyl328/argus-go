package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"rear/internal/utils"
	"strings"
)

// ExifData EXIF 数据
type ExifData map[string]interface{}

// GetExifData 获取 EXIF 数据
func GetExifData(ctx context.Context, input string) (ExifData, error) {
	if err := utils.EnsureInitialized(); err != nil {
		return nil, err
	}

	result, err := utils.ExecuteCommand(ctx, utils.ExifToolPath, "-j", "-n", input)
	if err != nil {
		return nil, fmt.Errorf("exiftool failed: %w, stderr: %s", err, string(result.Stderr))
	}

	// 解析 JSON 输出
	var data []ExifData
	if err := json.Unmarshal(result.Stdout, &data); err != nil {
		return nil, fmt.Errorf("failed to parse exif data: %w", err)
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("no exif data found")
	}

	return data[0], nil
}

// GetExifField 获取特定的 EXIF 字段
func GetExifField(ctx context.Context, input string, fields ...string) (map[string]string, error) {
	if err := utils.EnsureInitialized(); err != nil {
		return nil, err
	}

	args := []string{"-s", "-s", "-s"}
	for _, field := range fields {
		args = append(args, "-"+field)
	}
	args = append(args, input)

	result, err := utils.ExecuteCommand(ctx, utils.ExifToolPath, args...)
	if err != nil {
		return nil, fmt.Errorf("exiftool failed: %w, stderr: %s", err, string(result.Stderr))
	}

	// 解析输出
	data := make(map[string]string)
	lines := strings.Split(string(result.Stdout), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			// 简单解析，假设最后一个字段是值
			parts := strings.Fields(line)
			if len(parts) > 0 {
				field := fields[len(data)]
				data[field] = parts[len(parts)-1]
			}
		}
	}

	return data, nil
}

// RemoveExifData 移除 EXIF 数据
func RemoveExifData(ctx context.Context, input string, backup bool) error {
	if err := utils.EnsureInitialized(); err != nil {
		return err
	}

	args := []string{"-all="}
	if !backup {
		args = append(args, "-overwrite_original")
	}
	args = append(args, input)

	result, err := utils.ExecuteCommand(ctx, utils.ExifToolPath, args...)
	if err != nil {
		return fmt.Errorf("remove exif failed: %w, stderr: %s", err, string(result.Stderr))
	}

	return nil
}

// CopyExifData 复制 EXIF 数据从一个文件到另一个文件
func CopyExifData(ctx context.Context, source, target string) error {
	if err := utils.EnsureInitialized(); err != nil {
		return err
	}

	args := []string{
		"-TagsFromFile", source,
		"-all:all",
		"-overwrite_original",
		target,
	}

	result, err := utils.ExecuteCommand(ctx, utils.ExifToolPath, args...)
	if err != nil {
		return fmt.Errorf("copy exif failed: %w, stderr: %s", err, string(result.Stderr))
	}

	return nil
}

// SetExifField 设置 EXIF 字段
func SetExifField(ctx context.Context, input string, fields map[string]string) error {
	if err := utils.EnsureInitialized(); err != nil {
		return err
	}

	args := []string{}
	for key, value := range fields {
		args = append(args, fmt.Sprintf("-%s=%s", key, value))
	}
	args = append(args, "-overwrite_original", input)

	result, err := utils.ExecuteCommand(ctx, utils.ExifToolPath, args...)
	if err != nil {
		return fmt.Errorf("set exif failed: %w, stderr: %s", err, string(result.Stderr))
	}

	return nil
}

// IsExifToolAvailable 检查 ExifTool 是否可用
func IsExifToolAvailable() bool {
	if err := utils.EnsureInitialized(); err != nil {
		return false
	}
	return utils.ExifToolPath != ""
}

// GetToolPaths 获取工具路径（用于调试）
func GetToolPaths() (imageMagick, exifTool string) {
	utils.EnsureInitialized()
	return utils.ImageMagickPath, utils.ExifToolPath
}
