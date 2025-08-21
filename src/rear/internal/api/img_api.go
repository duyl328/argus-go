package api

import (
	"context"
	"fmt"
	"github.com/h2non/filetype"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"rear/internal/config"
	"rear/internal/consts"
	"rear/internal/model"
	"rear/internal/utils/tools"
	"rear/pkg/logger"
	"rear/pkg/utils"
	"strconv"
	"strings"
)

type ImageAPI struct {
	// 哈希值
	hash string
	// 照片格式
	format string
	// 照片路径
	path string
	// 照片字节数组
	fileBytes []byte
	// Exif 数据
	exifData *model.ParsedExif
}

func NewImageAPI(path string) (*ImageAPI, error) {
	// 判断路径是否是本软件的安装目录或缓存目录

	// 检测图片是否存在
	exists := utils.FileUtils.Exists(path)
	if !exists {
		logger.Error(
			"指定文件不存在!",
			zap.String("path", path),
		)
		return nil, fmt.Errorf("指定文件不存在: %s", path)
	}

	i := &ImageAPI{
		path: path,
	}

	// 读取文件
	buf, err := os.ReadFile(path)
	if err != nil {
		logger.Error(
			"文件读取失败!",
			zap.String("path", path),
		)
		return nil, fmt.Errorf("文件读取失败: %s", path)
	}
	i.fileBytes = buf

	// 检测照片格式
	kind, err := filetype.Match(i.fileBytes)
	if err != nil {
		logger.Error(
			"文件类型匹配失败!",
			zap.String("path", i.path),
		)
		return nil, fmt.Errorf("检测照片格式失败: %s", path)
	}
	i.format = kind.Extension

	// 获取 Hash
	hash, err := utils.HashUtils.HashFile(path, utils.SHA256)
	if err != nil {
		logger.Error(
			"Hash获取失败!",
			zap.String("path", path),
		)
		return nil, fmt.Errorf("hash获取失败: %s", path)
	}
	i.hash = hash

	return i, nil
}

// GetExif 获取 exif
func (api *ImageAPI) GetExif() error {
	// todo 尝试从数据库读取对应的 exif 数据，如果存在则直接返回

	// 使用 exiftool 获取 exif 信息
	ctx := context.Background()
	exifData, err := tools.GetExifData(ctx, api.path)
	if err != nil {
		logger.Error(
			"获取 EXIF 数据失败!",
			zap.String("path", api.path),
			zap.Error(err),
		)
		return fmt.Errorf("获取 EXIF 数据失败: %s", api.path)
	}

	// 分割 EXIF 数据
	splitExifData := model.SplitExifData(exifData)
	api.exifData = splitExifData

	// 打印 EXIF 信息
	if splitExifData != nil {
		logger.Info("EXIF 信息提取完成",
			zap.String("path", api.path),
			zap.String("make", splitExifData.Exif.Make),
			zap.String("model", splitExifData.Exif.Model),
			zap.String("date_time_orig", splitExifData.Exif.DateTimeOrig),
			zap.Int("width", splitExifData.BaseInfo.ImageWidth),
			zap.Int("height", splitExifData.BaseInfo.ImageHeight),
		)
	}

	// todo 将 exif 数据存储到数据库中

	return nil
}

