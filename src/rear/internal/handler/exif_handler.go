package handler

import (
	"net/http"
	"strconv"

	"rear/internal/container"
	baseModel "rear/internal/model"

	"github.com/gin-gonic/gin"
)

type ExifHandler struct {
	container *container.DbContainer
}

func NewExifHandler(container *container.DbContainer) *ExifHandler {
	return &ExifHandler{
		container: container,
	}
}

// GetExifByHash 根据Hash获取EXIF信息
func (h *ExifHandler) GetExifByHash(c *gin.Context) {
	hash := c.Param("hash")
	if hash == "" {
		c.JSON(http.StatusBadRequest, baseModel.Response{
			Code:    http.StatusBadRequest,
			Message: "Hash不能为空",
		})
		return
	}

	exif, err := h.container.ExifRepo.GetByHash(hash)
	if err != nil {
		c.JSON(http.StatusNotFound, baseModel.Response{
			Code:    http.StatusNotFound,
			Message: "未找到EXIF信息",
		})
		return
	}

	// 转换为ParsedExif格式返回
	parsedExif := exif.ToParsedExif()

	c.JSON(http.StatusOK, baseModel.Response{
		Code:    http.StatusOK,
		Message: "成功获取EXIF信息",
		Data:    parsedExif,
	})
}

// GetCameraStatistics 获取相机统计信息
func (h *ExifHandler) GetCameraStatistics(c *gin.Context) {
	cameras, err := h.container.ExifRepo.GetCameraStatistics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, baseModel.Response{
			Code:    http.StatusInternalServerError,
			Message: "获取相机统计信息失败",
		})
		return
	}

	c.JSON(http.StatusOK, baseModel.Response{
		Code:    http.StatusOK,
		Message: "成功获取相机统计信息",
		Data:    cameras,
	})
}

// GetExifsByCamera 根据相机型号获取EXIF记录
func (h *ExifHandler) GetExifsByCamera(c *gin.Context) {
	make := c.Query("make")
	model := c.Query("model")
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	exifs, err := h.container.ExifRepo.GetExifsByCamera(make, model, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, baseModel.Response{
			Code:    http.StatusInternalServerError,
			Message: "获取EXIF记录失败",
		})
		return
	}

	c.JSON(http.StatusOK, baseModel.Response{
		Code:    http.StatusOK,
		Message: "成功获取EXIF记录",
		Data:    exifs,
	})
}

// GetExifsWithGPS 获取包含GPS信息的EXIF记录
func (h *ExifHandler) GetExifsWithGPS(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	exifs, err := h.container.ExifRepo.GetExifsWithGPS(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, baseModel.Response{
			Code:    http.StatusInternalServerError,
			Message: "获取GPS EXIF记录失败",
		})
		return
	}

	c.JSON(http.StatusOK, baseModel.Response{
		Code:    http.StatusOK,
		Message: "成功获取GPS EXIF记录",
		Data:    exifs,
	})
}

// SearchExifs 搜索EXIF记录
func (h *ExifHandler) SearchExifs(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		c.JSON(http.StatusBadRequest, baseModel.Response{
			Code:    http.StatusBadRequest,
			Message: "搜索关键词不能为空",
		})
		return
	}

	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	exifs, err := h.container.ExifRepo.SearchExifs(keyword, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, baseModel.Response{
			Code:    http.StatusInternalServerError,
			Message: "搜索EXIF记录失败",
		})
		return
	}

	c.JSON(http.StatusOK, baseModel.Response{
		Code:    http.StatusOK,
		Message: "成功搜索EXIF记录",
		Data:    exifs,
	})
}

// GetUniqueCameras 获取所有唯一的相机信息
func (h *ExifHandler) GetUniqueCameras(c *gin.Context) {
	cameras, err := h.container.ExifRepo.GetUniqueCameras()
	if err != nil {
		c.JSON(http.StatusInternalServerError, baseModel.Response{
			Code:    http.StatusInternalServerError,
			Message: "获取相机列表失败",
		})
		return
	}

	c.JSON(http.StatusOK, baseModel.Response{
		Code:    http.StatusOK,
		Message: "成功获取相机列表",
		Data:    cameras,
	})
}

// GetExifsByISO 根据ISO范围获取EXIF记录
func (h *ExifHandler) GetExifsByISO(c *gin.Context) {
	minISOStr := c.DefaultQuery("min_iso", "0")
	maxISOStr := c.DefaultQuery("max_iso", "51200")
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	minISO, _ := strconv.Atoi(minISOStr)
	maxISO, _ := strconv.Atoi(maxISOStr)
	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	exifs, err := h.container.ExifRepo.GetExifsByISO(minISO, maxISO, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, baseModel.Response{
			Code:    http.StatusInternalServerError,
			Message: "获取ISO EXIF记录失败",
		})
		return
	}

	c.JSON(http.StatusOK, baseModel.Response{
		Code:    http.StatusOK,
		Message: "成功获取ISO EXIF记录",
		Data:    exifs,
	})
}

// GetExifsByAperture 根据光圈范围获取EXIF记录
func (h *ExifHandler) GetExifsByAperture(c *gin.Context) {
	minApertureStr := c.DefaultQuery("min_aperture", "1.0")
	maxApertureStr := c.DefaultQuery("max_aperture", "32.0")
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	minAperture, _ := strconv.ParseFloat(minApertureStr, 64)
	maxAperture, _ := strconv.ParseFloat(maxApertureStr, 64)
	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	exifs, err := h.container.ExifRepo.GetExifsByAperture(minAperture, maxAperture, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, baseModel.Response{
			Code:    http.StatusInternalServerError,
			Message: "获取光圈 EXIF记录失败",
		})
		return
	}

	c.JSON(http.StatusOK, baseModel.Response{
		Code:    http.StatusOK,
		Message: "成功获取光圈 EXIF记录",
		Data:    exifs,
	})
}

// GetExifStatistics 获取EXIF统计信息
func (h *ExifHandler) GetExifStatistics(c *gin.Context) {
	count, err := h.container.ExifRepo.GetCount()
	if err != nil {
		c.JSON(http.StatusInternalServerError, baseModel.Response{
			Code:    http.StatusInternalServerError,
			Message: "获取统计信息失败",
		})
		return
	}

	stats := map[string]interface{}{
		"total_count": count,
	}

	c.JSON(http.StatusOK, baseModel.Response{
		Code:    http.StatusOK,
		Message: "成功获取统计信息",
		Data:    stats,
	})
}

// GetAllExifs 获取所有EXIF记录
func (h *ExifHandler) GetAllExifs(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	exifs, err := h.container.ExifRepo.GetAll(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, baseModel.Response{
			Code:    http.StatusInternalServerError,
			Message: "获取EXIF记录失败",
		})
		return
	}

	c.JSON(http.StatusOK, baseModel.Response{
		Code:    http.StatusOK,
		Message: "成功获取EXIF记录",
		Data:    exifs,
	})
}
