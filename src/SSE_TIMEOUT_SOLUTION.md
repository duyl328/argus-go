# SSE 连接断开问题解决方案

## 🔍 问题分析

### 错误现象
```
GET http://127.0.0.1:3001/api/v1/sse/connect net::ERR_INCOMPLETE_CHUNKED_ENCODING 200 (OK)
```

### 根本原因

**Go 的 `http.Server.WriteTimeout` 与 SSE 长连接冲突**

1. **WriteTimeout 机制**:
   - 从第一次写入开始计时
   - 整个响应必须在超时前完成
   - **不会因为持续写入而重置**

2. **SSE 的 c.Stream() 机制**:
   ```go
   c.Stream(func(w io.Writer) bool {
       select {
       case event := <-client.EventChan:
           writeEvent(w, event) // 写入
           return true
       }
   })
   ```
   - 每次循环等待新事件
   - 等待期间**没有写入操作**
   - 如果等待时间 > WriteTimeout → 连接关闭

3. **时间线示例** (WriteTimeout = 5分钟):
   ```
   0:00  - 连接建立，发送 connected 事件
   0:05  - 发送 keepalive (写入)
   0:10  - 发送 ping (写入)
   ...
   5:00  - 从第一次写入起已过 5 分钟
   5:00  - http.Server 强制关闭连接
   5:00  - 前端收到 ERR_INCOMPLETE_CHUNKED_ENCODING
   ```

---

## ✅ 解决方案

### 当前实现: WriteTimeout = 0

**配置文件**: `rear/internal/config/loader.go:125`

```go
App: AppConfig{
    Port:         getEnv("PORT", "8080"),
    Mode:         getEnv("GIN_MODE", "debug"),
    ReadTimeout:  30 * time.Second,
    WriteTimeout: 0, // 🔑 SSE 需要长连接，设置为 0（无限期）
    IdleTimeout:  10 * time.Minute,
}
```

**HTTP 服务器**: `rear/main.go:237-243`

```go
srv := &http.Server{
    Addr:         ":" + config.CONFIG.Port,
    Handler:      r,
    ReadTimeout:  config.CONFIG.ReadTimeout,  // 30秒
    WriteTimeout: config.CONFIG.WriteTimeout, // 0（无限期）
    IdleTimeout:  config.CONFIG.IdleTimeout,  // 10分钟
}
```

---

## ⚠️ WriteTimeout = 0 的影响

### 对所有路由的影响

| 路由类型 | 影响 | 风险 |
|----------|------|------|
| 普通 API | ✅ 无影响 | ⚠️ 慢速客户端可能占用连接 |
| 文件上传 | ✅ 支持大文件 | ⚠️ 恶意客户端可慢速上传 |
| 文件下载 | ✅ 支持大文件 | ⚠️ 慢速客户端占用资源 |
| SSE | ✅✅ 正常工作 | ✅ 无风险 |

### 风险缓解措施

#### 1. 应用层超时控制

普通 API 使用 Context 超时：

```go
func (h *Handler) SomeAPI(c *gin.Context) {
    ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
    defer cancel()

    // 使用 ctx 进行数据库查询等操作
    result, err := db.QueryContext(ctx, ...)
}
```

#### 2. 限流中间件

```go
import "golang.org/x/time/rate"

var limiter = rate.NewLimiter(100, 200) // 100 req/s, burst 200

func RateLimitMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        if !limiter.Allow() {
            c.JSON(429, gin.H{"error": "Too many requests"})
            c.Abort()
            return
        }
        c.Next()
    }
}
```

#### 3. 连接数限制

当前配置:
```go
// SSE Manager
opts.MaxClients = 500 // 最多 500 个 SSE 连接
```

#### 4. IdleTimeout 保护

```go
IdleTimeout: 10 * time.Minute // 10分钟无活动则关闭
```

---

## 🚀 生产环境部署建议

### 使用 Nginx 反向代理

#### Nginx 配置示例

