package handler

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"rear/internal/config"
	"rear/internal/container"
	"rear/internal/model"
	"rear/internal/model/tables"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"rear/pkg/logger"
)

type PhotoHandler struct {
	container  *container.DbContainer
	imgContain *container.TaskContainer
}

func NewPhotoHandler(container *container.DbContainer, imgContain *container.TaskContainer) *PhotoHandler {
	return &PhotoHandler{
		container:  container,
		imgContain: imgContain,
	}
}

// PhotoDetailResponse 图像详细信息响应
type PhotoDetailResponse struct {
	// Photo基础信息
	Hash            string     `json:"hash"`
	ImgPath         string     `json:"imgPath"`
	ImgName         string     `json:"imgName"`
	Width           int        `json:"width"`
	Height          int        `json:"height"`
	AspectRatio     float32    `json:"aspectRatio"`
	FileSize        int64      `json:"fileSize"`
	Format          string     `json:"format"`
	Notes           *string    `json:"notes,omitempty"`
	FileCreatedAt   *time.Time `json:"fileCreatedAt,omitempty"`
	TakenAt         *time.Time `json:"takenAt,omitempty"`
	LastModified    *time.Time `json:"lastModified,omitempty"`
	Rating          int        `json:"rating"`
	LastViewedAt    *time.Time `json:"lastViewedAt,omitempty"`
	ViewCount       int        `json:"viewCount"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`

	// EXIF信息
	ExifInfo *PhotoExifInfo `json:"exifInfo,omitempty"`
}

type PhotoExifInfo struct {
	FileName         string                 `json:"fileName,omitempty"`
	FileSize         int64                  `json:"fileSize,omitempty"`
	ImageWidth       int                    `json:"imageWidth,omitempty"`
	ImageHeight      int                    `json:"imageHeight,omitempty"`
	ImageSize        string                 `json:"imageSize,omitempty"`
	MimeType         string                 `json:"mimeType,omitempty"`
	FileType         string                 `json:"fileType,omitempty"`
	FileTypeExt      string                 `json:"fileTypeExt,omitempty"`
	ModifyDate       string                 `json:"modifyDate,omitempty"`
	CreateDate       string                 `json:"createDate,omitempty"`
	ColorSpace       string                 `json:"colorSpace,omitempty"`
	BitsPerSample    int                    `json:"bitsPerSample,omitempty"`
	Resolution       string                 `json:"resolution,omitempty"`
	Quality          int                    `json:"quality,omitempty"`
	Model            string                 `json:"model,omitempty"`
	Make             string                 `json:"make,omitempty"`
	ISO              int                    `json:"iso,omitempty"`
	GPSLatitude      float64                `json:"gpsLatitude,omitempty"`
	GPSLongitude     float64                `json:"gpsLongitude,omitempty"`
	ExposureTime     float64                `json:"exposureTime,omitempty"`
	Aperture         float64                `json:"aperture,omitempty"`
	FNumber          float64                `json:"fNumber,omitempty"`
	FocalLength      float64                `json:"focalLength,omitempty"`
	LensID           string                 `json:"lensID,omitempty"`
	Title            string                 `json:"title,omitempty"`
	Description      string                 `json:"description,omitempty"`
	DateTimeOriginal string                 `json:"dateTimeOriginal,omitempty"`
	OtherFields      map[string]interface{} `json:"otherFields,omitempty"`
}

