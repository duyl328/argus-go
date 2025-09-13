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

	// 这里可以添加创建目录的逻辑
	// 暂时返回成功响应，实际实现需要调用utils.FileUtils.CreateDir
	c.JSON(http.StatusOK, model.Response{
		Code:    http.StatusOK,
		Message: "创建目录功能正在开发中",
		Data: map[string]string{
			"path": req.Path,
		},
	})
}

// DeleteItem 删除文件或目录
// DELETE /api/v1/filesystem/item?path=/path/to/item
func (h *FileOperationsHandler) DeleteItem(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    http.StatusBadRequest,
			Message: "path参数不能为空",
		})
		return
	}

	logger.Info("Delete item request",
		zap.String("path", path),
		zap.String("client_ip", c.ClientIP()),
	)

	// 这里可以添加删除逻辑
	// 暂时返回成功响应，实际实现需要调用utils.FileUtils.Delete
	c.JSON(http.StatusOK, model.Response{
		Code:    http.StatusOK,
		Message: "删除功能正在开发中",
		Data: map[string]string{
			"path": path,
		},
	})
}

// MoveItem 移动/重命名文件或目录
// PUT /api/v1/filesystem/item/move
// Body: {"source": "/old/path", "destination": "/new/path"}
func (h *FileOperationsHandler) MoveItem(c *gin.Context) {
	var req struct {
		Source      string `json:"source" binding:"required"`
		Destination string `json:"destination" binding:"required"`
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
		zap.String("client_ip", c.ClientIP()),
	)

	// 这里可以添加移动逻辑
	// 暂时返回成功响应，实际实现需要调用utils.FileUtils.MoveFile
	c.JSON(http.StatusOK, model.Response{
		Code:    http.StatusOK,
		Message: "移动功能正在开发中",
		Data: map[string]interface{}{
			"source":      req.Source,
			"destination": req.Destination,
		},
	})
}

// CopyItem 复制文件或目录
// POST /api/v1/filesystem/item/copy
// Body: {"source": "/source/path", "destination": "/dest/path", "overwrite": false}
func (h *FileOperationsHandler) CopyItem(c *gin.Context) {
	var req struct {
		Source      string `json:"source" binding:"required"`
		Destination string `json:"destination" binding:"required"`
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
		zap.Bool("overwrite", req.Overwrite),
		zap.String("client_ip", c.ClientIP()),
	)

	// 这里可以添加复制逻辑
	// 暂时返回成功响应，实际实现需要调用utils.FileUtils.CopyFile或CopyDir
	c.JSON(http.StatusOK, model.Response{
		Code:    http.StatusOK,
		Message: "复制功能正在开发中",
		Data: map[string]interface{}{
			"source":      req.Source,
			"destination": req.Destination,
			"overwrite":   req.Overwrite,
		},
	})
}

// SearchFiles 搜索文件
// GET /api/v1/filesystem/search?path=/search/path&pattern=*.txt&recursive=true
func (h *FileSystemHandler) SearchFiles(c *gin.Context) {
	path := c.Query("path")
	pattern := c.Query("pattern")
	recursiveStr := c.Query("recursive")

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
		zap.Bool("recursive", recursive),
		zap.String("client_ip", c.ClientIP()),
	)

	// 这里可以添加搜索逻辑
	// 暂时返回成功响应，实际实现需要调用utils.FileUtils.SearchFiles
	c.JSON(http.StatusOK, model.Response{
		Code:    http.StatusOK,
		Message: "搜索功能正在开发中",
		Data: map[string]interface{}{
			"path":      path,
			"pattern":   pattern,
			"recursive": recursive,
		},
	})
}