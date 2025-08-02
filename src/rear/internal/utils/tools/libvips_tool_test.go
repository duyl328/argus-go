package tools

import (
	"context"
	"os"
	"path/filepath"
	"rear/internal/consts"
	"rear/internal/utils"
	"testing"
	"time"
)

// 测试图片路径
const (
	testImageDir = "D:\\go-argus\\rear\\internal\\img"
	outputDir    = "D:\\go-argus\\rear\\internal\\imgtest_output"
)

func TestMain(m *testing.M) {
	// 创建输出目录
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		panic("Failed to create output directory: " + err.Error())
	}

	// 运行测试
	code := m.Run()

	// 清理（可选）
	// os.RemoveAll(outputDir)

	os.Exit(code)
}

func getTestImagePath3(filename string) string {
	return filepath.Join(testImageDir, filename)
}

func getOutputPath(filename string) string {
	return filepath.Join(outputDir, filename)
}

func TestIsVipsAvailable(t *testing.T) {
	available := IsVipsAvailable()
	if !available {
		t.Skip("VIPS is not available, skipping tests")
	}
	t.Log("VIPS is available")
}

func TestVipsVersion(t *testing.T) {
	dir := "D:\\go-argus\\rear\\output\\tools"
	err := utils.Initialize(nil, &dir)
	if err != nil {
		return
	}
	if !IsVipsAvailable() {
		t.Skip("VIPS is not available")
	}

	ctx := context.Background()
	version, err := VipsVersion(ctx)
	if err != nil {
		t.Fatalf("Failed to get VIPS version: %v", err)
	}

	t.Logf("VIPS version: %s", version)
}

func TestGetVipsImageInfo(t *testing.T) {
	dir := "D:\\go-argus\\rear\\output\\tools"
	err := utils.Initialize(nil, &dir)
	if err != nil {
		return
	}
	if !IsVipsAvailable() {
		t.Skip("VIPS is not available")
	}

	ctx := context.Background()

	tests := []struct {
		name     string
		filename string
	}{
		{"WEBP Image", "cyberpunk2077_world_third@3x.webp"},
		{"JPEG Image", "image-1.JPG"},
		{"JPG Image", "img_1.jpg"},
		{"PNG Image", "Snipaste_2022-12-17_20-13-15.png"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputPath := getTestImagePath3(tt.filename)

			// 检查文件是否存在
			if _, err := os.Stat(inputPath); os.IsNotExist(err) {
				t.Skipf("Test image %s not found", inputPath)
			}

			info, err := GetVipsImageInfo(ctx, inputPath)
			if err != nil {
				t.Fatalf("Failed to get image info for %s: %v", tt.filename, err)
			}

			t.Logf("Image: %s", tt.filename)
			t.Logf("  Format: %s", info.Format)
			t.Logf("  Dimensions: %dx%d", info.Width, info.Height)
			t.Logf("  Bands: %d", info.Bands)
			t.Logf("  Has Alpha: %t", info.HasAlpha)
			t.Logf("  Color Space: %s", info.ColorSpace)
			t.Logf("  File Size: %d bytes", info.FileSize)
			t.Logf("  DPI: %.2f", info.DPI)

			// 验证基本信息
			if info.Width <= 0 || info.Height <= 0 {
				t.Errorf("Invalid dimensions: %dx%d", info.Width, info.Height)
			}
			if info.Bands <= 0 {
				t.Errorf("Invalid bands: %d", info.Bands)
			}
		})
	}
}

