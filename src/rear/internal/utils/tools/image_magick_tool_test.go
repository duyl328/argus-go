package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"rear/internal/utils"
	"testing"
	"time"
)

const (
	// 测试图片基础路径
	testImageBasePath = "D:\\go-argus\\rear\\internal\\img"

	// 测试输出目录
	testOutputDir = "test_output"
)

// 测试用的图片文件
var testImages1 = []string{
	"cyberpunk2077_world_third@3x.webp",
	"image-1.JPG",
	"image-1-1.JPG",
	"image-2.JPG",
	"image-3.JPG",
	"image-4.JPG",
	"img_1.jpg",
	"img_2.jpg",
	"Snipaste_2022-12-17_20-13-15.png",
}

// setupTest 设置测试环境
func setupTest(t *testing.T) (string, func()) {
	// 创建测试输出目录
	outputDir := filepath.Join(".", testOutputDir)
	err := os.MkdirAll(outputDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test output directory: %v", err)
	}

	// 返回清理函数
	cleanup := func() {
		os.RemoveAll(outputDir)
	}

	return outputDir, cleanup
}

// getTestImagePath 获取测试图片的完整路径
func getTestImagePath(imageName string) string {
	return filepath.Join(testImageBasePath, imageName)
}

// fileExists 检查文件是否存在
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// TestIsImageMagickAvailable 测试 ImageMagick 是否可用
func TestIsImageMagickAvailable(t *testing.T) {
	available := IsImageMagickAvailable()
	if !available {
		t.Skip("ImageMagick is not available, skipping ImageMagick tests")
	}
	t.Log("ImageMagick is available")
}

// TestConvertImage 测试基础图片转换功能
func TestConvertImage(t *testing.T) {
	dir := "D:\\go-argus\\rear\\output\\tools"
	err := utils.Initialize(nil, &dir)
	if err != nil {
		return
	}
	if !IsImageMagickAvailable() {
		t.Skip("ImageMagick is not available")
	}

	outputDir, cleanup := setupTest(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tests := []struct {
		name    string
		input   string
		output  string
		options []string
		wantErr bool
	}{
		{
			name:    "Convert JPG to PNG",
			input:   getTestImagePath("image-1.JPG"),
			output:  filepath.Join(outputDir, "converted_jpg_to_png.png"),
			options: []string{"-quality", "90"},
			wantErr: false,
		},
		{
			name:    "Convert WebP to JPG",
			input:   getTestImagePath("cyberpunk2077_world_third@3x.webp"),
			output:  filepath.Join(outputDir, "converted_webp_to_jpg.jpg"),
			options: []string{"-quality", "85"},
			wantErr: false,
		},
		{
			name:    "Convert PNG to WebP",
			input:   getTestImagePath("Snipaste_2022-12-17_20-13-15.png"),
			output:  filepath.Join(outputDir, "converted_png_to_webp.webp"),
			options: []string{"-quality", "80"},
			wantErr: false,
		},
		{
			name:    "Invalid input file",
			input:   "nonexistent.jpg",
			output:  filepath.Join(outputDir, "should_not_exist.png"),
			options: []string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 检查输入文件是否存在（除了测试错误情况）
			if !tt.wantErr && !fileExists(tt.input) {
				t.Skipf("Input file %s does not exist", tt.input)
			}

			err := ConvertImage(ctx, tt.input, tt.output, tt.options...)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ConvertImage() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("ConvertImage() error = %v", err)
				return
			}

			// 检查输出文件是否存在
			if !fileExists(tt.output) {
				t.Errorf("Output file %s was not created", tt.output)
			}
		})
	}
}

// TestResizeImage 测试调整图片大小功能
func TestResizeImage(t *testing.T) {
	dir := "D:\\go-argus\\rear\\output\\tools"
	err := utils.Initialize(nil, &dir)
	if err != nil {
		return
	}
	if !IsImageMagickAvailable() {
		t.Skip("ImageMagick is not available")
	}

	outputDir, cleanup := setupTest(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tests := []struct {
		name    string
		input   string
		width   int
		height  int
		wantErr bool
	}{
		{
			name:    "Resize to 800x600",
			input:   getTestImagePath("image-1.JPG"),
			width:   800,
			height:  600,
			wantErr: false,
		},
		{
			name:    "Resize to 1920x1080",
			input:   getTestImagePath("img_1.jpg"),
			width:   1920,
			height:  1080,
			wantErr: false,
		},
		{
			name:    "Resize to small size 100x100",
			input:   getTestImagePath("Snipaste_2022-12-17_20-13-15.png"),
			width:   100,
			height:  100,
			wantErr: false,
		},
		{
			name:    "Invalid dimensions",
			input:   getTestImagePath("image-1.JPG"),
			width:   -100,
			height:  -100,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !fileExists(tt.input) {
				t.Skipf("Input file %s does not exist", tt.input)
			}

			output := filepath.Join(outputDir, fmt.Sprintf("resized_%dx%d_%s", tt.width, tt.height, filepath.Base(tt.input)))

			err := ResizeImage(ctx, tt.input, output, tt.width, tt.height)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ResizeImage() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("ResizeImage() error = %v", err)
				return
			}

			if !fileExists(output) {
				t.Errorf("Output file %s was not created", output)
			}
		})
	}
}

