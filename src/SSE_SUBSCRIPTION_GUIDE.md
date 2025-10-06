# SSE 订阅管理系统使用指南

## 📋 概述

本系统实现了**客户端级别的 SSE 订阅管理**，解决了以下问题：

1. ✅ **精准推送** - 只向订阅了特定路径的客户端发送事件
2. ✅ **自动清理** - 客户端断开时自动清理订阅
3. ✅ **独立管理** - 每个客户端独立管理自己的订阅列表
4. ✅ **减少噪音** - 避免无关事件干扰客户端

---

## 🏗️ 架构设计

### 核心组件

```
┌─────────────────────────────────────────────┐
│           Frontend (Vue3)                   │
│  ┌──────────────────────────────────────┐  │
│  │  sseService.ts                       │  │
│  │  - connect()                         │  │
│  │  - subscribe(path)                   │  │
│  │  - unsubscribe(path)                 │  │
│  │  - getSubscriptions()                │  │
│  └──────────────────────────────────────┘  │
└──────────────────┬──────────────────────────┘
                   │ HTTP + EventSource
┌──────────────────▼──────────────────────────┐
│           Backend (Go/Gin)                  │
│  ┌──────────────────────────────────────┐  │
│  │  SSE Manager                         │  │
│  │  - RegisterClient()                  │  │
│  │  - Subscribe(clientID, path)         │  │
│  │  - Unsubscribe(clientID, path)       │  │
│  │  - BroadcastToSubscribers(path, ...) │  │
│  └──────────────────────────────────────┘  │
│  ┌──────────────────────────────────────┐  │
│  │  FileSystem Watcher                  │  │
│  │  - Watch(path)                       │  │
│  │  - Unwatch(path)                     │  │
│  └──────────────────────────────────────┘  │
└─────────────────────────────────────────────┘
```

---

## 🚀 使用流程

### 1. 前端连接 SSE

```typescript
import { sseService } from '@/services/sseService'

// 连接 SSE 服务器
await sseService.connect()

// 等待客户端 ID 初始化
await new Promise(resolve => setTimeout(resolve, 1000))

// 获取客户端 ID
const clientID = sseService.getClientID()
console.log('客户端 ID:', clientID)
```

### 2. 订阅文件夹

```typescript
// 订阅指定路径
const path = 'D:\\EdgeDownload'
const success = await sseService.subscribe(path)

if (success) {
  console.log('✅ 订阅成功')
}
```

### 3. 监听文件系统变化

```typescript
// 注册回调函数
const unsubscribe = sseService.onFileSystemChange((event) => {
  console.log('文件系统变化:', event)
  console.log('类型:', event.type)  // create, modify, delete, rename
  console.log('路径:', event.path)
  console.log('名称:', event.name)
})

// 组件卸载时取消监听
onUnmounted(() => {
  unsubscribe()
})
```

### 4. 取消订阅

```typescript
// 取消订阅指定路径
await sseService.unsubscribe(path)
```

### 5. 查看当前订阅

```typescript
// 获取当前所有订阅
const subscriptions = await sseService.getSubscriptions()
console.log('当前订阅:', subscriptions)
```

---

## 📦 完整 Vue 组件示例

```vue
<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { sseService } from '@/services/sseService'

interface FileSystemEvent {
  type: 'create' | 'modify' | 'delete' | 'rename'
  path: string
  name: string
  timestamp: string
  is_dir: boolean
}

const currentPath = ref('D:\\EdgeDownload')
const events = ref<FileSystemEvent[]>([])
const isSubscribed = ref(false)

// 订阅回调清理函数
let unsubscribeCallback: (() => void) | null = null

onMounted(async () => {
  // 连接 SSE
  await sseService.connect()

  // 等待客户端 ID 初始化
  await new Promise(resolve => setTimeout(resolve, 1000))

  // 订阅当前路径
  const success = await sseService.subscribe(currentPath.value)

  if (success) {
    isSubscribed.value = true

    // 监听文件系统变化
    unsubscribeCallback = sseService.onFileSystemChange((event) => {
      events.value.unshift(event)

      // 保留最近 100 个事件
      if (events.value.length > 100) {
        events.value = events.value.slice(0, 100)
      }
    })
  }
})

onUnmounted(async () => {
  // 取消订阅
  if (isSubscribed.value) {
    await sseService.unsubscribe(currentPath.value)
  }

  // 取消回调监听
  if (unsubscribeCallback) {
    unsubscribeCallback()
  }
})

// 切换监听的文件夹
async function watchFolder(newPath: string) {
  // 取消旧订阅
  await sseService.unsubscribe(currentPath.value)

  // 添加新订阅
  const success = await sseService.subscribe(newPath)

  if (success) {
    currentPath.value = newPath
    events.value = []
  }
}
</script>

<template>
  <div>
    <h3>当前监听: {{ currentPath }}</h3>
    <p>订阅状态: {{ isSubscribed ? '✅ 已订阅' : '❌ 未订阅' }}</p>

    <div>
      <h4>文件系统事件 ({{ events.length }})</h4>
      <ul>
        <li v-for="(event, index) in events" :key="index">
          <strong>{{ event.type }}</strong>: {{ event.name }}
          <span v-if="event.is_dir"> (文件夹)</span>
        </li>
      </ul>
    </div>
  </div>
</template>
```

