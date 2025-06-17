package handler

import (
	"fmt"
	"log"
	"net/http"
	"rear/internal/container"
	"rear/internal/model"
	"strings"

	"github.com/gin-gonic/gin"
)

type LibraryHandler struct {
	container *container.Container
}

func NewLibraryHandler(container *container.Container) *LibraryHandler {
	return &LibraryHandler{container: container}
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
	if path == "" {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    http.StatusBadRequest,
			Message: "Path cannot be empty",
		})
		return
	}

	library := &model.LibraryTable{
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