// TestResizeImageKeepAspect 测试按比例调整图片大小功能
func TestResizeImageKeepAspect(t *testing.T) {
	dir := "D:\\go-argus\\rear\\output\\tools"
	err := utils.Initialize(nil, &dir)
	if err != nil {
		return
	}
	if !IsImageMagickAvailable() {
		t.Skip("ImageMagick is not available")
	}

	outputDir, cleanup := setupTest(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tests := []struct {
		name      string
		input     string
		maxWidth  int
		maxHeight int
		wantErr   bool
	}{
		{
			name:      "Keep aspect ratio 1920x1080",
			input:     getTestImagePath("image-2.JPG"),
			maxWidth:  1920,
			maxHeight: 1080,
			wantErr:   false,
		},
		{
			name:      "Keep aspect ratio 800x600",
			input:     getTestImagePath("cyberpunk2077_world_third@3x.webp"),
			maxWidth:  800,
			maxHeight: 600,
			wantErr:   false,
		},
		{
			name:      "Keep aspect ratio thumbnail 200x200",
			input:     getTestImagePath("img_2.jpg"),
			maxWidth:  200,
			maxHeight: 200,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !fileExists(tt.input) {
				t.Skipf("Input file %s does not exist", tt.input)
			}

			output := filepath.Join(outputDir, fmt.Sprintf("resized_aspect_%dx%d_%s", tt.maxWidth, tt.maxHeight, filepath.Base(tt.input)))

			err := ResizeImageKeepAspect(ctx, tt.input, output, tt.maxWidth, tt.maxHeight)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ResizeImageKeepAspect() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("ResizeImageKeepAspect() error = %v", err)
				return
			}

			if !fileExists(output) {
				t.Errorf("Output file %s was not created", output)
			}
		})
	}
}