// GetImagePath 获取原图路径、获取缩略图路径
/*
1. 判断是不是支持默认支持的格式，如 JPG、Webp，如果是默认支持则直接返回原始位置
2. 如果不是默认支持的格式，判断是否是支持格式，如 Png、Gif，如果不是则返回错误
3. 如果是支持格式，则将图片转换为 JPG 或 Webp 格式，并返回转换后的路径【该图片存储在对应 Hash 路径下】
4. 如果 最长边 的值小于等于 0，则返回原图路径，反之则返回对应宽度的缩略图路径
5. 如果是缩略图，则判断是否存在对应宽度的缩略图，如果存在则返回该路径，如果不存在则生成缩略图并返回（缩略图宽度会是固定的值）
6. 如果传入的参数很大，比如超过原图的宽度，则返回原图路径
*/
func (api *ImageAPI) GetImagePath(longSide int) (string, error) {
	ctx := context.Background()

	// 小于等于0获取原图
	if longSide <= 0 {
		return api.GetSupportOriginalImagePath(ctx)
	}

	// 获取原图信息以判断是否需要缩放
	supportedPath, err := api.GetSupportOriginalImagePath(ctx)
	if err != nil {
		return "", err
	}

	// 获取图像信息判断尺寸
	imageInfo, err := tools.GetVipsImageInfo(ctx, supportedPath)
	if err != nil {
		return "", fmt.Errorf("获取图像信息失败: %w", err)
	}

	// 如果请求的尺寸大于或等于原图的最长边，直接返回原图
	maxSide := imageInfo.Width
	if imageInfo.Height > maxSide {
		maxSide = imageInfo.Height
	}
	if longSide >= maxSide {
		return supportedPath, nil
	}

	// 生成缩略图
	return api.generateThumbnail(ctx, supportedPath, longSide)
}

// GetSupportOriginalImagePath 获取支持的原图路径（如果设置的为 JPG 则返回 JPG ，如果是 Webp 则会是 Webp）
func (api *ImageAPI) GetSupportOriginalImagePath(ctx context.Context) (string, error) {
	targetFormat := config.CONFIG.ImageCompressionOption.ThumbnailFormat

	// 判断是否是支持的格式
	if api.isSupportedFormat() {
		return api.path, nil
	}

	// ============ 到这里不是默认格式，需要转换 ==========

	// 生成转换后的文件路径
	convertedPath := api.getConvertedImagePath(targetFormat)

	// 检查转换后的文件是否已存在
	if utils.FileUtils.Exists(convertedPath) {
		return convertedPath, nil
	}

	// 需要转换图像
	if err := api.convertImageToSupported(ctx, convertedPath, targetFormat); err != nil {
		return "", fmt.Errorf("图像转换失败: %w", err)
	}

	return convertedPath, nil
}

// 辅助函数

// isSupportedFormat 判断是否是直接支持的格式
func (api *ImageAPI) isSupportedFormat() bool {
	return api.format == string(consts.FormatJPG) ||
		api.format == string(consts.FormatJPEG) ||
		api.format == string(consts.FormatWEBP)
}

// getConvertedImagePath 获取转换后的图像路径
func (api *ImageAPI) getConvertedImagePath(targetFormat consts.ImageFormat) string {
	// 使用与 main.go createCachePath 相同的逻辑构建完整缩略图路径
	thumbnailBasePath := filepath.Join(config.CONFIG.AppDir, config.CONFIG.PathConfig.CachePath, config.CONFIG.PathConfig.ThumbnailPath)

	ext := string(targetFormat)
	return utils.HashUtils.HashThumbPath(thumbnailBasePath, api.hash, "original", ext)
}

// convertImageToSupported 转换图像到支持的格式
func (api *ImageAPI) convertImageToSupported(ctx context.Context, outputPath string, targetFormat consts.ImageFormat) error {
	// 确保输出目录存在
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}

	// 根据原图格式选择转换策略
	switch strings.ToLower(api.format) {
	case "png":
		// PNG直接转换为目标格式
		return api.convertPNGToTarget(ctx, outputPath, targetFormat)
	case "raw", "cr2", "nef", "arw", "dng":
		// RAW格式：RAW → PNG → 目标格式
		return api.convertRAWToTarget(ctx, outputPath, targetFormat)
	case "heic", "heif":
		// HEIC格式转换
		return api.convertHEICToTarget(ctx, outputPath, targetFormat)
	default:
		// 其他格式尝试用ImageMagick转换
		return api.convertOtherToTarget(ctx, outputPath, targetFormat)
	}
}

