# Argus - 图像处理与管理系统

一个用Go语言开发的高性能图像处理与管理系统，支持多种图像格式处理、EXIF信息提取、缩略图生成等功能。

## 功能特性

- 🖼️ **多格式图像支持**: 支持 JPG、PNG、TIFF、BMP、GIF、HEIC/HEIF、WebP、AVIF、JXL等格式
- 📊 **EXIF信息提取**: 基于ExifTool的强大元数据提取功能
- 🎨 **高质量图像处理**: 集成ImageMagick和LibVips双重处理引擎
- 🚀 **高性能缩略图生成**: 可配置的缩略图格式和质量
- 🔄 **异步任务处理**: 基于协程的高并发图像处理
- 💾 **灵活的数据库支持**: 支持SQLite和MySQL数据库
- 🌐 **RESTful API**: 提供完整的HTTP API接口
- 📈 **性能监控**: 内置pprof性能分析工具

## 环境要求

### 运行环境
- **Go**: 1.24 或更高版本
- **操作系统**: Windows (x64)、Linux (x64)、macOS (x64/ARM64)

### 外部依赖工具
该项目依赖以下外部工具，需要在构建时正确配置：

1. **ExifTool** - EXIF信息提取
   - 版本要求: 13.30 或更高
   - 用途: 图像元数据提取和处理

2. **ImageMagick** - 图像处理
   - 版本要求: 7.1.1 或更高  
   - 配置: Portable Q16-HDRI版本
   - 用途: 图像格式转换、缩放、质量调整

3. **LibVips** - 高性能图像处理
   - 版本要求: 8.17.0 或更高
   - 用途: 高性能图像处理和缩略图生成

## 快速开始

### 1. 克隆项目

```bash
git clone <your-repo-url>
cd go-argus/src/rear
```

### 2. 安装Go依赖

```bash
go mod download
```

### 3. 准备工具链

项目使用自动工具链管理系统。在开发环境下，程序会自动从 `tools/` 目录解压所需的工具：

#### 目录结构说明
```
tools/
├── windows_amd64/          # Windows 64位工具
│   ├── exiftool/
│   │   └── exiftool-13.30_64.zip
│   ├── imagemagick/
│   │   └── ImageMagick-7.1.1-47-portable-Q16-HDRI-x64.zip
│   └── libvips/
│       └── vips-dev-w64-all-8.17.0.zip
├── linux_amd64/           # Linux 64位工具 (待添加)
└── darwin_amd64/          # macOS 64位工具 (待添加)
    └── darwin_arm64/      # macOS ARM64工具 (待添加)
```

#### Windows用户 (已提供)
工具包已包含在项目中，首次运行时会自动解压到应用目录。

#### Linux/macOS用户 (需要手动准备)
请下载对应平台的工具包并按以下结构放置：

**Linux x64:**
```bash
# 创建目录
mkdir -p tools/linux_amd64/{exiftool,imagemagick,libvips}

# 下载并放置工具包
# ExifTool
wget -O tools/linux_amd64/exiftool/exiftool-13.30.tar.gz https://exiftool.org/Image-ExifTool-13.30.tar.gz

# ImageMagick (示例)
# 请根据官方文档下载适合的Linux版本

# LibVips (示例)  
# 请根据官方文档下载适合的Linux版本
```

**macOS:**
```bash
# 创建目录
mkdir -p tools/darwin_amd64/{exiftool,imagemagick,libvips}
mkdir -p tools/darwin_arm64/{exiftool,imagemagick,libvips}

# 使用Homebrew安装或手动下载对应工具包
```

### 4. 配置文件

复制配置文件并根据需要修改：

```bash
cp config.yaml.example config.yaml
```

主要配置项：
- `app.port`: 服务端口 (默认: 8080)
- `app.mode`: 运行模式 debug/release (默认: debug)
- `app.development`: 开发模式开关 (默认: true)
- `database.type`: 数据库类型 sqlite/mysql (默认: sqlite)
- `image.thumbnail_format`: 缩略图格式 (默认: jpg)

### 5. 运行应用

#### 开发模式
```bash
go run main.go
```

#### 编译后运行
```bash
go build -o rear.exe .
./rear.exe
```

应用启动后，访问 `http://localhost:8080` 即可使用API。

### 6. 验证安装

应用启动时会显示工具版本信息，确认所有工具都正确检测：

```
=== 工具信息 ===
ImageMagick : Version: ImageMagick 7.1.1-47 Q16-HDRI x64
ExifTool    : ExifTool 13.30
LibVips     : vips-8.17.0
================
```

