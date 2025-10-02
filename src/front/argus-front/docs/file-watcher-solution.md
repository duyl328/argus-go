# 文件监听与实时同步技术方案

## 📋 文档信息

- **创建日期**: 2025-10-02
- **项目**: Argus Front - 照片管理系统
- **模块**: FileManager 文件管理器
- **作者**: 架构设计文档
- **版本**: v1.0

---

## 🎯 背景与目标

### 业务背景
用户在双面板文件管理器中浏览本地文件系统时，无法感知外部程序或其他用户对文件的修改（新增、删除、重命名等），导致界面显示内容与实际文件系统状态不一致。

### 核心目标
1. **实时性**: 文件变化后 < 500ms 内前端感知
2. **准确性**: 100% 捕获用户关注目录的文件变化
3. **性能**: 支持双面板同时监听不同目录，不影响UI流畅度
4. **可靠性**: 异常情况下降级为手动刷新，不影响核心功能

---

## 🏗️ 整体架构

### 架构图

```
┌─────────────────────────────────────────────────────────────┐
│                        前端层 (Vue3)                         │
├─────────────────┬───────────────────┬───────────────────────┤
│  FilePane (左)  │   FilePane (右)   │   FieldView (容器)   │
│  - 当前路径     │   - 当前路径      │   - SSE连接管理      │
│  - 本地状态     │   - 本地状态      │   - 全局事件分发     │
└────────┬────────┴─────────┬─────────┴──────────┬────────────┘
         │                  │                     │
         │ REST API         │                     │ SSE Stream
         ▼                  ▼                     ▼
┌─────────────────────────────────────────────────────────────┐
│                     后端层 (Go + Gin)                        │
├──────────────┬────────────────────┬───────────────────────┬─┤
│ FileSystem   │   WatchManager     │   SSE Manager        │ │
│ Handler      │   - 监听器池       │   - 客户端管理       │ │
│ (HTTP API)   │   - 事件聚合       │   - 事件广播         │ │
└──────┬───────┴──────┬─────────────┴──────┬────────────────┘ │
       │              │                     │                   │
       │ 文件操作     │ fsnotify           │                   │
       ▼              ▼                     ▼                   │
┌─────────────────────────────────────────────────────────────┤
│                    操作系统文件系统                          │
│   - inotify (Linux)                                         │
│   - ReadDirectoryChangesW (Windows)                         │
│   - FSEvents (macOS)                                        │
└─────────────────────────────────────────────────────────────┘
```

---

## 🔍 问题深度分析

### 一、监听范围与生命周期管理

#### 1.1 监听粒度问题

| 问题 | 场景描述 | 影响 | 解决方案 |
|------|---------|------|---------|
| **当前目录监听** | 用户浏览 `D:\Photos` | 只需监听此目录的直接子项变化 | ✅ 仅监听当前目录（非递归） |
| **子目录变化** | `D:\Photos\2024\` 内文件变化 | 用户未展开子目录，无需感知 | ❌ 不监听子目录内容 |
| **父目录变化** | 用户在 `D:\Photos` 时，`D:\` 新增文件 | 不影响当前视图 | ❌ 不监听父目录 |
| **双面板同目录** | 左右面板都显示 `D:\Photos` | 同一个监听器被两次引用 | 引用计数管理，共享监听器 |
| **快速切换目录** | 用户在5秒内切换10个目录 | 创建大量监听器 | 延迟销毁策略（30秒后清理） |

**决策**:
- ✅ 采用 **非递归监听** + **引用计数** + **延迟清理**
- ✅ 监听器池大小限制：**20个** (超过LRU淘汰)

#### 1.2 监听器生命周期

```go
type WatcherLifecycle {
    Path          string
    RefCount      int           // 引用计数
    Watcher       *fsnotify.Watcher
    LastAccess    time.Time     // 最后访问时间
    CleanupTimer  *time.Timer   // 清理定时器
}

// 生命周期状态机
[创建] → [活跃] → [空闲(30s)] → [清理] → [销毁]
   ↑_____________↓
      (重新激活)