// TestCropImage 测试裁剪图片功能
func TestCropImage(t *testing.T) {
	dir := "D:\\go-argus\\rear\\output\\tools"
	err := utils.Initialize(nil, &dir)
	if err != nil {
		return
	}
	if !IsImageMagickAvailable() {
		t.Skip("ImageMagick is not available")
	}

	outputDir, cleanup := setupTest(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tests := []struct {
		name    string
		input   string
		width   int
		height  int
		x       int
		y       int
		wantErr bool
	}{
		{
			name:    "Crop center 400x300",
			input:   getTestImagePath("image-3.JPG"),
			width:   400,
			height:  300,
			x:       100,
			y:       100,
			wantErr: false,
		},
		{
			name:    "Crop square 500x500",
			input:   getTestImagePath("image-4.JPG"),
			width:   500,
			height:  500,
			x:       50,
			y:       50,
			wantErr: false,
		},
		{
			name:    "Crop from origin",
			input:   getTestImagePath("Snipaste_2022-12-17_20-13-15.png"),
			width:   200,
			height:  200,
			x:       0,
			y:       0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !fileExists(tt.input) {
				t.Skipf("Input file %s does not exist", tt.input)
			}

			output := filepath.Join(outputDir, fmt.Sprintf("cropped_%dx%d_%s", tt.width, tt.height, filepath.Base(tt.input)))

			err := CropImage(ctx, tt.input, output, tt.width, tt.height, tt.x, tt.y)

			if tt.wantErr {
				if err == nil {
					t.Errorf("CropImage() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("CropImage() error = %v", err)
				return
			}

			if !fileExists(output) {
				t.Errorf("Output file %s was not created", output)
			}
		})
	}
}

// TestRotateImage 测试旋转图片功能
func TestRotateImage(t *testing.T) {
	dir := "D:\\go-argus\\rear\\output\\tools"
	err := utils.Initialize(nil, &dir)
	if err != nil {
		return
	}
	if !IsImageMagickAvailable() {
		t.Skip("ImageMagick is not available")
	}

	outputDir, cleanup := setupTest(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tests := []struct {
		name    string
		input   string
		degrees float64
		wantErr bool
	}{
		{
			name:    "Rotate 90 degrees",
			input:   getTestImagePath("image-1-1.JPG"),
			degrees: 90.0,
			wantErr: false,
		},
		{
			name:    "Rotate 180 degrees",
			input:   getTestImagePath("img_1.jpg"),
			degrees: 180.0,
			wantErr: false,
		},
		{
			name:    "Rotate 270 degrees",
			input:   getTestImagePath("img_2.jpg"),
			degrees: 270.0,
			wantErr: false,
		},
		{
			name:    "Rotate 45 degrees",
			input:   getTestImagePath("Snipaste_2022-12-17_20-13-15.png"),
			degrees: 45.0,
			wantErr: false,
		},
		{
			name:    "Rotate negative degrees",
			input:   getTestImagePath("image-2.JPG"),
			degrees: -90.0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !fileExists(tt.input) {
				t.Skipf("Input file %s does not exist", tt.input)
			}

			output := filepath.Join(outputDir, fmt.Sprintf("rotated_%.0f_%s", tt.degrees, filepath.Base(tt.input)))

			err := RotateImage(ctx, tt.input, output, tt.degrees)

			if tt.wantErr {
				if err == nil {
					t.Errorf("RotateImage() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("RotateImage() error = %v", err)
				return
			}

			if !fileExists(output) {
				t.Errorf("Output file %s was not created", output)
			}
		})
	}
}

// TestContextCancellation1 测试上下文取消功能
func TestContextCancellation1(t *testing.T) {
	if !IsImageMagickAvailable() {
		t.Skip("ImageMagick is not available")
	}

	outputDir, cleanup := setupTest(t)
	defer cleanup()

	// 创建一个会立即取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 立即取消

	input := getTestImagePath("image-1.JPG")
	if !fileExists(input) {
		t.Skip("Test image does not exist")
	}

	output := filepath.Join(outputDir, "should_not_be_created.png")

	err := ConvertImage(ctx, input, output, "-quality", "90")
	if err == nil {
		t.Error("Expected error due to context cancellation, got nil")
	}
}

// TestBatchOperations 测试批量操作
func TestBatchOperations(t *testing.T) {
	if !IsImageMagickAvailable() {
		t.Skip("ImageMagick is not available")
	}

	outputDir, cleanup := setupTest(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 获取所有存在的测试图片
	var existingImages []string
	for _, imgName := range testImages1 {
		imgPath := getTestImagePath(imgName)
		if fileExists(imgPath) {
			existingImages = append(existingImages, imgPath)
		}
	}

	if len(existingImages) == 0 {
		t.Skip("No test images found")
	}

	t.Logf("Processing %d images", len(existingImages))

	// 批量转换为不同格式
	for i, imgPath := range existingImages {
		t.Run(fmt.Sprintf("BatchConvert_%d", i), func(t *testing.T) {
			baseName := filepath.Base(imgPath)
			ext := filepath.Ext(baseName)
			nameWithoutExt := baseName[:len(baseName)-len(ext)]

			// 转换为WebP
			webpOutput := filepath.Join(outputDir, fmt.Sprintf("batch_%s.webp", nameWithoutExt))
			err := ConvertImage(ctx, imgPath, webpOutput, "-quality", "80")
			if err != nil {
				t.Errorf("Failed to convert %s to WebP: %v", baseName, err)
			}

			// 创建缩略图
			thumbnailOutput := filepath.Join(outputDir, fmt.Sprintf("thumbnail_%s.jpg", nameWithoutExt))
			err = ResizeImageKeepAspect(ctx, imgPath, thumbnailOutput, 200, 200)
			if err != nil {
				t.Errorf("Failed to create thumbnail for %s: %v", baseName, err)
			}
		})
	}
}

// BenchmarkConvertImage 性能测试
func BenchmarkConvertImage(b *testing.B) {
	if !IsImageMagickAvailable() {
		b.Skip("ImageMagick is not available")
	}

	outputDir, cleanup := setupTest(&testing.T{})
	defer cleanup()

	input := getTestImagePath("image-1.JPG")
	if !fileExists(input) {
		b.Skip("Test image does not exist")
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		output := filepath.Join(outputDir, fmt.Sprintf("bench_%d.png", i))
		err := ConvertImage(ctx, input, output, "-quality", "90")
		if err != nil {
			b.Fatalf("ConvertImage failed: %v", err)
		}
		// 清理输出文件
		os.Remove(output)
	}
}

// BenchmarkResizeImage 性能测试
func BenchmarkResizeImage(b *testing.B) {
	if !IsImageMagickAvailable() {
		b.Skip("ImageMagick is not available")
	}

	outputDir, cleanup := setupTest(&testing.T{})
	defer cleanup()

	input := getTestImagePath("image-1.JPG")
	if !fileExists(input) {
		b.Skip("Test image does not exist")
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		output := filepath.Join(outputDir, fmt.Sprintf("bench_resize_%d.jpg", i))
		err := ResizeImage(ctx, input, output, 800, 600)
		if err != nil {
			b.Fatalf("ResizeImage failed: %v", err)
		}
		// 清理输出文件
		os.Remove(output)
	}
}