func TestConvertVipsImage(t *testing.T) {
	if !IsVipsAvailable() {
		t.Skip("VIPS is not available")
	}

	ctx := context.Background()

	tests := []struct {
		name    string
		input   string
		output  string
		format  consts.ImageFormat
		options *VipsConvertOptions
	}{
		{
			name:   "JPEG to PNG",
			input:  "image-1.JPG",
			output: "converted_jpg_to_png.png",
			format: consts.FormatPNG,
			options: &VipsConvertOptions{
				Quality:   85,
				StripMeta: true,
			},
		},
		{
			name:   "PNG to JPEG",
			input:  "Snipaste_2022-12-17_20-13-15.png",
			output: "converted_png_to_jpg.jpg",
			format: consts.FormatJPG,
			options: &VipsConvertOptions{
				Quality:     90,
				StripMeta:   true,
				Progressive: true,
			},
		},
		{
			name:   "JPEG to WEBP",
			input:  "img_1.jpg",
			output: "converted_jpg_to_webp.webp",
			format: consts.FormatWEBP,
			options: &VipsConvertOptions{
				Quality:   80,
				StripMeta: true,
				Lossless:  false,
			},
		},
		{
			name:   "WEBP to PNG",
			input:  "cyberpunk2077_world_third@3x.webp",
			output: "converted_webp_to_png.png",
			format: consts.FormatPNG,
			options: &VipsConvertOptions{
				Quality:   95,
				StripMeta: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputPath := getTestImagePath3(tt.input)
			outputPath := getOutputPath(tt.output)

			// 检查输入文件是否存在
			if _, err := os.Stat(inputPath); os.IsNotExist(err) {
				t.Skipf("Test image %s not found", inputPath)
			}

			err := ConvertVipsImage(ctx, inputPath, outputPath, tt.format, tt.options)
			if err != nil {
				t.Fatalf("Failed to convert image: %v", err)
			}

			// 验证输出文件存在
			if _, err := os.Stat(outputPath); os.IsNotExist(err) {
				t.Errorf("Output file %s was not created", outputPath)
			}

			t.Logf("Successfully converted %s to %s", tt.input, tt.output)
		})
	}
}

func TestResizeVipsImage(t *testing.T) {
	if !IsVipsAvailable() {
		t.Skip("VIPS is not available")
	}

	ctx := context.Background()

	tests := []struct {
		name    string
		input   string
		output  string
		options *VipsResizeOptions
	}{
		{
			name:   "Resize by scale",
			input:  "image-1.JPG",
			output: "resized_scale_0.5.jpg",
			options: &VipsResizeOptions{
				Scale: 0.5,
			},
		},
		{
			name:   "Resize keep aspect ratio",
			input:  "img_1.jpg",
			output: "resized_keep_aspect_800x600.jpg",
			options: &VipsResizeOptions{
				Width:      800,
				Height:     600,
				KeepAspect: true,
				NoEnlarge:  true,
			},
		},
		{
			name:   "Resize fixed dimensions",
			input:  "Snipaste_2022-12-17_20-13-15.png",
			output: "resized_fixed_400x300.png",
			options: &VipsResizeOptions{
				Width:      400,
				Height:     300,
				KeepAspect: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputPath := getTestImagePath3(tt.input)
			outputPath := getOutputPath(tt.output)

			// 检查输入文件是否存在
			if _, err := os.Stat(inputPath); os.IsNotExist(err) {
				t.Skipf("Test image %s not found", inputPath)
			}

			err := ResizeVipsImage(ctx, inputPath, outputPath, tt.options)
			if err != nil {
				t.Fatalf("Failed to resize image: %v", err)
			}

			// 验证输出文件存在
			if _, err := os.Stat(outputPath); os.IsNotExist(err) {
				t.Errorf("Output file %s was not created", outputPath)
			}

			// 验证输出图片信息
			info, err := GetVipsImageInfo(ctx, outputPath)
			if err != nil {
				t.Errorf("Failed to get output image info: %v", err)
			} else {
				t.Logf("Resized image dimensions: %dx%d", info.Width, info.Height)
			}

			t.Logf("Successfully resized %s to %s", tt.input, tt.output)
		})
	}
}

