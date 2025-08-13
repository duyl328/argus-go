package sse

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

func ExampleUsage() {
	r := gin.Default()

	opts := DefaultOptions()
	opts.PingInterval = 30 * time.Second
	opts.MaxClients = 100

	manager := NewManager(opts)

	r.GET("/events", func(c *gin.Context) {
		manager.RegisterClient(c)
	})

	r.POST("/broadcast", func(c *gin.Context) {
		var req struct {
			Message string `json:"message"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		manager.Broadcast(req.Message)
		c.JSON(200, gin.H{"status": "sent"})
	})

	eventChan := manager.GetEventChannel()

	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				data := map[string]interface{}{
					"timestamp": time.Now().Format(time.RFC3339),
					"message":   "Heartbeat from business logic",
					"clients":   manager.GetClientCount(),
				}
				jsonData, _ := json.Marshal(data)
				eventChan <- string(jsonData)
			}
		}
	}()

	r.Run(":8080")
}

type BusinessService struct {
	eventChannel chan<- string
}

func NewBusinessService(eventChannel chan<- string) *BusinessService {
	return &BusinessService{
		eventChannel: eventChannel,
	}
}

func (bs *BusinessService) ProcessData(data interface{}) {
	message := fmt.Sprintf("Data processed: %v at %s", data, time.Now().Format(time.RFC3339))

	select {
	case bs.eventChannel <- message:
		fmt.Println("Event sent to SSE manager")
	case <-time.After(1 * time.Second):
		fmt.Println("Failed to send event: channel full or closed")
	}
}

func (bs *BusinessService) NotifyUsers(userIDs []string, message string) {
	notification := map[string]interface{}{
		"type":      "notification",
		"users":     userIDs,
		"message":   message,
		"timestamp": time.Now().Format(time.RFC3339),
	}

	jsonData, _ := json.Marshal(notification)

	select {
	case bs.eventChannel <- string(jsonData):
		fmt.Printf("Notification sent to %d users\n", len(userIDs))
	case <-time.After(1 * time.Second):
		fmt.Println("Failed to send notification: timeout")
	}
}