// convertPNGToTarget PNG转换为目标格式
func (api *ImageAPI) convertPNGToTarget(ctx context.Context, outputPath string, targetFormat consts.ImageFormat) error {
	options := &tools.VipsConvertOptions{
		Quality:   config.CONFIG.ImageCompressionOption.ThumbnailQuality,
		StripMeta: true,
	}

	return tools.ConvertVipsImage(ctx, api.path, outputPath, targetFormat, options)
}

// convertRAWToTarget RAW转换为目标格式
func (api *ImageAPI) convertRAWToTarget(ctx context.Context, outputPath string, targetFormat consts.ImageFormat) error {
	// 第一步：使用ImageMagick将RAW转换为PNG
	tempPNGPath := filepath.Join(config.CONFIG.PathConfig.PngTempPath, api.hash+".png")
	if err := os.MkdirAll(filepath.Dir(tempPNGPath), 0755); err != nil {
		return fmt.Errorf("创建临时目录失败: %w", err)
	}

	// RAW → PNG
	if err := tools.ConvertImage(ctx, api.path, tempPNGPath); err != nil {
		return fmt.Errorf("RAW到PNG转换失败: %w", err)
	}

	// 确保在函数结束时删除临时PNG文件
	defer func() {
		if err := os.Remove(tempPNGPath); err != nil {
			logger.Warn("删除临时PNG文件失败", zap.String("path", tempPNGPath), zap.Error(err))
		}
	}()

	// 第二步：使用LibVips将PNG转换为目标格式
	options := &tools.VipsConvertOptions{
		Quality:   config.CONFIG.ImageCompressionOption.ThumbnailQuality,
		StripMeta: true,
	}

	return tools.ConvertVipsImage(ctx, tempPNGPath, outputPath, targetFormat, options)
}

// convertHEICToTarget HEIC转换为目标格式
func (api *ImageAPI) convertHEICToTarget(ctx context.Context, outputPath string, targetFormat consts.ImageFormat) error {
	// 可以使用ImageMagick或libvips处理HEIC
	options := &tools.VipsConvertOptions{
		Quality:   config.CONFIG.ImageCompressionOption.ThumbnailQuality,
		StripMeta: true,
	}

	return tools.ConvertVipsImage(ctx, api.path, outputPath, targetFormat, options)
}

// convertOtherToTarget 其他格式转换为目标格式
func (api *ImageAPI) convertOtherToTarget(ctx context.Context, outputPath string, targetFormat consts.ImageFormat) error {
	// 尝试使用LibVips直接转换
	options := &tools.VipsConvertOptions{
		Quality:   config.CONFIG.ImageCompressionOption.ThumbnailQuality,
		StripMeta: true,
	}

	err := tools.ConvertVipsImage(ctx, api.path, outputPath, targetFormat, options)
	if err != nil {
		// 如果LibVips失败，尝试ImageMagick
		logger.Warn("LibVips转换失败，尝试ImageMagick", zap.Error(err))

		// 先转换为PNG，再用LibVips转换为目标格式
		tempPNGPath := filepath.Join(config.CONFIG.PathConfig.PngTempPath, api.hash+".png")
		if err := os.MkdirAll(filepath.Dir(tempPNGPath), 0755); err != nil {
			return fmt.Errorf("创建临时目录失败: %w", err)
		}

		if err := tools.ConvertImage(ctx, api.path, tempPNGPath); err != nil {
			return fmt.Errorf("ImageMagick转换失败: %w", err)
		}

		defer func() {
			if err := os.Remove(tempPNGPath); err != nil {
				logger.Warn("删除临时PNG文件失败", zap.String("path", tempPNGPath), zap.Error(err))
			}
		}()

		return tools.ConvertVipsImage(ctx, tempPNGPath, outputPath, targetFormat, options)
	}

	return nil
}

