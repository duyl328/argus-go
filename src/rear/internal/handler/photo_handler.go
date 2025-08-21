package handler

import (
	"fmt"
	"net/http"
	"rear/internal/api"
	"rear/internal/config"
	"rear/internal/container"
	"rear/internal/db"
	"rear/internal/model"
	"rear/internal/repositories"
	"strconv"
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
	Hash          string     `json:"hash"`
	ImgPath       string     `json:"imgPath"`
	ImgName       string     `json:"imgName"`
	Width         int        `json:"width"`
	Height        int        `json:"height"`
	AspectRatio   float32    `json:"aspectRatio"`
	FileSize      int64      `json:"fileSize"`
	Format        string     `json:"format"`
	Notes         *string    `json:"notes,omitempty"`
	FileCreatedAt *time.Time `json:"fileCreatedAt,omitempty"`
	TakenAt       *time.Time `json:"takenAt,omitempty"`
	LastModified  *time.Time `json:"lastModified,omitempty"`
	Rating        int        `json:"rating"`
	LastViewedAt  *time.Time `json:"lastViewedAt,omitempty"`
	ViewCount     int        `json:"viewCount"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`

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
// GET api/v1/photo/{hash}?format=thumbnail&size=400
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

	// 异步更新访问计数和时间
	task := &db.UpdateViewCountTask{Hash: hash}
	callback := db.TaskCallback{
		OnError: func(taskID string, err error) {
			logger.Warn("异步更新访问计数失败",
				zap.String("hash", hash),
				zap.String("taskID", taskID),
				zap.Error(err))
		},
	}
	if err := db.GetManger().SubmitWriteTask(task, callback); err != nil {
		logger.Warn("提交更新访问计数任务失败", zap.String("hash", hash), zap.Error(err))
	}

	// 使用ImageAPI处理图像
	imageAPI, err := api.NewImageAPI(photo.ImgPath)
	if err != nil {
		logger.Error("创建ImageAPI失败", zap.String("path", photo.ImgPath), zap.Error(err))
		c.JSON(http.StatusInternalServerError, model.Response{
			Code:    http.StatusInternalServerError,
			Message: "Failed to process image",
		})
		return
	}

	var imagePath string

	switch format {
	case "thumbnail":
		// 解析尺寸参数
		size, err := h.parseSize(sizeParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, model.Response{
				Code:    http.StatusBadRequest,
				Message: fmt.Sprintf("Invalid size parameter: %v", err),
			})
			return
		}

		// 使用ImageAPI获取缩略图
		imagePath, err = imageAPI.GetImagePath(size)
		if err != nil {
			logger.Error("获取缩略图失败", zap.String("hash", hash), zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.Response{
				Code:    http.StatusInternalServerError,
				Message: "Failed to generate thumbnail",
			})
			return
		}
	case "original":
		// 使用ImageAPI获取支持的原图格式
		imagePath, err = imageAPI.GetSupportOriginalImagePath(c.Request.Context())
		if err != nil {
			logger.Error("获取原图失败", zap.String("hash", hash), zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.Response{
				Code:    http.StatusInternalServerError,
				Message: "Failed to get original image",
			})
			return
		}
	default:
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid format parameter. Supported: original, thumbnail",
		})
		return
	}

	// 设置响应头
	h.setImageHeaders(c, imagePath, imageAPI.GetFormat())

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

// parseSize 解析尺寸参数，返回长边尺寸
func (h *PhotoHandler) parseSize(sizeParam string) (int, error) {
	if sizeParam == "" {
		// 使用默认缩略图尺寸
		defaultSize := config.GetDefaultThumbnailSize()
		return defaultSize, nil
	}

	size, err := strconv.Atoi(sizeParam)
	if err != nil {
		return 0, fmt.Errorf("invalid size parameter: %v", err)
	}

	if size <= 0 {
		return 0, fmt.Errorf("size must be positive")
	}

	return size, nil
}

