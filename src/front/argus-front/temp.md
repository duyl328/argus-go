# PhotosView API 接口需求文档

## 概述
PhotosView 是照片时间线页面，展示按月份分组的照片，支持虚拟滚动和时间线导航。以下是需要后端提供的API接口和数据结构。

## 核心数据结构

### Photo (照片对象)
```typescript
interface Photo {
  id: string              // 照片唯一标识符
  url: string             // 照片URL路径 (如: /api/photos/thumb/123.jpg)
  fullUrl?: string        // 原图URL路径 (如: /api/photos/full/123.jpg)
  width: number           // 照片宽度 (像素)
  height: number          // 照片高度 (像素)
  takenAt: string         // 拍摄时间 (ISO 8601格式)
  fileName: string        // 原始文件名
  size?: number           // 文件大小 (字节)
  metadata?: {            // 可选元数据
    camera?: string       // 相机型号
    iso?: number          // ISO值
    aperture?: string     // 光圈值
    shutterSpeed?: string // 快门速度
  }
}
```

### PhotoPeriod (时期分组)
```typescript
interface PhotoPeriod {
  id: string              // 时期标识符 (如 "2024-08")
  period: string          // 时期字符串 "YYYY-MM"
  displayName: string     // 显示名称 "2024年8月"
  photos: Photo[]         // 该时期的照片列表
  photoCount: number      // 照片总数
}
```

## 需要的API接口

### 1. 获取照片时间线数据
**接口路径**: `GET /api/photos/timeline`

**请求参数**: 无

**响应数据**:
```json
{
  "success": true,
  "data": {
    "periods": [
      {
        "id": "2024-08",
        "period": "2024-08", 
        "displayName": "2024年8月",
        "photoCount": 342,
        "photos": [
          {
            "id": "photo_001",
            "url": "/api/photos/thumb/photo_001.jpg",
            "fullUrl": "/api/photos/full/photo_001.jpg",
            "width": 4032,
            "height": 3024,
            "takenAt": "2024-08-15T14:30:25Z",
            "fileName": "IMG_001.jpg",
            "size": 2458624
          }
        ]
      }
    ],
    "totalPhotos": 10847
  }
}
```

### 2. 获取照片缩略图
**接口路径**: `GET /api/photos/thumb/{photoId}.jpg`

### 3. 获取原图
**接口路径**: `GET /api/photos/full/{photoId}.jpg`

## 数据要求

### 照片处理
- **缩略图**: 建议300x300像素，JPEG格式，质量80%
- **原图**: 保持原始尺寸和格式
- **路径**: 统一使用相对路径，前端会自动添加base URL

### 时间分组
- **分组逻辑**: 按拍摄时间的年月分组
- **排序**: 最新月份在前，同月内按拍摄时间降序
- **格式**: 时期ID使用"YYYY-MM"格式

### 前端适配说明
当前前端会根据照片的width和height计算宽高比进行响应式布局：
- 桌面端: 每行5-7张照片
- 移动端: 每行2-3张照片

## 现有前端代码修改点
需要将PhotosView.vue中的模拟数据替换为真实API调用：

```typescript
// 替换 generatePhotoPeriods() 函数为:
const loadPhotoPeriods = async (): Promise<PhotoPeriod[]> => {
  try {
    const response = await fetch('/api/photos/timeline')
    const result = await response.json()
    if (result.success) {
      return result.data.periods
    }
  } catch (error) {
    console.error('加载照片数据失败:', error)
  }
  return []
}
```

## 优先级
1. **高优先级**: 照片时间线接口 + 缩略图接口
2. **中优先级**: 原图接口 + 基础元数据
3. **低优先级**: 高级元数据(相机信息等)

## 已完成的集成

### ✅ API对接状态
- **时间线接口**: 已完成 `GET /api/v1/photos/timeline`
- **照片数据接口**: 已完成 `GET /api/v1/photos`  
- **缩略图接口**: 已完成 `GET /api/v1/photo/{hash}?format=thumbnail&size=400x400`
- **HTTP配置**: 已更新baseURL为 `http://localhost:9482`

### ✅ 前端实现特性
1. **时间线加载**: 按月分组显示照片统计
2. **懒加载**: 滚动时自动加载对应月份的照片数据
3. **缩略图懒加载**: 可视区域照片自动加载缩略图
4. **虚拟滚动**: 支持大量照片的高性能展示
5. **响应式布局**: 根据屏幕尺寸自适应照片排列

### 📝 数据流程
1. 页面初始化 → 加载时间线统计数据
2. 滚动到某月份 → 懒加载该月份的照片基础数据(hash、宽高比等)
3. 进入可视区域 → 加载实际缩略图显示

### 🔧 配置说明
- 后端端口: 9482
- 缩略图尺寸: 400x400
- 每批加载: 100张照片
- 照片高度: 统一200px，宽度按比例计算

---
*✅ 前端API对接已完成，可以与后端进行实时数据交互。*