func TestCropVipsImage(t *testing.T) {
	if !IsVipsAvailable() {
		t.Skip("VIPS is not available")
	}

	ctx := context.Background()

	tests := []struct {
		name    string
		input   string
		output  string
		width   int
		height  int
		options *VipsCropOptions
	}{
		{
			name:   "Center crop",
			input:  "image-1.JPG",
			output: "cropped_center_500x500.jpg",
			width:  500,
			height: 500,
			options: &VipsCropOptions{
				Position: VipsCropCenter,
			},
		},
		{
			name:   "Top crop",
			input:  "img_1.jpg",
			output: "cropped_top_400x400.jpg",
			width:  400,
			height: 400,
			options: &VipsCropOptions{
				Position: VipsCropTop,
			},
		},
		{
			name:   "Smart crop",
			input:  "cyberpunk2077_world_third@3x.webp",
			output: "cropped_smart_600x400.webp",
			width:  600,
			height: 400,
			options: &VipsCropOptions{
				Position: VipsCropSmart,
			},
		},
		{
			name:   "Custom crop",
			input:  "Snipaste_2022-12-17_20-13-15.png",
			output: "cropped_custom_300x300.png",
			width:  300,
			height: 300,
			options: &VipsCropOptions{
				Position: VipsCropCustom,
				X:        100,
				Y:        50,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputPath := getTestImagePath3(tt.input)
			outputPath := getOutputPath(tt.output)

			// 检查输入文件是否存在
			if _, err := os.Stat(inputPath); os.IsNotExist(err) {
				t.Skipf("Test image %s not found", inputPath)
			}

			err := CropVipsImage(ctx, inputPath, outputPath, tt.width, tt.height, tt.options)
			if err != nil {
				t.Fatalf("Failed to crop image: %v", err)
			}

			// 验证输出文件存在
			if _, err := os.Stat(outputPath); os.IsNotExist(err) {
				t.Errorf("Output file %s was not created", outputPath)
			}

			// 验证输出图片尺寸
			info, err := GetVipsImageInfo(ctx, outputPath)
			if err != nil {
				t.Errorf("Failed to get output image info: %v", err)
			} else if info.Width != tt.width || info.Height != tt.height {
				t.Errorf("Expected dimensions %dx%d, got %dx%d", tt.width, tt.height, info.Width, info.Height)
			}

			t.Logf("Successfully cropped %s to %s (%dx%d)", tt.input, tt.output, tt.width, tt.height)
		})
	}
}

func TestCompressVipsImage(t *testing.T) {
	if !IsVipsAvailable() {
		t.Skip("VIPS is not available")
	}

	ctx := context.Background()

	tests := []struct {
		name         string
		input        string
		output       string
		format       consts.ImageFormat
		targetWidth  int
		targetHeight int
		quality      int
	}{
		{
			name:         "Compress JPEG",
			input:        "image-1.JPG",
			output:       "compressed_image-1_800x600_q75.jpg",
			format:       consts.FormatJPG,
			targetWidth:  800,
			targetHeight: 600,
			quality:      75,
		},
		{
			name:         "Compress PNG",
			input:        "Snipaste_2022-12-17_20-13-15.png",
			output:       "compressed_screenshot_600x400_q80.png",
			format:       consts.FormatPNG,
			targetWidth:  600,
			targetHeight: 400,
			quality:      80,
		},
		{
			name:         "Compress to WEBP",
			input:        "img_1.jpg",
			output:       "compressed_img1_500x500_q70.webp",
			format:       consts.FormatWEBP,
			targetWidth:  500,
			targetHeight: 500,
			quality:      70,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputPath := getTestImagePath3(tt.input)
			outputPath := getOutputPath(tt.output)

			// 检查输入文件是否存在
			if _, err := os.Stat(inputPath); os.IsNotExist(err) {
				t.Skipf("Test image %s not found", inputPath)
			}

			// 获取原始文件信息
			originalInfo, err := os.Stat(inputPath)
			if err != nil {
				t.Fatalf("Failed to get original file stats: %v", err)
			}

			err = CompressVipsImage(ctx, inputPath, outputPath, tt.format, tt.targetWidth, tt.targetHeight, tt.quality)
			if err != nil {
				t.Fatalf("Failed to compress image: %v", err)
			}

			// 验证输出文件存在
			compressedInfo, err := os.Stat(outputPath)
			if err != nil {
				t.Errorf("Output file %s was not created: %v", outputPath, err)
				return
			}

			// 验证文件大小减小（通常情况下）
			compressionRatio := float64(compressedInfo.Size()) / float64(originalInfo.Size())
			t.Logf("Compression ratio: %.2f (original: %d bytes, compressed: %d bytes)",
				compressionRatio, originalInfo.Size(), compressedInfo.Size())

			t.Logf("Successfully compressed %s to %s", tt.input, tt.output)
		})
	}
}

