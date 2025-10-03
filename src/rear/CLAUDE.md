# Argus 照片管理系统 - Claude 开发指南

> **文档说明**: 此文档专为 Claude AI 助手设计，用于快速理解项目架构、功能模块、开发规范等关键信息。
> **最后更新**: 2025-10-03

---

## 📋 项目概览

### 基本信息
- **项目名称**: Argus 照片管理系统 (后端服务)
- **开发语言**: Go 1.24
- **项目类型**: HTTP RESTful API 服务
- **核心功能**: 照片管理、图像处理、EXIF 提取、缩略图生成、实时通信
- **部署方式**: 独立可执行文件 (支持 Windows/Linux/macOS)
- **数据库**: SQLite (默认) / MySQL (可选)
- **Web 框架**: Gin (v1.10.1)
- **ORM**: GORM (v1.30.0)

### 设计目标
1. **高性能图像处理**: 利用外部工具链 (ExifTool/ImageMagick/LibVips) 实现高效图像处理
2. **并发任务管理**: 基于 Go 协程的高并发任务调度系统，支持动态并发数调整
3. **可扩展架构**: 清晰的分层架构，便于功能扩展和维护
4. **实时通信**: SSE (Server-Sent Events) 实现任务进度实时推送
5. **跨平台支持**: 自动管理平台特定的外部工具依赖

---

## 🏗️ 项目架构

### 目录结构
```
D:\go-argus\src\rear\
├── main.go                     # 应用入口，服务启动逻辑
├── go.mod/go.sum              # Go 模块依赖
├── config.yaml                # 配置文件
├── CLAUDE.md                  # Claude AI 开发指南 (本文件)
├── README.md                  # 用户文档
│
├── internal/                  # 内部业务代码 (不对外暴露)
│   ├── api/                   # API 层 - 对外接口定义
│   ├── config/                # 配置管理 - 配置加载/解析/验证
│   ├── container/             # 依赖注入容器 - 管理仓库和任务容器
│   ├── db/                    # 数据库管理 - 连接/迁移/写任务队列
│   ├── handler/               # HTTP 处理器 - 处理请求和响应
│   ├── model/                 # 数据模型 - 业务模型和数据表定义
│   ├── repositories/          # 数据访问层 - 数据库 CRUD 操作
│   ├── router/                # 路由配置 - HTTP 路由注册
│   ├── service/               # 业务服务 - 中间件和业务逻辑
│   ├── utils/                 # 内部工具 - 任务调度/系统工具/外部工具管理
│   └── workflow/              # 工作流 - 图像处理任务流程
│
├── pkg/                       # 公共包 (可对外暴露)
│   ├── img_utils/             # 图像工具 - 格式检测/转换
│   ├── logger/                # 日志工具 - Zap 日志封装
│   ├── sse/                   # SSE 实时通信 - 客户端/管理器
│   ├── utils/                 # 通用工具 - 文件/哈希/系统工具
│   └── system/                # 系统信息 - CPU/内存/磁盘/网络监控
│
├── tools/                     # 外部工具链 (平台特定)
│   ├── windows_amd64/         # Windows 64位工具
│   ├── linux_amd64/           # Linux 64位工具
│   └── darwin_arm64/          # macOS ARM64工具
│
├── scripts/                   # 构建脚本
├── logs/                      # 日志文件
├── thumbnail/                 # 缩略图存储
├── db/                        # 数据库文件 (SQLite)
└── examples/                  # 示例代码
```

### 分层架构设计

```
┌─────────────────────────────────────────────────────┐
│                   HTTP Layer                        │
│  (Gin Router + Middleware + CORS + Error Handler)  │
└─────────────────┬───────────────────────────────────┘
                  │
┌─────────────────▼───────────────────────────────────┐
│                Handler Layer                        │
│  (LibraryHandler, ExifHandler, PhotoHandler, etc)  │
└─────────────────┬───────────────────────────────────┘
                  │
┌─────────────────▼───────────────────────────────────┐
│              Repository Layer                       │
│  (LibraryRepo, ExifRepo, PhotoRepo, UserRepo)      │
└─────────────────┬───────────────────────────────────┘
                  │
┌─────────────────▼───────────────────────────────────┐
│               Database Layer                        │
│         (GORM + SQLite/MySQL + 写队列)              │
└─────────────────────────────────────────────────────┘

并行的任务处理流:
┌─────────────────────────────────────────────────────┐
│             Task Workflow Layer                     │
│  (ImgTask + TaskManager + External Tools)          │
└─────────────────────────────────────────────────────┘
```

---

## 🔧 核心技术栈

### Go 依赖库

#### Web 框架
- **gin-gonic/gin v1.10.1** - 主 HTTP 框架
- **gin-contrib/cors v1.7.5** - CORS 跨域处理
- **gin-contrib/pprof v1.5.3** - 性能分析工具

#### 数据库 ORM
- **gorm.io/gorm v1.30.0** - ORM 框架
- **gorm.io/driver/sqlite v1.6.0** - SQLite 驱动
- **gorm.io/driver/mysql v1.6.0** - MySQL 驱动

