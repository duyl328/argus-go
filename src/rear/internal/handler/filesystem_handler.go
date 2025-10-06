package handler

import (
	"encoding/json"
	"net/http"
	"rear/internal/model"
	"rear/internal/service"
	"rear/internal/utils"
	"rear/pkg/logger"
	"rear/pkg/sse"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// FileSystemHandler 文件系统处理器
type FileSystemHandler struct {
	fileSystemService *service.FileSystemService
	fileWatcher       *utils.FileSystemWatcher
	sseManager        *sse.Manager
	eventThrottle     chan struct{} // 事件节流通道
}

// NewFileSystemHandler 创建文件系统处理器实例
func NewFileSystemHandler(sseManager *sse.Manager) *FileSystemHandler {
	// 创建事件节流通道（最多缓存10个待发送事件）
	eventThrottle := make(chan struct{}, 10)

	// 创建文件系统监听器，事件回调通过 SSE 推送
	fileWatcher, err := utils.NewFileSystemWatcher(func(event *utils.FileSystemEvent) {
		// 防御性编程：捕获回调中的panic
		defer func() {
			if r := recover(); r != nil {
				logger.Error("文件系统事件回调panic",
					zap.Any("panic", r),
					zap.String("type", event.Type),
					zap.String("path", event.Path))
			}
		}()

		// 验证事件有效性
		if event == nil {
			logger.Warn("收到空的文件系统事件")
			return
		}

		// 将事件转换为 JSON
		eventData, err := json.Marshal(event)
		if err != nil {
			logger.Error("序列化文件系统事件失败",
				zap.Error(err),
				zap.String("type", event.Type),
				zap.String("path", event.Path))
			return
		}

		// 验证JSON字符串的有效性
		eventString := string(eventData)
		if len(eventString) == 0 {
			logger.Error("序列化后的JSON为空",
				zap.String("type", event.Type),
				zap.String("path", event.Path))
			return
		}

		// 检查是否包含非法控制字符
		for i, r := range eventString {
			if r < 32 && r != '\n' && r != '\r' && r != '\t' {
				logger.Error("JSON包含非法控制字符",
					zap.Int("position", i),
					zap.Int32("char", r),
					zap.String("json", eventString))
				return
			}
		}

		// 非阻塞发送：如果节流通道已满，丢弃事件（避免阻塞监听器）
		select {
		case eventThrottle <- struct{}{}:
			// 异步广播事件，避免阻塞 fsnotify 循环
			go func(eventStr string, ev *utils.FileSystemEvent) {
				defer func() {
					<-eventThrottle // 释放节流槽位
					if r := recover(); r != nil {
						logger.Error("SSE广播panic", zap.Any("panic", r))
					}
				}()

				// 使用新的订阅广播方法：只向订阅了该路径的客户端发送
				// 使用事件的 WatchedPath 字段作为订阅路径
				sseManager.BroadcastToSubscribers(ev.WatchedPath, "filesystem-change", eventStr)

				// 🟢 关键日志: 记录每次文件系统变化推送
				logger.Info("📤 文件系统变化事件已推送",
					zap.String("event_type", ev.Type),
					zap.String("file_path", ev.Path),
					zap.String("file_name", ev.Name),
					zap.String("watched_path", ev.WatchedPath),
					zap.Bool("is_dir", ev.IsDir),
					zap.String("timestamp", ev.Timestamp.Format("2006-01-02 15:04:05")))
			}(eventString, event)
		default:
			// 节流通道已满，丢弃事件
			logger.Warn("文件系统事件频率过高，跳过事件",
				zap.String("type", event.Type),
				zap.String("path", event.Path))
		}
	})

	if err != nil {
		logger.Error("创建文件系统监听器失败", zap.Error(err))
		return nil
	}

	// 启动监听器
	fileWatcher.Start()

	return &FileSystemHandler{
		fileSystemService: service.NewFileSystemService(),
		fileWatcher:       fileWatcher,
		sseManager:        sseManager,
		eventThrottle:     eventThrottle,
	}
}

// GetFileWatcher 获取文件监听器（用于清理）
func (h *FileSystemHandler) GetFileWatcher() *utils.FileSystemWatcher {
	return h.fileWatcher
}

// BrowseFileSystem 浏览文件系统
// GET /api/v1/filesystem/browse?path=/path/to/dir
func (h *FileSystemHandler) BrowseFileSystem(c *gin.Context) {
	path := c.Query("path")

	logger.Info("🔄 浏览文件系统请求",
		zap.String("path", path),
		zap.String("client_ip", c.ClientIP()),
	)

	result, err := h.fileSystemService.Browse(path)
	if err != nil {
		logger.Error("❌ 浏览文件系统失败",
			zap.String("path", path),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.Info("✅ 浏览文件系统成功",
		zap.String("current_path", result.CurrentPath),
		zap.Int("items_count", len(result.Items)),
	)

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

// WatchPath 订阅文件夹监听
// POST /api/v1/filesystem/watch
// Body: {"path": "/path/to/watch"}
func (h *FileSystemHandler) WatchPath(c *gin.Context) {
	var req struct {
		Path string `json:"path" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    http.StatusBadRequest,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	logger.Info("👀 订阅文件夹监听请求",
		zap.String("path", req.Path),
		zap.String("client_ip", c.ClientIP()),
	)

	// 添加监听
	err := h.fileWatcher.Watch(req.Path)
	if err != nil {
		logger.Error("❌ 添加监听失败",
			zap.String("path", req.Path),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, model.Response{
			Code:    http.StatusInternalServerError,
			Message: "添加监听失败: " + err.Error(),
		})
		return
	}

	watchedPaths := h.fileWatcher.GetWatchedPaths()
	logger.Info("✅ 已开始监听文件夹",
		zap.String("path", req.Path),
		zap.Int("total_watched", len(watchedPaths)),
	)

	c.JSON(http.StatusOK, model.Response{
		Code:    http.StatusOK,
		Message: "已开始监听文件夹",
		Data: map[string]interface{}{
			"path":          req.Path,
			"watched_paths": watchedPaths,
		},
	})
}

// UnwatchPath 取消文件夹监听
// POST /api/v1/filesystem/unwatch
// Body: {"path": "/path/to/unwatch"}
func (h *FileSystemHandler) UnwatchPath(c *gin.Context) {
	var req struct {
		Path string `json:"path" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    http.StatusBadRequest,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	logger.Info("🚫 取消文件夹监听请求",
		zap.String("path", req.Path),
		zap.String("client_ip", c.ClientIP()),
	)

	// 移除监听
	err := h.fileWatcher.Unwatch(req.Path)
	if err != nil {
		logger.Error("❌ 移除监听失败",
			zap.String("path", req.Path),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, model.Response{
			Code:    http.StatusInternalServerError,
			Message: "移除监听失败: " + err.Error(),
		})
		return
	}

	watchedPaths := h.fileWatcher.GetWatchedPaths()
	logger.Info("✅ 已停止监听文件夹",
		zap.String("path", req.Path),
		zap.Int("total_watched", len(watchedPaths)),
	)

	c.JSON(http.StatusOK, model.Response{
		Code:    http.StatusOK,
		Message: "已停止监听文件夹",
		Data: map[string]interface{}{
			"path":          req.Path,
			"watched_paths": watchedPaths,
		},
	})
}

// GetWatchedPaths 获取当前所有监听的路径
// GET /api/v1/filesystem/watched
func (h *FileSystemHandler) GetWatchedPaths(c *gin.Context) {
	paths := h.fileWatcher.GetWatchedPaths()

	c.JSON(http.StatusOK, model.Response{
		Code:    http.StatusOK,
		Message: "获取监听路径成功",
		Data: map[string]interface{}{
			"count": len(paths),
			"paths": paths,
		},
	})
}