func TestStripVipsImageMeta(t *testing.T) {
	if !IsVipsAvailable() {
		t.Skip("VIPS is not available")
	}

	ctx := context.Background()

	tests := []struct {
		name   string
		input  string
		output string
	}{
		{
			name:   "Strip JPEG metadata",
			input:  "image-1.JPG",
			output: "stripped_image-1.jpg",
		},
		{
			name:   "Strip PNG metadata",
			input:  "Snipaste_2022-12-17_20-13-15.png",
			output: "stripped_screenshot.png",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputPath := getTestImagePath3(tt.input)
			outputPath := getOutputPath(tt.output)

			// 检查输入文件是否存在
			if _, err := os.Stat(inputPath); os.IsNotExist(err) {
				t.Skipf("Test image %s not found", inputPath)
			}

			err := StripVipsImageMeta(ctx, inputPath, outputPath)
			if err != nil {
				t.Fatalf("Failed to strip metadata: %v", err)
			}

			// 验证输出文件存在
			if _, err := os.Stat(outputPath); os.IsNotExist(err) {
				t.Errorf("Output file %s was not created", outputPath)
			}

			t.Logf("Successfully stripped metadata from %s to %s", tt.input, tt.output)
		})
	}
}

// 测试快捷函数
func TestQuickFunctions(t *testing.T) {
	if !IsVipsAvailable() {
		t.Skip("VIPS is not available")
	}

	ctx := context.Background()

	t.Run("QuickConvertPNGToJPEG", func(t *testing.T) {
		inputPath := getTestImagePath3("Snipaste_2022-12-17_20-13-15.png")
		outputPath := getOutputPath("quick_png_to_jpeg.jpg")

		if _, err := os.Stat(inputPath); os.IsNotExist(err) {
			t.Skip("Test PNG image not found")
		}

		err := QuickConvertPNGToJPEG(ctx, inputPath, outputPath, 85)
		if err != nil {
			t.Fatalf("QuickConvertPNGToJPEG failed: %v", err)
		}

		if _, err := os.Stat(outputPath); os.IsNotExist(err) {
			t.Error("Output file was not created")
		}

		t.Log("QuickConvertPNGToJPEG succeeded")
	})

	t.Run("QuickConvertJPEGToPNG", func(t *testing.T) {
		inputPath := getTestImagePath3("image-1.JPG")
		outputPath := getOutputPath("quick_jpeg_to_png.png")

		if _, err := os.Stat(inputPath); os.IsNotExist(err) {
			t.Skip("Test JPEG image not found")
		}

		err := QuickConvertJPEGToPNG(ctx, inputPath, outputPath)
		if err != nil {
			t.Fatalf("QuickConvertJPEGToPNG failed: %v", err)
		}

		if _, err := os.Stat(outputPath); os.IsNotExist(err) {
			t.Error("Output file was not created")
		}

		t.Log("QuickConvertJPEGToPNG succeeded")
	})

	t.Run("QuickConvertToWEBP", func(t *testing.T) {
		inputPath := getTestImagePath3("img_1.jpg")
		outputPath := getOutputPath("quick_to_webp.webp")

		if _, err := os.Stat(inputPath); os.IsNotExist(err) {
			t.Skip("Test image not found")
		}

		err := QuickConvertToWEBP(ctx, inputPath, outputPath, 80, false)
		if err != nil {
			t.Fatalf("QuickConvertToWEBP failed: %v", err)
		}

		if _, err := os.Stat(outputPath); os.IsNotExist(err) {
			t.Error("Output file was not created")
		}

		t.Log("QuickConvertToWEBP succeeded")
	})

	t.Run("QuickResizeKeepAspect", func(t *testing.T) {
		inputPath := getTestImagePath3("image-2.JPG")
		outputPath := getOutputPath("quick_resize_keep_aspect.jpg")

		if _, err := os.Stat(inputPath); os.IsNotExist(err) {
			t.Skip("Test image not found")
		}

		err := QuickResizeKeepAspect(ctx, inputPath, outputPath, 800, 600)
		if err != nil {
			t.Fatalf("QuickResizeKeepAspect failed: %v", err)
		}

		if _, err := os.Stat(outputPath); os.IsNotExist(err) {
			t.Error("Output file was not created")
		}

		t.Log("QuickResizeKeepAspect succeeded")
	})

	t.Run("QuickCropCenter", func(t *testing.T) {
		inputPath := getTestImagePath3("image-3.JPG")
		outputPath := getOutputPath("quick_crop_center.jpg")

		if _, err := os.Stat(inputPath); os.IsNotExist(err) {
			t.Skip("Test image not found")
		}

		err := QuickCropCenter(ctx, inputPath, outputPath, 400, 400)
		if err != nil {
			t.Fatalf("QuickCropCenter failed: %v", err)
		}

		if _, err := os.Stat(outputPath); os.IsNotExist(err) {
			t.Error("Output file was not created")
		}

		t.Log("QuickCropCenter succeeded")
	})

	t.Run("QuickSmartCrop", func(t *testing.T) {
		inputPath := getTestImagePath3("image-4.JPG")
		outputPath := getOutputPath("quick_smart_crop.jpg")

		if _, err := os.Stat(inputPath); os.IsNotExist(err) {
			t.Skip("Test image not found")
		}

		err := QuickSmartCrop(ctx, inputPath, outputPath, 500, 300)
		if err != nil {
			t.Fatalf("QuickSmartCrop failed: %v", err)
		}

		if _, err := os.Stat(outputPath); os.IsNotExist(err) {
			t.Error("Output file was not created")
		}

		t.Log("QuickSmartCrop succeeded")
	})
}