// GetPhoto 获取图像文件
// GET api/v1/photo/{hash}?format=thumbnail&size=400x400
func (h *PhotoHandler) GetPhoto(c *gin.Context) {
	hash := c.Param("hash")
	if hash == "" {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    http.StatusBadRequest,
			Message: "Missing hash parameter",
		})
		return
	}

	// 获取查询参数
	format := c.DefaultQuery("format", "original") // original, thumbnail
	sizeParam := c.Query("size")

	// 解析尺寸参数
	maxWidth, maxHeight, err := h.parseSize(sizeParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("Invalid size parameter: %v", err),
		})
		return
	}

	// 查询照片信息
	photo, err := h.container.PhotoRepo.GetByHash(hash)
	if err != nil {
		logger.Error("获取照片信息失败", zap.String("hash", hash), zap.Error(err))
		c.JSON(http.StatusNotFound, model.Response{
			Code:    http.StatusNotFound,
			Message: "Photo not found",
		})
		return
	}

	// 更新访问计数和时间
	if err := h.container.PhotoRepo.UpdateViewCount(hash); err != nil {
		logger.Warn("更新访问计数失败", zap.String("hash", hash), zap.Error(err))
	}

	var imagePath string
	
	switch format {
	case "thumbnail":
		// 生成或获取缩略图路径
		imagePath, err = h.getThumbnailPath(photo, maxWidth, maxHeight)
		if err != nil {
			logger.Error("获取缩略图失败", zap.String("hash", hash), zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.Response{
				Code:    http.StatusInternalServerError,
				Message: "Failed to generate thumbnail",
			})
			return
		}
	case "original":
		// 返回原图
		imagePath = photo.ImgPath
	default:
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid format parameter. Supported: original, thumbnail",
		})
		return
	}

	// 检查文件是否存在
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		logger.Error("图像文件不存在", zap.String("path", imagePath))
		c.JSON(http.StatusNotFound, model.Response{
			Code:    http.StatusNotFound,
			Message: "Image file not found",
		})
		return
	}

	// 设置响应头
	h.setImageHeaders(c, imagePath, photo.Format)

	// 返回文件
	c.File(imagePath)
}

// GetPhotoAssets 获取图像详细信息
// GET api/v1/assets/{hash}
func (h *PhotoHandler) GetPhotoAssets(c *gin.Context) {
	hash := c.Param("hash")
	if hash == "" {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    http.StatusBadRequest,
			Message: "Missing hash parameter",
		})
		return
	}

	// 查询照片基础信息
	photo, err := h.container.PhotoRepo.GetByHash(hash)
	if err != nil {
		logger.Error("获取照片信息失败", zap.String("hash", hash), zap.Error(err))
		c.JSON(http.StatusNotFound, model.Response{
			Code:    http.StatusNotFound,
			Message: "Photo not found",
		})
		return
	}

	// 构建响应对象
	response := &PhotoDetailResponse{
		Hash:          photo.Hash,
		ImgPath:       photo.ImgPath,
		ImgName:       photo.ImgName,
		Width:         photo.Width,
		Height:        photo.Height,
		AspectRatio:   photo.AspectRatio,
		FileSize:      photo.FileSize,
		Format:        photo.Format,
		Notes:         photo.Notes,
		FileCreatedAt: photo.FileCreatedAt,
		TakenAt:       photo.TakenAt,
		LastModified:  photo.LastModified,
		Rating:        photo.Rating,
		LastViewedAt:  photo.LastViewedAt,
		ViewCount:     photo.ViewCount,
		CreatedAt:     photo.CreatedAt,
		UpdatedAt:     photo.UpdatedAt,
	}

	// 查询EXIF信息
	if exif, err := h.container.ExifRepo.GetByHash(hash); err == nil {
		response.ExifInfo = &PhotoExifInfo{
			FileName:         exif.FileName,
			FileSize:         exif.FileSize,
			ImageWidth:       exif.ImageWidth,
			ImageHeight:      exif.ImageHeight,
			ImageSize:        exif.ImageSize,
			MimeType:         exif.MIMEType,
			FileType:         exif.FileType,
			FileTypeExt:      exif.FileTypeExt,
			ModifyDate:       exif.ModifyDate,
			CreateDate:       exif.CreateDate,
			ColorSpace:       exif.ColorSpace,
			BitsPerSample:    exif.BitsPerSample,
			Resolution:       exif.Resolution,
			Quality:          exif.Quality,
			Model:            exif.Model,
			Make:             exif.Make,
			ISO:              exif.ISO,
			GPSLatitude:      exif.GPSLatitude,
			GPSLongitude:     exif.GPSLongitude,
			ExposureTime:     exif.ExposureTime,
			Aperture:         exif.Aperture,
			FNumber:          exif.FNumber,
			FocalLength:      exif.FocalLength,
			LensID:           exif.LensID,
			Title:            exif.Title,
			Description:      exif.Description,
			DateTimeOriginal: exif.DateTimeOrig,
		}

		// 解析OtherFields JSON
		if exif.OtherFields != nil {
			response.ExifInfo.OtherFields = map[string]interface{}(exif.OtherFields)
		}
	} else {
		logger.Debug("未找到EXIF信息", zap.String("hash", hash), zap.Error(err))
	}

	c.JSON(http.StatusOK, model.Response{
		Code:    http.StatusOK,
		Message: "Success",
		Data:    response,
	})
}

