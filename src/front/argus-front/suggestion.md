# 后端API建议

## 当前API使用体验

基于对 `timeline.http` 和 `photo.http` API文档的分析，当前API设计整体合理，能够满足照片时间线展示的需求。以下是一些建议改进点：

## 建议改进的API设计

### 1. 照片详情API缺失

**问题**：
- 目前只有获取照片文件的接口（`/api/v1/photo/{hash}`），缺少获取照片详细信息的API
- 前端需要照片的元数据信息（如拍摄地点、设备信息、标签等）用于详情展示

**建议**：
```http
GET /api/v1/photo/{hash}/metadata
```
返回格式：
```json
{
  "code": 200,
  "message": "Success",
  "data": {
    "hash": "abc123...",
    "filename": "IMG_1234.jpg",
    "fileSize": 1024000,
    "takenAt": "2024-10-05T10:30:00",
    "width": 1920,
    "height": 1080,
    "location": {
      "latitude": 39.9042,
      "longitude": 116.4074,
      "address": "北京市"
    },
    "camera": {
      "make": "Apple",
      "model": "iPhone 15",
      "lens": "6mm f/1.6"
    },
    "tags": ["风景", "旅行"],
    "people": ["张三", "李四"]
  }
}
```

### 2. 批量获取照片缩略图优化

**问题**：
- 当前需要为每张照片单独请求缩略图URL
- 可能导致大量HTTP请求，影响性能

**建议**：
考虑添加批量获取缩略图的接口：
```http
POST /api/v1/photos/thumbnails
Content-Type: application/json

{
  "hashes": ["hash1", "hash2", "hash3"],
  "size": "400x400"
}
```
返回zip压缩包或者JSON格式的base64图片数据。

### 3. 时间线API增强

**问题**：
- 当前时间线API只返回日期和数量，前端需要额外请求每天的照片详情
- 可能导致瀑布式请求，影响加载性能

**建议**：
添加一个组合API，一次性返回时间线和对应的照片信息：
```http
GET /api/v1/photos/timeline/full?start_date=2024-01-01&end_date=2024-12-31&photos_per_day=10
```
返回格式：
```json
{
  "code": 200,
  "message": "Success", 
  "data": {
    "2024-10-05": {
      "count": 1202,
      "photos": {
        "hash": ["hash1", "hash2"],
        "ratio": [1.5, 0.8],
        "takenAt": ["2024-10-05T10:30:00", "2024-10-05T11:30:00"]
      }
    }
  }
}
```

### 4. 缓存策略优化

**建议**：
- 为缩略图接口添加适当的缓存头（Cache-Control, ETag等）
- 支持条件请求（If-None-Match）减少重复传输

### 5. 错误处理统一

**观察**：
- 当前API错误处理基本统一，但建议添加更详细的错误码
- 可以考虑添加国际化支持的错误消息

**建议错误码**：
```json
{
  "code": 4001,
  "message": "Photo not found",
  "details": "The requested photo hash does not exist in the database"
}
```

## 前端适配情况

✅ **已适配**：
- 时间线数据获取和展示
- 照片列表获取和分页
- 缩略图URL生成和显示
- 错误处理和加载状态
- 响应式布局重计算

⏳ **待实现**（需要后端支持）：
- 照片详情弹窗（需要metadata API）
- 照片搜索功能
- 照片标签和人物识别展示
- 地理位置信息展示

## 总结

当前API设计已经可以支持基本的照片时间线功能，前端接口对接已完成。建议优先考虑实现照片元数据API，以支持更丰富的用户交互体验。