package handler

import (
	"github.com/gin-gonic/gin"
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
