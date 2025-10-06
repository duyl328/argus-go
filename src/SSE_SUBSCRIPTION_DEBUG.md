# SSE 订阅和文件监听诊断指南

## 问题描述
前端打开某个文件夹后，通过文件管理器删除/改名文件，页面不自动更新。

## 诊断流程

### 步骤 1: 检查前端订阅请求是否发送

**操作**:
1. 打开浏览器控制台（F12）
2. 切换到 **Network** 标签页
3. 在前端进入某个文件夹（例如 `D:\test`）

**预期看到的请求**:
```
POST http://localhost:9484/api/v1/filesystem/watch
Request Payload: {"path":"D:\\test"}
Response:
{
  "code": 200,
  "message": "已开始监听文件夹",
  "data": {
    "path": "D:\\test",
    "watched_paths": ["D:\\test"]
  }
}
```

**同时检查 Console 日志**:
```
🔄 [useFileSystemAPI.browse] 开始刷新: D:\test
✅ [useFileSystemAPI.browse] API 返回: { ... }
✅ [useFileSystemAPI.browse] items.value 已更新: 5 个项目
✅ 已开始监听: D:\test
```

**如果没有看到 `✅ 已开始监听`**:
- 检查 `watchCurrentPath()` 是否被调用
- 检查是否有错误日志 `⚠️ [useFileSystemAPI.browse] 订阅监听失败`
- 检查 Network 是否有 `/filesystem/watch` 请求

---

### 步骤 2: 检查后端是否收到订阅请求

**操作**:
查看后端日志 `D:\go-argus\src\rear\logs\app.log`

**预期日志**:
```
👀 订阅文件夹监听请求 | path=D:\test | client_ip=127.0.0.1
开始监听文件夹: D:\test
✅ 已开始监听文件夹 | path=D:\test | total_watched=1
```

**如果没有这些日志**:
- 前端请求没有发送
- 检查前端代码 `watchCurrentPath()` 逻辑
- 检查路径是否为空: `if (!currentPath.value || currentPath.value === '')`

**如果有订阅日志，但 total_watched=0**:
- 后端 `fileWatcher.Watch()` 失败
- 检查路径是否有效
- 检查文件权限

---

### 步骤 3: 检查文件变化是否被监听到

**操作**:
1. 确认步骤 1 和 2 都通过
2. 在文件管理器中删除 `D:\test` 下的一个文件（例如 `test.txt`）
3. 立即查看后端日志

**预期日志**:
```
📤 文件系统变化事件已推送 | event_type=delete | file_path=D:\test\test.txt | file_name=test.txt | watched_path=D:\test | is_dir=false | timestamp=2025-10-05 14:30:45
```

**如果没有这条日志**:
- **问题**: fsnotify 没有监听到文件变化
- **可能原因**:
  1. Windows 路径问题：`D:\test` vs `D:/test`
  2. 文件监听器没有启动
  3. 路径未正确添加到监听列表

**调试方法**: 在后端添加调试日志
```go
// 在 filesystem_handler.go 的 WatchPath 中添加
fmt.Printf("DEBUG: 监听路径 = %s\n", req.Path)
fmt.Printf("DEBUG: 当前监听列表 = %v\n", h.fileWatcher.GetWatchedPaths())
```

---

### 步骤 4: 检查 SSE 事件是否推送到前端

**操作**:
1. 删除文件后，检查浏览器控制台 Console 日志

**预期日志**:
```
📨 [useFileSystemAPI.handleFileSystemChange] 收到 SSE 事件: {
  type: "delete",
  path: "D:\\test\\test.txt",
  name: "test.txt",
  is_dir: false,
  timestamp: "2025-10-05T14:30:45Z"
}
```

**如果没有这条日志**:
- **问题**: SSE 事件没有推送到前端
- **可能原因**:
  1. SSE 连接断开
  2. 事件没有通过 SSE Manager 广播
  3. 前端 `sseService.onFileSystemChange()` 没有注册

