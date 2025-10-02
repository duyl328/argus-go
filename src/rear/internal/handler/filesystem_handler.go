package handler

import (
	"net/http"
	"rear/internal/model"
	"rear/internal/service"
	"rear/pkg/logger"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// FileSystemHandler 文件系统处理器
type FileSystemHandler struct {
	fileSystemService *service.FileSystemService
}

// NewFileSystemHandler 创建文件系统处理器实例
func NewFileSystemHandler() *FileSystemHandler {
	return &FileSystemHandler{
		fileSystemService: service.NewFileSystemService(),
	}
}

// BrowseFileSystem 浏览文件系统
// GET /api/v1/filesystem/browse?path=/path/to/dir
func (h *FileSystemHandler) BrowseFileSystem(c *gin.Context) {
	path := c.Query("path")

	logger.Info("Browse file system request",
		zap.String("path", path),
		zap.String("client_ip", c.ClientIP()),
	)

	result, err := h.fileSystemService.Browse(path)
	if err != nil {
		logger.Error("Failed to browse file system",
			zap.String("path", path),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Code:    http.StatusOK,
		Message: "获取文件系统信息成功",
		Data:    result,
	})
}

// GetDiskUsage 获取磁盘使用情况
// GET /api/v1/filesystem/disk-usage?path=/path/to/drive
func (h *FileSystemHandler) GetDiskUsage(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    http.StatusBadRequest,
			Message: "path参数不能为空",
		})
		return
	}

	logger.Info("Get disk usage request",
		zap.String("path", path),
		zap.String("client_ip", c.ClientIP()),
	)

	usage, err := h.fileSystemService.GetDiskUsage(path)
	if err != nil {
		logger.Error("Failed to get disk usage",
			zap.String("path", path),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Code:    http.StatusOK,
		Message: "获取磁盘使用情况成功",
		Data:    usage,
	})
}

// GetFileSystemItem 获取单个文件系统项目信息
// GET /api/v1/filesystem/item?path=/path/to/item
func (h *FileSystemHandler) GetFileSystemItem(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    http.StatusBadRequest,
			Message: "path参数不能为空",
		})
		return
	}

	logger.Info("Get file system item request",
		zap.String("path", path),
		zap.String("client_ip", c.ClientIP()),
	)

	// 使用Browse方法获取父目录，然后筛选出指定项目
	result, err := h.fileSystemService.Browse(path)
	if err != nil {
		logger.Error("Failed to get file system item",
			zap.String("path", path),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Code:    http.StatusOK,
		Message: "获取文件系统项目信息成功",
		Data:    result,
	})
}

// FileOperationsHandler 文件操作处理器，用于未来扩展
type FileOperationsHandler struct {
	fileSystemService *service.FileSystemService
}

// NewFileOperationsHandler 创建文件操作处理器实例
func NewFileOperationsHandler() *FileOperationsHandler {
	return &FileOperationsHandler{
		fileSystemService: service.NewFileSystemService(),
	}
}

