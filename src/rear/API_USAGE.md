# Photo API 使用指南

## 📸 图像接口

### 获取图像文件
```
GET /api/v1/photo/{hash}
```

#### 查询参数
- `format`: 图像格式 (可选)
  - `original`: 原图 (默认)
  - `thumbnail`: 缩略图
- `size`: 缩略图尺寸 (可选，仅在format=thumbnail时有效)
  - 格式: `WIDTHxHEIGHT` (例如: `400x400`)
  - 如果不提供，使用配置文件中的 `DefaultThumbnailSize` (默认720)

#### 使用示例
```bash
# 获取原图
curl "http://localhost:8080/api/v1/photo/abc123?format=original"

# 获取默认尺寸缩略图
curl "http://localhost:8080/api/v1/photo/abc123?format=thumbnail"

# 获取指定尺寸缩略图 (最大400x400，按比例缩放)
curl "http://localhost:8080/api/v1/photo/abc123?format=thumbnail&size=400x400"

# 获取自定义尺寸缩略图
curl "http://localhost:8080/api/v1/photo/abc123?format=thumbnail&size=800x600"
```

#### 响应
- 成功: 返回图像文件（二进制数据）
- 失败: 返回JSON错误信息

## 📊 图像详情接口

### 获取图像详细信息
```
GET /api/v1/assets/{hash}
```

#### 使用示例
```bash
curl "http://localhost:8080/api/v1/assets/abc123"
```

#### 响应示例
```json
{
  "code": 200,
  "message": "Success",
  "data": {
    "hash": "abc123",
    "imgPath": "/path/to/image.jpg",
    "imgName": "image.jpg",
    "width": 1920,
    "height": 1080,
    "aspectRatio": 1.777,
    "fileSize": 2048576,
    "format": "jpg",
    "notes": null,
    "fileCreatedAt": "2025-01-01T10:00:00Z",
    "takenAt": "2024-12-25T15:30:00Z",
    "lastModified": "2025-01-01T10:00:00Z",
    "rating": 0,
    "lastViewedAt": null,
    "viewCount": 0,
    "createdAt": "2025-01-01T10:00:00Z",
    "updatedAt": "2025-01-01T10:00:00Z",
    "exifInfo": {
      "fileName": "image.jpg",
      "fileSize": 2048576,
      "imageWidth": 1920,
      "imageHeight": 1080,
      "mimeType": "image/jpeg",
      "fileType": "JPEG",
      "make": "Canon",
      "model": "EOS R5",
      "iso": 100,
      "gpsLatitude": 39.9042,
      "gpsLongitude": 116.4074,
      "exposureTime": 0.008,
      "aperture": 2.8,
      "fNumber": 2.8,
      "focalLength": 85.0,
      "dateTimeOriginal": "2024:12:25 15:30:00",
      "otherFields": {}
    }
  }
}
```

## 🎯 缩略图生成逻辑

### 尺寸限制
- `size` 参数指定的是**最大尺寸限制**，不是精确尺寸
- 系统会按原图比例缩放，确保图像不超过指定的宽度和高度
- 例如: `size=400x400` 对于 1920x1080 的图像，会生成 400x225 的缩略图

### 缓存机制
- 缩略图会被缓存到 `{AppDir}/cache/thumbnail/` 目录
- 文件名格式: `{hash}_{width}x{height}.jpg`
- 如果缩略图已存在，直接返回缓存文件

### 默认尺寸
- 如果未指定 `size` 参数，使用配置文件中的 `defaultThumbnailSize`
- 默认值: 720像素

## 🚀 快速开始

1. **启动服务器**
   ```bash
   go run .
   ```

2. **添加图像库**
   ```bash
   curl -X POST "http://localhost:8080/api/v1/library" \
     -H "Content-Type: application/json" \
     -d '{"path": "/path/to/your/photos"}'
   ```

3. **开始索引**
   ```bash
   curl -X POST "http://localhost:8080/api/v1/library/indexed"
   ```

4. **获取图像列表**
   ```bash
   curl "http://localhost:8080/api/v1/library"
   ```

5. **使用图像hash测试接口**
   ```bash
   # 使用实际的hash值
   curl "http://localhost:8080/api/v1/assets/{actual_hash}"
   curl "http://localhost:8080/api/v1/photo/{actual_hash}?format=thumbnail&size=400x400"
   ```

## ⚠️ 注意事项

- hash参数是必需的，必须是有效的图像hash值
- 图像文件必须存在且可访问
- 缩略图生成可能需要一些时间，特别是对于大图像
- API会自动更新图像的访问计数和访问时间
- 所有时间字段都以ISO 8601格式返回

## 📝 错误代码

- `400 Bad Request`: 参数错误（如size格式错误）
- `404 Not Found`: 图像不存在或文件不可访问
- `500 Internal Server Error`: 服务器内部错误（如缩略图生成失败）