// GetTimeline 获取照片时间线统计
// GET /api/v1/photos/timeline?start_date=2023-01-01&end_date=2023-12-31
func (h *PhotoHandler) GetTimeline(c *gin.Context) {
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	var timeline []repositories.PhotoTimelineItem
	var err error

	// 如果提供了时间范围参数
	if startDateStr != "" && endDateStr != "" {
		startDate, err1 := time.Parse("2006-01-02", startDateStr)
		endDate, err2 := time.Parse("2006-01-02", endDateStr)

		if err1 != nil || err2 != nil {
			c.JSON(http.StatusBadRequest, model.Response{
				Code:    http.StatusBadRequest,
				Message: "Invalid date format. Use YYYY-MM-DD format",
			})
			return
		}

		timeline, err = h.container.PhotoRepo.GetPhotoTimelineByDateRange(startDate, endDate)
	} else {
		// 获取全部时间线数据
		timeline, err = h.container.PhotoRepo.GetPhotoTimeline()
	}

	if err != nil {
		logger.Error("获取时间线数据失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, model.Response{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get timeline data",
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Code:    http.StatusOK,
		Message: "Success",
		Data:    timeline,
	})
}

// PhotoColumnResponse 照片列式存储响应格式
type PhotoColumnResponse struct {
	Hash    []string  `json:"hash"`
	IsImage []bool    `json:"isImage"`
	TakenAt []string  `json:"takenAt"`
	Ratio   []float32 `json:"ratio"`
}

// GetPhotos 获取照片列表（列式存储格式）
// GET /api/v1/photos?limit=100&offset=0&start_date=2023-01-01&end_date=2023-12-31
func (h *PhotoHandler) GetPhotos(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}
	if limit > 1000 { // 限制最大数量
		limit = 1000
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	var photos []repositories.PhotoListItem

	// 如果提供了时间范围参数
	if startDateStr != "" && endDateStr != "" {
		startDate, err1 := time.Parse("2006-01-02", startDateStr)
		endDate, err2 := time.Parse("2006-01-02", endDateStr)

		if err1 != nil || err2 != nil {
			c.JSON(http.StatusBadRequest, model.Response{
				Code:    http.StatusBadRequest,
				Message: "Invalid date format. Use YYYY-MM-DD format",
			})
			return
		}

		photos, err = h.container.PhotoRepo.GetPhotoListItemsByDateRange(startDate, endDate, limit, offset)
	} else {
		// 获取全部照片列表
		photos, err = h.container.PhotoRepo.GetPhotoListItems(limit, offset)
	}

	if err != nil {
		logger.Error("获取照片列表失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, model.Response{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get photos list",
		})
		return
	}

	// 转换为列式存储格式
	columnData := PhotoColumnResponse{
		Hash:    make([]string, len(photos)),
		IsImage: make([]bool, len(photos)),
		TakenAt: make([]string, len(photos)),
		Ratio:   make([]float32, len(photos)),
	}

	for i, photo := range photos {
		columnData.Hash[i] = photo.Hash
		columnData.IsImage[i] = true // 所有照片都是图像

		// 处理拍摄时间
		if photo.TakenAt != nil {
			columnData.TakenAt[i] = photo.TakenAt.Format("2006-01-02T15:04:05")
		} else {
			columnData.TakenAt[i] = ""
		}

		columnData.Ratio[i] = photo.AspectRatio
	}

	c.JSON(http.StatusOK, model.Response{
		Code:    http.StatusOK,
		Message: "Success",
		Data:    columnData,
	})
}

// setImageHeaders 设置图像响应头
func (h *PhotoHandler) setImageHeaders(c *gin.Context, imagePath, format string) {
	// 设置Content-Type
	var contentType string
	switch format {
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

	// 设置文件名
	fileName := fmt.Sprintf("image.%s", format)
	c.Header("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, fileName))
}
