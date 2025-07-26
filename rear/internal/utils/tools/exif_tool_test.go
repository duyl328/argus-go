package tools

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

// 测试用的图片文件路径
var testImages = []string{
	"../../img/cyberpunk2077_world_third@3x.webp",
	"../../img/image-1.JPG",
	"../../img/image-1-1.JPG",
	"../../img/image-2.JPG",
	"../../img/image-3.JPG",
	"../../img/image-4.JPG",
	"../../img/Snipaste_2022-12-17_20-13-15.png",
}

// TestIsExifToolAvailable 测试ExifTool是否可用
func TestIsExifToolAvailable(t *testing.T) {
	available := IsExifToolAvailable()
	if !available {
		t.Skip("ExifTool not available, skipping EXIF tests")
	}
	t.Logf("ExifTool is available")
}

// TestGetToolPaths 测试获取工具路径
func TestGetToolPaths(t *testing.T) {
	imageMagick, exifTool := GetToolPaths()
	t.Logf("ImageMagick path: %s", imageMagick)
	t.Logf("ExifTool path: %s", exifTool)

	if exifTool == "" {
		t.Skip("ExifTool path is empty, skipping remaining tests")
	}
}

// TestGetExifData 测试获取EXIF数据
func TestGetExifData(t *testing.T) {
	if !IsExifToolAvailable() {
		t.Skip("ExifTool not available")
	}

	ctx := context.Background()

	for _, imgPath := range testImages {
		t.Run(filepath.Base(imgPath), func(t *testing.T) {
			// 检查文件是否存在
			if _, err := os.Stat(imgPath); os.IsNotExist(err) {
				t.Skipf("Test image %s does not exist", imgPath)
				return
			}

			data, err := GetExifData(ctx, imgPath)
			if err != nil {
				t.Logf("Warning: Failed to get EXIF data for %s: %v", imgPath, err)
				return
			}

			if len(data) == 0 {
				t.Logf("No EXIF data found for %s", imgPath)
				return
			}

			t.Logf("EXIF data for %s:", imgPath)
			for key, value := range data {
				t.Logf("  %s: %v", key, value)
			}

			// 验证常见的EXIF字段
			commonFields := []string{"FileName", "FileSize", "ImageWidth", "ImageHeight"}
			for _, field := range commonFields {
				if _, exists := data[field]; exists {
					t.Logf("Found expected field: %s", field)
				}
			}
		})
	}
}

// TestGetExifDataWithInvalidFile 测试无效文件的情况
func TestGetExifDataWithInvalidFile(t *testing.T) {
	if !IsExifToolAvailable() {
		t.Skip("ExifTool not available")
	}

	ctx := context.Background()

	// 测试不存在的文件
	_, err := GetExifData(ctx, "nonexistent.jpg")
	if err == nil {
		t.Error("Expected error for nonexistent file, but got nil")
	}
	t.Logf("Expected error for nonexistent file: %v", err)

	// 测试空路径
	_, err = GetExifData(ctx, "")
	if err == nil {
		t.Error("Expected error for empty path, but got nil")
	}
	t.Logf("Expected error for empty path: %v", err)
}

// TestGetExifField 测试获取特定EXIF字段
func TestGetExifField(t *testing.T) {
	if !IsExifToolAvailable() {
		t.Skip("ExifTool not available")
	}

	ctx := context.Background()

	// 找一个存在的图片文件进行测试
	var testImg string
	for _, imgPath := range testImages {
		if _, err := os.Stat(imgPath); err == nil {
			testImg = imgPath
			break
		}
	}

	if testImg == "" {
		t.Skip("No test images available")
	}

	t.Run("SingleField", func(t *testing.T) {
		fields, err := GetExifField(ctx, testImg, "ImageWidth")
		if err != nil {
			t.Logf("Failed to get single field: %v", err)
			return
		}
		t.Logf("Single field result: %+v", fields)
	})

	t.Run("MultipleFields", func(t *testing.T) {
		fields, err := GetExifField(ctx, testImg, "ImageWidth", "ImageHeight", "FileSize")
		if err != nil {
			t.Logf("Failed to get multiple fields: %v", err)
			return
		}
		t.Logf("Multiple fields result: %+v", fields)
	})

	t.Run("NonexistentField", func(t *testing.T) {
		fields, err := GetExifField(ctx, testImg, "NonexistentField")
		if err != nil {
			t.Logf("Error getting nonexistent field (expected): %v", err)
		} else {
			t.Logf("Nonexistent field result: %+v", fields)
		}
	})
}

