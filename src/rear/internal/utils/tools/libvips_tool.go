package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"rear/internal/consts"
	"rear/internal/utils"
	"strconv"
	"strings"
)

// VipsImageInfo VIPS 图片信息
type VipsImageInfo struct {
	Format     string  `json:"format"`
	Width      int     `json:"width"`
	Height     int     `json:"height"`
	Bands      int     `json:"bands"`      // 通道数
	HasAlpha   bool    `json:"has_alpha"`  // 是否有透明通道
	ColorSpace string  `json:"colorspace"` // 颜色空间
	FileSize   int64   `json:"file_size"`  // 文件大小（字节）
	DPI        float64 `json:"dpi"`        // DPI
}

// VipsCropPosition 裁剪位置枚举
type VipsCropPosition string

const (
	VipsCropTop    VipsCropPosition = "top"
	VipsCropCenter VipsCropPosition = "center"
	VipsCropBottom VipsCropPosition = "bottom"
	VipsCropSmart  VipsCropPosition = "smart"  // 智能裁剪
	VipsCropCustom VipsCropPosition = "custom" // 自定义位置
)

// VipsConvertOptions 转换选项
type VipsConvertOptions struct {
	Quality     int  // 质量 (1-100)
	StripMeta   bool // 删除元数据
	Progressive bool // 渐进式 (仅JPEG)
	Lossless    bool // 无损压缩 (仅WEBP)
}

// VipsResizeOptions 调整大小选项
type VipsResizeOptions struct {
	Width      int     // 目标宽度
	Height     int     // 目标高度
	KeepAspect bool    // 保持宽高比
	Scale      float64 // 缩放比例 (优先级高于宽高)
	NoEnlarge  bool    // 不放大图片
}

// VipsCropOptions 裁剪选项
type VipsCropOptions struct {
	Position VipsCropPosition // 裁剪位置
	X        int              // 自定义裁剪起始X坐标
	Y        int              // 自定义裁剪起始Y坐标
}

// GetVipsImageInfo 获取图片基础信息
func GetVipsImageInfo(ctx context.Context, inputPath string) (*VipsImageInfo, error) {
	if err := utils.EnsureInitialized(nil, nil); err != nil {
		return nil, err
	}
	// 构建 vipsheader 的路径（跨平台兼容）
	vipsHeaderPath := filepath.Join(filepath.Dir(utils.VipsPath), "vipsheader")

	// 执行 vipsheader -a 命令
	result, err := utils.ExecuteCommand(ctx, vipsHeaderPath, "-a", inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get image info: %w, stderr: %s", err, string(result.Stderr))
	}

	output := string(result.Stdout)
	lines := strings.Split(strings.TrimSpace(output), "\n")

	info := &VipsImageInfo{}

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) != 2 {
				continue
			}
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			switch key {
			case "width":
				if w, err := strconv.Atoi(value); err == nil {
					info.Width = w
				}
			case "height":
				if h, err := strconv.Atoi(value); err == nil {
					info.Height = h
				}
			case "bands":
				if b, err := strconv.Atoi(value); err == nil {
					info.Bands = b
				}
			case "format":
				info.Format = value
			case "interpretation":
				info.ColorSpace = value
			case "alpha":
				info.HasAlpha = (value == "yes")
			}
		}
	}

	return info, nil
}

