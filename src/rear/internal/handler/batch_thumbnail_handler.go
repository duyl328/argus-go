package handler

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"rear/internal/api"
	"rear/pkg/logger"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// BatchThumbnailHandler 批量缩略图处理器
type BatchThumbnailHandler struct{}

// NewBatchThumbnailHandler 创建批量缩略图处理器
func NewBatchThumbnailHandler() *BatchThumbnailHandler {
	return &BatchThumbnailHandler{}
}

// BatchThumbnailRequest 批量缩略图请求
type BatchThumbnailRequest struct {
	Paths []string `json:"paths" binding:"required"` // 文件路径列表
	Size  int      `json:"size"`                     // 缩略图尺寸（统一）
}

// GetBatchThumbnails 批量获取缩略图（HTTP Multipart Streaming）
// POST /api/v1/photo/batch-preview
// Body: {"paths": ["/path/1.jpg", "/path/2.jpg"], "size": 512}
// Response: multipart/mixed，每个part包含一张缩略图
func (h *BatchThumbnailHandler) GetBatchThumbnails(c *gin.Context) {
	var req BatchThumbnailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Invalid request: %v", err),
		})
		return
	}

	// 验证路径数量
	if len(req.Paths) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No paths provided",
		})
		return
	}

	if len(req.Paths) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Too many paths (max 100)",
		})
		return
	}

	// 默认尺寸
	if req.Size <= 0 {
		req.Size = 512
	}

	logger.Info("批量缩略图请求",
		zap.Int("count", len(req.Paths)),
		zap.Int("size", req.Size))

	// 设置multipart响应头
	boundary := "argus-thumbnail-boundary"
	c.Header("Content-Type", fmt.Sprintf("multipart/mixed; boundary=%s", boundary))
	c.Header("Transfer-Encoding", "chunked")
	c.Status(http.StatusOK)

	// 创建multipart writer
	mw := multipart.NewWriter(c.Writer)
	mw.SetBoundary(boundary)
	defer mw.Close()

	ctx := context.Background()
	successCount := 0
	errorCount := 0

	// 使用并发处理（限制并发数）
	maxConcurrency := 5
	sem := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup
	var mu sync.Mutex // 保护writer写入

	for idx, path := range req.Paths {
		wg.Add(1)
		sem <- struct{}{} // 获取信号量

		go func(index int, filePath string) {
			defer wg.Done()
			defer func() { <-sem }() // 释放信号量

			// 处理单个缩略图
			err := h.processSingleThumbnail(ctx, mw, &mu, index, filePath, req.Size)

			mu.Lock()
			if err != nil {
				errorCount++
				logger.Warn("缩略图生成失败",
					zap.String("path", filePath),
					zap.Error(err))
			} else {
				successCount++
			}
			mu.Unlock()

			// 刷新缓冲区，立即发送给客户端
			if flusher, ok := c.Writer.(http.Flusher); ok {
				flusher.Flush()
			}
		}(idx, path)
	}

	// 等待所有任务完成
	wg.Wait()

	logger.Info("批量缩略图处理完成",
		zap.Int("total", len(req.Paths)),
		zap.Int("success", successCount),
		zap.Int("errors", errorCount))
}

// processSingleThumbnail 处理单个缩略图
func (h *BatchThumbnailHandler) processSingleThumbnail(
	ctx context.Context,
	mw *multipart.Writer,
	mu *sync.Mutex,
	index int,
	filePath string,
	size int,
) error {
	// 创建ImageAPI
	imageAPI, err := api.NewImageAPI(filePath)
	if err != nil {
		return h.writeThumbnailError(mw, mu, index, filePath, err)
	}

	// 获取缩略图路径
	thumbnailPath, err := imageAPI.GetImagePath(size)
	if err != nil {
		return h.writeThumbnailError(mw, mu, index, filePath, err)
	}

	// 读取缩略图文件
	thumbnailData, err := api.ReadFileWithLimit(thumbnailPath, 10*1024*1024) // 限制10MB
	if err != nil {
		return h.writeThumbnailError(mw, mu, index, filePath, err)
	}

	// 获取格式
	format := imageAPI.GetFormat()
	var contentType string
	switch format {
	case "jpg", "jpeg":
		contentType = "image/jpeg"
	case "png":
		contentType = "image/png"
	case "webp":
		contentType = "image/webp"
	default:
		contentType = "application/octet-stream"
	}

	// 写入multipart part（需要加锁）
	mu.Lock()
	defer mu.Unlock()

	// 创建part header
	partHeader := make(textproto.MIMEHeader)
	partHeader.Set("Content-Type", contentType)
	partHeader.Set("Content-Disposition", fmt.Sprintf(`inline; filename="thumb_%d.%s"`, index, format))
	partHeader.Set("X-Original-Path", filePath)
	partHeader.Set("X-Thumbnail-Size", strconv.Itoa(size))
	partHeader.Set("X-Index", strconv.Itoa(index))
	partHeader.Set("Content-Length", strconv.Itoa(len(thumbnailData)))

	// 创建part
	part, err := mw.CreatePart(partHeader)
	if err != nil {
		return fmt.Errorf("create part failed: %w", err)
	}

	// 写入图片数据
	_, err = part.Write(thumbnailData)
	if err != nil {
		return fmt.Errorf("write thumbnail data failed: %w", err)
	}

	return nil
}

// writeThumbnailError 写入错误响应part
func (h *BatchThumbnailHandler) writeThumbnailError(
	mw *multipart.Writer,
	mu *sync.Mutex,
	index int,
	filePath string,
	err error,
) error {
	mu.Lock()
	defer mu.Unlock()

	// 创建错误part
	partHeader := make(textproto.MIMEHeader)
	partHeader.Set("Content-Type", "application/json")
	partHeader.Set("X-Original-Path", filePath)
	partHeader.Set("X-Index", strconv.Itoa(index))
	partHeader.Set("X-Error", "true")

	part, createErr := mw.CreatePart(partHeader)
	if createErr != nil {
		return fmt.Errorf("create error part failed: %w", createErr)
	}

	errorData := fmt.Sprintf(`{"index": %d, "path": "%s", "error": "%s"}`, index, filePath, err.Error())
	_, writeErr := io.WriteString(part, errorData)
	if writeErr != nil {
		return fmt.Errorf("write error data failed: %w", writeErr)
	}

	return err
}