// generateThumbnail 生成缩略图
func (api *ImageAPI) generateThumbnail(ctx context.Context, sourcePath string, longSide int) (string, error) {
	// 获取缩略图路径
	thumbnailPath := api.GetThumbnailPath(longSide)

	// 检查缩略图是否已存在
	if utils.FileUtils.Exists(thumbnailPath) {
		return thumbnailPath, nil
	}

	// 确保输出目录存在
	if err := os.MkdirAll(filepath.Dir(thumbnailPath), 0755); err != nil {
		return "", fmt.Errorf("创建缩略图目录失败: %w", err)
	}

	// 获取原图信息
	imageInfo, err := tools.GetVipsImageInfo(ctx, sourcePath)
	if err != nil {
		return "", fmt.Errorf("获取图像信息失败: %w", err)
	}

	// 计算缩略图尺寸
	targetWidth, targetHeight := api.calculateThumbnailSize(imageInfo.Width, imageInfo.Height, longSide)

	// 检查是否需要特殊裁剪处理（针对极端比例的图片）
	if api.needsSpecialCropping(imageInfo.Width, imageInfo.Height) {
		return api.generateCroppedThumbnail(ctx, sourcePath, thumbnailPath, targetWidth, targetHeight)
	}

	// 标准等比例缩放
	return api.generateStandardThumbnail(ctx, sourcePath, thumbnailPath, targetWidth, targetHeight)
}

// GetThumbnailPath 获取缩略图路径（公开方法）
func (api *ImageAPI) GetThumbnailPath(longSide int) string {
	// 使用与 main.go createCachePath 相同的逻辑构建完整缩略图路径
	thumbnailPath := filepath.Join(config.CONFIG.AppDir, config.CONFIG.PathConfig.CachePath, config.CONFIG.PathConfig.ThumbnailPath)

	sizeStr := strconv.Itoa(longSide) + "X" + strconv.Itoa(longSide)
	ext := string(config.CONFIG.ImageCompressionOption.ThumbnailFormat)

	return utils.HashUtils.HashThumbPath(thumbnailPath, api.hash, sizeStr, ext)
}

// calculateThumbnailSize 计算缩略图尺寸
func (api *ImageAPI) calculateThumbnailSize(originalWidth, originalHeight, longSide int) (int, int) {
	if originalWidth >= originalHeight {
		// 宽度为长边
		scale := float64(longSide) / float64(originalWidth)
		return longSide, int(float64(originalHeight) * scale)
	} else {
		// 高度为长边
		scale := float64(longSide) / float64(originalHeight)
		return int(float64(originalWidth) * scale), longSide
	}
}

// needsSpecialCropping 判断是否需要特殊裁剪（极端比例图片）
func (api *ImageAPI) needsSpecialCropping(width, height int) bool {
	ratio := float64(width) / float64(height)
	// 比例超过1:10或10:1认为是极端比例
	return ratio > 10.0 || ratio < 0.1
}

// generateCroppedThumbnail 生成裁剪缩略图
func (api *ImageAPI) generateCroppedThumbnail(ctx context.Context, sourcePath, outputPath string, targetWidth, targetHeight int) (string, error) {
	// 对于极端比例的图片，使用智能裁剪
	options := &tools.VipsCropOptions{
		Position: tools.VipsCropSmart,
	}

	// 裁剪为更合理的比例，比如最大1:3
	maxWidth := targetHeight * 3
	maxHeight := targetWidth * 3

	if targetWidth > maxWidth {
		targetWidth = maxWidth
	}
	if targetHeight > maxHeight {
		targetHeight = maxHeight
	}

	return outputPath, tools.CropVipsImage(ctx, sourcePath, outputPath, targetWidth, targetHeight, options)
}

// generateStandardThumbnail 生成标准缩略图
func (api *ImageAPI) generateStandardThumbnail(ctx context.Context, sourcePath, outputPath string, targetWidth, targetHeight int) (string, error) {
	// 使用libvips的综合压缩功能
	quality := config.CONFIG.ImageCompressionOption.ThumbnailQuality
	targetFormat := config.CONFIG.ImageCompressionOption.ThumbnailFormat

	err := tools.CompressVipsImage(ctx, sourcePath, outputPath, targetFormat, targetWidth, targetHeight, quality)
	if err != nil {
		return "", fmt.Errorf("生成缩略图失败: %w", err)
	}

	return outputPath, nil
}