**检查 SSE 连接**:
在 Network 标签查看是否有持续的请求:
```
GET http://localhost:9484/api/v1/sse/connect
Status: 200 (持续连接)
Type: eventsource
```

---

### 步骤 5: 检查路径匹配逻辑

**操作**:
如果步骤 4 有日志，但页面不刷新，检查路径匹配日志

**预期日志**:
```
🔍 [useFileSystemAPI.handleFileSystemChange] 路径匹配检查: {
  eventDir: "D:/test",
  currentDir: "D:/test",
  match: true
}
✅ [useFileSystemAPI.handleFileSystemChange] 当前文件夹内容发生变化 (delete): test.txt
🔄 [useFileSystemAPI.browse] 开始刷新: D:\test
```

**如果 match: false**:
- **问题**: 路径不匹配
- **可能原因**:
  1. 路径分隔符不一致 (`\` vs `/`)
  2. 路径大小写不一致
  3. eventDir 和 currentDir 值不同

**调试方法**:
修改 `handleFileSystemChange` 添加详细日志:
```typescript
console.log('DEBUG: event.path =', event.path)
console.log('DEBUG: currentPath.value =', currentPath.value)
console.log('DEBUG: eventDir =', eventDir)
console.log('DEBUG: currentDir =', currentDir)
```

---

## 常见问题和解决方案

### 问题 1: 订阅请求没有发送

**症状**: Network 中没有 `/filesystem/watch` 请求

**原因**:
- `currentPath.value === ''` (根目录跳过监听)
- `currentPath.value === previousWatchedPath.value` (重复路径跳过)

**解决**:
清除 `previousWatchedPath` 缓存
```typescript
// 在 browse() 开始时强制重置
previousWatchedPath.value = ''
```

---

### 问题 2: 后端收到订阅但没有监听到文件变化

**症状**: 有 `👀 订阅文件夹监听请求` 日志，但删除文件后没有 `📤 文件系统变化事件已推送`

**原因**:
- `fileWatcher.Watch()` 内部失败
- Windows 路径规范化问题

**调试步骤**:
1. 检查 `filesystem_watcher.go` 的 `Watch()` 方法
2. 打印 `absPath` 和 `fsw.watchedPaths`
3. 确认 `fsw.watcher.Add(absPath)` 没有返回错误

**临时解决**:
在 `Watch()` 中添加详细日志:
```go
func (fsw *FileSystemWatcher) Watch(path string) error {
    absPath, err := filepath.Abs(path)
    fmt.Printf("DEBUG: 原始路径=%s, 绝对路径=%s\n", path, absPath)

    if err := fsw.watcher.Add(absPath); err != nil {
        fmt.Printf("ERROR: fsnotify.Add() 失败: %v\n", err)
        return err
    }

    fmt.Printf("SUCCESS: 已添加监听: %s\n", absPath)
    return nil
}
```

---

### 问题 3: SSE 事件推送了但前端没收到

**症状**: 后端有 `📤 文件系统变化事件已推送`，但前端 Console 没有对应日志

**原因**:
- SSE 连接已断开
- 事件类型不匹配
- `sseService.onFileSystemChange()` 回调没有注册

**检查 SSE 连接**:
```typescript
// 在 useFileSystemAPI onMounted 中添加
console.log('SSE 连接状态:', sseService.isConnected())
```

**检查事件订阅**:
```typescript
// 在 sseService.ts 中添加
onFileSystemChange(callback: (event: FileSystemChangeEvent) => void) {
  console.log('注册 FileSystemChange 回调')
  this.eventHandlers['filesystem_change'] = callback
}
```

---

### 问题 4: 路径匹配失败

**症状**: 有 SSE 事件，但 `match: false`

**原因**:
- Windows 路径 `D:\test` vs `D:/test`
- 事件路径是子文件路径 `D:\test\file.txt`
- currentPath 是文件夹路径 `D:\test`

**当前实现**:
```typescript
const eventDir = eventPath.substring(0, eventPath.lastIndexOf('/'))
if (eventDir === currentDir || eventPath.startsWith(currentDir + '/'))
```

**可能问题**:
- `eventDir` 提取错误
- `currentDir` 格式不一致

**改进建议**:
```typescript
// 更健壮的路径匹配
const normalize = (p: string) => {
  return p.replace(/\\/g, '/').replace(/\/+$/, '').toLowerCase()
}

const normalizedEvent = normalize(event.path)
const normalizedCurrent = normalize(currentPath.value)

// 检查事件是否发生在当前目录或其子目录
const isInCurrentDir = normalizedEvent.startsWith(normalizedCurrent + '/') ||
                       normalizedEvent === normalizedCurrent

// 对于删除/重命名，只关心直接子项（不包括子文件夹内的文件）
const eventDir = normalizedEvent.substring(0, normalizedEvent.lastIndexOf('/'))
const isDirectChild = (eventDir === normalizedCurrent)

if (isDirectChild) {
  console.log('✅ 当前文件夹的直接子项发生变化')
  debouncedRefresh(currentPath.value)
}
```

---

## 快速诊断命令

### 检查后端监听状态
```bash
# Windows
curl http://localhost:9484/api/v1/filesystem/watched

# 预期返回
{
  "code": 200,
  "message": "获取监听路径成功",
  "data": {
    "count": 1,
    "paths": ["D:\\test"]
  }
}
```

### 手动订阅测试
```bash
# 订阅
curl -X POST http://localhost:9484/api/v1/filesystem/watch ^
  -H "Content-Type: application/json" ^
  -d "{\"path\":\"D:\\test\"}"

# 取消订阅
curl -X POST http://localhost:9484/api/v1/filesystem/unwatch ^
  -H "Content-Type: application/json" ^
  -d "{\"path\":\"D:\\test\"}"
```

### 检查 SSE 连接
```bash
# 使用 curl 监听 SSE
curl -N http://localhost:9484/api/v1/sse/connect

# 预期看到心跳事件
event: heartbeat
data: {"timestamp":"2025-10-05T14:30:00Z"}
```

---

## 诊断检查清单

完成以下检查，标记每一步的结果:

- [ ] **步骤 1**: 前端发送订阅请求 `/filesystem/watch`
  - [ ] Network 中看到请求
  - [ ] Response 返回 200
  - [ ] Console 显示 `✅ 已开始监听`

- [ ] **步骤 2**: 后端收到订阅请求
  - [ ] 日志中有 `👀 订阅文件夹监听请求`
  - [ ] 日志中有 `✅ 已开始监听文件夹`
  - [ ] `total_watched >= 1`

- [ ] **步骤 3**: 文件变化被监听到
  - [ ] 删除文件后日志中有 `📤 文件系统变化事件已推送`
  - [ ] `event_type` 正确 (delete/create/modify/rename)
  - [ ] `file_path` 正确

- [ ] **步骤 4**: SSE 事件推送到前端
  - [ ] Console 中有 `📨 [useFileSystemAPI.handleFileSystemChange] 收到 SSE 事件`
  - [ ] 事件内容正确
  - [ ] Network 中 SSE 连接保持活跃

- [ ] **步骤 5**: 路径匹配和刷新
  - [ ] Console 中有 `🔍 路径匹配检查`
  - [ ] `match: true`
  - [ ] Console 中有 `✅ 当前文件夹内容发生变化`
  - [ ] Console 中有 `🔄 [useFileSystemAPI.browse] 开始刷新`

---

## 下一步操作

根据上面的检查结果，找到第一个失败的步骤:

1. **如果步骤 1 失败**: 修复前端订阅逻辑
2. **如果步骤 2 失败**: 检查后端路由和 Handler
3. **如果步骤 3 失败**: 检查 `filesystem_watcher.go` 实现
4. **如果步骤 4 失败**: 检查 SSE Manager 和事件广播
5. **如果步骤 5 失败**: 修复路径匹配逻辑

**请按照上面的诊断流程，逐步检查并告诉我哪一步失败了。**
