package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"rear/pkg/sse"

	"github.com/gin-gonic/gin"
)

type SSEHandler struct {
	Manager *sse.Manager // 公开字段以便其他 Handler 访问
}

type SSEResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type ClientInfo struct {
	ID          string    `json:"id"`
	RemoteAddr  string    `json:"remote_addr"`
	UserAgent   string    `json:"user_agent"`
	ConnectedAt time.Time `json:"connected_at"`
	LastPing    time.Time `json:"last_ping"`
}

func NewSSEHandler() *SSEHandler {
	opts := sse.DefaultOptions()
	opts.PingInterval = 10 * time.Second  // 10秒心跳，配合 keepalive 的 5秒
	opts.MaxClients = 500
	opts.EventBufferSize = 200

	manager := sse.NewManager(opts)

	return &SSEHandler{
		Manager: manager,
	}
}

func (h *SSEHandler) HandleSSEConnection(c *gin.Context) {
	// 添加日志确认方法被调用
	fmt.Printf("==== SSE HandleSSEConnection called, client IP: %s ====\n", c.ClientIP())

	// 对于 SSE 连接，需要设置无限期的超时
	// 通过设置 deadline 为很远的未来时间来绕过 WriteTimeout
	c.Request.Context()

	h.Manager.RegisterClient(c)
}

func (h *SSEHandler) BroadcastMessage(c *gin.Context) {
	var request struct {
		Message   string `json:"message" binding:"required"`
		EventType string `json:"event_type,omitempty"`
	}
	
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, SSEResponse{
			Success: false,
			Message: "Invalid request format: " + err.Error(),
		})
		return
	}
	
	if request.EventType != "" {
		h.Manager.BroadcastEvent(request.EventType, request.Message)
	} else {
		h.Manager.Broadcast(request.Message)
	}
	
	c.JSON(http.StatusOK, SSEResponse{
		Success: true,
		Message: "Message broadcasted successfully",
		Data: map[string]interface{}{
			"client_count": h.Manager.GetClientCount(),
			"message":      request.Message,
			"event_type":   request.EventType,
		},
	})
}