---

## 🔧 后端 API

### 1. 订阅路径

```http
POST /api/v1/sse/subscribe
Content-Type: application/json

{
  "client_id": "your-client-id",
  "path": "D:\\EdgeDownload"
}
```

**响应:**
```json
{
  "success": true,
  "message": "Subscribed successfully",
  "data": {
    "client_id": "your-client-id",
    "path": "D:\\EdgeDownload"
  }
}
```

### 2. 取消订阅

```http
POST /api/v1/sse/unsubscribe
Content-Type: application/json

{
  "client_id": "your-client-id",
  "path": "D:\\EdgeDownload"
}
```

### 3. 获取订阅列表

```http
GET /api/v1/sse/subscriptions/:clientId
```

**响应:**
```json
{
  "success": true,
  "message": "Subscriptions retrieved successfully",
  "data": {
    "client_id": "your-client-id",
    "subscriptions": ["D:\\EdgeDownload", "D:\\Projects"],
    "count": 2
  }
}
```

---

## 🔍 调试技巧

### 前端调试

```typescript
// 1. 检查客户端 ID
console.log('客户端 ID:', sseService.getClientID())

// 2. 检查连接状态
console.log('连接状态:', sseService.isConnected())

// 3. 查看当前订阅
const subs = await sseService.getSubscriptions()
console.log('当前订阅:', subs)
```

### 后端日志

后端会输出以下日志：

```bash
# 客户端连接
Client connected: abc-123 (Total: 1)

# 订阅路径
客户端 abc-123 订阅了路径: D:\EdgeDownload

# 广播事件
向 1 个订阅了路径 D:\EdgeDownload 的客户端广播事件

# 取消订阅
客户端 abc-123 取消订阅路径: D:\EdgeDownload

# 客户端断开
Client disconnected: abc-123 (Total: 0)
```

---

## ⚠️ 注意事项

### 1. 客户端 ID 初始化

客户端 ID 是在 SSE 连接建立后通过 `connected` 事件传递的，需要等待：

```typescript
await sseService.connect()
await new Promise(resolve => setTimeout(resolve, 1000)) // 等待初始化
```

### 2. 组件卸载时清理

**必须在组件卸载时取消订阅和回调**：

```typescript
onUnmounted(async () => {
  await sseService.unsubscribe(path)
  unsubscribeCallback()
})
```

### 3. 路径格式

- ✅ Windows: `D:\\EdgeDownload`（双反斜杠）
- ✅ Linux/Mac: `/home/user/Downloads`

### 4. SSE 重连处理

SSE 连接会自动重连，但订阅信息保留在服务端，**无需重新订阅**。

---

## 🎯 常见问题

### Q1: 订阅后收不到事件？

检查清单：
- [ ] 客户端 ID 是否正确获取
- [ ] 路径是否已正确订阅（调用 `getSubscriptions()` 确认）
- [ ] 后端是否正在监听该路径（检查文件系统监听器）
- [ ] 监听路径是否与实际文件变化路径一致

### Q2: SSE 连接频繁断开？

- 当前后端配置：WriteTimeout = 5 分钟
- 前端心跳超时：90 秒
- Keepalive 间隔：5 秒
- Ping 间隔：10 秒

如果仍然断开，检查：
- 网络代理或防火墙设置
- 浏览器是否限制长连接
- 服务器资源是否充足

### Q3: 客户端断开后订阅会自动清理吗？

✅ **是的**！客户端断开时：
1. SSE Manager 自动注销客户端
2. 客户端的订阅映射自动清理
3. 不会再收到任何事件

---

## 📊 性能优化

### 1. 事件节流

后端使用事件节流通道（容量 10）：
- 高频事件会被自动丢弃
- 避免 SSE 通道阻塞

### 2. 精准推送

只向订阅了特定路径的客户端发送事件：
- 减少无效网络传输
- 降低客户端处理负担

### 3. 并发发送

后端使用 goroutine 并发向多个客户端发送事件：
- 单个客户端阻塞不影响其他客户端
- 100ms 超时自动断开慢速客户端

---

## 🔒 安全建议

1. **客户端 ID 验证** - 服务端验证客户端 ID 是否存在
2. **路径验证** - 限制订阅路径范围（防止订阅敏感目录）
3. **订阅数量限制** - 限制单个客户端最大订阅数
4. **连接数限制** - 当前限制 500 个并发连接

---

## 🎉 总结

### 优势

- ✅ **精准推送** - 只收到感兴趣的事件
- ✅ **自动清理** - 断开连接自动清理订阅
- ✅ **易于使用** - 简洁的 API 设计
- ✅ **高性能** - 节流 + 并发优化
- ✅ **可扩展** - 支持多客户端、多路径订阅

### 下一步

1. 实现路径权限验证
2. 添加订阅数量限制
3. 支持路径模糊匹配（如 `D:\\**\\*.jpg`）
4. 添加订阅统计和监控

---

**文档版本**: 1.0
**最后更新**: 2025-10-04
**维护者**: Claude Code
