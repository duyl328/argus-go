# 优雅关闭问题修复说明

## 问题描述

后端服务器在接收到关闭信号（Ctrl+C 或 SIGTERM）后，总是显示 "Shutting down server..." 但无法立即关闭，需要等待 30 秒超时才能退出。

## 根本原因

在调用 `srv.Shutdown()` 之前，没有清理以下长连接资源：

1. **SSE (Server-Sent Events) 连接** - SSE Manager 维护的长连接客户端
2. **文件系统监听器** - fsnotify 创建的文件监听 goroutine

这些未关闭的连接会阻止 HTTP 服务器优雅关闭，导致必须等待超时。

## 解决方案

### 1. 修改 `internal/router/router.go`

添加清理函数返回值：

```go
// SetupRoutes 设置路由
// 返回一个清理函数，用于在服务器关闭时清理资源
func SetupRoutes(r *gin.Engine, contain *container.DbContainer, imgContain *container.TaskContainer) CleanupFunc {
    // ... 路由设置代码 ...

    // 返回清理函数
    return func() {
        // 关闭 SSE Manager
        if sseHandler != nil && sseHandler.Manager != nil {
            sseHandler.Manager.Close()
        }

        // 关闭文件系统监听器
        if fileSystemHandler != nil && fileSystemHandler.GetFileWatcher() != nil {
            fileSystemHandler.GetFileWatcher().Close()
        }
    }
}
```

### 2. 修改 `internal/handler/filesystem_handler.go`

添加 `GetFileWatcher()` 方法暴露内部监听器：

```go
// GetFileWatcher 获取文件监听器（用于清理）
func (h *FileSystemHandler) GetFileWatcher() *utils.FileSystemWatcher {
    return h.fileWatcher
}
```

### 3. 修改 `internal/utils/filesystem_watcher.go`

添加 `Close()` 方法作为 `Stop()` 的别名：

```go
// Close 关闭监听器（Stop 的别名，用于资源清理）
func (fsw *FileSystemWatcher) Close() {
    fsw.Stop()
}
```

### 4. 修改 `main.go`

在服务器关闭前调用清理函数：

```go
// 设置路由，并获取清理函数
cleanup := router.SetupRoutes(r, con, imgContain)

// ... 启动服务器 ...

//  阻塞主goroutine，等待信号
<-quit
logger.Info("Shutting down server...")

// 先清理资源（关闭 SSE 连接和文件监听器）
logger.Info("Cleaning up resources...")
cleanup()
logger.Info("Resources cleaned up")

// 然后关闭 HTTP 服务器（超时时间从 30s 改为 5s）
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

if err := srv.Shutdown(ctx); err != nil {
    logger.Errorf("Server forced to shutdown: %v", err)
} else {
    logger.Info("Server exited gracefully")
}
```

## 关闭流程

修复后的优雅关闭流程：

1. **接收信号** (SIGINT/SIGTERM)
2. **清理长连接资源**
   - 关闭所有 SSE 客户端连接
   - 停止文件系统监听器
3. **关闭 HTTP 服务器**
   - 等待现有请求完成
   - 最多等待 5 秒（之前是 30 秒）
4. **退出程序**

## 测试方法

1. 启动后端服务器
2. 打开前端，建立 SSE 连接
3. 在前端打开一个文件夹（触发文件监听）
4. 在后端控制台按 `Ctrl+C`
5. 观察日志输出

**预期输出**:
```
Shutting down server...
Cleaning up resources...
Resources cleaned up
Server exited gracefully
```

**预期时间**: < 1 秒（之前需要 30 秒）

## 修改的文件

- `rear/main.go` - 调用清理函数
- `rear/internal/router/router.go` - 返回清理函数
- `rear/internal/handler/filesystem_handler.go` - 暴露文件监听器
- `rear/internal/utils/filesystem_watcher.go` - 添加 Close 方法

## 注意事项

1. **清理顺序很重要**: 必须先清理长连接资源，再关闭 HTTP 服务器
2. **超时时间调整**: 从 30 秒改为 5 秒，因为长连接已提前关闭
3. **并发安全**: SSE Manager 和 FileSystemWatcher 的 Close 方法都是并发安全的
4. **幂等性**: Close 方法可以多次调用而不会出错

## 日志说明

关闭时的日志输出：

```
2025-10-05 15:30:00.000    INFO    Shutting down server...
2025-10-05 15:30:00.100    INFO    Cleaning up resources...
2025-10-05 15:30:00.150    INFO    Resources cleaned up
2025-10-05 15:30:00.200    INFO    Server exited gracefully
```

如果仍然超时，日志会显示：
```
2025-10-05 15:30:00.000    INFO    Shutting down server...
2025-10-05 15:30:00.100    INFO    Cleaning up resources...
2025-10-05 15:30:00.150    INFO    Resources cleaned up
2025-10-05 15:30:05.200    ERROR   Server forced to shutdown: context deadline exceeded
```

这种情况下需要检查是否有其他未关闭的连接或阻塞的 goroutine。