// TestRemoveExifData 测试移除EXIF数据
func TestRemoveExifData(t *testing.T) {
	if !IsExifToolAvailable() {
		t.Skip("ExifTool not available")
	}

	ctx := context.Background()

	// 找一个JPEG文件作为测试目标（因为JPEG通常有EXIF数据）
	var testImg string
	for _, imgPath := range testImages {
		if _, err := os.Stat(imgPath); err == nil && filepath.Ext(imgPath) == ".JPG" {
			testImg = imgPath
			break
		}
	}

	if testImg == "" {
		t.Skip("No JPEG test images available")
	}

	// 创建测试文件的副本
	tempFile := filepath.Join(os.TempDir(), "test_remove_exif.jpg")
	defer os.Remove(tempFile)

	// 复制原文件
	if err := copyFile(testImg, tempFile); err != nil {
		t.Fatalf("Failed to create test file copy: %v", err)
	}

	t.Run("RemoveWithBackup", func(t *testing.T) {
		// 获取移除前的EXIF数据
		beforeData, err := GetExifData(ctx, tempFile)
		if err != nil {
			t.Logf("No EXIF data before removal: %v", err)
		} else {
			t.Logf("EXIF data before removal: %d fields", len(beforeData))
		}

		// 移除EXIF数据（保留备份）
		err = RemoveExifData(ctx, tempFile, true)
		if err != nil {
			t.Fatalf("Failed to remove EXIF data with backup: %v", err)
		}

		// 检查备份文件是否存在
		backupFile := tempFile + "_original"
		if _, err := os.Stat(backupFile); err == nil {
			defer os.Remove(backupFile)
			t.Logf("Backup file created successfully")
		}

		// 获取移除后的EXIF数据
		afterData, err := GetExifData(ctx, tempFile)
		if err != nil {
			t.Logf("No EXIF data after removal (expected): %v", err)
		} else {
			t.Logf("EXIF data after removal: %d fields", len(afterData))
		}
	})

	t.Run("RemoveWithoutBackup", func(t *testing.T) {
		// 重新创建测试文件
		if err := copyFile(testImg, tempFile); err != nil {
			t.Fatalf("Failed to recreate test file: %v", err)
		}

		// 移除EXIF数据（不保留备份）
		err := RemoveExifData(ctx, tempFile, false)
		if err != nil {
			t.Fatalf("Failed to remove EXIF data without backup: %v", err)
		}

		// 检查备份文件不存在
		backupFile := tempFile + "_original"
		if _, err := os.Stat(backupFile); err == nil {
			t.Error("Backup file should not exist when backup=false")
			os.Remove(backupFile)
		}
	})

	t.Run("RemoveFromNonexistentFile", func(t *testing.T) {
		err := RemoveExifData(ctx, "nonexistent.jpg", false)
		if err == nil {
			t.Error("Expected error for nonexistent file, but got nil")
		}
		t.Logf("Expected error for nonexistent file: %v", err)
	})
}

