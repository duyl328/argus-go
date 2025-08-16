package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"rear/internal/handler"
)

func main1() {
	r := gin.Default()

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization", "X-Request-ID"}
	r.Use(cors.New(corsConfig))

	sseHandler := handler.NewSSEHandler()

	setupSSERoutes(r, sseHandler)

	simulateBusinessEvents(sseHandler.GetEventChannel())

	fmt.Println("Server starting on :8080")
	fmt.Println("SSE endpoint: http://localhost:8080/api/v1/events")
	fmt.Println("Admin panel: http://localhost:8080/admin/sse/stats")

	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func setupSSERoutes(r *gin.Engine, sseHandler *handler.SSEHandler) {
	api := r.Group("/api/v1")
	{
		api.GET("/events", sseHandler.HandleSSEConnection)

		api.POST("/events/broadcast", sseHandler.BroadcastMessage)

		api.POST("/events/send/:clientId", sseHandler.SendToClient)

		api.GET("/events/clients", sseHandler.GetClients)
		api.GET("/events/clients/:clientId", sseHandler.GetClientInfo)
		api.DELETE("/events/clients/:clientId", sseHandler.DisconnectClient)

		api.GET("/events/stats", sseHandler.GetStats)

		api.POST("/events/test", sseHandler.TestEvent)
	}

	admin := r.Group("/admin")
	{
		admin.Static("/sse", "./examples/sse_admin")

		admin.GET("/sse/stats", sseHandler.GetStats)
	}

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "SSE Server is running",
			"endpoints": map[string]string{
				"sse_connection": "/api/v1/events",
				"broadcast":      "POST /api/v1/events/broadcast",
				"send_to_client": "POST /api/v1/events/send/:clientId",
				"get_clients":    "GET /api/v1/events/clients",
				"stats":          "GET /api/v1/events/stats",
				"test_events":    "POST /api/v1/events/test",
			},
		})
	})
}

func simulateBusinessEvents(eventChannel chan<- string) {
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()

		counter := 0
		for {
			select {
			case <-ticker.C:
				counter++
				event := map[string]interface{}{
					"id":        fmt.Sprintf("business-event-%d", counter),
					"type":      "business_update",
					"message":   fmt.Sprintf("定期业务更新 #%d", counter),
					"timestamp": time.Now().Format(time.RFC3339),
					"data": map[string]interface{}{
						"counter":     counter,
						"server_time": time.Now().Unix(),
						"status":      "active",
					},
				}

				eventJSON, _ := json.Marshal(event)

				select {
				case eventChannel <- string(eventJSON):
					fmt.Printf("Business event %d sent to SSE manager\n", counter)
				case <-time.After(1 * time.Second):
					fmt.Printf("Failed to send business event %d: timeout\n", counter)
				}
			}
		}
	}()

	go func() {
		imageTicker := time.NewTicker(30 * time.Second)
		defer imageTicker.Stop()

		imageCounter := 0
		for {
			select {
			case <-imageTicker.C:
				imageCounter++
				imageEvent := map[string]interface{}{
					"id":      fmt.Sprintf("image-processed-%d", imageCounter),
					"type":    "image_processed",
					"message": fmt.Sprintf("图片处理完成 #%d", imageCounter),
					"data": map[string]interface{}{
						"image_id":     fmt.Sprintf("img_%d", imageCounter),
						"filename":     fmt.Sprintf("photo_%d.jpg", imageCounter),
						"status":       "completed",
						"thumbnail":    fmt.Sprintf("/thumbnails/photo_%d_thumb.jpg", imageCounter),
						"process_time": fmt.Sprintf("%.2fs", float64(imageCounter)*0.5),
					},
				}

				eventJSON, _ := json.Marshal(imageEvent)

				select {
				case eventChannel <- string(eventJSON):
					fmt.Printf("Image processing event %d sent\n", imageCounter)
				case <-time.After(1 * time.Second):
					fmt.Printf("Failed to send image event %d\n", imageCounter)
				}
			}
		}
	}()

	go func() {
		userTicker := time.NewTicker(45 * time.Second)
		defer userTicker.Stop()

		actions := []string{"登录", "上传照片", "创建相册", "分享照片", "删除照片"}

		for {
			select {
			case <-userTicker.C:
				action := actions[time.Now().Second()%len(actions)]
				userEvent := map[string]interface{}{
					"id":      fmt.Sprintf("user-action-%d", time.Now().Unix()),
					"type":    "user_activity",
					"message": fmt.Sprintf("用户活动: %s", action),
					"data": map[string]interface{}{
						"action":    action,
						"user_id":   fmt.Sprintf("user_%d", (time.Now().Unix()%1000)+1),
						"timestamp": time.Now().Unix(),
					},
				}

				eventJSON, _ := json.Marshal(userEvent)

				select {
				case eventChannel <- string(eventJSON):
					fmt.Printf("User activity event sent: %s\n", action)
				case <-time.After(1 * time.Second):
					fmt.Printf("Failed to send user activity event\n")
				}
			}
		}
	}()
}

type PhotoProcessingService struct {
	eventChannel chan<- string
}

func NewPhotoProcessingService(eventChannel chan<- string) *PhotoProcessingService {
	return &PhotoProcessingService{
		eventChannel: eventChannel,
	}
}

func (pps *PhotoProcessingService) ProcessPhoto(photoID, filename string) {
	fmt.Printf("开始处理照片: %s\n", filename)

	startEvent := map[string]interface{}{
		"type":    "photo_processing_started",
		"message": fmt.Sprintf("开始处理照片: %s", filename),
		"data": map[string]interface{}{
			"photo_id": photoID,
			"filename": filename,
			"status":   "processing",
		},
	}
	pps.sendEvent(startEvent)

	time.Sleep(2 * time.Second)

	completedEvent := map[string]interface{}{
		"type":    "photo_processing_completed",
		"message": fmt.Sprintf("照片处理完成: %s", filename),
		"data": map[string]interface{}{
			"photo_id":  photoID,
			"filename":  filename,
			"status":    "completed",
			"thumbnail": fmt.Sprintf("/thumbnails/%s_thumb.jpg", photoID),
		},
	}
	pps.sendEvent(completedEvent)
}

func (pps *PhotoProcessingService) sendEvent(event map[string]interface{}) {
	event["id"] = fmt.Sprintf("photo-event-%d", time.Now().UnixNano())
	event["timestamp"] = time.Now().Format(time.RFC3339)

	eventJSON, _ := json.Marshal(event)

	select {
	case pps.eventChannel <- string(eventJSON):
	case <-time.After(1 * time.Second):
		fmt.Println("Failed to send photo processing event")
	}
}
