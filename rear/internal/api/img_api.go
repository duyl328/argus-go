package api

import (
	"context"
	"fmt"
	"github.com/h2non/filetype"
	"go.uber.org/zap"
	"log"
	"os"
	"rear/internal/config"
	"rear/internal/consts"
	"rear/internal/model"
	"rear/internal/utils/tools"
	"rear/pkg/logger"
	"rear/pkg/utils"
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
func (api *ImageAPI) GetImagePath(longSide int) {
	// 文件类型是否匹配
	s := string(config.CONFIG.ImageCompressionOption.ThumbnailFormat)
	fileType := api.format
	// 类型相同则返回源路径； 类型不同将图片转换为 PNG 再转换为 指定格式 “原图”
	if fileType == s {

	}

	// 小于 0 获取原图
	if longSide <= 0 {
	} else {
		// 获取指定质量

	}
}

// CompressedImg 压缩图片 longSide 为图片的最长边
func (api *ImageAPI) CompressedImg(longSide int) {

	// 获取支持的原图路径
	// 判断是否需要压缩，如果超过指定大小就进行压缩，如果超过指定比例就裁切，裁切分为截取头部和截取中间（第一版截取中间）

}

// GetOriginalImagePath 获取原图（真的是原图）
func (api *ImageAPI) GetOriginalImagePath() string {
	return api.path
}

// GetSupportOriginalImagePath 获取支持的原图路径（如果设置的为 JPG 则返回 JPG ，如果是 Webp 则会是 Webp）
func (api *ImageAPI) GetSupportOriginalImagePath() string {
	// 判断是否是支持的格式
	if api.format == string(consts.FormatJPG) || api.format == string(consts.FormatJPEG) || api.format == string(consts.FormatWEBP) {
		return api.path
	}

	// ============ 到这里不是默认格式，要么是 PNG，要么是其他格式，所以一定会存在或最终转换为 PNG ==========

	// 假设认为他是 PNG 格式
	//pngPath := api.path

	// 如果不是 PNG 则转换为 PNG
	if api.format != string(consts.FormatPNG) {
		logger.Error("[TODO] 获取到不支持的图片格式!", zap.String("format", api.format))
		log.Fatal("[TODO] 获取到不支持的图片格式!", api.format)
	}

	// 到这里一定是 PNG 了，转换为 JPG 或 Webp 格式
	logger.Error("[TODO] 现在不支持 PNG 格式转换!", zap.String("format", api.format))
	log.Fatal("[TODO] 现在不支持 PNG 格式转换!", api.format)
	return ""
}

// GetImage 获取照片
func (api *ImageAPI) GetImage() {
	/*
		获取图片路径的逻辑：

		1. 判断照片类型，如果照片类型是非 JPG 和 Webp 则寻找 JPG 和 Webp 所在路径
		2. 如果没有 JPG 和 Webp 则将照片转换为 JPG 或 Webp 并返回
	*/

}

// 获取图片路径
//func (api *ImageAPI) getImageFormat(path string) (string, error) {
//buf, err := os.Read
//}