#### 日志系统
- **go.uber.org/zap v1.27.0** - 高性能结构化日志
- **gopkg.in/natefinch/lumberjack.v2 v2.2.1** - 日志轮转

#### 工具库
- **google/uuid v1.6.0** - UUID 生成
- **h2non/filetype v1.1.3** - 文件类型检测
- **gopkg.in/yaml.v3 v3.0.1** - YAML 配置解析

### 外部工具链 (跨平台)

#### 图像处理工具
1. **ExifTool** - EXIF 元数据提取 (Perl 工具)
2. **ImageMagick** - 图像格式转换和处理 (magick.exe)
3. **LibVips** - 高性能图像处理库 (vips.exe)

#### 工具管理策略
- **开发模式**: 自动从 `tools/{platform}/` 解压工具到 `AppDir/tools/`
- **生产模式**: 手动配置工具路径或使用系统安装的工具
- **平台检测**: 自动识别 `windows_amd64`, `linux_amd64`, `darwin_arm64` 等

---

## 📦 功能模块详解

### 1. 配置管理模块 (`internal/config/`)

**文件**: `config.go`, `loader.go`, `database.go`

**功能**:
- 加载 `config.yaml` 配置文件
- 支持默认配置和用户自定义配置合并
- 配置热更新 (未来扩展)

**关键配置项**:
```yaml
app:
  port: "9484"                 # HTTP 端口
  mode: "release"              # debug/release
  development: true            # 开发/生产模式切换

database:
  type: "sqlite"               # sqlite/mysql
  path: "db/db.sqlite"         # SQLite 路径

image:
  thumbnail_format: "jpg"      # jpg/webp
  thumbnail_size: [512, 720]   # 缩略图尺寸
  thumbnail_quality: 80        # 质量 (1-100)

task:
  concurrency: 0               # 并发数 (0=自动)
  auto_adjust: true            # 动态调整
```

**设计目标**: 灵活配置，支持多环境部署

---

### 2. 数据模型 (`internal/model/tables/`)

#### 核心数据表

**Photo 表** (`photo_table.go`)
```go
type Photo struct {
    ImgPath      string     // 图片路径
    ImgName      string     // 图片名称
    Hash         string     // SHA256 哈希 (唯一索引)
    Width        int        // 宽度
    Height       int        // 高度
    AspectRatio  float32    // 宽高比
    FileSize     int64      // 文件大小
    Format       string     // 格式 (jpg/png/etc)
    TakenAt      *time.Time // 拍摄时间
    Rating       int        // 评分 (0-5)
    ViewCount    int        // 访问次数
}
```

**PhotoExif 表** (`exif_table.go`)
```go
type PhotoExif struct {
    Hash         string  // 照片哈希 (主键)

    // 基础文件信息
    FileName     string
    FileSize     int64
    ImageWidth   int
    ImageHeight  int
    MIMEType     string
    ModifyDate   string

    // EXIF 摄影参数
    Model        string  // 相机型号 (索引)
    Make         string  // 相机厂商 (索引)
    ISO          int     // ISO 值 (索引)
    GPSLatitude  float64 // GPS 纬度 (索引)
    GPSLongitude float64 // GPS 经度 (索引)
    ExposureTime float64 // 曝光时间
    Aperture     float64 // 光圈
    FocalLength  float64 // 焦距
    LensID       string  // 镜头ID
    DateTimeOrig string  // 拍摄时间 (索引)

    // 扩展字段 (JSON)
    OtherFields  JSONMap // 其他 EXIF 字段
}
```

**设计特点**:
- **Hash 作为关联键**: Photo 和 PhotoExif 通过 Hash 关联，简化设计
- **索引优化**: 常用查询字段 (相机型号、ISO、GPS、时间) 建立索引
- **JSON 扩展字段**: 灵活存储非标准 EXIF 字段
- **无外键约束**: 避免级联删除复杂性，提升性能

---

### 3. 数据访问层 (`internal/repositories/`)

**BaseService** (`base_service.go`)
- 管理写操作协程，避免数据库锁竞争
- 单例模式，全局唯一写队列

**LibraryRepository** (`library_repository.go`)
- 图片库路径的 CRUD 操作
- 支持启用/禁用图片库

**PhotoRepository** (`photo_repository.go`)
- 照片基础信息 CRUD
- 支持按时间范围、评分、格式查询

**ExifRepository** (`exif_repository.go`)
- EXIF 信息查询和统计
- 支持多种筛选: 相机型号、ISO 范围、GPS、时间范围
- 提供统计接口: 相机使用统计、时间线分布

**设计目标**: 封装数据库操作，提供清晰的业务接口

---

### 4. HTTP 处理器 (`internal/handler/`)

**LibraryHandler** (`library_handler.go`)
- 图片库管理: 添加/删除/更新/查询
- 启动图片索引任务

**PhotoHandler** (`photo_handler.go`)
- 照片文件获取 (原图/缩略图)
- 照片列表查询 (分页、时间范围)
- 时间线统计

**ExifHandler** (`exif_handler.go`)
- EXIF 信息查询 (单个/批量/分页)
- EXIF 筛选: 相机、ISO、光圈、GPS
- EXIF 统计: 相机使用量、时间分布

