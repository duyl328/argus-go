# Argus HTTP API 测试文件

本目录包含了 Argus 照片管理系统的全套 HTTP API 测试文件，按功能模块进行分类。

## 📁 文件结构

```
tests/
├── README.md          # 本文档
├── base.http          # 基础功能测试（健康检查、环境配置）
├── library.http       # 图像库管理测试
├── photo.http         # 照片文件和详情接口测试
├── exif.http         # EXIF信息查询测试
├── user.http         # 用户管理接口测试
└── dev.http          # 开发调试接口测试
```

## 🚀 快速开始

### 1. 环境准备

在项目根目录创建 `http-client.env.json` 文件：

```json
{
  "dev": {
    "baseUrl": "http://localhost:8080",
    "sampleHash": "your_actual_hash_here"
  },
  "prod": {
    "baseUrl": "https://your-production-domain.com", 
    "sampleHash": "production_hash_here"
  }
}
```

### 2. 启动服务器

```bash
go run .
```

### 3. 测试顺序

建议按以下顺序执行测试：

1. **基础功能** (`base.http`)
   - 健康检查
   - API基础路由

2. **图像库管理** (`library.http`) 
   - 添加图像库路径
   - 启用图像库
   - 开始索引
   - 获取图像hash值

3. **照片接口** (`photo.http`)
   - 使用实际hash测试图像访问
   - 测试各种缩略图尺寸
   - 获取图像详情

4. **EXIF查询** (`exif.http`)
   - 查询EXIF信息
   - 按相机、参数筛选
   - GPS信息查询

5. **用户管理** (`user.http`)
   - 用户CRUD操作
   - 用户查询和筛选

6. **开发调试** (`dev.http`)
   - 系统状态监控
   - 性能测试
   - 开发工具

## 📝 测试说明

### 变量替换

测试文件中使用了以下变量，需要替换为实际值：

- `{{baseUrl}}`: 服务器地址 (如: http://localhost:8080)
- `{{apiBase}}`: API基础路径 ({{baseUrl}}/api/v1)
- `{{sampleHash}}`: 实际的图像hash值

### 获取实际Hash值

执行 `library.http` 中的图像库索引后，从响应中获取实际的图像hash值，然后替换测试文件中的 `sample_hash_12345`。

### 文件路径注意事项

- Windows: 使用反斜杠 `D:\\Pictures\\sample.jpg` 或正斜杠 `D:/Pictures/sample.jpg`
- Linux/Mac: 使用正斜杠 `/home/user/Pictures/sample.jpg`
- 确保路径中的图像文件真实存在

## 🔧 IDE 配置

### VS Code

安装 `REST Client` 扩展，可以直接在编辑器中执行HTTP请求。

### IntelliJ IDEA / WebStorm

内置HTTP客户端支持，可以直接运行 `.http` 文件。

### 命令行工具

也可以使用 curl 执行测试：

```bash
# 健康检查
curl http://localhost:8080/health

# 获取图像
curl "http://localhost:8080/api/v1/photo/your_hash?format=thumbnail&size=400x400"

# 获取图像详情
curl "http://localhost:8080/api/v1/assets/your_hash"
```

## 📊 响应格式

### 成功响应

```json
{
  "code": 200,
  "message": "Success", 
  "data": { ... }
}
```

### 错误响应

```json
{
  "code": 404,
  "message": "Photo not found",
  "error": "详细错误信息"
}
```

### 图像文件响应

图像接口直接返回二进制数据，HTTP头包含正确的 `Content-Type`。

## ⚠️ 注意事项

1. **数据安全**: 测试前备份重要数据
2. **路径检查**: 确保文件路径存在且可访问
3. **性能影响**: 大量并发测试可能影响系统性能
4. **开发接口**: `dev.http` 中的接口仅供开发环境使用
5. **权限要求**: 某些操作可能需要特定权限

## 🐛 常见问题

### 连接失败

- 检查服务器是否启动
- 确认端口号正确
- 检查防火墙设置

### 404 错误

- 确认API路径正确
- 检查图像hash是否有效
- 验证图像文件是否存在

### 缩略图生成失败

- 检查ImageMagick/LibVips安装
- 确认原图文件可读
- 查看系统日志获取详细错误

### EXIF解析失败

- 检查ExifTool安装
- 确认图像文件包含EXIF数据
- 验证文件格式支持

## 📚 相关文档

- [API使用指南](../API_USAGE.md)
- [项目说明](../CLAUDE.md)
- [配置文档](../config.yaml)