// parseSize 解析尺寸参数
func (h *PhotoHandler) parseSize(sizeParam string) (int, int, error) {
	if sizeParam == "" {
		// 使用默认缩略图尺寸
		defaultSize := config.GetDefaultThumbnailSize()
		return defaultSize, defaultSize, nil
	}

	// 解析 "400x400" 格式
	parts := strings.Split(sizeParam, "x")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid size format, expected WIDTHxHEIGHT")
	}

	width, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid width: %v", err)
	}

	height, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid height: %v", err)
	}

	if width <= 0 || height <= 0 {
		return 0, 0, fmt.Errorf("width and height must be positive")
	}

	return width, height, nil
}

// getThumbnailPath 获取或生成缩略图路径
func (h *PhotoHandler) getThumbnailPath(photo *tables.Photo, maxWidth, maxHeight int) (string, error) {
	// 构建缩略图文件名
	// 格式: {hash}_{maxWidth}x{maxHeight}.jpg
	thumbnailName := fmt.Sprintf("%s_%dx%d.jpg", photo.Hash, maxWidth, maxHeight)
	
	// 构建缩略图完整路径
	appDir := config.CONFIG.AppDir
	thumbnailDir := filepath.Join(appDir, config.CONFIG.PathConfig.CachePath, config.CONFIG.PathConfig.ThumbnailPath)
	thumbnailPath := filepath.Join(thumbnailDir, thumbnailName)

	// 检查缩略图是否已存在
	if _, err := os.Stat(thumbnailPath); err == nil {
		return thumbnailPath, nil
	}

	// 缩略图不存在，需要生成
	err := h.generateThumbnail(photo.ImgPath, thumbnailPath, maxWidth, maxHeight)
	if err != nil {
		return "", fmt.Errorf("生成缩略图失败: %w", err)
	}

	return thumbnailPath, nil
}

// generateThumbnail 生成缩略图
func (h *PhotoHandler) generateThumbnail(originalPath, thumbnailPath string, maxWidth, maxHeight int) error {
	// 确保缩略图目录存在
	thumbnailDir := filepath.Dir(thumbnailPath)
	if err := os.MkdirAll(thumbnailDir, 0755); err != nil {
		return fmt.Errorf("创建缩略图目录失败: %w", err)
	}

	// TODO: 这里应该调用图像处理库生成缩略图
	// 目前先简单复制原图作为占位符
	return h.copyFile(originalPath, thumbnailPath)
}

// copyFile 复制文件（临时方案）
func (h *PhotoHandler) copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// setImageHeaders 设置图像响应头
func (h *PhotoHandler) setImageHeaders(c *gin.Context, imagePath, format string) {
	// 设置Content-Type
	var contentType string
	switch strings.ToLower(format) {
	case "jpg", "jpeg":
		contentType = "image/jpeg"
	case "png":
		contentType = "image/png"
	case "webp":
		contentType = "image/webp"
	case "gif":
		contentType = "image/gif"
	default:
		contentType = "application/octet-stream"
	}
	
	c.Header("Content-Type", contentType)
	
	// 设置缓存头
	c.Header("Cache-Control", "public, max-age=86400") // 缓存1天
	c.Header("ETag", fmt.Sprintf(`"%s"`, filepath.Base(imagePath)))
	
	// 设置文件名
	fileName := filepath.Base(imagePath)
	c.Header("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, fileName))
}