**FileSystemHandler** (`filesystem_handler.go`)
- 文件系统浏览 (新增功能)
- 磁盘使用情况查询
- 文件搜索

**FileOperationsHandler** (对应功能)
- 创建目录/删除/移动/复制文件

**SSEHandler** (`sse_handler.go`)
- SSE 连接管理
- 实时推送任务进度

**设计目标**: 薄层处理器，主要负责参数验证和响应格式化

---

### 5. 图像处理工作流 (`internal/workflow/`)

**ImgTask** (`img_task.go`)

**任务状态管理**:
```go
const (
    StatusPending  = "pending"  // 等待处理
    StatusRunning  = "running"  // 正在处理
    StatusPaused   = "paused"   // 已暂停
    StatusFailed   = "failed"   // 处理失败
    StatusDone     = "done"     // 处理完成
)
```

**处理步骤**:
1. **Initializing** - 初始化任务
2. **Validating** - 验证文件格式
3. **ReadingFile** - 读取文件内容
4. **CalculatingHash** - 计算 SHA256 哈希
5. **ExtractingExif** - 提取 EXIF 信息
6. **ConvertingFormat** - 格式转换 (如需)
7. **GeneratingThumbnails** - 生成多尺寸缩略图
8. **SavingToDatabase** - 保存到数据库
9. **IntelligentAnalysis** - 智能分析 (AI 标签等)

**TaskManager** (任务管理器)
- **并发控制**: 根据 CPU 核心数自动调整并发数
- **系统监控**: 监控 CPU/内存使用率，动态调整
- **暂停/恢复**: 支持任务暂停和恢复
- **进度通知**: 通过 SSE 推送实时进度

**设计目标**: 高效、可监控、可恢复的图像处理流水线

---

### 6. 外部工具管理 (`internal/utils/tools/`)

**ExifTool** (`exif_tool.go`)
- 功能: 提取 EXIF 元数据
- 输出: JSON 格式
- 性能: 快速读取，无需解码图像

**ImageMagick** (`image_magick_tool.go`)
- 功能: 图像格式转换、裁剪、缩放
- 命令: `magick convert`
- 支持: 200+ 图像格式

**LibVips** (`libvips_tool.go`)
- 功能: 高性能缩略图生成
- 特点: 流式处理，低内存占用
- 性能: 比 ImageMagick 快 4-8 倍

**ToolManager** (`tool_manager.go`)
- 工具路径查找和验证
- 版本检测
- 跨平台兼容性处理

**设计目标**: 统一的工具调用接口，易于扩展新工具

---

### 7. SSE 实时通信 (`pkg/sse/`)

**Manager** (`manager.go`)
- 管理所有 SSE 客户端连接
- 支持广播和单点发送
- 心跳检测 (30s 间隔)
- 自动清理超时连接

**Event** (事件结构)
```go
type Event struct {
    ID    string // 事件ID
    Event string // 事件类型
    Data  string // 事件数据 (JSON)
    Retry int    // 重试间隔 (ms)
}
```

**使用场景**:
- 图片索引任务进度推送
- 图像处理状态更新
- 系统通知

**设计目标**: 低延迟、高并发的实时通信

---

## 🛠️ 工具函数库

### pkg/utils/ (通用工具)

#### FileUtils (`file_utils.go`)
**功能**: 文件和目录操作
```go
// 基础操作
FileUtils.Exists(path string) bool              // 检查文件/目录是否存在
FileUtils.IsDir(path string) bool               // 是否为目录
FileUtils.IsFile(path string) bool              // 是否为文件

// 目录操作
FileUtils.GetDirectories(dirPath string)        // 获取所有子目录
FileUtils.GetAllFiles(dirPath, recursive)       // 递归获取文件
FileUtils.GetFilteredFiles(path, recursive, types) // 按类型筛选文件
FileUtils.CreateDir(dirPath string)             // 创建目录

// 文件操作
FileUtils.CopyFile(src, dst string)             // 复制文件
FileUtils.CopyDir(src, dst string)              // 复制目录
FileUtils.MoveFile(src, dst string)             // 移动/重命名
FileUtils.Delete(path string)                   // 删除文件/目录

// 文件读写
FileUtils.ReadFile(filePath string)             // 读取文件内容
FileUtils.ReadFileLines(filePath string)        // 按行读取
FileUtils.WriteFile(filePath, content)          // 写入文件
FileUtils.AppendFile(filePath, content)         // 追加内容

// 文件信息
FileUtils.GetFileSize(filePath string)          // 获取文件大小
FileUtils.GetModTime(filePath string)           // 获取修改时间
FileUtils.GetExtension(filePath string)         // 获取扩展名
FileUtils.GetFileName(filePath string)          // 获取文件名 (不含扩展名)
FileUtils.GetDirSize(dirPath string)            // 获取目录大小
FileUtils.FormatFileSize(bytes int64)           // 格式化文件大小

// 哈希计算
FileUtils.GetFileMD5(filePath string)           // 计算 MD5
FileUtils.GetFileSHA256(filePath string)        // 计算 SHA256

// 搜索
FileUtils.SearchFiles(dirPath, pattern, recursive) // 搜索文件
```