```

**边界问题**:
- ❓ 用户快速来回切换同一目录？ → 延迟清理避免频繁创建
- ❓ 面板A监听，面板B也监听同目录？ → 引用计数共享实例
- ❓ 面板关闭但监听器未销毁？ → 前端需主动发送 `unwatch` 请求

---

### 二、性能与资源优化

#### 2.1 事件风暴处理

| 场景 | 触发量 | 不处理后果 | 优化方案 |
|------|--------|-----------|---------|
| **批量复制** | 10000个文件 | 10000个SSE推送，前端卡死 | 500ms防抖+按目录聚合 |
| **大文件写入** | 1GB视频保存 | 每次写缓冲触发事件 (100+次) | 同文件500ms内只处理最后一次 |
| **文件夹重命名** | 包含1000个文件 | 同时触发1000个RENAME事件 | 检测目录重命名，单个事件 |
| **递归删除** | 删除多层目录 | 每个文件触发DELETE事件 | 仅通知父目录变化 |

**事件聚合算法**:
```go
type EventAggregator {
    buffer     map[string]*AggregatedEvent  // path -> 聚合事件
    timer      *time.Timer
    debounce   time.Duration = 500ms
}

type AggregatedEvent {
    Path          string
    EventTypes    []FileEventType  // [created, deleted, modified]
    AffectedFiles []string         // 受影响的文件名列表
    Count         int              // 总数
    FirstSeen     time.Time
    LastSeen      time.Time
}

// 聚合逻辑
func (ea *EventAggregator) AddEvent(event FileEvent) {
    key := event.ParentDir

    if existing, ok := ea.buffer[key]; ok {
        existing.Merge(event)
        ea.timer.Reset(ea.debounce)
    } else {
        ea.buffer[key] = NewAggregatedEvent(event)
        ea.timer.Reset(ea.debounce)
    }
}

