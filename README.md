# Argus - 智能相册管理系统

一个基于 Vue 3 + Go + Gin 构建的现代化相册管理系统，专注于图片的智能处理、管理和展示。

## 🌟 特性

- **现代化界面**: 基于 Vue 3 + Naive UI 的响应式前端界面
- **智能图片处理**: 支持图片压缩、格式转换、缩略图生成
- **EXIF 信息管理**: 读取、编辑和删除图片的 EXIF 元数据
- **多路径管理**: 支持添加多个照片存储路径，灵活管理相册
- **任务调度**: 后台任务处理系统，支持大批量图片处理
- **跨平台**: 支持 Windows、Linux、macOS

## 🛠 技术栈

### 前端 (Vue 3)
- **框架**: Vue 3.5.13 + TypeScript 5.8
- **UI 库**: Naive UI 2.41.1
- **路由**: Vue Router 4.5.0
- **状态管理**: Pinia 3.0.1
- **HTTP 客户端**: Axios 1.9.0
- **图标**: @vicons/ionicons5 0.13.0
- **构建工具**: Vite 6.2.4
- **开发工具**: 
  - ESLint 9.22.0 (代码规范)
  - Prettier 3.5.3 (代码格式化)
  - Vitest 3.1.1 (单元测试)

### 后端 (Go 1.24)
- **Web 框架**: Gin 1.10.1
- **ORM**: GORM 1.30.0
- **日志**: Zap 1.27.0 + Lumberjack 2.2.1
- **UUID**: Google UUID 1.6.0
- **文件类型检测**: h2non/filetype 1.1.3
- **跨域**: gin-contrib/cors 1.7.5
- **性能监控**: gin-contrib/pprof 1.5.3

### 数据库
- **SQLite**: 默认数据库，适合轻量级部署
- **MySQL**: 可选的关系型数据库支持

### 图片处理工具
- **ExifTool**: 图片 EXIF 信息读取和编辑
- **ImageMagick**: 图片格式转换和基础处理
- **libvips**: 高性能图片压缩和处理库

## 🚀 主要功能

### 已完成功能
1. **资料库管理**
   - 添加/删除图片存储路径
   - 启用/禁用指定路径
   - 路径有效性验证

2. **图片处理**
   - EXIF 信息读取、编辑、删除
   - 图片格式转换 (JPEG, PNG, WebP, AVIF 等)
   - 图片压缩和优化
   - 缩略图生成
   - 批量处理支持

3. **后台任务系统**
   - 任务队列管理
   - 图片索引和处理任务
   - 任务状态监控

### 计划中功能
- 照片相似度检测
- 人脸识别和分类
- 智能标签系统
- 照片地理位置管理
- 全文搜索功能

## 📦 安装和配置

### 环境要求
- Node.js 22.x 或更高版本
- Go 1.24 或更高版本
- SQLite 3.x

### 开发环境搭建

1. **克隆项目**
```bash
git clone https://github.com/yourusername/argus-go.git
cd argus-go
```

2. **安装外部工具依赖**

根据你的操作系统，下载并安装以下工具：

**Windows:**
- 下载 ExifTool 并解压到 `src/rear/tools/windows_amd64/exiftool/`
- 下载 ImageMagick Portable 并解压到 `src/rear/tools/windows_amd64/imagemagick/`
- 下载 libvips 并解压到 `src/rear/tools/windows_amd64/libvips/`

**Linux/macOS:**
```bash
# Ubuntu/Debian
sudo apt-get install exiftool imagemagick libvips-tools

# macOS
brew install exiftool imagemagick vips
```

3. **启动后端服务**
```bash
cd src/rear
go mod download
go run main.go
```

4. **启动前端服务**
```bash
cd src/front/argus-front
npm install
npm run dev
```

5. **访问应用**
- 前端地址: http://localhost:5173
- 后端 API: http://localhost:8080

### 生产环境部署

1. **构建前端**
```bash
cd src/front/argus-front
npm run build
```

2. **构建后端**
```bash
cd src/rear
go build -o argus-server main.go
```

3. **配置数据库** (可选，默认使用 SQLite)
- 修改配置文件中的数据库连接信息
- 支持 MySQL 数据库

4. **启动服务**
```bash
./argus-server
```

## 🤝 贡献指南

欢迎贡献代码！请遵循以下规范：

### 开发规范
1. **代码风格**
   - 前端: ESLint + Prettier 配置
   - 后端: Go 标准格式化工具

2. **提交规范**
   - 使用语义化提交信息
   - 每个 commit 应该是一个完整的功能点

3. **测试要求**
   - 新功能需要添加相应的单元测试
   - 确保现有测试通过

### 贡献流程
1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建 Pull Request

## 📄 许可证

本项目采用 GPL 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情

## 📞 联系方式

如有问题或建议，请通过以下方式联系：
- 提交 Issue: https://github.com/duyl328/argus-go/issues

## 🙏 致谢

感谢以下开源项目的支持：
- [Vue.js](https://vuejs.org/)
- [Gin](https://gin-gonic.com/)
- [Naive UI](https://www.naiveui.com/)
- [ExifTool](https://exiftool.org/)
- [ImageMagick](https://imagemagick.org/)
- [libvips](https://libvips.github.io/libvips/)