// TestCopyExifData 测试复制EXIF数据
func TestCopyExifData(t *testing.T) {
	if !IsExifToolAvailable() {
		t.Skip("ExifTool not available")
	}

	ctx := context.Background()

	// 找两个不同的图片文件
	var sourceImg, targetImg string
	jpegCount := 0
	for _, imgPath := range testImages {
		if _, err := os.Stat(imgPath); err == nil {
			if filepath.Ext(imgPath) == ".JPG" {
				jpegCount++
				if sourceImg == "" {
					sourceImg = imgPath
				} else if targetImg == "" && imgPath != sourceImg {
					targetImg = imgPath
					break
				}
			}
		}
	}

	if sourceImg == "" || targetImg == "" {
		t.Skip("Need at least 2 JPEG test images")
	}

	// 创建目标文件的副本
	tempTarget := filepath.Join(os.TempDir(), "test_copy_target.jpg")
	defer os.Remove(tempTarget)

	if err := copyFile(targetImg, tempTarget); err != nil {
		t.Fatalf("Failed to create target file copy: %v", err)
	}

	t.Run("CopyValidFiles", func(t *testing.T) {
		// 获取源文件的EXIF数据
		sourceData, err := GetExifData(ctx, sourceImg)
		if err != nil {
			t.Logf("No EXIF data in source file: %v", err)
		}

		// 复制EXIF数据
		err = CopyExifData(ctx, sourceImg, tempTarget)
		if err != nil {
			t.Fatalf("Failed to copy EXIF data: %v", err)
		}

		// 获取目标文件的EXIF数据
		targetData, err := GetExifData(ctx, tempTarget)
		if err != nil {
			t.Logf("No EXIF data in target file after copy: %v", err)
		} else {
			t.Logf("Successfully copied EXIF data, target now has %d fields", len(targetData))

			// 比较一些关键字段
			if sourceData != nil {
				for _, field := range []string{"Make", "Model", "DateTime"} {
					if sourceVal, exists := sourceData[field]; exists {
						if targetVal, exists := targetData[field]; exists {
							if reflect.DeepEqual(sourceVal, targetVal) {
								t.Logf("Field %s copied correctly: %v", field, sourceVal)
							} else {
								t.Logf("Field %s differs: source=%v, target=%v", field, sourceVal, targetVal)
							}
						}
					}
				}
			}
		}
	})

	t.Run("CopyFromNonexistentSource", func(t *testing.T) {
		err := CopyExifData(ctx, "nonexistent.jpg", tempTarget)
		if err == nil {
			t.Error("Expected error for nonexistent source file, but got nil")
		}
		t.Logf("Expected error for nonexistent source: %v", err)
	})

	t.Run("CopyToNonexistentTarget", func(t *testing.T) {
		err := CopyExifData(ctx, sourceImg, "nonexistent.jpg")
		if err == nil {
			t.Error("Expected error for nonexistent target file, but got nil")
		}
		t.Logf("Expected error for nonexistent target: %v", err)
	})
}

// TestSetExifField 测试设置EXIF字段
func TestSetExifField(t *testing.T) {
	if !IsExifToolAvailable() {
		t.Skip("ExifTool not available")
	}

	ctx := context.Background()

	// 找一个JPEG文件作为测试目标
	var testImg string
	for _, imgPath := range testImages {
		if _, err := os.Stat(imgPath); err == nil && filepath.Ext(imgPath) == ".JPG" {
			testImg = imgPath
			break
		}
	}

	if testImg == "" {
		t.Skip("No JPEG test images available")
	}

	// 创建测试文件的副本
	tempFile := filepath.Join(os.TempDir(), "test_set_exif.jpg")
	defer os.Remove(tempFile)

	if err := copyFile(testImg, tempFile); err != nil {
		t.Fatalf("Failed to create test file copy: %v", err)
	}

	t.Run("SetSingleField", func(t *testing.T) {
		fields := map[string]string{
			"Artist": "Test Artist",
		}

		err := SetExifField(ctx, tempFile, fields)
		if err != nil {
			t.Fatalf("Failed to set single field: %v", err)
		}

		// 验证字段是否设置成功
		data, err := GetExifData(ctx, tempFile)
		if err != nil {
			t.Fatalf("Failed to get EXIF data after setting: %v", err)
		}

		if artist, exists := data["Artist"]; exists {
			t.Logf("Artist field set successfully: %v", artist)
		} else {
			t.Error("Artist field not found after setting")
		}
	})

	t.Run("SetMultipleFields", func(t *testing.T) {
		currentTime := time.Now().Format("2006:01:02 15:04:05")
		fields := map[string]string{
			"Artist":   "Test Artist Updated",
			"Comment":  "Test Comment",
			"DateTime": currentTime,
		}

		err := SetExifField(ctx, tempFile, fields)
		if err != nil {
			t.Fatalf("Failed to set multiple fields: %v", err)
		}

		// 验证字段是否设置成功
		data, err := GetExifData(ctx, tempFile)
		if err != nil {
			t.Fatalf("Failed to get EXIF data after setting: %v", err)
		}

		for key, _ := range fields {
			if actualValue, exists := data[key]; exists {
				t.Logf("Field %s set successfully: %v", key, actualValue)
			} else {
				t.Errorf("Field %s not found after setting", key)
			}
		}
	})

	t.Run("SetFieldWithSpecialCharacters", func(t *testing.T) {
		fields := map[string]string{
			"Comment": "测试中文注释 & Special chars: !@#$%",
		}

		err := SetExifField(ctx, tempFile, fields)
		if err != nil {
			t.Logf("Failed to set field with special characters (may be expected): %v", err)
		} else {
			t.Logf("Successfully set field with special characters")
		}
	})

	t.Run("SetFieldOnNonexistentFile", func(t *testing.T) {
		fields := map[string]string{
			"Artist": "Test Artist",
		}

		err := SetExifField(ctx, "nonexistent.jpg", fields)
		if err == nil {
			t.Error("Expected error for nonexistent file, but got nil")
		}
		t.Logf("Expected error for nonexistent file: %v", err)
	})
}