**使用规范**:
- ✅ **始终使用 FileUtils**，不要直接调用 `os` 包
- ✅ **错误处理**: 所有方法都返回 error，必须检查
- ✅ **路径安全**: 避免路径遍历攻击

#### HashUtils (`hash_utils.go`)
**功能**: 文件和字符串哈希计算
```go
// 文件哈希
HashUtils.HashFile(filename, hashType)          // 通用文件哈希
HashUtils.MD5File(filename)                     // MD5 快捷方法
HashUtils.SHA256File(filename)                  // SHA256 快捷方法
HashUtils.SHA512File(filename)                  // SHA512 快捷方法

// 字符串哈希
HashUtils.HashString(data, hashType)            // 通用字符串哈希
HashUtils.MD5String(data)                       // MD5 快捷方法
HashUtils.SHA256String(data)                    // SHA256 快捷方法

// 高级功能
HashUtils.HashFileWithProgress(filename, hashType, callback) // 带进度回调
HashUtils.HashMultipleFiles(filenames, hashType) // 并发计算多个文件
HashUtils.CompareFiles(file1, file2, hashType)  // 比较文件
HashUtils.FindDuplicates(filenames, hashType)   // 查找重复文件
HashUtils.HashFileMultipleAlgorithms(filename, types) // 多算法计算

// 缩略图路径生成
HashUtils.HashThumbPath(basePath, hash, filename, ext) // 生成缩略图路径
```

**使用规范**:
- ✅ **照片哈希**: 统一使用 `SHA256`
- ✅ **并发计算**: 批量处理时使用 `HashMultipleFiles`
- ✅ **进度反馈**: 大文件使用 `HashFileWithProgress`

#### SysUtils (`sys_utils.go` + 平台特定文件)
**功能**: 系统信息获取
```go
sysUtil := NewSysUtils()

// 基础信息
sysUtil.GetSystemInfo()         // 操作系统/架构/CPU核心数
sysUtil.GetCurrentProcessInfo() // 当前进程信息
sysUtil.GetSystemUptime()       // 系统运行时间

// 硬件信息
sysUtil.GetCPUInfo()            // CPU 型号/核心数/频率
sysUtil.GetMemoryInfo()         // 内存使用情况
sysUtil.GetDiskUsage(path)      // 磁盘使用情况
sysUtil.GetNetworkInterfaces()  // 网络接口信息

// 格式化
sysUtil.FormatBytes(bytes)      // 格式化字节数 (1.2GB)

// 完整打印
sysUtil.PrintSystemInfo()       // 打印所有系统信息
```

**使用规范**:
- ✅ **性能监控**: 用于任务管理器动态调整并发数
- ✅ **跨平台**: 自动选择平台特定实现

### pkg/system/ (系统信息模块)

**新增功能**: 完整的系统信息获取和监控

```go
// 设备管理
DeviceManager.GetAllDevices()       // 获取所有设备
DeviceManager.GetDeviceByPath(path) // 按路径获取设备
DeviceManager.GetDeviceUsage(path)  // 设备使用情况

// 进程管理
ProcessManager.GetCurrentProcess()  // 当前进程信息
ProcessManager.GetProcessByPID(pid) // 按PID获取进程
ProcessManager.ListProcesses()      // 列出所有进程

// 网络工具
NetworkUtils.GetLocalIP()           // 获取本地IP
NetworkUtils.GetMACAddress()        // 获取MAC地址
NetworkUtils.IsPortAvailable(port)  // 检查端口是否可用
```

**使用规范**:
- ✅ **监控**: 用于系统监控和性能分析
- ✅ **集成**: 与 `pkg/utils/SysUtils` 互补使用

### pkg/logger/ (日志工具)

**功能**: 基于 Zap 的高性能日志
```go
// 初始化
logConfig := logger.DefaultConfig()
logConfig.Level = logger.DebugLevel
logConfig.LogPath = "./logs"
logger.InitDefaultLogger(logConfig)

// 基础日志
logger.Debug("调试信息")
logger.Info("普通信息")
logger.Warn("警告信息")
logger.Error("错误信息")
logger.Fatal("致命错误")

// 格式化日志
logger.Infof("处理文件: %s", filename)
logger.Errorf("处理失败: %v", err)

// 结构化日志
logger.Info("任务完成",
    zap.String("task_id", id),
    zap.Int("duration", elapsed),
    zap.Error(err))
```

**配置项**:
```go
type LogConfig struct {
    Level           LogLevel // 日志级别
    LogPath         string   // 日志路径
    MaxSize         int      // 单文件最大 (MB)
    MaxBackups      int      // 保留文件数
    MaxAge          int      // 保留天数
    Compress        bool     // 是否压缩
    EnableConsole   bool     // 是否输出到控制台
    EnableCaller    bool     // 是否显示调用者
}
```

**使用规范**:
- ✅ **统一使用 logger 包**，不要使用 `log` 或 `fmt.Println`
- ✅ **结构化日志**: 关键信息使用 `zap` 字段
- ✅ **日志级别**: Debug (开发) / Info (生产) / Warn (警告) / Error (错误)

