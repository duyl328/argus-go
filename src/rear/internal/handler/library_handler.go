package handler

import (
	"fmt"
	"go.uber.org/zap"
	"log"
	"net/http"
	"rear/internal/config"
	"rear/internal/container"
	"rear/internal/model"
	"rear/internal/model/tables"
	"rear/pkg/logger"
	"rear/pkg/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

type LibraryHandler struct {
	container  *container.DbContainer
	imgContain *container.TaskContainer
}

func NewLibraryHandler(container *container.DbContainer, imgContain *container.TaskContainer) *LibraryHandler {
	return &LibraryHandler{container: container, imgContain: imgContain}
}

// GetLibrary 获取存储
func (h *LibraryHandler) GetLibrary(c *gin.Context) {
	libraries, err := h.container.LibraryRepo.GetAllLibrary()
	if err != nil {
		log.Printf("Failed to get all libraries: %v", err)
		c.JSON(http.StatusInternalServerError, model.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Data:    nil,
		})
		return
	}

	log.Printf("Total libraries: %d", len(libraries))
	c.JSON(http.StatusOK, model.Response{
		Code:    http.StatusOK,
		Message: "Success",
		Data:    libraries,
	})
}
func (h *LibraryHandler) AddLibrary(c *gin.Context) {
	type AddLibraryRequest struct {
		Path string `json:"path"`
	}

	var req AddLibraryRequest

	// 绑定JSON到结构体
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("Invalid request body: %v", err),
		})
		return
	}

	// 验证路径
	path := strings.TrimSpace(req.Path)
	log.Printf("Add library path: %s", path)
	if path == "" {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    http.StatusBadRequest,
			Message: "Path cannot be empty",
		})
		return
	}

	library := &tables.LibraryTable{
		ImgPath:  path,
		IsEnable: true,
	}

	if err := h.container.LibraryRepo.AddLibrary(library); err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Code:    http.StatusOK,
		Message: "Success",
		Data:    nil,
	})
}

func (h *LibraryHandler) UpdateLibrary(c *gin.Context) {
	type UpdateLibraryRequest struct {
		Path     string `json:"path"`
		IsEnable bool   `json:"is_enable"`
	}

	var req UpdateLibraryRequest

	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Path) == "" {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid request body or missing path",
		})
		return
	}

	err := h.container.LibraryRepo.UpdateLibrary(req.Path, req.IsEnable)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Code:    http.StatusOK,
		Message: "Library updated successfully",
	})
}

func (h *LibraryHandler) DeleteLibrary(c *gin.Context) {
	path := c.Query("path")
	if strings.TrimSpace(path) == "" {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    http.StatusBadRequest,
			Message: "Missing path query parameter",
		})
		return
	}

	if err := h.container.LibraryRepo.DeleteLibrary(path); err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Code:    http.StatusOK,
		Message: "Library deleted successfully",
	})
}

// LibraryIndex 开始图片检索【缩略图生成】
func (h *LibraryHandler) LibraryIndex(c *gin.Context) {
	// 获取所有已添加路径
	library, err := h.container.LibraryRepo.GetAllLibrary()
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}
	if len(library) == 0 {
		c.JSON(http.StatusInternalServerError, model.Response{
			Code:    http.StatusOK,
			Message: "无可用路径, 任务取消!",
		})
		return
	}

	// 存在可用路径，检索开始
	var dirs []string // 创建一个空的 string 切片
	for i := range library {
		path := library[i]
		isDir := utils.FileUtils.IsDir(path.ImgPath)
		if isDir {
			dirs = append(dirs, path.ImgPath)
		}
	}
	// 无用文件夹是否要进行提示【TODO】
	if len(dirs) == 0 {
		c.JSON(http.StatusInternalServerError, model.Response{
			Code:    http.StatusOK,
			Message: "无可用路径, 任务结束!",
		})
		return
	}

	// 获取可用路径下所有的照片【指定类型】
	logger.Info(fmt.Sprintf("检索的列表: %v", dirs))
	var fileList []string // 创建一个空的 string 切片
	for i := range dirs {
		dir := dirs[i]

		files, err := utils.FileUtils.GetFilteredFiles(dir, true, config.CONFIG.BaseSupportedFileTypes)
		if err != nil {
			logger.Error("文件获取失败！", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.Response{
				Code:    http.StatusInternalServerError,
				Message: fmt.Sprintf("指定路径 %v 文件获取失败!", dir),
			})
			return
		}
		for i2 := range files.SupportedFiles {
			info := files.SupportedFiles[i2]
			fileList = append(fileList, info.Path)
		}
	}
	logger.Info(fmt.Sprintf("得到的照片总数: %d", len(fileList)))

	// 检查哪些文件已经在数据库中存在，过滤出需要处理的文件
	var newFiles []string
	skippedCount := 0

	for _, file := range fileList {
		// 检查数据库中是否已经存在该文件记录
		_, err := h.container.PhotoRepo.GetByPath(file)
		if err != nil {
			// 数据库中不存在该记录，需要处理
			newFiles = append(newFiles, file)
		} else {
			// 数据库中已存在该记录，跳过
			skippedCount++
		}
	}

	logger.Info(fmt.Sprintf("需要处理的新文件: %d, 跳过已存在的文件: %d", len(newFiles), skippedCount))

	// 异步启动后台任务，开始处理新文件
	go func() {
		for _, file := range newFiles {
			h.imgContain.ImgTaskManager.AddTask(file)
		}
		logger.Info(fmt.Sprintf("已添加 %d 个任务到处理队列", len(newFiles)))
	}()

	// 立即返回响应
	c.JSON(http.StatusOK, model.Response{
		Code: http.StatusOK,
		Message: fmt.Sprintf("索引任务已启动，共发现 %d 个文件，其中 %d 个新文件需要处理，%d 个文件已存在跳过",
			len(fileList), len(newFiles), skippedCount),
		Data: map[string]interface{}{
			"total_files":   len(fileList),
			"new_files":     len(newFiles),
			"skipped_files": skippedCount,
		},
	})
}