// CreateDirectory 创建目录
// POST /api/v1/filesystem/directory
// Body: {"path": "/path/to/new/directory"}
func (h *FileOperationsHandler) CreateDirectory(c *gin.Context) {
	var req struct {
		Path string `json:"path" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    http.StatusBadRequest,
			Message: "参数错误: " + err.Error(),
		})
		return
	}

	logger.Info("Create directory request",
		zap.String("path", req.Path),
		zap.String("client_ip", c.ClientIP()),
	)

	// 调用service创建目录
	result, err := h.fileSystemService.CreateDirectory(req.Path)
	if err != nil {
		logger.Error("Failed to create directory",
			zap.String("path", req.Path),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Code:    http.StatusOK,
		Message: "创建目录成功",
		Data:    result,
	})
}

// DeleteItem 删除文件或目录
// DELETE /api/v1/filesystem/item?path=/path/to/item&operation_id=uuid
func (h *FileOperationsHandler) DeleteItem(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    http.StatusBadRequest,
			Message: "path参数不能为空",
		})
		return
	}

	operationID := c.Query("operation_id")
	recursiveStr := c.DefaultQuery("recursive", "true")
	recursive := recursiveStr == "true"

	logger.Info("Delete item request",
		zap.String("path", path),
		zap.String("operation_id", operationID),
		zap.Bool("recursive", recursive),
		zap.String("client_ip", c.ClientIP()),
	)

	// 调用service删除
	result, err := h.fileSystemService.DeleteItem(path, operationID)
	if err != nil {
		logger.Error("Failed to delete item",
			zap.String("path", path),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Code:    http.StatusOK,
		Message: "删除成功",
		Data:    result,
	})
}

// MoveItem 移动/重命名文件或目录
// PUT /api/v1/filesystem/item/move
// Body: {"source": "/old/path", "destination": "/new/path", "operation_id": "uuid", "overwrite": false}
func (h *FileOperationsHandler) MoveItem(c *gin.Context) {
	var req struct {
		Source      string `json:"source" binding:"required"`
		Destination string `json:"destination" binding:"required"`
		OperationID string `json:"operation_id"`
		Overwrite   bool   `json:"overwrite"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    http.StatusBadRequest,
			Message: "参数错误: " + err.Error(),
		})
		return
	}

	logger.Info("Move item request",
		zap.String("source", req.Source),
		zap.String("destination", req.Destination),
		zap.String("operation_id", req.OperationID),
		zap.Bool("overwrite", req.Overwrite),
		zap.String("client_ip", c.ClientIP()),
	)

	// 调用service移动
	result, err := h.fileSystemService.MoveItem(req.Source, req.Destination, req.OperationID, req.Overwrite)
	if err != nil {
		logger.Error("Failed to move item",
			zap.String("source", req.Source),
			zap.String("destination", req.Destination),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Code:    http.StatusOK,
		Message: "移动成功",
		Data:    result,
	})
}

// CopyItem 复制文件或目录
// POST /api/v1/filesystem/item/copy
// Body: {"source": "/source/path", "destination": "/dest/path", "operation_id": "uuid", "overwrite": false}
func (h *FileOperationsHandler) CopyItem(c *gin.Context) {
	var req struct {
		Source      string `json:"source" binding:"required"`
		Destination string `json:"destination" binding:"required"`
		OperationID string `json:"operation_id"`
		Overwrite   bool   `json:"overwrite"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    http.StatusBadRequest,
			Message: "参数错误: " + err.Error(),
		})
		return
	}

	logger.Info("Copy item request",
		zap.String("source", req.Source),
		zap.String("destination", req.Destination),
		zap.String("operation_id", req.OperationID),
		zap.Bool("overwrite", req.Overwrite),
		zap.String("client_ip", c.ClientIP()),
	)

	// 调用service复制
	result, err := h.fileSystemService.CopyItem(req.Source, req.Destination, req.OperationID, req.Overwrite)
	if err != nil {
		logger.Error("Failed to copy item",
			zap.String("source", req.Source),
			zap.String("destination", req.Destination),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Code:    http.StatusOK,
		Message: "复制成功",
		Data:    result,
	})
}

// SearchFiles 搜索文件
// GET /api/v1/filesystem/search?path=/search/path&pattern=*.txt&recursive=true&type=photo
func (h *FileSystemHandler) SearchFiles(c *gin.Context) {
	path := c.Query("path")
	pattern := c.Query("pattern")
	recursiveStr := c.Query("recursive")
	fileType := c.Query("type")

	if path == "" {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    http.StatusBadRequest,
			Message: "path参数不能为空",
		})
		return
	}

	if pattern == "" {
		pattern = "*" // 默认搜索所有文件
	}

	recursive := false
	if recursiveStr != "" {
		if parsed, err := strconv.ParseBool(recursiveStr); err == nil {
			recursive = parsed
		}
	}

	logger.Info("Search files request",
		zap.String("path", path),
		zap.String("pattern", pattern),
		zap.String("type", fileType),
		zap.Bool("recursive", recursive),
		zap.String("client_ip", c.ClientIP()),
	)

	// 调用service搜索
	result, err := h.fileSystemService.SearchFiles(path, pattern, fileType, recursive)
	if err != nil {
		logger.Error("Failed to search files",
			zap.String("path", path),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Code:    http.StatusOK,
		Message: "搜索完成",
		Data:    result,
	})
}