### pkg/sse/ (SSE 实时通信)

**功能**: Server-Sent Events 管理
```go
// 创建管理器
opts := sse.DefaultOptions()
manager := sse.NewManager(opts)

// 注册客户端
manager.HandleSSE(c *gin.Context)

// 发送事件
event := &sse.Event{
    Event: "progress",
    Data:  `{"percent": 50, "message": "处理中..."}`,
}
manager.Broadcast(event)           // 广播
manager.SendToClient(clientID, event) // 单点发送

// 关闭
manager.Close()
```

**使用规范**:
- ✅ **任务进度**: 图像处理任务实时推送进度
- ✅ **事件命名**: 使用清晰的事件名 (`progress`, `error`, `complete`)
- ✅ **JSON 数据**: Data 字段使用 JSON 格式

---

## 📐 开发规范

### 1. 代码组织原则

#### 不要重复造轮子
- ✅ **优先使用已有工具**: 检查 `pkg/utils/` 和 `internal/utils/` 是否已有实现
- ✅ **扩展而非重写**: 如果功能不足，扩展现有工具函数
- ❌ **禁止重复实现**: 不要在多处实现相同功能

#### 工具函数使用规范
```go
// ✅ 正确: 使用 FileUtils
if !utils.FileUtils.Exists(filePath) {
    return errors.New("文件不存在")
}

// ❌ 错误: 直接使用 os 包
if _, err := os.Stat(filePath); os.IsNotExist(err) {
    return errors.New("文件不存在")
}
```

#### 引入规范
```go
// ✅ 正确: 分组引入，清晰分类
import (
    // 标准库
    "context"
    "fmt"
    "time"

    // 第三方库
    "github.com/gin-gonic/gin"
    "go.uber.org/zap"

    // 项目内部包
    "rear/internal/config"
    "rear/pkg/logger"
    "rear/pkg/utils"
)

// ✅ 正确: 别名避免冲突
import (
    "rear/internal/utils"
    utilsPkg "rear/pkg/utils"
)
```

### 2. 错误处理规范

```go
// ✅ 正确: 详细的错误上下文
hash, err := utils.HashUtils.SHA256File(filePath)
if err != nil {
    logger.Errorf("计算文件哈希失败: %v", err,
        zap.String("file_path", filePath))
    return fmt.Errorf("计算文件哈希失败: %w", err)
}

// ✅ 正确: 使用 errors.Is/As 判断错误类型
if errors.Is(err, os.ErrNotExist) {
    return fmt.Errorf("文件不存在: %s", filePath)
}

// ❌ 错误: 忽略错误
hash, _ := utils.HashUtils.SHA256File(filePath)
```

### 3. 日志规范

```go
// ✅ 正确: 结构化日志
logger.Info("开始处理图片",
    zap.String("file_path", filePath),
    zap.String("hash", hash),
    zap.Int("width", width),
    zap.Int("height", height))

// ✅ 正确: 错误日志包含上下文
logger.Errorf("EXIF提取失败: %v", err,
    zap.String("file_path", filePath),
    zap.String("tool", "exiftool"))

// ❌ 错误: 使用 fmt.Println
fmt.Println("开始处理图片:", filePath)
```

### 4. 并发安全规范

```go
// ✅ 正确: 使用互斥锁保护共享数据
type TaskManager struct {
    mu    sync.Mutex
    tasks map[string]*Task
}

func (tm *TaskManager) AddTask(task *Task) {
    tm.mu.Lock()
    defer tm.mu.Unlock()
    tm.tasks[task.ID] = task
}

// ✅ 正确: 使用 context 控制协程生命周期
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

go func(ctx context.Context) {
    select {
    case <-ctx.Done():
        return
    case result := <-resultChan:
        // 处理结果
    }
}(ctx)
```

### 5. 数据库操作规范

```go
// ✅ 正确: 使用 Repository 层
photo := &tables.Photo{
    ImgPath: filePath,
    Hash:    hash,
}
err := photoRepo.Create(photo)

// ✅ 正确: 使用写队列避免锁竞争
photoRepo.AsyncCreate(photo)

// ❌ 错误: 直接使用 GORM
db.Create(&photo)
```

### 6. HTTP 接口规范

```go
// ✅ 正确: 统一的响应格式
c.JSON(http.StatusOK, model.Response{
    Code:    http.StatusOK,
    Message: "Success",
    Data:    result,
})

// ✅ 正确: 参数验证
var req AddLibraryRequest
if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(http.StatusBadRequest, model.Response{
        Code:    http.StatusBadRequest,
        Message: fmt.Sprintf("Invalid request: %v", err),
    })
    return
}

// ✅ 正确: 使用 gin 的参数绑定
hash := c.Param("hash")
page := c.DefaultQuery("page", "1")
```

### 7. 配置管理规范

```go
// ✅ 正确: 使用全局配置
port := config.CONFIG.Port
thumbnailSize := config.GetDefaultThumbnailSize()

// ✅ 正确: 环境判断
if config.IsDevelopment() {
    // 开发环境逻辑
} else {
    // 生产环境逻辑
}

// ❌ 错误: 硬编码配置
port := "8080"
```

