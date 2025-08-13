# SSE (Server-Sent Events) 管理器

一个功能完整、生产就绪的 Go SSE 管理器，用于实现服务端到客户端的实时数据推送。

## 功能特性

- ✅ **动态客户端管理**: 自动添加/删除客户端连接，支持断线检测
- ✅ **多种推送模式**: 广播消息、按客户端ID定向推送
- ✅ **业务解耦设计**: 通过 `chan string` 接收业务层事件
- ✅ **断线重连支持**: 客户端自动重连，服务端自动清理失效连接
- ✅ **高可用设计**: 内置心跳机制、超时处理、panic恢复
- ✅ **配置灵活**: 支持自定义选项配置
- ✅ **CORS支持**: 跨域访问支持
- ✅ **并发安全**: 使用读写锁保护并发访问

## 快速开始

### 基本使用

```go
package main

import (
    "rear/pkg/sse"
    "github.com/gin-gonic/gin"
)

func main() {
    r := gin.Default()
    
    // 创建 SSE 管理器
    manager := sse.NewManager(nil) // 使用默认配置
    
    // 设置 SSE 连接端点
    r.GET("/events", func(c *gin.Context) {
        manager.RegisterClient(c)
    })
    
    // 广播消息端点
    r.POST("/broadcast", func(c *gin.Context) {
        var req struct {
            Message string `json:"message"`
        }
        c.ShouldBindJSON(&req)
        
        manager.Broadcast(req.Message)
        c.JSON(200, gin.H{"status": "sent"})
    })
    
    r.Run(":8080")
}
```

### 集成到现有项目

1. **创建 SSE 处理器**

```go
// internal/handler/sse_handler.go
sseHandler := handler.NewSSEHandler()
```

2. **添加路由**

```go
// internal/router/router.go
func SetupRoutes(r *gin.Engine, contain *container.DbContainer, imgContain *container.TaskContainer) {
    // 现有路由...
    
    // SSE 相关路由
    api := r.Group("/api/v1")
    {
        api.GET("/events", sseHandler.HandleSSEConnection)
        api.POST("/events/broadcast", sseHandler.BroadcastMessage)
        api.POST("/events/send/:clientId", sseHandler.SendToClient)
        api.GET("/events/clients", sseHandler.GetClients)
    }
}
```

3. **业务层集成**

```go
// 获取事件通道
eventChannel := sseHandler.GetEventChannel()

// 在业务逻辑中推送事件
func ProcessPhoto(photoID string) {
    // 处理照片...
    
    // 发送处理完成事件
    eventData := map[string]interface{}{
        "type": "photo_processed",
        "photo_id": photoID,
        "status": "completed",
    }
    eventJSON, _ := json.Marshal(eventData)
    
    select {
    case eventChannel <- string(eventJSON):
        // 事件发送成功
    case <-time.After(1 * time.Second):
        // 超时处理
    }
}
```

## 配置选项

```go
opts := &sse.Options{
    PingInterval:    30 * time.Second,  // 心跳间隔
    MaxClients:      1000,              // 最大客户端数
    AllowedOrigins:  []string{"*"},     // 允许的来源
    EnableCORS:      true,              // 启用CORS
    EventBufferSize: 100,               // 事件缓冲区大小
    ClientTimeout:   5 * time.Minute,   // 客户端超时时间
}

manager := sse.NewManager(opts)
```

## API 接口

### 客户端连接
```
GET /api/v1/events
```

### 广播消息
```
POST /api/v1/events/broadcast
Content-Type: application/json

{
    "message": "Hello World",
    "event_type": "notification"  // 可选
}
```

### 定向发送
```
POST /api/v1/events/send/{clientId}
Content-Type: application/json

{
    "message": "Private message"
}
```

### 获取客户端列表
```
GET /api/v1/events/clients
```

### 获取统计信息
```
GET /api/v1/events/stats
```

## 客户端连接示例

### JavaScript
```javascript
const eventSource = new EventSource('/api/v1/events');

eventSource.onopen = function(event) {
    console.log('Connected to SSE');
};

eventSource.onmessage = function(event) {
    const data = JSON.parse(event.data);
    console.log('Received:', data);
};

// 监听特定事件类型
eventSource.addEventListener('photo_processed', function(event) {
    const data = JSON.parse(event.data);
    console.log('Photo processed:', data);
});

eventSource.onerror = function(event) {
    console.log('SSE error:', event);
};
```

### cURL 测试
```bash
# 连接 SSE
curl -N -H "Accept: text/event-stream" http://localhost:8080/api/v1/events

# 发送广播消息
curl -X POST http://localhost:8080/api/v1/events/broadcast \
  -H "Content-Type: application/json" \
  -d '{"message": "Hello World", "event_type": "notification"}'
```

## 架构设计

```
业务层 (Business Logic)
    ↓ (通过 chan string)
SSE 管理器 (Manager)
    ↓ (WebSocket/HTTP连接)
客户端 (Frontend/Mobile App)
```

### 核心组件

1. **Manager**: 核心管理器，负责客户端生命周期管理
2. **Client**: 客户端连接抽象，包含连接信息和通信通道
3. **Event**: 事件数据结构，支持 SSE 标准格式
4. **Options**: 配置选项，支持自定义行为

### 线程安全

- 使用 `sync.RWMutex` 保护客户端集合
- 通过 channel 实现 goroutine 间通信
- 自动清理失效连接和资源

## 生产环境考虑

### 性能优化
- 事件缓冲区避免阻塞
- 读写锁优化并发访问
- 超时机制防止资源泄露

### 监控指标
- 客户端连接数
- 事件发送成功率
- 连接异常次数

### 扩展性
- 支持集群部署（需要外部消息队列）
- 支持客户端分组
- 支持事件过滤

## 故障排除

### 常见问题

1. **客户端连接失败**
   - 检查 CORS 设置
   - 确认路由配置正确
   - 检查防火墙设置

2. **消息丢失**
   - 增大事件缓冲区
   - 检查网络稳定性
   - 实现客户端重连机制

3. **内存占用过高**
   - 限制最大客户端数
   - 减少事件缓冲区大小
   - 启用客户端超时清理

### 调试技巧

```go
// 启用详细日志
manager.SetDebugMode(true)

// 获取连接统计
stats := manager.GetStats()
fmt.Printf("Active clients: %d\n", stats.ClientCount)
```

## 示例项目

查看 `examples/sse_integration_example.go` 获取完整的集成示例。

运行示例：
```bash
cd examples
go run sse_integration_example.go
```

访问管理面板：http://localhost:8080/admin/sse/

## 许可证

本项目采用 MIT 许可证。