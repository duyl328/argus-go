# SSE 订阅系统更新日志

## 版本 2.0 - 客户端级别订阅管理 (2025-10-04)

### 🎯 核心改进

实现了**客户端级别的 SSE 订阅管理系统**，解决了以下问题：

1. ✅ **精准推送** - 只向订阅了特定路径的客户端发送事件
2. ✅ **自动清理** - 客户端断开时自动清理订阅
3. ✅ **独立管理** - 每个客户端独立管理自己的订阅列表
4. ✅ **减少噪音** - 避免无关事件干扰客户端

---

## 📝 详细更改

### 后端更改 (Go)

#### 1. SSE Manager 增强 (`pkg/sse/manager.go`)

**Client 结构新增字段:**
```go
type Client struct {
    // ... 原有字段

    // 新增：订阅管理
    subscriptions     map[string]bool // key: 订阅的路径, value: true
    subscriptionMutex sync.RWMutex    // 保护 subscriptions
}
```

**新增方法:**
- `Subscribe(clientID, path)` - 客户端订阅指定路径
- `Unsubscribe(clientID, path)` - 客户端取消订阅
- `GetSubscriptions(clientID)` - 获取客户端所有订阅
- `IsSubscribed(clientID, path)` - 检查客户端是否订阅
- `BroadcastToSubscribers(path, eventType, message)` - 向订阅者广播事件

**修改:**
- `RegisterClient()` - 初始化 `subscriptions` 映射

---

#### 2. SSE Handler 增强 (`internal/handler/sse_handler.go`)

**新增 HTTP 接口:**

1. **订阅路径**
   - 路由: `POST /api/v1/sse/subscribe`
   - 请求体: `{ "client_id": "...", "path": "..." }`

2. **取消订阅**
   - 路由: `POST /api/v1/sse/unsubscribe`
   - 请求体: `{ "client_id": "...", "path": "..." }`

3. **获取订阅列表**
   - 路由: `GET /api/v1/sse/subscriptions/:clientId`

---

#### 3. FileSystemEvent 增强 (`internal/utils/filesystem_watcher.go`)

**新增字段:**
```go
type FileSystemEvent struct {
    // ... 原有字段

    WatchedPath string `json:"watched_path"` // 被监听的根路径
}
```

**新增方法:**
```go
// findWatchedPath 查找事件路径对应的被监听根路径
func (fsw *FileSystemWatcher) findWatchedPath(eventPath string) string
```

**修改逻辑:**
- `handleEvent()` - 自动填充 `WatchedPath` 字段

---

#### 4. 文件系统 Handler 修改 (`internal/handler/filesystem_handler.go`)

**修改:**
- 使用 `BroadcastToSubscribers()` 替代 `BroadcastEvent()`
- 根据 `event.WatchedPath` 精准推送事件

**代码对比:**
```go
// 修改前
sseManager.BroadcastEvent("filesystem-change", eventStr)

// 修改后
sseManager.BroadcastToSubscribers(ev.WatchedPath, "filesystem-change", eventStr)
```

---

### 前端更改 (Vue3/TypeScript)

#### 1. SSEService 增强 (`services/sseService.ts`)

**新增字段:**
```typescript
class SSEService {
    private clientID: string | null = null // 存储客户端 ID
    // ...
}
```

**新增方法:**
```typescript
getClientID(): string | null
subscribe(path: string): Promise<boolean>
unsubscribe(path: string): Promise<boolean>
getSubscriptions(): Promise<string[]>
```

**新增事件监听:**
```typescript
// 监听 connected 事件获取客户端 ID
this.eventSource.addEventListener('connected', (event) => {
    const data = JSON.parse(event.data)
    this.clientID = data.client_id
})
```

---

#### 2. FileSystemChangeEvent 增强

**新增字段:**
```typescript
interface FileSystemChangeEvent {
    // ... 原有字段
    watched_path?: string // 被监听的根路径
}
```

---

## 🚀 使用示例

### 前端使用