// 性能测试
func BenchmarkGetVipsImageInfo(b *testing.B) {
	if !IsVipsAvailable() {
		b.Skip("VIPS is not available")
	}

	ctx := context.Background()
	inputPath := getTestImagePath3("image-1.JPG")

	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		b.Skip("Test image not found")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GetVipsImageInfo(ctx, inputPath)
		if err != nil {
			b.Fatalf("GetVipsImageInfo failed: %v", err)
		}
	}
}

func BenchmarkResizeVipsImage(b *testing.B) {
	if !IsVipsAvailable() {
		b.Skip("VIPS is not available")
	}

	ctx := context.Background()
	inputPath := getTestImagePath3("image-1.JPG")

	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		b.Skip("Test image not found")
	}

	options := &VipsResizeOptions{
		Scale: 0.5,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		outputPath := getOutputPath("bench_resize_" + time.Now().Format("20060102150405") + ".jpg")
		err := ResizeVipsImage(ctx, inputPath, outputPath, options)
		if err != nil {
			b.Fatalf("ResizeVipsImage failed: %v", err)
		}
		// 清理输出文件
		os.Remove(outputPath)
	}
}

// 错误处理测试
func TestErrorHandling(t *testing.T) {
	if !IsVipsAvailable() {
		t.Skip("VIPS is not available")
	}

	ctx := context.Background()

	t.Run("Invalid input path", func(t *testing.T) {
		_, err := GetVipsImageInfo(ctx, "nonexistent.jpg")
		if err == nil {
			t.Error("Expected error for nonexistent file")
		}
	})

	t.Run("Invalid resize options", func(t *testing.T) {
		inputPath := getTestImagePath3("image-1.JPG")
		outputPath := getOutputPath("invalid_resize.jpg")

		if _, err := os.Stat(inputPath); os.IsNotExist(err) {
			t.Skip("Test image not found")
		}

		err := ResizeVipsImage(ctx, inputPath, outputPath, nil)
		if err == nil {
			t.Error("Expected error for nil resize options")
		}
	})

	t.Run("Invalid crop dimensions", func(t *testing.T) {
		inputPath := getTestImagePath3("image-1.JPG")
		outputPath := getOutputPath("invalid_crop.jpg")

		if _, err := os.Stat(inputPath); os.IsNotExist(err) {
			t.Skip("Test image not found")
		}

		err := CropVipsImage(ctx, inputPath, outputPath, 0, 0, nil)
		if err == nil {
			t.Error("Expected error for invalid crop dimensions")
		}
	})

	t.Run("Invalid compress quality", func(t *testing.T) {
		inputPath := getTestImagePath3("image-1.JPG")
		outputPath := getOutputPath("invalid_compress.jpg")

		if _, err := os.Stat(inputPath); os.IsNotExist(err) {
			t.Skip("Test image not found")
		}

		err := CompressVipsImage(ctx, inputPath, outputPath, consts.FormatJPG, 100, 100, 101)
		if err == nil {
			t.Error("Expected error for invalid quality value")
		}
	})
}
