package handler

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
	"rear/internal/container"
	"rear/internal/model"
)

type Handler struct {
	container *container.Container
}

func NewLibraryHandler(container *container.Container) *Handler {
	return &Handler{container: container}
}

// GetLibrary 获取存储哭
func (h *Handler) GetLibrary(c *gin.Context) {
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

func (h *Handler) AddLibrary(c *gin.Context) {
	// 定义请求体结构
	type AddLibraryRequest struct {
		Path string `json:"path"`
	}

	var req AddLibraryRequest

	// 读取原始JSON数据（用于调试）
	rawData, _ := c.GetRawData()
	// 重新设置请求体，因为GetRawData()会消耗掉body
	c.Request.Body = io.NopCloser(bytes.NewBuffer(rawData))

	// 绑定JSON到结构体
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid request body",
		})
		return
	}

	// 现在可以使用 req.Path
	path := req.Path

	library := &model.LibraryTable{
		ImgPath:  path,
		IsEnable: true,
	}
	err := h.container.LibraryRepo.AddLibrary(library)
	if err != nil {
		c.JSON(http.StatusOK, model.Response{
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
func (h *Handler) UpdateLibrary(c *gin.Context) {
	var path string
	if err := c.ShouldBindJSON(&path); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid request body",
		})
		return
	}
	library := &model.LibraryTable{
		ImgPath:  path,
		IsEnable: true,
	}
	err := h.container.LibraryRepo.AddLibrary(library)
	if err != nil {
		c.JSON(http.StatusOK, model.Response{
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
func (h *Handler) DeleteLibrary(c *gin.Context) {
	var path string
	if err := c.ShouldBindJSON(&path); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid request body",
		})
		return
	}
	library := &model.LibraryTable{
		ImgPath:  path,
		IsEnable: true,
	}
	err := h.container.LibraryRepo.AddLibrary(library)
	if err != nil {
		c.JSON(http.StatusOK, model.Response{
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