### 8. 命名规范

```go
// ✅ 正确: 清晰的命名
func (h *PhotoHandler) GetPhotoByHash(c *gin.Context) {}
func calculateFileHash(filePath string) (string, error) {}

// ✅ 正确: 常量使用全大写
const (
    StatusPending = "pending"
    StatusRunning = "running"
)

// ✅ 正确: 接口使用 -er 后缀
type Repository interface {
    Create(entity interface{}) error
    Update(entity interface{}) error
}

// ❌ 错误: 模糊的命名
func handle(c *gin.Context) {}
func process(data string) error {}
```

---

## 🔍 常见开发场景

### 场景 1: 添加新的 API 接口

**步骤**:
1. 在 `internal/handler/` 创建或扩展 Handler
2. 在 `internal/router/router.go` 注册路由
3. 如果需要数据库操作，在 `internal/repositories/` 添加方法
4. 使用统一的响应格式 `model.Response`

**示例**: 添加获取照片详情接口
```go
// 1. 在 PhotoHandler 添加方法
func (h *PhotoHandler) GetPhotoDetail(c *gin.Context) {
    hash := c.Param("hash")

    photo, err := h.photoRepo.GetByHash(hash)
    if err != nil {
        c.JSON(http.StatusNotFound, model.Response{
            Code:    http.StatusNotFound,
            Message: "照片不存在",
        })
        return
    }

    c.JSON(http.StatusOK, model.Response{
        Code:    http.StatusOK,
        Message: "Success",
        Data:    photo,
    })
}

// 2. 在 router.go 注册
photo.GET("/:hash/detail", photoHandler.GetPhotoDetail)
```

### 场景 2: 添加新的工具函数

**步骤**:
1. 确定工具类型: 通用工具 (`pkg/utils/`) 还是内部工具 (`internal/utils/`)
2. 选择合适的文件或创建新文件
3. 遵循现有命名规范
4. 添加完整的注释和错误处理

**示例**: 添加图片尺寸检测
```go
// pkg/utils/image_utils.go

type imageUtilsStruct struct{}

var ImageUtils = imageUtilsStruct{}

// GetImageDimensions 获取图片尺寸 (不解码整个图像)
func (imageUtilsStruct) GetImageDimensions(filePath string) (width, height int, err error) {
    file, err := os.Open(filePath)
    if err != nil {
        return 0, 0, fmt.Errorf("打开文件失败: %w", err)
    }
    defer file.Close()

    config, _, err := image.DecodeConfig(file)
    if err != nil {
        return 0, 0, fmt.Errorf("解码图片配置失败: %w", err)
    }

    return config.Width, config.Height, nil
}
```

**使用**:
```go
width, height, err := utils.ImageUtils.GetImageDimensions(filePath)
if err != nil {
    logger.Errorf("获取图片尺寸失败: %v", err)
}
```

### 场景 3: 添加新的数据表

**步骤**:
1. 在 `internal/model/tables/` 创建表结构
2. 在 `internal/db/migration.go` 添加迁移
3. 在 `internal/repositories/` 创建 Repository
4. 在 `internal/container/container.go` 注册 Repository

**示例**: 添加标签表
```go
// 1. internal/model/tables/tag_table.go
type Tag struct {
    BaseModel
    Name   string `gorm:"uniqueIndex;size:50" json:"name"`
    Color  string `gorm:"size:20" json:"color"`
    Count  int    `json:"count"`
}

func (Tag) TableName() string {
    return "tags"
}

// 2. internal/db/migration.go
func AutoMigrate() error {
    return GetDB().AutoMigrate(
        // ... 现有表
        &tables.Tag{},
    )
}

// 3. internal/repositories/tag_repository.go
type TagRepository struct {
    BaseRepository
}

func NewTagRepository() *TagRepository {
    return &TagRepository{
        BaseRepository: BaseRepository{tableName: "tags"},
    }
}

func (r *TagRepository) GetAllTags() ([]tables.Tag, error) {
    var tags []tables.Tag
    err := GetDB().Find(&tags).Error
    return tags, err
}

// 4. internal/container/container.go
type DbContainer struct {
    // ... 现有 Repo
    TagRepo *repositories.TagRepository
}

func NewContainer() *DbContainer {
    return &DbContainer{
        // ... 现有 Repo
        TagRepo: repositories.NewTagRepository(),
    }
}
```

### 场景 4: 添加新的外部工具支持

**步骤**:
1. 在 `internal/utils/tools/` 创建工具封装
2. 在 `tool_manager.go` 注册工具
3. 下载工具到 `tools/{platform}/toolname/`

**示例**: 添加 FFmpeg 支持
```go
// internal/utils/tools/ffmpeg_tool.go
type FFmpegTool struct {
    execPath string
}

func NewFFmpegTool(execPath string) *FFmpegTool {
    return &FFmpegTool{execPath: execPath}
}

func (t *FFmpegTool) ExtractVideoThumbnail(videoPath, outputPath string) error {
    cmd := exec.Command(t.execPath,
        "-i", videoPath,
        "-ss", "00:00:01",
        "-vframes", "1",
        outputPath)

    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("FFmpeg 执行失败: %w, 输出: %s", err, output)
    }

    return nil
}

// internal/utils/tool_manager.go
func InitializeFromConfig(appDir *string) error {
    // ... 现有工具初始化

    // FFmpeg
    ffmpegPath := findToolPath("ffmpeg", "ffmpeg.exe", appDir)
    FFmpeg = NewFFmpegTool(ffmpegPath)

    return nil
}
```

