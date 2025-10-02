# Argus 照片管理系统 HTTP API 规范文档

## 📋 文档信息

- **版本**: v1.0
- **创建日期**: 2025-10-02
- **API 基础路径**: `http://localhost:8726/api/v1`
- **协议**: HTTP/1.1, SSE (Server-Sent Events)
- **数据格式**: JSON
- **字符编码**: UTF-8

---

## 📑 目录

1. [通用规范](#通用规范)
2. [文件系统 API](#文件系统-api)
3. [照片管理 API](#照片管理-api)
4. [资料库管理 API](#资料库管理-api)
5. [实时推送 SSE](#实时推送-sse)
6. [错误处理](#错误处理)
7. [前端集成示例](#前端集成示例)

---

## 通用规范

### 请求头

```http
Content-Type: application/json
Accept: application/json
Authorization: Bearer <token>  # 如果需要认证
```

### 响应格式

所有API返回统一的响应格式：

```typescript
interface Response<T> {
  code: number        // HTTP状态码
  message: string     // 响应消息
  data?: T           // 响应数据（可选）
}
```

### HTTP 状态码

| 状态码 | 说明 |
|--------|------|
| 200 | 成功 |
| 201 | 创建成功 |
| 204 | 成功但无内容返回 |
| 400 | 请求参数错误 |
| 401 | 未授权 |
| 403 | 禁止访问 |
| 404 | 资源不存在 |
| 409 | 资源冲突 |
| 500 | 服务器内部错误 |

---

## 文件系统 API

### 1. 浏览文件系统

**功能**: 获取指定路径的文件和文件夹列表

#### 请求

```http
GET /api/v1/filesystem/browse?path={path}
```

**查询参数**:

| 参数 | 类型 | 必需 | 说明 |
|------|------|------|------|
| path | string | 否 | 文件路径，空或"/"返回驱动器列表 |
| sort | string | 否 | 排序字段: name, size, date, type |
| order | string | 否 | 排序顺序: asc, desc |
| filter | string | 否 | 文件类型过滤: all, photo, video, folder |

#### 响应

```typescript
interface FileSystemResponse {
  current_path: string      // 当前路径
  parent_path: string       // 父路径
  items: FileSystemItem[]   // 文件/文件夹列表
  summary: PathSummary      // 路径摘要信息
}

interface FileSystemItem {
  id: string                // 唯一标识
  name: string              // 名称
  path: string              // 完整路径
  type: "drive" | "directory" | "file"  // 类型
  size: number              // 大小（字节）
  mod_time: string          // 修改时间 (ISO 8601)
  is_accessible: boolean    // 是否可访问

  // 驱动器特有属性
  drive_info?: {
    label: string           // 卷标
    file_system: string     // 文件系统类型
    total_space: number     // 总空间
    free_space: number      // 可用空间
    used_space: number      // 已用空间
    usage_percent: number   // 使用百分比
    is_removable: boolean   // 是否可移动
    drive_type: string      // 驱动器类型
  }

  // 文件特有属性
  file_info?: {
    extension: string       // 文件扩展名
    mime_type: string       // MIME类型
    is_photo: boolean       // 是否为照片
    is_video: boolean       // 是否为视频
    thumbnail_path?: string // 缩略图路径（如果有）
  }

  // 目录特有属性
  directory_info?: {
    item_count: number      // 子项目数量
    has_photos: boolean     // 是否包含照片
    has_videos: boolean     // 是否包含视频
  }
}

interface PathSummary {
  total_items: number       // 总项目数
  directory_count: number   // 目录数量
  file_count: number        // 文件数量
  drive_count: number       // 驱动器数量
  photo_count: number       // 照片数量
  video_count: number       // 视频数量
}
```

#### 示例

**请求**:
```http
GET /api/v1/filesystem/browse?path=D:\Photos\2024
```

**响应**:
```json
{
  "code": 200,
  "message": "获取文件系统信息成功",
  "data": {
    "current_path": "D:\\Photos\\2024",
    "parent_path": "D:\\Photos",
    "items": [
      {
        "id": "dir_D_Photos_2024_January",
        "name": "January",
        "path": "D:\\Photos\\2024\\January",
        "type": "directory",
        "size": 0,
        "mod_time": "2024-01-15T10:30:00Z",
        "is_accessible": true,
        "directory_info": {
          "item_count": 156,
          "has_photos": true,
          "has_videos": false
        }
      },
      {
        "id": "file_D_Photos_2024_sunset.jpg",
        "name": "sunset.jpg",
        "path": "D:\\Photos\\2024\\sunset.jpg",
        "type": "file",
        "size": 2458624,
        "mod_time": "2024-02-20T15:45:30Z",
        "is_accessible": true,
        "file_info": {
          "extension": ".jpg",
          "mime_type": "image/jpeg",
          "is_photo": true,
          "is_video": false,
          "thumbnail_path": "/thumbnails/sunset_thumb.jpg"
        }
      }
    ],
    "summary": {
      "total_items": 15,
      "directory_count": 5,
      "file_count": 10,
      "drive_count": 0,
      "photo_count": 8,
      "video_count": 2
    }
  }
}
```

---

### 2. 创建文件夹

#### 请求

```http
POST /api/v1/filesystem/directory
Content-Type: application/json

{
  "path": "D:\\Photos\\2025\\NewFolder"
}
```

#### 响应

```json
{
  "code": 200,
  "message": "创建目录成功",
  "data": {
    "path": "D:\\Photos\\2025\\NewFolder",
    "created_at": "2025-10-02T10:30:00Z"
  }
}
```

---

### 3. 删除文件/文件夹

#### 请求

```http
DELETE /api/v1/filesystem/item?path={path}&operation_id={uuid}
```

**查询参数**:

| 参数 | 类型 | 必需 | 说明 |
|------|------|------|------|
| path | string | 是 | 要删除的文件/文件夹路径 |
| operation_id | string | 否 | 操作ID（用于SSE事件去重） |
| recursive | boolean | 否 | 是否递归删除（默认true） |

#### 响应

```json
{
  "code": 200,
  "message": "删除成功",
  "data": {
    "path": "D:\\Photos\\old_file.jpg",
    "deleted_at": "2025-10-02T10:35:00Z",
    "operation_id": "uuid-1234-5678"
  }
}
```

---

### 4. 移动/重命名

#### 请求

```http
PUT /api/v1/filesystem/item/move
Content-Type: application/json

{
  "source": "D:\\Photos\\old_name.jpg",
  "destination": "D:\\Photos\\2025\\new_name.jpg",
  "operation_id": "uuid-1234-5678",
  "overwrite": false
}
```

**请求体参数**:

| 参数 | 类型 | 必需 | 说明 |
|------|------|------|------|
| source | string | 是 | 源路径 |
| destination | string | 是 | 目标路径 |
| operation_id | string | 否 | 操作ID |
| overwrite | boolean | 否 | 目标存在时是否覆盖 |

#### 响应

```json
{
  "code": 200,
  "message": "移动成功",
  "data": {
    "source": "D:\\Photos\\old_name.jpg",
    "destination": "D:\\Photos\\2025\\new_name.jpg",
    "operation_id": "uuid-1234-5678"
  }
}
```

---

### 5. 复制文件/文件夹

#### 请求

```http
POST /api/v1/filesystem/item/copy
Content-Type: application/json

{
  "source": "D:\\Photos\\photo.jpg",
  "destination": "D:\\Backup\\photo.jpg",
  "operation_id": "uuid-1234-5678",
  "overwrite": false
}
```

#### 响应

```json
{
  "code": 200,
  "message": "复制成功",
  "data": {
    "source": "D:\\Photos\\photo.jpg",
    "destination": "D:\\Backup\\photo.jpg",
    "size": 2458624,
    "operation_id": "uuid-1234-5678"
  }
}
```

---

### 6. 搜索文件

#### 请求

```http
GET /api/v1/filesystem/search?path={path}&pattern={pattern}&recursive={true}
```

**查询参数**:

| 参数 | 类型 | 必需 | 说明 |
|------|------|------|------|
| path | string | 是 | 搜索路径 |
| pattern | string | 否 | 文件名模式（支持通配符） |
| recursive | boolean | 否 | 是否递归搜索 |
| type | string | 否 | 文件类型过滤: photo, video, all |

#### 响应

```json
{
  "code": 200,
  "message": "搜索完成",
  "data": {
    "results": [
      {
        "name": "sunset.jpg",
        "path": "D:\\Photos\\2024\\sunset.jpg",
        "type": "file",
        "size": 2458624,
        "mod_time": "2024-02-20T15:45:30Z"
      }
    ],
    "total_count": 1,
    "search_duration_ms": 150
  }
}
```

---

### 7. 获取磁盘使用情况

#### 请求

```http
GET /api/v1/filesystem/disk-usage?path={path}
```

#### 响应

```json
{
  "code": 200,
  "message": "获取磁盘使用情况成功",
  "data": {
    "label": "Windows",
    "file_system": "NTFS",
    "total_space": 512110190592,
    "free_space": 123456789012,
    "used_space": 388653401580,
    "usage_percent": 75.89,
    "is_removable": false,
    "drive_type": "Fixed"
  }
}
```

---

## 照片管理 API

### 8. 获取照片列表（时间线）

#### 请求

```http
GET /api/v1/photos?start={start}&end={end}&limit={limit}&offset={offset}
```

**查询参数**:

| 参数 | 类型 | 必需 | 说明 |
|------|------|------|------|
| start | string | 否 | 开始时间 (ISO 8601) |
| end | string | 否 | 结束时间 (ISO 8601) |
| limit | number | 否 | 每页数量（默认50） |
| offset | number | 否 | 偏移量（默认0） |
| sort | string | 否 | 排序字段: date, name, size |
| order | string | 否 | 排序顺序: asc, desc |

#### 响应

```json
{
  "code": 200,
  "message": "获取照片列表成功",
  "data": {
    "photos": [
      {
        "hash": "abc123def456",
        "path": "D:\\Photos\\sunset.jpg",
        "file_name": "sunset.jpg",
        "file_size": 2458624,
        "width": 4032,
        "height": 3024,
        "taken_at": "2024-02-20T15:45:30Z",
        "created_at": "2024-02-20T15:50:00Z",
        "thumbnail_url": "/api/v1/photo/abc123def456?size=thumb",
        "preview_url": "/api/v1/photo/abc123def456?size=preview"
      }
    ],
    "total": 1580,
    "limit": 50,
    "offset": 0,
    "has_more": true
  }
}
```

---

### 9. 获取照片时间线统计

#### 请求

```http
GET /api/v1/photos/timeline?granularity={month}
```

**查询参数**:

| 参数 | 类型 | 必需 | 说明 |
|------|------|------|------|
| granularity | string | 否 | 时间粒度: year, month, day |

#### 响应

```json
{
  "code": 200,
  "message": "获取时间线统计成功",
  "data": {
    "timeline": [
      {
        "period": "2024-02",
        "count": 156,
        "start_date": "2024-02-01T00:00:00Z",
        "end_date": "2024-02-29T23:59:59Z"
      },
      {
        "period": "2024-01",
        "count": 203,
        "start_date": "2024-01-01T00:00:00Z",
        "end_date": "2024-01-31T23:59:59Z"
      }
    ],
    "total_photos": 1580,
    "earliest_date": "2020-01-01T00:00:00Z",
    "latest_date": "2024-12-31T23:59:59Z"
  }
}
```

---

### 10. 获取单张照片详情

#### 请求

```http
GET /api/v1/assets/{hash}
```

#### 响应

```json
{
  "code": 200,
  "message": "获取照片详情成功",
  "data": {
    "hash": "abc123def456",
    "path": "D:\\Photos\\sunset.jpg",
    "file_name": "sunset.jpg",
    "file_size": 2458624,
    "width": 4032,
    "height": 3024,
    "taken_at": "2024-02-20T15:45:30Z",
    "created_at": "2024-02-20T15:50:00Z",
    "exif": {
      "camera_make": "Canon",
      "camera_model": "EOS R5",
      "lens": "RF 24-105mm F4 L IS USM",
      "iso": 400,
      "aperture": 5.6,
      "shutter_speed": "1/250",
      "focal_length": 50,
      "gps": {
        "latitude": 39.9042,
        "longitude": 116.4074,
        "altitude": 50.5
      }
    },
    "thumbnails": {
      "small": "/api/v1/photo/abc123def456?size=small",
      "medium": "/api/v1/photo/abc123def456?size=medium",
      "large": "/api/v1/photo/abc123def456?size=large"
    }
  }
}
```

---

### 11. 获取照片文件

#### 请求

```http
GET /api/v1/photo/{hash}?size={size}
```

**查询参数**:

| 参数 | 类型 | 必需 | 说明 |
|------|------|------|------|
| size | string | 否 | 尺寸: thumb, small, medium, large, original |

#### 响应

返回图片二进制数据，Content-Type为图片MIME类型

---

## 资料库管理 API

### 12. 获取资料库列表

#### 请求

```http
GET /api/v1/library
```

#### 响应

```json
{
  "code": 200,
  "message": "获取资料库列表成功",
  "data": {
    "libraries": [
      {
        "id": "lib-001",
        "name": "主照片库",
        "path": "D:\\Photos",
        "photo_count": 1580,
        "total_size": 15728640000,
        "last_indexed": "2025-10-02T08:00:00Z",
        "is_active": true,
        "watch_enabled": true
      }
    ],
    "total": 1
  }
}
```

---

### 13. 添加资料库

#### 请求

```http
POST /api/v1/library
Content-Type: application/json

{
  "name": "新照片库",
  "path": "E:\\MyPhotos",
  "watch_enabled": true,
  "auto_index": true
}
```

#### 响应

```json
{
  "code": 201,
  "message": "添加资料库成功",
  "data": {
    "id": "lib-002",
    "name": "新照片库",
    "path": "E:\\MyPhotos",
    "created_at": "2025-10-02T10:30:00Z"
  }
}
```

---

### 14. 开始索引资料库

#### 请求

```http
POST /api/v1/library/indexed
Content-Type: application/json

{
  "library_id": "lib-001",
  "force": false
}
```

**请求体参数**:

| 参数 | 类型 | 必需 | 说明 |
|------|------|------|------|
| library_id | string | 是 | 资料库ID |
| force | boolean | 否 | 是否强制重新索引 |

#### 响应

```json
{
  "code": 200,
  "message": "索引任务已启动",
  "data": {
    "task_id": "task-001",
    "library_id": "lib-001",
    "started_at": "2025-10-02T10:35:00Z",
    "estimated_duration_seconds": 300
  }
}
```

---

## 实时推送 SSE

### 15. 建立SSE连接

#### 请求

```http
GET /api/sse
Accept: text/event-stream
```

#### 响应

持续的事件流：

```
event: ping
data: {"time":"2025-10-02T10:30:00Z"}

event: file-change
data: {"type":"created","path":"D:\\Photos\\2025","affected_count":1,"timestamp":"2025-10-02T10:30:05Z"}

event: file-change
data: {"type":"deleted","path":"D:\\Photos","affected_count":3,"operation_id":"uuid-1234","timestamp":"2025-10-02T10:30:10Z"}

event: indexing-progress
data: {"library_id":"lib-001","progress":45,"total":1000,"current_file":"sunset.jpg"}

event: indexing-complete
data: {"library_id":"lib-001","total_indexed":1000,"duration_seconds":120}
```

### SSE 事件类型

#### ping 事件
保持连接活跃

```json
{
  "time": "2025-10-02T10:30:00Z"
}
```

#### file-change 事件
文件系统变化通知

```json
{
  "type": "created" | "modified" | "deleted" | "renamed",
  "path": "D:\\Photos\\2025",
  "affected_count": 3,
  "operation_id": "uuid-1234-5678",  // 可选，用于去重
  "timestamp": "2025-10-02T10:30:00Z",
  "details": {
    "files": ["photo1.jpg", "photo2.jpg"]  // 可选
  }
}
```

#### indexing-progress 事件
索引进度更新

```json
{
  "library_id": "lib-001",
  "progress": 45,
  "total": 1000,
  "current_file": "sunset.jpg",
  "speed": 5.2,  // 文件/秒
  "eta_seconds": 180
}
```

#### indexing-complete 事件
索引完成

```json
{
  "library_id": "lib-001",
  "total_indexed": 1000,
  "new_photos": 50,
  "updated_photos": 10,
  "duration_seconds": 120
}
```

#### error 事件
错误通知

```json
{
  "error_type": "permission_denied" | "disk_full" | "network_error",
  "message": "无法访问文件夹",
  "path": "D:\\Photos\\Private",
  "timestamp": "2025-10-02T10:30:00Z"
}
```

---

## 错误处理

### 错误响应格式

```json
{
  "code": 400,
  "message": "请求参数错误",
  "data": {
    "error_type": "validation_error",
    "details": {
      "path": "路径不能为空"
    }
  }
}
```

### 常见错误类型

| error_type | 说明 |
|------------|------|
| validation_error | 参数验证错误 |
| not_found | 资源不存在 |
| permission_denied | 权限不足 |
| conflict | 资源冲突（如文件已存在） |
| disk_full | 磁盘空间不足 |
| network_error | 网络错误 |
| timeout | 请求超时 |
| internal_error | 服务器内部错误 |

---

## 前端集成示例

### 1. 浏览文件系统

```typescript
import { httpClient } from '@/utils/http'

interface FileSystemResponse {
  current_path: string
  parent_path: string
  items: FileSystemItem[]
  summary: PathSummary
}

async function browseFileSystem(path: string) {
  const response = await httpClient.get<FileSystemResponse>(
    '/filesystem/browse',
    { params: { path } }
  )
  return response.data
}

// 使用
const result = await browseFileSystem('D:\\Photos\\2024')
console.log('文件列表:', result.items)
console.log('照片数量:', result.summary.photo_count)
```

---

### 2. 文件操作

```typescript
// 删除文件
async function deleteFile(path: string) {
  const operationId = crypto.randomUUID()

  await httpClient.delete('/filesystem/item', {
    params: { path, operation_id: operationId }
  })

  return operationId
}

// 移动文件
async function moveFile(source: string, destination: string) {
  const operationId = crypto.randomUUID()

  await httpClient.put('/filesystem/item/move', {
    source,
    destination,
    operation_id: operationId,
    overwrite: false
  })

  return operationId
}

// 创建文件夹
async function createFolder(path: string) {
  await httpClient.post('/filesystem/directory', { path })
}
```

---

### 3. SSE 实时监听

```typescript
class FileSystemWatcher {
  private eventSource: EventSource | null = null
  private operationIds = new Set<string>()

  connect() {
    this.eventSource = new EventSource('http://localhost:8726/api/sse')

    // 监听文件变化
    this.eventSource.addEventListener('file-change', (e) => {
      const event = JSON.parse(e.data)

      // 忽略自己触发的操作
      if (event.operation_id && this.operationIds.has(event.operation_id)) {
        this.operationIds.delete(event.operation_id)
        return
      }

      // 刷新受影响的目录
      this.refreshDirectory(event.path)
    })

    // 监听索引进度
    this.eventSource.addEventListener('indexing-progress', (e) => {
      const event = JSON.parse(e.data)
      this.updateIndexingProgress(event)
    })

    // 错误处理
    this.eventSource.addEventListener('error', (e) => {
      console.error('SSE连接错误', e)
      this.reconnect()
    })
  }

  disconnect() {
    this.eventSource?.close()
  }

  // 记录自己的操作ID（用于去重）
  registerOperation(operationId: string) {
    this.operationIds.add(operationId)
  }

  private refreshDirectory(path: string) {
    // 刷新对应目录的文件列表
  }

  private reconnect() {
    setTimeout(() => this.connect(), 5000)
  }
}

// 使用
const watcher = new FileSystemWatcher()
watcher.connect()

// 执行操作时注册ID
const opId = await deleteFile('D:\\Photos\\old.jpg')
watcher.registerOperation(opId)
```

---

### 4. 完整的FileManager集成

```typescript
class FileManagerAPI {
  private watcher: FileSystemWatcher

  constructor() {
    this.watcher = new FileSystemWatcher()
    this.watcher.connect()
  }

  // 浏览目录
  async browse(path: string) {
    return await browseFileSystem(path)
  }

  // 删除（带乐观更新）
  async delete(path: string, onOptimisticUpdate: () => void) {
    // 1. 立即更新UI
    onOptimisticUpdate()

    // 2. 调用API
    const opId = await deleteFile(path)
    this.watcher.registerOperation(opId)
  }

  // 移动（带乐观更新）
  async move(source: string, dest: string, onOptimisticUpdate: () => void) {
    onOptimisticUpdate()
    const opId = await moveFile(source, dest)
    this.watcher.registerOperation(opId)
  }

  // 创建文件夹
  async createFolder(path: string) {
    await createFolder(path)
  }

  destroy() {
    this.watcher.disconnect()
  }
}
```

---

## 📝 待实现功能清单

### 高优先级
- [ ] 文件操作API的完整实现（删除、移动、复制、创建）
- [ ] 文件系统监听与SSE推送集成
- [ ] 照片缩略图生成和缓存
- [ ] 批量操作支持

### 中优先级
- [ ] 搜索结果分页优化
- [ ] EXIF信息提取和缓存
- [ ] 文件操作撤销/恢复
- [ ] 文件版本控制

### 低优先级
- [ ] 文件标签和收藏功能
- [ ] 智能相册分类
- [ ] 人脸识别集成
- [ ] 照片编辑API

---

## 📅 版本历史

| 版本 | 日期 | 变更内容 |
|------|------|---------|
| v1.0 | 2025-10-02 | 初始版本，定义核心API规范 |

---

**文档维护者**: 开发团队
**最后更新**: 2025-10-02