// 定时器触发后
func (ea *EventAggregator) Flush() {
    for _, aggregated := range ea.buffer {
        SSEManager.BroadcastEvent("file-change", aggregated.ToJSON())
    }
    ea.buffer = make(map[string]*AggregatedEvent)
}
```

#### 2.2 内存占用评估

| 组件 | 单位占用 | 最大数量 | 总占用 | 优化措施 |
|------|---------|---------|--------|---------|
| **fsnotify.Watcher** | ~1MB | 20个 | 20MB | LRU淘汰策略 |
| **事件缓冲区** | ~100KB | 1个 | 100KB | 固定大小环形缓冲 |
| **SSE客户端** | ~50KB/客户端 | 100个 | 5MB | 超时自动断开 |
| **聚合事件缓存** | ~10KB/目录 | 100个 | 1MB | 定时清理 |
| **总计** | - | - | **~26MB** | 可接受 |

#### 2.3 CPU开销分析

| 操作 | 频率 | 单次耗时 | CPU占用 | 优化 |
|------|------|---------|---------|------|
| **fsnotify轮询** | 系统调度 | < 1ms | < 0.1% | 操作系统优化 |
| **事件聚合** | 500ms/次 | < 5ms | < 1% | 批量处理 |
| **SSE推送** | 按需 | < 10ms | < 1% | goroutine异步 |
| **JSON序列化** | 每次推送 | < 2ms | < 0.5% | 预分配buffer |

---

### 三、并发与同步问题

#### 3.1 竞态条件识别

| 场景 | 竞态描述 | 后果 | 解决方案 |
|------|---------|------|---------|
| **双goroutine写map** | WatchManager的watchers map无锁 | panic | sync.RWMutex保护 |
| **文件删除vs读取** | 前端请求文件详情时文件被删除 | 404错误 | API返回文件不存在 |
| **重复添加监听** | 两个面板同时watch同目录 | 创建重复监听器 | 双重检查锁定 |
| **SSE推送时客户端断开** | 写channel时客户端已关闭 | panic | select超时保护 |

**关键代码模式**:
```go
// 双重检查锁定 (Double-Checked Locking)
func (wm *WatchManager) AddWatch(path string) error {
    // 第一次检查（无锁，快速路径）
    wm.mu.RLock()
    if watcher, exists := wm.watchers[path]; exists {
        wm.mu.RUnlock()
        watcher.IncRef()
        return nil
    }
    wm.mu.RUnlock()

    // 获取写锁
    wm.mu.Lock()
    defer wm.mu.Unlock()

    // 第二次检查（持有锁）
    if watcher, exists := wm.watchers[path]; exists {
        watcher.IncRef()
        return nil
    }

    // 创建新监听器
    watcher, err := fsnotify.NewWatcher()
    // ...
}
```

#### 3.2 死锁风险

| 锁顺序 | 场景 | 风险 | 预防 |
|--------|------|------|------|
| `WatchManager.mu` → `SSEManager.clientsMutex` | 事件推送时需获取两个锁 | 可能死锁 | 统一锁顺序 |
| `EventAggregator.mu` → `WatchManager.mu` | 聚合器访问监听器 | 可能死锁 | 避免嵌套锁 |

**锁顺序规范**:
```
全局锁顺序: WatchManager → EventAggregator → SSEManager
```

---

### 四、跨平台兼容性

#### 4.1 文件系统差异

| 特性 | Windows | Linux | macOS | 处理方案 |
|------|---------|-------|-------|---------|
| **路径分隔符** | `\` | `/` | `/` | `filepath.Clean()` 标准化 |
| **大小写敏感** | 不敏感 | 敏感 | 不敏感 | 路径比对时统一转小写 |
| **盘符** | `C:\` | 无 | 无 | 抽象为根路径 `/` |
| **权限模型** | ACL | POSIX | POSIX + ACL | 操作前检查可访问性 |
| **文件锁** | 独占锁 | 建议锁 | 建议锁 | 删除失败返回详细错误 |
| **符号链接** | 支持(需管理员) | 支持 | 支持 | 检测并特殊处理 |

#### 4.2 fsnotify平台差异

| 事件类型 | Windows | Linux | macOS | 统一处理 |
|---------|---------|-------|-------|---------|
| **CREATE** | ✅ | ✅ | ✅ | 直接映射 |
| **WRITE** | ✅ (多次触发) | ✅ | ✅ | 防抖合并 |
| **REMOVE** | ✅ | ✅ | ✅ | 直接映射 |
| **RENAME** | ✅ 两个事件 | ✅ 两个事件 | ⚠️ 单个事件 | 配对检测 |
| **CHMOD** | ❌ | ✅ | ✅ | 忽略或可选支持 |

**RENAME特殊处理**:
```go
type RenameTracker {
    pending map[string]time.Time  // oldPath -> timestamp
    timeout time.Duration = 100ms
}

// Windows/Linux: 收到两个事件 (REMOVE + CREATE)
// macOS: 收到一个 RENAME 事件
func (rt *RenameTracker) HandleEvent(event fsnotify.Event) {
    if event.Op == fsnotify.Rename {
        if runtime.GOOS == "darwin" {
            // macOS直接处理
            return RenamedEvent{OldPath: event.Name}
        } else {
            // Windows/Linux: 等待配对
            rt.pending[event.Name] = time.Now()
        }
    }
}
```

#### 4.3 网络文件系统

| 类型 | Windows | Linux | 支持方案 |
|------|---------|-------|---------|
| **SMB/CIFS** | ✅ 本地映射盘符 | ⚠️ 挂载点 | 检测远程路径，降级轮询 |
| **NFS** | ❌ | ✅ | 降级轮询 |
| **WebDAV** | ⚠️ | ⚠️ | 不支持监听 |
| **云同步(OneDrive/iCloud)** | ⚠️ 延迟通知 | ⚠️ | 接受延迟 |

**检测逻辑**:
```go
func IsNetworkPath(path string) bool {
    if runtime.GOOS == "windows" {
        return strings.HasPrefix(path, `\\`) // UNC路径
    }
    // Linux: 检查挂载点类型
    return isNFSMount(path)
}