// ConvertVipsImage 使用 VIPS 转换图片格式
func ConvertVipsImage(ctx context.Context, inputPath, outputPath string, format consts.ImageFormat, options *VipsConvertOptions) error {
	if err := utils.EnsureInitialized(nil, nil); err != nil {
		return err
	}

	if options == nil {
		options = &VipsConvertOptions{Quality: 85}
	}

	var args []string

	switch format {
	case consts.FormatJPG:
		args = append(args, "jpegsave", inputPath, outputPath)
		if options.Quality > 0 {
			args = append(args, "--Q", strconv.Itoa(options.Quality))
		}
		if options.Progressive {
			args = append(args, "--interlace")
		}
	case consts.FormatPNG:
		args = append(args, "pngsave", inputPath, outputPath)
		if options.Quality > 0 {
			compression := 9 - (options.Quality-1)*9/99 // 转换为0-9范围
			args = append(args, "--compression", strconv.Itoa(compression))
		}
	case consts.FormatWEBP:
		args = append(args, "webpsave", inputPath, outputPath)
		if options.Quality > 0 {
			args = append(args, "--Q", strconv.Itoa(options.Quality))
		}
		if options.Lossless {
			args = append(args, "--lossless")
		}
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	// 删除元数据
	if options.StripMeta {
		args = append(args, "--strip")
	}

	result, err := utils.ExecuteCommand(ctx, utils.VipsPath, args...)
	if err != nil {
		return fmt.Errorf("vips convert failed: %w, stderr: %s", err, string(result.Stderr))
	}

	return nil
}

// ResizeVipsImage 调整图片大小
func ResizeVipsImage(ctx context.Context, inputPath, outputPath string, options *VipsResizeOptions) error {
	if err := utils.EnsureInitialized(nil, nil); err != nil {
		return err
	}

	if options == nil {
		return fmt.Errorf("resize options cannot be nil")
	}

	var args []string

	// 如果指定了缩放比例，优先使用缩放
	if options.Scale > 0 {
		args = []string{"resize", inputPath, outputPath, strconv.FormatFloat(options.Scale, 'f', 2, 64)}
	} else {
		// 获取原图信息以计算缩放比例
		info, err := GetVipsImageInfo(ctx, inputPath)
		if err != nil {
			return fmt.Errorf("failed to get image info: %w", err)
		}

		var scale float64
		if options.KeepAspect {
			// 保持宽高比，计算最小缩放比例
			scaleX := float64(options.Width) / float64(info.Width)
			scaleY := float64(options.Height) / float64(info.Height)

			if scaleX > 0 && scaleY > 0 {
				scale = scaleX
				if scaleY < scaleX {
					scale = scaleY
				}
			} else if scaleX > 0 {
				scale = scaleX
			} else if scaleY > 0 {
				scale = scaleY
			} else {
				return fmt.Errorf("invalid resize dimensions")
			}
		} else {
			// 直接使用指定的宽高（可能会变形）
			if options.Width > 0 && options.Height > 0 {
				_ = float64(options.Width) / float64(info.Width)
				_ = float64(options.Height) / float64(info.Height)
				// 使用 thumbnail 命令可以指定具体尺寸
				args = []string{"thumbnail", inputPath, outputPath,
					fmt.Sprintf("%dx%d!", options.Width, options.Height)}
			} else {
				return fmt.Errorf("width and height must be specified when not keeping aspect ratio")
			}
		}

		if len(args) == 0 {
			// 检查是否需要防止放大
			if options.NoEnlarge && scale > 1.0 {
				scale = 1.0
			}
			args = []string{"resize", inputPath, outputPath, strconv.FormatFloat(scale, 'f', 2, 64)}
		}
	}

	result, err := utils.ExecuteCommand(ctx, utils.VipsPath, args...)
	if err != nil {
		return fmt.Errorf("vips resize failed: %w, stderr: %s", err, string(result.Stderr))
	}

	return nil
}

// CropVipsImage 裁剪图片
func CropVipsImage(ctx context.Context, inputPath, outputPath string, width, height int, options *VipsCropOptions) error {
	if err := utils.EnsureInitialized(nil, nil); err != nil {
		return err
	}

	if width <= 0 || height <= 0 {
		return fmt.Errorf("invalid crop dimensions: width and height must be positive")
	}

	if options == nil {
		options = &VipsCropOptions{Position: VipsCropCenter}
	}

	var args []string

	switch options.Position {
	case VipsCropSmart:
		// 智能裁剪，使用 smartcrop
		args = []string{"smartcrop", inputPath, outputPath,
			strconv.Itoa(width), strconv.Itoa(height)}
	case VipsCropCustom:
		// 自定义位置裁剪
		args = []string{"crop", inputPath, outputPath,
			strconv.Itoa(options.X), strconv.Itoa(options.Y),
			strconv.Itoa(width), strconv.Itoa(height)}
	default:
		// 使用 thumbnail 进行位置裁剪
		var gravity string
		switch options.Position {
		case VipsCropTop:
			gravity = "attention" // top -> attention
		case VipsCropBottom:
			gravity = "low" // bottom -> low
		default:
			gravity = "centre"
		}

		args = []string{"thumbnail", inputPath, outputPath,
			fmt.Sprintf("%dx%d^", width, height),
			"--crop", gravity}

	}

	result, err := utils.ExecuteCommand(ctx, utils.VipsPath, args...)
	if err != nil {
		return fmt.Errorf("vips crop failed: %w, stderr: %s", err, string(result.Stderr))
	}

	return nil
}

// CompressVipsImage 综合压缩图片（删除EXIF，调整大小，质量压缩，格式转换）
func CompressVipsImage(ctx context.Context, inputPath, outputPath string,
	targetFormat consts.ImageFormat, targetWidth, targetHeight, quality int) error {

	if err := utils.EnsureInitialized(nil, nil); err != nil {
		return err
	}

	if targetWidth <= 0 || targetHeight <= 0 {
		return fmt.Errorf("invalid target dimensions: width and height must be positive")
	}
	if quality < 1 || quality > 100 {
		return fmt.Errorf("invalid quality: must be between 1 and 100")
	}

	// 创建一个中间临时文件
	ext := filepath.Ext(outputPath)
	tmpOutput := strings.TrimSuffix(outputPath, ext) + ".tmp" + ext

	// 第一步：缩略图处理（不带 strip）
	argsThumb := []string{
		"thumbnail", inputPath, tmpOutput,
		fmt.Sprintf("%dx%d>", targetWidth, targetHeight),
	}

	if result, err := utils.ExecuteCommand(ctx, utils.VipsPath, argsThumb...); err != nil {
		return fmt.Errorf("vips thumbnail failed: %w, stderr: %s", err, string(result.Stderr))
	}

	// 第二步：拷贝并压缩，带 strip 和质量设置
	var argsCopy []string
	switch targetFormat {
	case consts.FormatJPG:
		argsCopy = append(argsCopy, "jpegsave", tmpOutput, outputPath)
		if quality > 0 {
			argsCopy = append(argsCopy, "--Q", strconv.Itoa(quality))
		}
		argsCopy = append(argsCopy, "--strip")

	case consts.FormatPNG:
		argsCopy = append(argsCopy, "pngsave", tmpOutput, outputPath)
		compression := 9 - (quality-1)*9/99
		argsCopy = append(argsCopy, "--compression", strconv.Itoa(compression))
		argsCopy = append(argsCopy, "--strip")

	case consts.FormatWEBP:
		argsCopy = append(argsCopy, "webpsave", tmpOutput, outputPath)
		if quality > 0 {
			argsCopy = append(argsCopy, "--Q", strconv.Itoa(quality))
		}
		argsCopy = append(argsCopy, "--strip")

	default:
		return fmt.Errorf("unsupported target format: %s", targetFormat)
	}

	if result, err := utils.ExecuteCommand(ctx, utils.VipsPath, argsCopy...); err != nil {
		return fmt.Errorf("vips copy failed: %w, stderr: %s", err, string(result.Stderr))
	}

	// 可选：删除中间文件（如果你没有自动管理临时文件）
	_ = os.Remove(tmpOutput)

	return nil
}

// StripVipsImageMeta 删除图片元数据
func StripVipsImageMeta(ctx context.Context, inputPath, outputPath string) error {
	if err := utils.EnsureInitialized(nil, nil); err != nil {
		return err
	}

	ext := strings.ToLower(filepath.Ext(outputPath))

	var args []string
	switch ext {
	case ".jpg", ".jpeg":
		args = []string{"jpegsave", inputPath, outputPath, "--strip"}
	case ".png":
		args = []string{"pngsave", inputPath, outputPath, "--strip"}
	case ".webp":
		args = []string{"webpsave", inputPath, outputPath, "--strip"}
	default:
		return fmt.Errorf("unsupported output format for stripping metadata: %s", ext)
	}

	result, err := utils.ExecuteCommand(ctx, utils.VipsPath, args...)
	if err != nil {
		return fmt.Errorf("vips strip meta failed: %w, stderr: %s", err, string(result.Stderr))
	}

	return nil
}

// IsVipsAvailable 检查 VIPS 是否可用
func IsVipsAvailable() bool {
	if err := utils.EnsureInitialized(nil, nil); err != nil {
		return false
	}
	return utils.VipsPath != ""
}

// VipsVersion 获取 VIPS 版本信息
func VipsVersion(ctx context.Context) (string, error) {
	if err := utils.EnsureInitialized(nil, nil); err != nil {
		return "", err
	}

	result, err := utils.ExecuteCommand(ctx, utils.VipsPath, "--version")
	if err != nil {
		return "", fmt.Errorf("failed to get vips version: %w", err)
	}

	return strings.TrimSpace(string(result.Stdout)), nil
}

// 快捷函数

// QuickConvertPNGToJPEG 快速将PNG转换为JPEG
func QuickConvertPNGToJPEG(ctx context.Context, inputPath, outputPath string, quality int) error {
	options := &VipsConvertOptions{
		Quality:   quality,
		StripMeta: true,
	}
	return ConvertVipsImage(ctx, inputPath, outputPath, consts.FormatJPG, options)
}

// QuickConvertJPEGToPNG 快速将JPEG转换为PNG
func QuickConvertJPEGToPNG(ctx context.Context, inputPath, outputPath string) error {
	options := &VipsConvertOptions{
		StripMeta: true,
	}
	return ConvertVipsImage(ctx, inputPath, outputPath, consts.FormatPNG, options)
}

// QuickConvertToWEBP 快速转换为WEBP格式
func QuickConvertToWEBP(ctx context.Context, inputPath, outputPath string, quality int, lossless bool) error {
	options := &VipsConvertOptions{
		Quality:   quality,
		StripMeta: true,
		Lossless:  lossless,
	}
	return ConvertVipsImage(ctx, inputPath, outputPath, consts.FormatWEBP, options)
}

// QuickResizeKeepAspect 快速等比例调整大小
func QuickResizeKeepAspect(ctx context.Context, inputPath, outputPath string, maxWidth, maxHeight int) error {
	options := &VipsResizeOptions{
		Width:      maxWidth,
		Height:     maxHeight,
		KeepAspect: true,
		NoEnlarge:  true,
	}
	return ResizeVipsImage(ctx, inputPath, outputPath, options)
}

// QuickCropCenter 快速居中裁剪
func QuickCropCenter(ctx context.Context, inputPath, outputPath string, width, height int) error {
	options := &VipsCropOptions{
		Position: VipsCropCenter,
	}
	return CropVipsImage(ctx, inputPath, outputPath, width, height, options)
}

// QuickSmartCrop 快速智能裁剪
func QuickSmartCrop(ctx context.Context, inputPath, outputPath string, width, height int) error {
	options := &VipsCropOptions{
		Position: VipsCropSmart,
	}
	return CropVipsImage(ctx, inputPath, outputPath, width, height, options)
}

// TileGenerationOptions 瓦片生成选项
type TileGenerationOptions struct {
	TileSize   int    // 瓦片大小（正方形边长，默认256）
	Overlap    int    // 瓦片重叠像素（默认0）
	Layout     string // 布局方式: "dz" (DeepZoom), "google", "zoomify"
	Quality    int    // JPEG质量（1-100）
	OutputDir  string // 输出目录
	Background string // 背景颜色（用于填充不完整的瓦片）
}

// GenerateTiles 生成图片瓦片（用于超大图的渐进式加载）
// 使用 Deep Zoom Image (DZI) 格式
func GenerateTiles(ctx context.Context, inputPath string, options *TileGenerationOptions) error {
	if err := utils.EnsureInitialized(nil, nil); err != nil {
		return err
	}

	if options == nil {
		options = &TileGenerationOptions{
			TileSize: 256,
			Quality:  85,
			Layout:   "dz",
		}
	}

	// 设置默认值
	if options.TileSize <= 0 {
		options.TileSize = 256
	}
	if options.Quality <= 0 {
		options.Quality = 85
	}
	if options.Layout == "" {
		options.Layout = "dz"
	}

	// 构建vips dzsave命令参数
	args := []string{"dzsave", inputPath, options.OutputDir}

	// 添加瓦片大小参数
	args = append(args, "--tile-size", strconv.Itoa(options.TileSize))

	// 添加重叠参数
	if options.Overlap > 0 {
		args = append(args, "--overlap", strconv.Itoa(options.Overlap))
	}

	// 添加质量参数
	if options.Quality > 0 {
		args = append(args, "--Q", strconv.Itoa(options.Quality))
	}

	// 添加布局参数
	switch options.Layout {
	case "google":
		args = append(args, "--layout", "google")
	case "zoomify":
		args = append(args, "--layout", "zoomify")
	default:
		args = append(args, "--layout", "dz")
	}

	// 添加背景颜色（如果指定）
	if options.Background != "" {
		args = append(args, "--background", options.Background)
	}

	result, err := utils.ExecuteCommand(ctx, utils.VipsPath, args...)
	if err != nil {
		return fmt.Errorf("vips dzsave failed: %w, stderr: %s", err, string(result.Stderr))
	}

	return nil
}

// GenerateProgressive 生成渐进式JPEG（用于网络传输优化）
func GenerateProgressive(ctx context.Context, inputPath, outputPath string, quality int) error {
	if err := utils.EnsureInitialized(nil, nil); err != nil {
		return err
	}

	if quality <= 0 {
		quality = 85
	}

	args := []string{"jpegsave", inputPath, outputPath,
		"--Q", strconv.Itoa(quality),
		"--interlace", // 渐进式JPEG
		"--optimize-coding", // 优化编码
		"--strip", // 移除元数据
	}

	result, err := utils.ExecuteCommand(ctx, utils.VipsPath, args...)
	if err != nil {
		return fmt.Errorf("vips progressive jpeg failed: %w, stderr: %s", err, string(result.Stderr))
	}

	return nil
}

// GetImageLevels 计算图片金字塔层级数（用于瓦片生成）
func GetImageLevels(width, height, tileSize int) int {
	maxDimension := width
	if height > maxDimension {
		maxDimension = height
	}

	levels := 0
	for maxDimension > tileSize {
		maxDimension = maxDimension / 2
		levels++
	}

	return levels + 1
}