// 公共API函数

// GetOriginalImagePath 获取原图（真的是原图）
func (api *ImageAPI) GetOriginalImagePath() string {
	return api.path
}

// GetHash 获取图像哈希值
func (api *ImageAPI) GetHash() string {
	return api.hash
}

// GetFormat 获取图像格式
func (api *ImageAPI) GetFormat() string {
	return api.format
}

// GetExifData 获取EXIF数据
func (api *ImageAPI) GetExifData() *model.ParsedExif {
	return api.exifData
}

// GetImageInfo 获取图像基础信息
func (api *ImageAPI) GetImageInfo(ctx context.Context) (*tools.VipsImageInfo, error) {
	supportedPath, err := api.GetSupportOriginalImagePath(ctx)
	if err != nil {
		return nil, err
	}

	return tools.GetVipsImageInfo(ctx, supportedPath)
}

// IsSupported 判断图像格式是否被支持
func (api *ImageAPI) IsSupported() bool {
	supportedFormats := []string{"jpg", "jpeg", "png", "webp", "heic", "heif", "raw", "cr2", "nef", "arw", "dng", "tiff", "bmp", "gif"}
	format := strings.ToLower(api.format)

	for _, supported := range supportedFormats {
		if format == supported {
			return true
		}
	}

	return false
}

// ProcessImage 处理图像的主要入口方法
// 包含：格式转换、缩略图生成、EXIF提取等完整处理流程
func (api *ImageAPI) ProcessImage() error {
	ctx := context.Background()

	// 1. 提取EXIF信息
	if err := api.GetExif(); err != nil {
		logger.Warn("获取EXIF信息失败", zap.String("path", api.path), zap.Error(err))
		// EXIF获取失败不终止处理，但记录警告
	}

	// 2. 格式转换（如果需要）
	supportedPath, err := api.GetSupportOriginalImagePath(ctx)
	if err != nil {
		return fmt.Errorf("生成支持格式原图失败: %w", err)
	}
	logger.Info("格式转换完成", zap.String("path", supportedPath))

	// 3. 生成缩略图
	if err := api.generateAllThumbnails(ctx); err != nil {
		return fmt.Errorf("生成缩略图失败: %w", err)
	}

	return nil
}

// generateAllThumbnails 生成所有尺寸的缩略图
func (api *ImageAPI) generateAllThumbnails(ctx context.Context) error {
	// 从配置中获取需要生成的缩略图尺寸
	thumbnailSizes := api.getThumbnailSizes()

	for _, size := range thumbnailSizes {
		thumbnailPath, err := api.GetImagePath(size)
		if err != nil {
			logger.Warn("生成缩略图失败",
				zap.String("path", api.path),
				zap.Int("size", size),
				zap.Error(err))
			continue // 某个尺寸失败不影响其他尺寸
		}

		logger.Debug("生成缩略图完成",
			zap.String("original", api.path),
			zap.String("thumbnail", thumbnailPath),
			zap.Int("size", size))
	}

	return nil
}

// getThumbnailSizes 获取需要预生成的缩略图尺寸
func (api *ImageAPI) getThumbnailSizes() []int {
	// 从配置中获取，如果没有配置则使用默认值
	if len(config.CONFIG.ImageCompressionOption.ThumbnailSize) > 0 {
		return config.CONFIG.ImageCompressionOption.ThumbnailSize
	}

	// 默认缩略图尺寸
	return []int{800}
}

// CompressedImg 压缩图片获取路径
func (api *ImageAPI) CompressedImg(longSide int) (string, error) {
	return api.GetImagePath(longSide)
}
