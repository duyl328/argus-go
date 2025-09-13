# 文件系统管理API使用示例

本文档展示了如何使用Argus照片管理系统的文件系统管理API。

## API基础地址

```
基础URL: http://localhost:8080
API路径: /api/v1/filesystem
```

## 1. 文件系统浏览

### 1.1 获取硬盘/挂载点列表

**请求**：
```http
GET /api/v1/filesystem/browse
```

**响应示例**：
```json
{
  "code": 200,
  "message": "获取文件系统信息成功",
  "data": {
    "current_path": "/",
    "parent_path": "",
    "items": [
      {
        "id": "drive_C",
        "name": "C: (Windows)",
        "path": "C:\\",
        "type": "drive",
        "size": 500000000000,
        "mod_time": "2024-01-15T10:30:00Z",
        "is_accessible": true,
        "drive_info": {
          "label": "Windows",
          "file_system": "NTFS",
          "total_space": 500000000000,
          "free_space": 100000000000,
          "used_space": 400000000000,
          "usage_percent": 80.0,
          "is_removable": false,
          "drive_type": "fixed"
        }
      },
      {
        "id": "drive_D",
        "name": "D: (数据盘)",
        "path": "D:\\",
        "type": "drive",
        "size": 1000000000000,
        "mod_time": "2024-01-15T10:30:00Z",
        "is_accessible": true,
        "drive_info": {
          "label": "数据盘",
          "file_system": "NTFS",
          "total_space": 1000000000000,
          "free_space": 500000000000,
          "used_space": 500000000000,
          "usage_percent": 50.0,
          "is_removable": false,
          "drive_type": "fixed"
        }
      }
    ],
    "summary": {
      "total_items": 2,
      "directory_count": 0,
      "file_count": 0,
      "drive_count": 2
    }
  }
}
```

### 1.2 浏览指定目录

**请求**：
```http
GET /api/v1/filesystem/browse?path=C:\Windows
```

**响应示例**：
```json
{
  "code": 200,
  "message": "获取文件系统信息成功",
  "data": {
    "current_path": "C:\\Windows",
    "parent_path": "C:\\",
    "items": [
      {
        "id": "dir_C_Windows_System32",
        "name": "System32",
        "path": "C:\\Windows\\System32",
        "type": "directory",
        "size": 0,
        "mod_time": "2024-01-10T08:15:00Z",
        "is_accessible": true,
        "directory_info": {
          "item_count": 0
        }
      },
      {
        "id": "file_C_Windows_win.ini",
        "name": "win.ini",
        "path": "C:\\Windows\\win.ini",
        "type": "file",
        "size": 92,
        "mod_time": "2024-01-05T12:00:00Z",
        "is_accessible": true,
        "file_info": {
          "extension": ".ini",
          "mime_type": "text/plain"
        }
      }
    ],
    "summary": {
      "total_items": 2,
      "directory_count": 1,
      "file_count": 1,
      "drive_count": 0
    }
  }
}
```

## 2. 磁盘使用情况查询

### 2.1 获取特定磁盘使用情况

**请求**：
```http
GET /api/v1/filesystem/disk-usage?path=C:\
```

**响应示例**：
```json
{
  "code": 200,
  "message": "获取磁盘使用情况成功",
  "data": {
    "label": "Windows",
    "file_system": "NTFS",
    "total_space": 500000000000,
    "free_space": 100000000000,
    "used_space": 400000000000,
    "usage_percent": 80.0,
    "is_removable": false,
    "drive_type": "fixed"
  }
}
```

## 3. 文件搜索

### 3.1 搜索指定类型文件

**请求**：
```http
GET /api/v1/filesystem/search?path=D:\go-argus&pattern=*.go&recursive=true
```

**参数说明**：
- `path`: 搜索路径
- `pattern`: 搜索模式（支持通配符）
- `recursive`: 是否递归搜索子目录

## 4. 目录操作

### 4.1 创建目录

**请求**：
```http
POST /api/v1/filesystem/directory
Content-Type: application/json

{
  "path": "D:\\go-argus\\test-directory"
}
```

**响应示例**：
```json
{
  "code": 200,
  "message": "创建目录功能正在开发中",
  "data": {
    "path": "D:\\go-argus\\test-directory"
  }
}
```

## 5. 文件/目录删除

### 5.1 删除文件或目录

**请求**：
```http
DELETE /api/v1/filesystem/item?path=D:\go-argus\test-file.txt
```

## 6. 文件/目录移动

### 6.1 重命名文件

**请求**：
```http
PUT /api/v1/filesystem/item/move
Content-Type: application/json

{
  "source": "D:\\go-argus\\old-file.txt",
  "destination": "D:\\go-argus\\new-file.txt"
}
```

### 6.2 移动文件到其他目录

**请求**：
```http
PUT /api/v1/filesystem/item/move
Content-Type: application/json

{
  "source": "D:\\go-argus\\src\\file.txt",
  "destination": "D:\\go-argus\\backup\\file.txt"
}
```

## 7. 文件/目录复制

### 7.1 复制文件

**请求**：
```http
POST /api/v1/filesystem/item/copy
Content-Type: application/json

{
  "source": "D:\\go-argus\\important-file.txt",
  "destination": "D:\\go-argus\\backup\\important-file.txt",
  "overwrite": false
}
```

**参数说明**：
- `source`: 源文件路径
- `destination`: 目标文件路径
- `overwrite`: 是否覆盖已存在的文件

## 8. 错误处理

### 8.1 路径不存在

**请求**：
```http
GET /api/v1/filesystem/browse?path=X:\NonExistentPath
```