func (wm *WatchManager) AddWatch(path string) error {
    if IsNetworkPath(path) {
        logger.Warn("Network path detected, falling back to polling")
        return wm.StartPolling(path, 5*time.Second)
    }
    // 正常fsnotify监听
}
```

---

### 五、用户体验与交互设计

#### 5.1 乐观更新策略

| 操作 | 传统流程 | 乐观更新流程 | 回滚时机 |
|------|---------|-------------|---------|
| **删除文件** | API调用 → 等待响应 → 刷新列表 (500ms) | 立即从UI移除 → API调用 (50ms) | API失败时恢复 |
| **重命名** | 弹窗 → API → 刷新 (600ms) | 即时显示新名称 → API (100ms) | API失败时恢复原名称 |
| **创建文件夹** | API → 刷新 → 定位新项 (400ms) | 立即插入 → API (100ms) | API失败时移除 |

**冲突解决**:
```typescript
class OptimisticUpdateManager {
    pendingOps: Map<string, PendingOperation>

    async deleteFile(path: string) {
        const opId = generateUUID()

        // 1. 乐观更新UI
        this.removeFromUI(path)
        this.pendingOps.set(opId, { type: 'delete', path, rollback: fileData })

        try {
            // 2. 调用API
            await api.delete('/filesystem/item', { path, operationId: opId })
            this.pendingOps.delete(opId)
        } catch (error) {
            // 3. 失败回滚
            const op = this.pendingOps.get(opId)
            this.restoreToUI(op.rollback)
            this.showError('删除失败: ' + error.message)
        }
    }

    // SSE事件处理
    onSSEEvent(event: FileChangeEvent) {
        // 忽略自己触发的操作
        if (this.pendingOps.has(event.operationId)) {
            return
        }

        // 处理外部变化
        this.refresh()
    }
}
```

#### 5.2 进度反馈设计

| 操作类型 | 预计耗时 | 反馈方式 | 取消支持 |
|---------|---------|---------|---------|
| **删除单个文件** | < 100ms | 无需反馈 | ❌ |
| **删除100个文件** | 1-5s | 进度条 (已删除X/100) | ✅ |
| **复制大文件** | 取决于大小 | 传输速度 + 预计剩余时间 | ✅ |
| **批量操作** | 不确定 | Toast通知 + 后台任务列表 | ✅ |

**后台任务系统**:
```typescript
class TaskManager {
    tasks: Map<string, Task>

    // 大文件复制示例
    async copyLargeFile(src: string, dst: string) {
        const taskId = generateUUID()
        const task = new Task({
            id: taskId,
            type: 'copy',
            title: `复制 ${basename(src)}`,
            cancelable: true
        })

        this.tasks.set(taskId, task)

        // 后端支持分块传输 + 进度回调
        await api.copyFile(src, dst, {
            onProgress: (loaded, total) => {
                task.updateProgress(loaded / total * 100)
            },
            signal: task.abortController.signal
        })

        this.tasks.delete(taskId)
    }
}
```

#### 5.3 错误处理与降级

| 错误类型 | 场景 | 用户感知 | 降级方案 |
|---------|------|---------|---------|
| **SSE连接失败** | 网络中断 | Toast: "实时更新已断开" | 30秒轮询 + 手动刷新按钮 |
| **监听器创建失败** | 权限不足 | 静默失败，显示刷新按钮 | 手动刷新模式 |
| **事件推送失败** | 客户端断开 | 无（后端日志） | 自动清理断开的客户端 |
| **文件操作失败** | 被占用/权限不足 | 详细错误弹窗 | 提供重试选项 |

**错误信息设计**:
```typescript
// ❌ 糟糕的错误提示
"操作失败"

// ✅ 良好的错误提示
{
    title: "无法删除文件",
    message: "文件 'photo.jpg' 正被其他程序占用",
    detail: "请关闭打开此文件的程序后重试",
    actions: [
        { label: "重试", action: retry },
        { label: "强制删除", action: forceDelete, dangerous: true },
        { label: "取消", action: close }
    ]
}
```

---

### 六、安全与权限控制

#### 6.1 路径遍历攻击防护

| 攻击向量 | 示例 | 危害 | 防护措施 |
|---------|------|------|---------|
| **相对路径** | `../../../etc/passwd` | 访问系统文件 | `filepath.Clean()` + 白名单验证 |
| **符号链接逃逸** | `/safe/link -> /etc/` | 绕过目录限制 | `filepath.EvalSymlinks()` + 边界检查 |
| **UNC路径注入** | `\\evil.com\share` | 访问远程恶意服务器 | 拒绝UNC路径 |

**安全路径验证**:
```go
type PathValidator struct {
    allowedRoots []string  // 允许访问的根目录
}