func (h *SSEHandler) SendToClient(c *gin.Context) {
	clientID := c.Param("clientId")
	if clientID == "" {
		c.JSON(http.StatusBadRequest, SSEResponse{
			Success: false,
			Message: "Client ID is required",
		})
		return
	}
	
	var request struct {
		Message string `json:"message" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, SSEResponse{
			Success: false,
			Message: "Invalid request format: " + err.Error(),
		})
		return
	}
	
	success := h.Manager.SendToClient(clientID, request.Message)
	if !success {
		c.JSON(http.StatusNotFound, SSEResponse{
			Success: false,
			Message: "Client not found or disconnected",
		})
		return
	}
	
	c.JSON(http.StatusOK, SSEResponse{
		Success: true,
		Message: "Message sent successfully",
		Data: map[string]interface{}{
			"client_id": clientID,
			"message":   request.Message,
		},
	})
}

func (h *SSEHandler) GetClients(c *gin.Context) {
	clientIDs := h.Manager.GetClientIDs()
	clients := make([]ClientInfo, 0, len(clientIDs))
	
	for _, id := range clientIDs {
		if client, exists := h.Manager.GetClientInfo(id); exists {
			clients = append(clients, ClientInfo{
				ID:          client.ID,
				RemoteAddr:  client.RemoteAddr,
				UserAgent:   client.UserAgent,
				ConnectedAt: client.ConnectedAt,
				LastPing:    client.LastPing,
			})
		}
	}
	
	c.JSON(http.StatusOK, SSEResponse{
		Success: true,
		Message: "Clients retrieved successfully",
		Data: map[string]interface{}{
			"total_clients": len(clients),
			"clients":       clients,
		},
	})
}

func (h *SSEHandler) GetClientInfo(c *gin.Context) {
	clientID := c.Param("clientId")
	if clientID == "" {
		c.JSON(http.StatusBadRequest, SSEResponse{
			Success: false,
			Message: "Client ID is required",
		})
		return
	}
	
	client, exists := h.Manager.GetClientInfo(clientID)
	if !exists {
		c.JSON(http.StatusNotFound, SSEResponse{
			Success: false,
			Message: "Client not found",
		})
		return
	}
	
	clientInfo := ClientInfo{
		ID:          client.ID,
		RemoteAddr:  client.RemoteAddr,
		UserAgent:   client.UserAgent,
		ConnectedAt: client.ConnectedAt,
		LastPing:    client.LastPing,
	}
	
	c.JSON(http.StatusOK, SSEResponse{
		Success: true,
		Message: "Client info retrieved successfully",
		Data:    clientInfo,
	})
}

func (h *SSEHandler) DisconnectClient(c *gin.Context) {
	clientID := c.Param("clientId")
	if clientID == "" {
		c.JSON(http.StatusBadRequest, SSEResponse{
			Success: false,
			Message: "Client ID is required",
		})
		return
	}
	
	h.Manager.UnregisterClient(clientID)
	
	c.JSON(http.StatusOK, SSEResponse{
		Success: true,
		Message: "Client disconnected successfully",
		Data: map[string]interface{}{
			"client_id": clientID,
		},
	})
}

func (h *SSEHandler) GetStats(c *gin.Context) {
	stats := map[string]interface{}{
		"total_clients":   h.Manager.GetClientCount(),
		"client_ids":      h.Manager.GetClientIDs(),
		"server_time":     time.Now().Format(time.RFC3339),
		"uptime_seconds":  time.Since(time.Now().Add(-time.Hour)).Seconds(), // 示例值
	}
	
	c.JSON(http.StatusOK, SSEResponse{
		Success: true,
		Message: "Stats retrieved successfully",
		Data:    stats,
	})
}

func (h *SSEHandler) TestEvent(c *gin.Context) {
	eventType := c.DefaultQuery("type", "test")
	interval := c.DefaultQuery("interval", "5")
	count := c.DefaultQuery("count", "10")
	
	intervalDur, err := strconv.Atoi(interval)
	if err != nil {
		intervalDur = 5
	}
	
	maxCount, err := strconv.Atoi(count)
	if err != nil {
		maxCount = 10
	}
	
	go func() {
		for i := 0; i < maxCount; i++ {
			message := map[string]interface{}{
				"sequence":  i + 1,
				"timestamp": time.Now().Format(time.RFC3339),
				"message":   fmt.Sprintf("Test event %d of %d", i+1, maxCount),
			}
			
			messageJSON, _ := json.Marshal(message)
			h.Manager.BroadcastEvent(eventType, string(messageJSON))
			
			time.Sleep(time.Duration(intervalDur) * time.Second)
		}
	}()
	
	c.JSON(http.StatusOK, SSEResponse{
		Success: true,
		Message: "Test events started",
		Data: map[string]interface{}{
			"event_type": eventType,
			"interval":   intervalDur,
			"count":      maxCount,
		},
	})
}

func (h *SSEHandler) GetEventChannel() chan<- string {
	return h.Manager.GetEventChannel()
}

func (h *SSEHandler) Close() {
	if h.Manager != nil {
		h.Manager.Close()
	}
}

// ============ 订阅管理接口 ============

// Subscribe 客户端订阅指定路径
// POST /api/v1/sse/subscribe
func (h *SSEHandler) Subscribe(c *gin.Context) {
	var request struct {
		ClientID string `json:"client_id" binding:"required"`
		Path     string `json:"path" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, SSEResponse{
			Success: false,
			Message: "Invalid request format: " + err.Error(),
		})
		return
	}

	err := h.Manager.Subscribe(request.ClientID, request.Path)
	if err != nil {
		c.JSON(http.StatusNotFound, SSEResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SSEResponse{
		Success: true,
		Message: "Subscribed successfully",
		Data: map[string]interface{}{
			"client_id": request.ClientID,
			"path":      request.Path,
		},
	})
}

// Unsubscribe 客户端取消订阅指定路径
// POST /api/v1/sse/unsubscribe
func (h *SSEHandler) Unsubscribe(c *gin.Context) {
	var request struct {
		ClientID string `json:"client_id" binding:"required"`
		Path     string `json:"path" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, SSEResponse{
			Success: false,
			Message: "Invalid request format: " + err.Error(),
		})
		return
	}

	err := h.Manager.Unsubscribe(request.ClientID, request.Path)
	if err != nil {
		c.JSON(http.StatusNotFound, SSEResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SSEResponse{
		Success: true,
		Message: "Unsubscribed successfully",
		Data: map[string]interface{}{
			"client_id": request.ClientID,
			"path":      request.Path,
		},
	})
}

// GetClientSubscriptions 获取客户端的所有订阅
// GET /api/v1/sse/subscriptions/:clientId
func (h *SSEHandler) GetClientSubscriptions(c *gin.Context) {
	clientID := c.Param("clientId")
	if clientID == "" {
		c.JSON(http.StatusBadRequest, SSEResponse{
			Success: false,
			Message: "Client ID is required",
		})
		return
	}

	subscriptions, err := h.Manager.GetSubscriptions(clientID)
	if err != nil {
		c.JSON(http.StatusNotFound, SSEResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SSEResponse{
		Success: true,
		Message: "Subscriptions retrieved successfully",
		Data: map[string]interface{}{
			"client_id":     clientID,
			"subscriptions": subscriptions,
			"count":         len(subscriptions),
		},
	})
}