**响应示例**：
```json
{
  "code": 400,
  "message": "路径不存在: X:\\NonExistentPath"
}
```

### 8.2 缺少必需参数

**请求**：
```http
GET /api/v1/filesystem/disk-usage
```

**响应示例**：
```json
{
  "code": 400,
  "message": "path参数不能为空"
}
```

## 9. 前端集成示例

### 9.1 JavaScript/TypeScript示例

```typescript
// 文件系统API客户端
class FileSystemAPI {
  private baseUrl = 'http://localhost:8080/api/v1/filesystem';

  // 获取硬盘列表
  async getDrives(): Promise<FileSystemResponse> {
    const response = await fetch(`${this.baseUrl}/browse`);
    return response.json();
  }

  // 浏览目录
  async browseDirectory(path: string): Promise<FileSystemResponse> {
    const response = await fetch(`${this.baseUrl}/browse?path=${encodeURIComponent(path)}`);
    return response.json();
  }

  // 获取磁盘使用情况
  async getDiskUsage(path: string): Promise<DriveInfo> {
    const response = await fetch(`${this.baseUrl}/disk-usage?path=${encodeURIComponent(path)}`);
    const result = await response.json();
    return result.data;
  }

  // 搜索文件
  async searchFiles(path: string, pattern: string, recursive: boolean = false) {
    const params = new URLSearchParams({
      path,
      pattern,
      recursive: recursive.toString()
    });
    const response = await fetch(`${this.baseUrl}/search?${params}`);
    return response.json();
  }

  // 创建目录
  async createDirectory(path: string) {
    const response = await fetch(`${this.baseUrl}/directory`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ path })
    });
    return response.json();
  }
}

// 使用示例
const fsAPI = new FileSystemAPI();

// 获取硬盘列表并显示
fsAPI.getDrives().then(result => {
  result.data.items.forEach(drive => {
    if (drive.type === 'drive') {
      console.log(`${drive.name}: ${drive.drive_info?.usage_percent.toFixed(1)}% 已使用`);
    }
  });
});

// 浏览C盘内容
fsAPI.browseDirectory('C:\\').then(result => {
  console.log(`C盘包含 ${result.data.summary.directory_count} 个目录和 ${result.data.summary.file_count} 个文件`);
});
```

### 9.2 面包屑导航实现

```typescript
// 面包屑导航组件
class BreadcrumbNavigation {
  private currentPath: string = '/';
  private pathHistory: string[] = ['/'];

  // 导航到指定路径
  async navigateTo(path: string) {
    const fsAPI = new FileSystemAPI();

    try {
      const result = await fsAPI.browseDirectory(path);
      this.currentPath = result.data.current_path;
      this.pathHistory.push(path);
      this.updateBreadcrumb();
      this.displayItems(result.data.items);
    } catch (error) {
      console.error('导航失败:', error);
    }
  }

  // 返回上级目录
  async navigateUp() {
    if (this.pathHistory.length > 1) {
      this.pathHistory.pop(); // 移除当前路径
      const parentPath = this.pathHistory[this.pathHistory.length - 1];
      await this.navigateTo(parentPath);
    }
  }

  // 更新面包屑显示
  private updateBreadcrumb() {
    const breadcrumb = document.getElementById('breadcrumb');
    if (!breadcrumb) return;

    const pathParts = this.currentPath.split(/[/\\]/);
    breadcrumb.innerHTML = pathParts.map((part, index) => {
      if (index === 0 && part === '') return '<span>根目录</span>';
      return `<span onclick="navigateToIndex(${index})">${part}</span>`;
    }).join(' > ');
  }

  // 显示文件和目录项目
  private displayItems(items: FileSystemItem[]) {
    const itemsContainer = document.getElementById('items');
    if (!itemsContainer) return;

    itemsContainer.innerHTML = items.map(item => {
      const icon = item.type === 'directory' ? '📁' : '📄';
      const sizeStr = item.type === 'file' ? this.formatSize(item.size) : '';

      return `
        <div class="item" onclick="handleItemClick('${item.path}', '${item.type}')">
          <span class="icon">${icon}</span>
          <span class="name">${item.name}</span>
          <span class="size">${sizeStr}</span>
        </div>
      `;
    }).join('');
  }

  // 格式化文件大小
  private formatSize(bytes: number): string {
    const units = ['B', 'KB', 'MB', 'GB', 'TB'];
    let size = bytes;
    let unitIndex = 0;

    while (size >= 1024 && unitIndex < units.length - 1) {
      size /= 1024;
      unitIndex++;
    }

    return `${size.toFixed(1)} ${units[unitIndex]}`;
  }
}

// 项目点击处理
function handleItemClick(path: string, type: string) {
  const navigation = new BreadcrumbNavigation();

  if (type === 'directory' || type === 'drive') {
    navigation.navigateTo(path);
  } else {
    // 处理文件点击（例如下载、预览等）
    console.log('点击文件:', path);
  }
}
```

## 10. 使用场景总结

### 10.1 文件管理器界面
1. **首页展示**：调用 `browse` 接口不传参数，展示所有硬盘
2. **目录导航**：用户点击目录时调用 `browse?path=xxx`
3. **磁盘信息**：调用 `disk-usage` 显示磁盘使用率图表

### 10.2 文件操作功能
1. **新建文件夹**：使用 `POST /directory` 接口
2. **文件搜索**：使用 `GET /search` 接口
3. **文件管理**：使用移动、复制、删除接口

### 10.3 系统监控
1. **磁盘监控**：定期调用 `disk-usage` 监控磁盘使用情况
2. **系统信息**：结合硬盘列表和系统信息API

这个API设计提供了完整的文件系统管理功能，适合构建现代化的文件管理界面。