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
	manager *sse.Manager
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
	opts.PingInterval = 30 * time.Second
	opts.MaxClients = 500
	opts.EventBufferSize = 200
	
	manager := sse.NewManager(opts)
	
	return &SSEHandler{
		manager: manager,
	}
}

func (h *SSEHandler) HandleSSEConnection(c *gin.Context) {
	h.manager.RegisterClient(c)
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
		h.manager.BroadcastEvent(request.EventType, request.Message)
	} else {
		h.manager.Broadcast(request.Message)
	}
	
	c.JSON(http.StatusOK, SSEResponse{
		Success: true,
		Message: "Message broadcasted successfully",
		Data: map[string]interface{}{
			"client_count": h.manager.GetClientCount(),
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
	
	success := h.manager.SendToClient(clientID, request.Message)
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
	clientIDs := h.manager.GetClientIDs()
	clients := make([]ClientInfo, 0, len(clientIDs))
	
	for _, id := range clientIDs {
		if client, exists := h.manager.GetClientInfo(id); exists {
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
	
	client, exists := h.manager.GetClientInfo(clientID)
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
	
	h.manager.UnregisterClient(clientID)
	
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
		"total_clients":   h.manager.GetClientCount(),
		"client_ids":      h.manager.GetClientIDs(),
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
			h.manager.BroadcastEvent(eventType, string(messageJSON))
			
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
	return h.manager.GetEventChannel()
}

func (h *SSEHandler) Close() {
	if h.manager != nil {
		h.manager.Close()
	}
}