```typescript
import { sseService } from '@/services/sseService'

// 1. 连接 SSE
await sseService.connect()
await new Promise(resolve => setTimeout(resolve, 1000))

// 2. 订阅路径
await sseService.subscribe('D:\\EdgeDownload')

// 3. 监听事件
const unsubscribe = sseService.onFileSystemChange((event) => {
    console.log('文件变化:', event)
})

// 4. 清理
onUnmounted(async () => {
    await sseService.unsubscribe('D:\\EdgeDownload')
    unsubscribe()
})
```

### 后端 API 测试

```http
# 1. 订阅路径
POST http://127.0.0.1:9484/api/v1/sse/subscribe
Content-Type: application/json

{
  "client_id": "your-client-id",
  "path": "D:\\EdgeDownload"
}

# 2. 获取订阅列表
GET http://127.0.0.1:9484/api/v1/sse/subscriptions/your-client-id

# 3. 取消订阅
POST http://127.0.0.1:9484/api/v1/sse/unsubscribe
Content-Type: application/json

{
  "client_id": "your-client-id",
  "path": "D:\\EdgeDownload"
}
```

---

## 🔍 测试步骤

### 1. 启动后端

```bash
cd D:\go-argus\src\rear
go run main.go
```

### 2. 启动前端

```bash
cd D:\go-argus\src\front\argus-front
npm run dev
```

### 3. 测试订阅功能

1. 打开浏览器控制台
2. 等待 SSE 连接建立（看到客户端 ID）
3. 订阅路径：
   ```javascript
   await sseService.subscribe('D:\\EdgeDownload')
   ```
4. 在 `D:\EdgeDownload` 中创建/修改/删除文件
5. 观察控制台输出的文件系统事件

### 4. 验证精准推送

1. 打开多个浏览器标签页
2. 每个标签页订阅不同的路径
3. 修改某个路径下的文件
4. 验证只有订阅该路径的标签页收到事件

---

## 📊 性能改进

### 1. 减少网络传输

**修改前:**
- 所有客户端收到所有文件系统事件
- 客户端自己过滤无关事件

**修改后:**
- 服务端根据订阅精准推送
- 网络传输量大幅减少

### 2. 减少客户端处理负担

**修改前:**
```
事件 100 个 × 客户端 10 个 = 1000 次处理
```

**修改后:**
```
事件 100 个 × 订阅客户端 2 个 = 200 次处理
```

---

## ⚠️ 破坏性更改

### 1. FileSystemEvent 结构变更

新增了 `watched_path` 字段，旧版本客户端可能无法正确解析。

**解决方案:**
- 字段为可选 (`json:"-omitempty"`)
- 旧客户端可忽略该字段

### 2. 订阅模式变更

**修改前:**
- 连接即收到所有文件系统事件

**修改后:**
- 需要主动订阅才能收到事件
- 未订阅的客户端不会收到 `filesystem-change` 事件

**迁移指南:**
```typescript
// 在连接成功后立即订阅
await sseService.connect()
await new Promise(resolve => setTimeout(resolve, 1000))
await sseService.subscribe(defaultPath) // 添加这行
```

---

## 🐛 已修复的问题

### 1. 编译错误: `undefined: path`

**问题:**
`filesystem_handler.go:91` 使用了未定义的变量 `path`

**解决方案:**
- 在 `FileSystemEvent` 中添加 `WatchedPath` 字段
- 在 `FileSystemWatcher.handleEvent()` 中自动填充该字段
- 使用 `event.WatchedPath` 作为订阅路径

---

## 📚 相关文档

- **使用指南**: `SSE_SUBSCRIPTION_GUIDE.md`
- **HTTP 测试**: `rear/tests/sse_subscription.http`
- **代码示例**: `front/argus-front/src/examples/sseSubscriptionExample.ts`

---

## 🔮 未来改进

### 计划中的功能

1. **路径模式匹配**
   - 支持通配符订阅 (如 `D:\**\*.jpg`)
   - 支持正则表达式匹配

2. **订阅限制**
   - 单个客户端最大订阅数限制
   - 路径权限验证

3. **订阅统计**
   - 每个路径的订阅客户端数量
   - 事件发送统计

4. **持久化订阅**
   - 重连后自动恢复订阅
   - 订阅信息存储到数据库

---

## 👥 贡献者

- Claude Code

---

**版本**: 2.0
**发布日期**: 2025-10-04