// TestContextCancellation 测试上下文取消
func TestContextCancellation(t *testing.T) {
	if !IsExifToolAvailable() {
		t.Skip("ExifTool not available")
	}

	// 找一个存在的图片文件
	var testImg string
	for _, imgPath := range testImages {
		if _, err := os.Stat(imgPath); err == nil {
			testImg = imgPath
			break
		}
	}

	if testImg == "" {
		t.Skip("No test images available")
	}

	t.Run("CancelledContext", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // 立即取消上下文

		_, err := GetExifData(ctx, testImg)
		if err != nil {
			t.Logf("Operation correctly failed with cancelled context: %v", err)
		} else {
			t.Log("Operation completed despite cancelled context")
		}
	})

	t.Run("TimeoutContext", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		_, err := GetExifData(ctx, testImg)
		if err != nil {
			t.Logf("Operation correctly failed with timeout context: %v", err)
		} else {
			t.Log("Operation completed within timeout")
		}
	})
}

// TestDifferentImageFormats 测试不同图片格式
func TestDifferentImageFormats(t *testing.T) {
	if !IsExifToolAvailable() {
		t.Skip("ExifTool not available")
	}

	ctx := context.Background()

	formatTests := map[string][]string{
		"JPEG": {".JPG", ".jpg", ".jpeg"},
		"PNG":  {".PNG", ".png"},
		"WEBP": {".webp"},
	}

	for format, extensions := range formatTests {
		t.Run(format, func(t *testing.T) {
			found := false
			for _, imgPath := range testImages {
				ext := filepath.Ext(imgPath)
				for _, validExt := range extensions {
					if ext == validExt {
						if _, err := os.Stat(imgPath); err == nil {
							found = true
							data, err := GetExifData(ctx, imgPath)
							if err != nil {
								t.Logf("No EXIF data for %s format (%s): %v", format, imgPath, err)
							} else {
								t.Logf("%s format (%s) has %d EXIF fields", format, imgPath, len(data))
							}
							break
						}
					}
				}
				if found {
					break
				}
			}
			if !found {
				t.Logf("No %s format images found for testing", format)
			}
		})
	}
}

// 辅助函数：复制文件
func copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, input, 0644)
}

// BenchmarkGetExifData 性能测试
func BenchmarkGetExifData(b *testing.B) {
	if !IsExifToolAvailable() {
		b.Skip("ExifTool not available")
	}

	// 找一个存在的图片文件
	var testImg string
	for _, imgPath := range testImages {
		if _, err := os.Stat(imgPath); err == nil {
			testImg = imgPath
			break
		}
	}

	if testImg == "" {
		b.Skip("No test images available")
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GetExifData(ctx, testImg)
		if err != nil {
			b.Fatalf("Benchmark failed: %v", err)
		}
	}
}

// TestConcurrentAccess 并发访问测试
func TestConcurrentAccess(t *testing.T) {
	if !IsExifToolAvailable() {
		t.Skip("ExifTool not available")
	}

	// 找一个存在的图片文件
	var testImg string
	for _, imgPath := range testImages {
		if _, err := os.Stat(imgPath); err == nil {
			testImg = imgPath
			break
		}
	}

	if testImg == "" {
		t.Skip("No test images available")
	}

	ctx := context.Background()
	const numGoroutines = 5

	done := make(chan bool, numGoroutines)
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()

			_, err := GetExifData(ctx, testImg)
			if err != nil {
				errors <- err
				return
			}

			t.Logf("Goroutine %d completed successfully", id)
		}(i)
	}

	// 等待所有goroutine完成
	for i := 0; i < numGoroutines; i++ {
		select {
		case <-done:
			// 成功完成
		case err := <-errors:
			t.Errorf("Concurrent access error: %v", err)
		case <-time.After(30 * time.Second):
			t.Error("Concurrent access test timed out")
		}
	}
}
