package handler

import (
	"fmt"
	"rear/internal/container"
	"rear/internal/utils/tools"
	"rear/pkg/logger"

	"github.com/gin-gonic/gin"
)

type DevImageHandler struct {
	container *container.Container
}

func NewDevImageHandler(container *container.Container) *DevImageHandler {
	return &DevImageHandler{container: container}
}

// GetExif 获取图片的EXIF信息
func (h *DevImageHandler) GetExif(c *gin.Context) {
	// 从 query 参数中获取图片路径
	logger.Warn("GetExif called, processing request...")
	imagePath := c.Query("image_path")

	result := fmt.Sprintf("File generation done. %s\n", imagePath)
	logger.Warn(result)
	if imagePath == "" {
		c.JSON(400, gin.H{"error": "image_path query parameter is required"})
		return
	}
	// 调用工具函数获取 EXIF 数据
	exifData, err := tools.GetExifData(c, imagePath)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to get EXIF data"})
		return
	}

	c.JSON(200, gin.H{"exif_data": exifData})
}