func (pv *PathValidator) Validate(userPath string) error {
    // 1. 清理路径
    cleaned := filepath.Clean(userPath)

    // 2. 解析符号链接
    resolved, err := filepath.EvalSymlinks(cleaned)
    if err != nil {
        return fmt.Errorf("invalid path: %w", err)
    }

    // 3. 检查是否在允许的根目录下
    for _, root := range pv.allowedRoots {
        if strings.HasPrefix(resolved, root) {
            return nil
        }
    }

    return fmt.Errorf("access denied: path outside allowed roots")
}
```

#### 6.2 权限检查

| 操作 | 需要权限 | 检查时机 | 失败处理 |
|------|---------|---------|---------|
| **读取目录** | 读+执行 | API调用前 | 返回403 + 错误信息 |
| **删除文件** | 写+执行 | API调用前 | 提示权限不足 |
| **创建监听** | 读+执行 | 添加监听前 | 降级为轮询 |

```go
func (fs *FileSystemService) CheckPermission(path string, perm Permission) error {
    info, err := os.Stat(path)
    if err != nil {
        return err
    }

    mode := info.Mode()

    // Unix权限检查
    if runtime.GOOS != "windows" {
        uid := os.Getuid()
        gid := os.Getgid()

        // 检查owner/group/other权限
        // ... 详细权限位检查
    } else {
        // Windows ACL检查（复杂，可使用syscall）
        return checkWindowsACL(path, perm)
    }

    return nil
}
```

---

### 七、测试与可观测性

#### 7.1 测试场景矩阵

| 场景类别 | 具体测试项 | 验证点 |
|---------|-----------|--------|
| **正常流程** | 单文件创建/删除/重命名 | 事件正确触发 |
| **边界条件** | 空目录、单文件、10000文件 | 性能可接受 |
| **并发操作** | 双面板同时操作同文件 | 无竞态条件 |
| **异常恢复** | SSE断线、监听器崩溃 | 自动重连、降级 |
| **跨平台** | Windows/Linux/macOS | 行为一致 |
| **性能压测** | 1000次/秒事件触发 | CPU < 10%, 内存稳定 |

#### 7.2 监控指标

```go
type WatcherMetrics struct {
    ActiveWatchers      prometheus.Gauge      // 当前活跃监听器数量
    EventsProcessed     prometheus.Counter    // 已处理事件总数
    EventLatency        prometheus.Histogram  // 事件处理延迟
    SSEConnections      prometheus.Gauge      // SSE连接数
    FailedOperations    prometheus.Counter    // 失败操作计数
    AggregationRate     prometheus.Gauge      // 事件聚合率
}

// 关键告警
- ActiveWatchers > 20  → 监听器泄漏
- EventLatency > 1s    → 性能瓶颈
- FailedOperations增长 → 系统异常
```

#### 7.3 调试工具

```typescript
// 开发模式下的调试面板
class WatcherDebugPanel {
    showMetrics() {
        return {
            watchedPaths: watchManager.getPaths(),
            eventQueue: eventAggregator.getQueueSize(),
            sseClients: sseManager.getClientCount(),
            recentEvents: eventHistory.getLast(50)
        }
    }