```nginx
upstream argus_backend {
    server 127.0.0.1:9484;
    keepalive 64;
}

server {
    listen 80;
    server_name your-domain.com;

    # 普通 API 路由（有超时限制）
    location /api/ {
        proxy_pass http://argus_backend;
        proxy_http_version 1.1;
        proxy_set_header Connection "";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;

        # 普通请求超时设置
        proxy_connect_timeout 5s;
        proxy_send_timeout 30s;
        proxy_read_timeout 30s;
    }

    # SSE 路由（特殊配置）
    location /api/v1/sse/ {
        proxy_pass http://argus_backend;
        proxy_http_version 1.1;
        proxy_set_header Connection "";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;

        # SSE 专用配置
        proxy_buffering off;               # 🔑 禁用缓冲
        proxy_cache off;                   # 禁用缓存
        proxy_read_timeout 3600s;          # 1小时超时
        proxy_connect_timeout 3600s;
        chunked_transfer_encoding on;      # 启用分块传输
        proxy_set_header X-Accel-Buffering no; # 禁用 Nginx 缓冲
    }
}
```

### 架构图

```
客户端
   ↓
Nginx (80端口)
   ├─→ /api/*          (30秒超时)
   └─→ /api/v1/sse/*   (1小时超时)
       ↓
Go 服务器 (9484端口, WriteTimeout=0)
```

---

## 🔬 其他尝试过的方案

### ❌ 方案 1: 中间件修改 Deadline

**代码**:
```go
func NoWriteDeadlineMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        if conn, ok := getUnderlyingConn(c.Writer); ok {
            conn.SetWriteDeadline(time.Time{}) // 移除 deadline
        }
        c.Next()
    }
}
```

**问题**:
- 无法在不 Hijack 的情况下访问 `net.Conn`
- Hijack 会接管连接，无法继续使用 Gin 的 Stream

**结论**: ❌ 不可行

---

### ❌ 方案 2: 定期写入 comment

**代码**:
```go
// 每隔一段时间写入 SSE comment 保持连接
go func() {
    ticker := time.NewTicker(2 * time.Minute)
    for range ticker.C {
        fmt.Fprintf(w, ": keepalive\n\n")
        w.Flush()
    }
}()
```

**问题**:
- WriteTimeout 是从**第一次写入**开始计时
- 持续写入并不会重置超时
- 5分钟后依然会断开

**结论**: ❌ 无效

---

## 📊 测试结果

### 修改前 (WriteTimeout = 5分钟)

```
0:00 - 连接建立
0:05 - keepalive
0:10 - ping
...
5:00 - ❌ ERR_INCOMPLETE_CHUNKED_ENCODING
```

### 修改后 (WriteTimeout = 0)

```
0:00  - 连接建立
0:05  - keepalive
0:10  - ping
0:15  - keepalive
...
60:00 - ✅ 仍然连接
```

**测试命令**:
```bash
# 1. 启动后端
cd D:\go-argus\src\rear
go run main.go

# 2. 浏览器控制台测试
const es = new EventSource('http://127.0.0.1:9484/api/v1/sse/connect')
es.addEventListener('keepalive', e => console.log('keepalive:', e.data))
es.addEventListener('ping', e => console.log('ping:', e.data))

# 观察是否在5分钟后仍然收到事件
```

---

## 📚 参考资料

### Go 官方文档

- [http.Server](https://pkg.go.dev/net/http#Server)
  > WriteTimeout: 写入响应的最大持续时间（从请求头读取完成到响应写入完成）。**零值表示无超时**。

- [Server-Sent Events](https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events)

### 相关问题

- [Gin + SSE timeout issue](https://github.com/gin-gonic/gin/issues/1287)
- [Go http.Server WriteTimeout for SSE](https://stackoverflow.com/questions/37294519)

---

## 🎯 总结

### 最终方案

✅ **WriteTimeout = 0** (当前实现)

### 优点

1. ✅ 简单可靠，一行配置解决
2. ✅ SSE 连接稳定，不会断开
3. ✅ 适用于开发和小型部署

### 注意事项

1. ⚠️ 生产环境建议使用 Nginx 反向代理
2. ⚠️ 应用层需要自己控制超时（Context）
3. ⚠️ 配置限流和连接数限制

### 是否还会断开？

**仍可能断开的情况**:

1. **IdleTimeout 触发** (10分钟无活动)
   - **解决**: 后端每 5 秒发送 keepalive ✅

2. **浏览器/代理超时**
   - **解决**: 前端自动重连 ✅

3. **网络波动**
   - **解决**: 前端自动重连 ✅

**正常情况**: 连接应该保持稳定，偶尔断开会自动重连

---

**文档版本**: 1.0
**最后更新**: 2025-10-04
**作者**: Claude Code