---

## 📚 API 接口速查

### 基础接口
- `GET /` - 基础响应
- `GET /health` - 健康检查

### API v1 接口

#### 用户管理 (`/api/v1/users`)
- `GET /api/v1/users` - 获取用户列表
- `GET /api/v1/users/:id` - 获取特定用户
- `POST /api/v1/users` - 创建用户
- `PUT /api/v1/users/:id` - 更新用户
- `DELETE /api/v1/users/:id` - 删除用户

#### 图片库管理 (`/api/v1/library`)
- `GET /api/v1/library` - 获取图片库列表
- `POST /api/v1/library` - 添加图片库路径
  - Body: `{"path": "/path/to/library"}`
- `PUT /api/v1/library` - 更新图片库
  - Body: `{"path": "/path", "is_enable": true}`
- `DELETE /api/v1/library` - 删除图片库
  - Query: `?path=/path/to/library`
- `POST /api/v1/library/indexed` - 启动图片索引任务
  - Body: `{"path": "/path/to/library"}`

#### EXIF 信息管理 (`/api/v1/exif`)
- `GET /api/v1/exif` - 获取所有 EXIF 记录
  - Query: `?page=1&page_size=20`
- `GET /api/v1/exif/:hash` - 根据哈希获取 EXIF
- `GET /api/v1/exif/statistics` - 获取统计信息
- `GET /api/v1/exif/cameras` - 获取所有相机型号列表
- `GET /api/v1/exif/cameras/stats` - 获取相机使用统计
- `GET /api/v1/exif/search` - 搜索 EXIF
  - Query: `?keyword=keyword`
- `GET /api/v1/exif/gps` - 获取包含 GPS 的照片
- `GET /api/v1/exif/iso` - 根据 ISO 筛选
  - Query: `?min=100&max=3200`
- `GET /api/v1/exif/aperture` - 根据光圈筛选
  - Query: `?min=1.4&max=5.6`
- `GET /api/v1/exif/camera` - 根据相机筛选
  - Query: `?model=Canon EOS R5`

#### 照片管理 (`/api/v1/photo`, `/api/v1/photos`)
- `GET /api/v1/photo/:hash` - 获取图像文件
  - Query: `?size=720` (缩略图尺寸)
- `GET /api/v1/photos/timeline` - 获取时间线统计
- `GET /api/v1/photos` - 获取照片列表
  - Query: `?page=1&page_size=20&start_date=2024-01-01&end_date=2024-12-31`

#### 资产信息 (`/api/v1/assets`)
- `GET /api/v1/assets/:hash` - 获取图像详细信息

#### 文件系统 (`/api/v1/filesystem`)
- `GET /api/v1/filesystem/browse` - 浏览文件系统
  - Query: `?path=/path/to/dir`
- `GET /api/v1/filesystem/disk-usage` - 获取磁盘使用情况
  - Query: `?path=/path`
- `GET /api/v1/filesystem/item` - 获取文件系统项目信息
  - Query: `?path=/path/to/item`
- `GET /api/v1/filesystem/search` - 搜索文件
  - Query: `?path=/path&keyword=keyword`
- `POST /api/v1/filesystem/directory` - 创建目录
  - Body: `{"path": "/path/to/new/dir"}`
- `DELETE /api/v1/filesystem/item` - 删除文件或目录
  - Query: `?path=/path/to/item`
- `PUT /api/v1/filesystem/item/move` - 移动/重命名
  - Body: `{"src": "/path/src", "dst": "/path/dst"}`
- `POST /api/v1/filesystem/item/copy` - 复制文件或目录
  - Body: `{"src": "/path/src", "dst": "/path/dst"}`

#### 开发接口 (`/dev`)
- `GET /dev/tool/exiftool/get_exif` - 获取图片 EXIF 信息
  - Query: `?path=/path/to/image.jpg`

---

## 🚀 性能优化策略

### 1. 图像处理优化
- **并发处理**: 根据 CPU 核心数自动调整并发数
- **流式处理**: 使用 LibVips 的流式处理，避免大文件占用内存
- **缓存策略**: 缩略图按哈希存储，避免重复生成
- **智能跳过**: 检查哈希值，已存在的照片跳过处理

### 2. 数据库优化
- **索引**: 常用查询字段建立索引 (Hash, Model, ISO, GPS, DateTime)
- **写队列**: 所有写操作通过队列，避免锁竞争
- **批量操作**: 使用批量插入/更新
- **连接池**: 合理配置数据库连接池

### 3. 内存管理
- **缓冲区**: 根据文件大小动态调整缓冲区 (64KB - 1MB)
- **哈希计算**: 分块读取，避免一次加载整个文件
- **垃圾回收**: 定期清理临时文件和过期缓存