    // 可视化事件流
    visualizeEvents() {
        // 实时显示事件流动图
        // path -> aggregation -> SSE -> UI update
    }
}
```

---

## ❓ 深度问题清单

### 核心决策问题

#### 监听策略
- [ ] **Q1**: 仅监听当前目录 vs 递归监听子目录？
  - 推荐：**仅监听当前目录**
  - 理由：性能可控，用户未展开的子目录无需感知

- [ ] **Q2**: 监听器数量限制？
  - 推荐：**20个** (超过LRU淘汰)
  - 理由：平衡资源占用和用户体验

- [ ] **Q3**: 监听器清理时机？
  - 推荐：**30秒无访问后延迟清理**
  - 理由：避免频繁创建/销毁

#### 性能优化
- [ ] **Q4**: 事件防抖时间？
  - 推荐：**500ms**
  - 备选：100ms (更实时但可能风暴)

- [ ] **Q5**: 事件聚合策略？
  - 推荐：**按目录聚合 + 计数**
  - 备选：详细列出每个文件 (网络开销大)

- [ ] **Q6**: 大批量操作处理？
  - 推荐：**后台任务 + 进度通知**
  - 备选：阻塞UI直到完成

#### 用户体验
- [ ] **Q7**: 乐观更新策略？
  - 推荐：**所有写操作都乐观更新**
  - 备选：仅删除操作乐观更新

- [ ] **Q8**: SSE断线后？
  - 推荐：**降级为30秒轮询 + Toast提示**
  - 备选：完全禁用自动刷新

- [ ] **Q9**: 网络文件系统？
  - 推荐：**自动降级为轮询(5秒)**
  - 备选：不支持监听

#### 安全与权限
- [ ] **Q10**: 允许访问的路径范围？
  - 推荐：**用户主目录 + 已挂载的驱动器**
  - 备选：不限制 (存在安全风险)

- [ ] **Q11**: 权限不足时？
  - 推荐：**静默降级 + 显示刷新按钮**
  - 备选：弹窗报错

### 边界场景问题

#### 文件系统特殊情况
- [ ] **Q12**: 监听的目录被删除？
  - 方案：检测到删除事件 → 自动跳转到父目录

- [ ] **Q13**: 符号链接的目标被删除？
  - 方案：显示为损坏的链接，禁止操作

- [ ] **Q14**: 文件名包含特殊字符 (Unicode/Emoji)？
  - 方案：使用 `filepath` 标准库处理，支持所有合法字符

- [ ] **Q15**: 超长路径 (Windows 260字符限制)？
  - 方案：使用 `\\?\` 前缀绕过限制

- [ ] **Q16**: 大小写敏感文件系统？
  - 方案：路径比对时根据OS特性处理

#### 并发与同步
- [ ] **Q17**: 两个用户同时删除同一文件？
  - 方案：第二个用户收到404，提示文件已被删除

- [ ] **Q18**: 文件在监听器添加前就被修改？
  - 方案：接受轻微延迟，用户手动刷新

- [ ] **Q19**: 事件聚合期间用户刷新页面？
  - 方案：聚合器独立运行，不影响刷新

#### 性能极限
- [ ] **Q20**: 单目录10万个文件？
  - 方案：虚拟滚动 + 分页加载 + 监听器仍正常工作

- [ ] **Q21**: 每秒1000个文件变化？
  - 方案：500ms防抖 + 聚合 + 显示"目录变化频繁，已暂停刷新"

- [ ] **Q22**: 同时100个SSE客户端？
  - 方案：goroutine异步推送，可支持1000+客户端

#### 跨平台差异
- [ ] **Q23**: Windows文件被占用无法删除？
  - 方案：返回详细错误 "文件被 xx.exe 占用"

- [ ] **Q24**: macOS的 .DS_Store 文件？
  - 方案：可选过滤系统隐藏文件

- [ ] **Q25**: Linux的 /proc、/sys 虚拟文件系统？
  - 方案：检测并拒绝监听

#### 错误恢复
- [ ] **Q26**: fsnotify返回错误？
  - 方案：记录日志 + 自动重试3次 + 降级轮询

- [ ] **Q27**: SSE Manager崩溃？
  - 方案：recover捕获 + 重启SSE服务

- [ ] **Q28**: 磁盘空间不足？
  - 方案：创建/复制前检查可用空间

#### 数据一致性
- [ ] **Q29**: 前端显示的文件已被外部删除？
  - 方案：操作时API返回404，前端自动刷新

- [ ] **Q30**: 事件顺序错乱 (网络延迟)？
  - 方案：事件携带时间戳，前端忽略过期事件

---

## 🚀 实施计划

### 阶段一：核心功能 (预计4小时)

#### 后端 (2.5小时)
1. **文件操作API实现** (1小时)
   - 完善 `DeleteItem` (支持文件和目录)
   - 完善 `CreateDirectory`
   - 完善 `MoveItem` (重命名+移动)
   - 完善 `CopyItem`

2. **WatchManager实现** (1小时)
   - 基础fsnotify封装
   - 监听器生命周期管理
   - 引用计数机制

3. **SSE事件推送集成** (30分钟)
   - 文件变化事件定义
   - WatchManager → SSE Manager连接

#### 前端 (1.5小时)
4. **SSE连接管理** (30分钟)
   - EventSource封装
   - 断线重连逻辑

5. **文件操作UI** (30分钟)
   - 右键菜单 (删除/重命名/复制/粘贴)
   - 快捷键支持 (Del/F2/Ctrl+C/Ctrl+V)

6. **自动刷新逻辑** (30分钟)
   - 监听SSE事件
   - 判断是否影响当前视图
   - 触发刷新

### 阶段二：优化体验 (预计3小时)

7. **事件防抖与聚合** (1小时)
   - EventAggregator实现
   - 500ms防抖逻辑

8. **乐观更新** (1小时)
   - OptimisticUpdateManager
   - 操作回滚机制

9. **进度反馈** (1小时)
   - 批量操作进度条
   - 后台任务管理器

### 阶段三：边界处理 (预计2小时)

10. **错误处理** (1小时)
    - 详细错误信息
    - 降级策略
    - 重试机制

11. **跨平台适配** (1小时)
    - 路径分隔符处理
    - 网络文件系统检测
    - 权限检查

### 阶段四：测试与监控 (预计2小时)

12. **单元测试** (1小时)
    - WatchManager测试
    - EventAggregator测试

13. **集成测试** (30分钟)
    - 完整流程测试
    - 并发测试

14. **监控埋点** (30分钟)
    - 性能指标收集
    - 错误日志记录

---

## 📊 风险评估

| 风险项 | 概率 | 影响 | 等级 | 缓解措施 |
|--------|------|------|------|---------|
| **fsnotify跨平台差异** | 中 | 高 | 🟡 中 | 充分测试 + 适配层 |
| **性能瓶颈 (事件风暴)** | 低 | 高 | 🟡 中 | 防抖聚合 + 限流 |
| **SSE断线** | 高 | 中 | 🟡 中 | 自动重连 + 降级轮询 |
| **内存泄漏 (监听器未清理)** | 中 | 中 | 🟡 中 | 延迟清理 + 上限控制 |
| **并发竞态** | 低 | 高 | 🟢 低 | Mutex保护 + 代码审查 |
| **权限问题** | 中 | 低 | 🟢 低 | 静默降级 |
| **网络文件系统不支持** | 高 | 低 | 🟢 低 | 自动降级轮询 |

---

## 📝 未解决问题 (待讨论)

1. **多用户协作**: 如果未来支持多用户同时浏览同一网络路径，如何避免冲突？
2. **历史记录**: 是否需要记录文件操作历史 (用于撤销/恢复)？
3. **回收站**: 删除文件是否进入系统回收站 vs 永久删除？
4. **云同步集成**: 是否需要感知OneDrive/Dropbox等云盘的同步状态？
5. **大文件传输**: 复制超大文件 (>10GB) 是否需要断点续传？
6. **文件预览**: 监听到图片修改后，是否自动刷新预览缩略图？
7. **批量操作事务**: 批量操作是否支持"全部成功"或"全部回滚"？
8. **性能监控**: 是否需要在生产环境暴露监控指标API？

---

## 🔗 相关资源

- [fsnotify官方文档](https://github.com/fsnotify/fsnotify)
- [SSE规范](https://html.spec.whatwg.org/multipage/server-sent-events.html)
- [文件系统监听最佳实践](https://stackoverflow.com/questions/tagged/fsnotify)

---

## 📅 更新日志

- **2025-10-02**: 初始版本，完成深度问题分析