## 构建部署

### 本地构建

```bash
# 构建当前平台
go build -o rear .

# 跨平台构建示例
# Windows
GOOS=windows GOARCH=amd64 go build -o rear.exe .

# Linux
GOOS=linux GOARCH=amd64 go build -o rear .

# macOS
GOOS=darwin GOARCH=amd64 go build -o rear .
GOOS=darwin GOARCH=arm64 go build -o rear .
```

### 使用构建脚本

项目提供了自动化构建脚本：

```bash
chmod +x scripts/build.sh
./scripts/build.sh
```

该脚本会：
1. 构建多平台可执行文件
2. 复制对应平台的工具链
3. 创建发布包

### 生产部署

1. **准备运行环境**
   - 确保目标机器有对应平台的工具链文件
   - 配置合适的 `config.yaml`
   - 设置 `app.development: false` 启用生产模式

2. **部署步骤**
   ```bash
   # 1. 上传可执行文件和配置
   scp rear config.yaml target-server:/opt/argus/
   
   # 2. 上传工具链 (如果需要)
   scp -r tools/ target-server:/opt/argus/
   
   # 3. 启动服务
   ssh target-server "cd /opt/argus && ./rear"
   ```

3. **系统服务配置** (Linux)
   ```ini
   # /etc/systemd/system/argus.service
   [Unit]
   Description=Argus Image Processing Service
   After=network.target
   
   [Service]
   Type=simple
   User=argus
   WorkingDirectory=/opt/argus
   ExecStart=/opt/argus/rear
   Restart=always
   RestartSec=10
   
   [Install]
   WantedBy=multi-user.target
   ```

## API文档

### 基础接口

- `GET /api/health` - 健康检查
- `GET /api/version` - 版本信息

### 图像处理接口

- `POST /api/images/upload` - 上传图像
- `GET /api/images/{id}` - 获取图像信息
- `GET /api/images/{id}/thumbnail` - 获取缩略图
- `GET /api/images/{id}/exif` - 获取EXIF信息

详细API文档请查看项目的API文档或通过接口探索。

## 开发指南

### 项目结构

```
.
├── main.go                 # 应用入口
├── config.yaml.example     # 配置文件模板
├── internal/              # 内部包
│   ├── api/              # API处理器
│   ├── config/           # 配置管理
│   ├── container/        # 依赖注入容器
│   ├── db/              # 数据库操作
│   ├── handler/         # 业务处理器
│   ├── model/           # 数据模型
│   ├── repositories/    # 数据仓库
│   ├── router/          # 路由配置
│   ├── service/         # 业务服务
│   ├── utils/           # 工具函数
│   └── workflow/        # 工作流
├── pkg/                 # 公共包
│   ├── img_utils/       # 图像工具
│   ├── logger/          # 日志工具  
│   └── utils/           # 通用工具
├── tools/               # 外部工具链
└── scripts/             # 构建脚本
```

### 运行测试

```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./internal/utils -v

# 运行性能测试
go test -bench=. ./...
```

### 性能分析

开发模式下，应用会启用pprof性能分析：

```bash
# CPU性能分析
go tool pprof http://localhost:8080/debug/pprof/profile

# 内存分析  
go tool pprof http://localhost:8080/debug/pprof/heap

# 协程分析
go tool pprof http://localhost:8080/debug/pprof/goroutine
```

## 故障排除

### 常见问题

1. **工具未找到错误**
   ```
   Error: ImageMagick not found
   ```
   - 检查 `tools/` 目录下是否有对应平台的工具包
   - 确认工具包文件名正确
   - 在开发模式下确保 `app.development: true`

2. **权限错误** (Linux/macOS)
   ```
   permission denied
   ```
   - 给可执行文件添加执行权限: `chmod +x rear`
   - 检查工具链文件权限

3. **端口占用**
   ```
   bind: address already in use
   ```
   - 修改 `config.yaml` 中的端口配置
   - 或终止占用端口的进程

4. **数据库连接失败**
   - 检查数据库配置
   - 确认数据库文件路径权限 (SQLite)
   - 检查MySQL连接参数 (MySQL)

### 日志查看

应用日志保存在 `app-logs/` 目录下：

```bash
# 查看应用日志
tail -f app-logs/app.log

# 查看错误日志
grep "ERROR" app-logs/app.log
```

### 调试模式

设置环境变量启用详细日志：

```bash
# 设置日志级别为debug
export LOG_LEVEL=debug
./rear
```