### 4. 系统监控
- **动态调整**: 监控 CPU/内存使用率，自动调整并发数
- **任务暂停**: 系统资源不足时自动暂停任务
- **优雅降级**: 外部工具失败时使用备用方案

---

## 🔒 安全特性

### 1. 文件安全
- **路径验证**: 防止路径遍历攻击 (`.`, `..`, `~`)
- **文件类型检查**: 使用 `h2non/filetype` 验证真实文件类型
- **哈希验证**: 文件完整性检查

### 2. 网络安全
- **CORS 配置**: 可配置允许的来源
- **请求限制**: 防止恶意请求 (未来扩展)
- **日志审计**: 完整的访问日志

### 3. 数据安全
- **SQL 注入防护**: 使用 GORM 参数化查询
- **XSS 防护**: 输出数据转义
- **敏感信息**: 不在日志中输出敏感路径

---

## 🐛 调试技巧

### 1. 日志级别调整
```yaml
# config.yaml
logging:
  level: "debug"  # 开发: debug, 生产: info
```

### 2. 性能分析
```go
// 启用 pprof (debug 模式)
// 访问 http://localhost:9484/debug/pprof/
```

### 3. SSE 调试
```bash
# 使用 curl 测试 SSE
curl -N http://localhost:9484/api/v1/sse
```

### 4. 数据库调试
```go
// 在 GORM 查询前添加
db.Debug().Find(&photos)
```

---

## 📝 未来扩展方向

### 计划中的功能
1. **AI 智能分析**: 图像标签、人脸识别、场景分类
2. **视频支持**: FFmpeg 集成，视频缩略图生成
3. **相似图片查找**: 感知哈希 (pHash) 算法
4. **批量编辑**: EXIF 批量修改、批量重命名
5. **分享功能**: 生成分享链接、权限控制
6. **多用户支持**: 用户认证、权限管理
7. **插件系统**: 支持自定义处理流程

### 性能优化计划
1. **缓存层**: Redis 缓存热点数据
2. **CDN 集成**: 缩略图 CDN 加速
3. **分布式处理**: 多机分布式任务处理
4. **WebP 支持**: 更高压缩比的缩略图格式

---

## 📖 参考资源

### Go 语言资源
- [Go 官方文档](https://golang.org/doc/)
- [Effective Go](https://golang.org/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)

### 框架文档
- [Gin 文档](https://gin-gonic.com/docs/)
- [GORM 文档](https://gorm.io/docs/)
- [Zap 日志库](https://github.com/uber-go/zap)

### 外部工具文档
- [ExifTool](https://exiftool.org/)
- [ImageMagick](https://imagemagick.org/)
- [LibVips](https://www.libvips.org/)

---

## ✅ 开发检查清单

### 添加新功能前
- [ ] 检查是否有现成工具函数可用
- [ ] 确认设计符合现有架构
- [ ] 考虑错误处理和日志记录
- [ ] 评估性能影响

### 编写代码时
- [ ] 遵循命名规范
- [ ] 添加完整的错误处理
- [ ] 使用结构化日志
- [ ] 考虑并发安全
- [ ] 添加必要的注释

### 提交代码前
- [ ] 代码格式化 (`go fmt`)
- [ ] 运行测试 (`go test`)
- [ ] 检查日志输出
- [ ] 验证 API 接口
- [ ] 更新本文档 (如需)

---

## 🎯 快速定位代码

### 我想找...
- **添加 API 路由**: `internal/router/router.go`
- **修改配置**: `config.yaml` + `internal/config/`
- **数据库表结构**: `internal/model/tables/`
- **业务逻辑**: `internal/handler/`
- **数据库操作**: `internal/repositories/`
- **工具函数**: `pkg/utils/` 或 `internal/utils/`
- **日志配置**: `pkg/logger/logger.go`
- **任务处理**: `internal/workflow/img_task.go`
- **外部工具调用**: `internal/utils/tools/`

---

## 📞 开发规范总结

### 核心原则
1. **DRY (Don't Repeat Yourself)**: 不重复造轮子
2. **KISS (Keep It Simple, Stupid)**: 保持简单
3. **单一职责**: 每个函数/模块只做一件事
4. **错误处理**: 永远不要忽略错误
5. **日志记录**: 关键操作都要记录日志
6. **并发安全**: 共享数据必须加锁
7. **测试**: 关键功能编写单元测试

### 禁止事项
- ❌ 不要直接使用 `os` 包操作文件 (使用 `FileUtils`)
- ❌ 不要使用 `fmt.Println` 输出日志 (使用 `logger`)
- ❌ 不要忽略错误返回值
- ❌ 不要硬编码配置 (使用 `config.CONFIG`)
- ❌ 不要直接使用 GORM (使用 Repository)
- ❌ 不要在多处实现相同功能

### 推荐做法
- ✅ 使用已有工具函数
- ✅ 结构化日志
- ✅ 完整的错误处理
- ✅ 清晰的命名
- ✅ 合理的注释
- ✅ 并发安全
- ✅ 单元测试

---

**本文档由 Claude AI 维护，最后更新: 2025-10-03**
