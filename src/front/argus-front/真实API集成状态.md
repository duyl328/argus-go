# 真实 API 集成状态

**最后更新**: 2025-10-03

## ✅ 已完成的功能

### 基础浏览功能
- [x] 浏览驱动器列表 (根目录)
- [x] 双击进入文件夹
- [x] 显示文件和文件夹列表
- [x] 文件类型识别 (photo/video/file/folder)
- [x] 文件大小格式化显示
- [x] 修改时间格式化显示

### 导航功能
- [x] 面包屑导航点击跳转
- [x] 鼠标前进/后退按钮
- [x] 历史记录管理
- [x] 路径编辑功能 (手动输入路径)

### UI 显示
- [x] 面包屑路径显示
- [x] 文件统计 (照片数量、视频数量)
- [x] 网格/列表视图切换
- [x] 缩略图大小调整
- [x] 加载状态显示
- [x] 错误提示

### 调试工具
- [x] 调试面板显示 API 状态
- [x] 真实 API / Mock 数据切换
- [x] API 加载状态监控
- [x] 错误信息显示

## ⚠️ 部分支持的功能

### 面包屑下拉菜单
- **状态**: 真实 API 模式下已禁用
- **原因**: 需要额外的 API 调用获取同级目录列表
- **Mock 模式**: 正常工作
- **后续**: 可以添加一个新的 API 端点支持

### 搜索功能
- **状态**: UI 存在，但真实 API 未集成
- **后端 API**: 已实现 `/api/v1/filesystem/search`
- **待集成**: 前端调用和结果显示

## ❌ 未支持的功能

### 文件操作 (需要集成后端 API)
- [ ] 创建文件夹
- [ ] 删除文件/文件夹
- [ ] 重命名
- [ ] 复制文件/文件夹
- [ ] 移动文件/文件夹
- [ ] 拖放操作

### 高级功能
- [ ] 文件预览
- [ ] 批量选择操作
- [ ] 右键菜单集成真实 API
- [ ] 快速预览

## 📝 已修改的文件

### 核心文件
1. **`src/views/FieldView.vue`**
   - 添加 `:use-real-api="true"` 启用真实 API

2. **`src/components/FileManager/FileManager.vue`**
   - 添加 `useRealApi` prop
   - 传递给所有 FilePane 组件

3. **`src/components/FileManager/FilePane.vue`**
   - 接收 `useRealApi` prop
   - 修改所有导航函数支持真实 API:
     - `navigateToFolder` - 文件夹导航
     - `navigateToFolderByPath` - 使用完整路径导航
     - `navigateToIndex` - 面包屑点击导航
     - `goBack` / `goForward` - 历史记录前进后退
     - `navigateToPathArray` - 路径数组导航辅助函数
     - `convertPathToBreadcrumbs` - 路径转面包屑
     - `applyPathEdit` - 手动输入路径
     - `enterPathEditMode` - 路径编辑模式
   - 修改 `handleItemDoubleClick` - 使用 item.path
   - 修改 `onMounted` - 初始化加载真实文件系统
   - 修改 `currentFolder` computed - 支持真实 API 数据
   - 禁用 `toggleBreadcrumbDropdown` (真实 API 模式)

4. **`src/components/FileManager/types.ts`**
   - 添加 `path?: string` 字段到 FileItem

5. **`src/components/FileManager/DebugPanel.vue`**
   - 添加 API 集成状态显示
   - 添加 API 切换按钮
   - 显示加载状态和错误

### 新增文件
1. **`src/services/fileSystemService.ts`**
   - 文件系统 API 服务层
   - 封装所有后端 API 调用

2. **`src/composables/fileManager/useFileSystemAPI.ts`**
   - 文件系统 API Composable
   - 状态管理和数据转换

## 🔍 技术细节

### 路径处理

#### Windows 路径格式
- **驱动器**: `D:\`
- **子目录**: `D:\Users\Documents`
- **分隔符**: 反斜杠 `\`

#### 面包屑转换
```typescript
// 后端返回: "D:\Users\Documents"
// 前端显示: ["D:", "Users", "Documents"]

function convertPathToBreadcrumbs(path: string): string[] {
  if (!path) return ['根目录']

  // 处理纯驱动器: "D:\" -> ["D:\"]
  if (/^[A-Z]:\\?$/.test(path)) {
    return [path]
  }

  // 处理子路径: "D:\Users\Docs" -> ["D:", "Users", "Docs"]
  const parts = path.split('\\').filter(p => p)
  return parts.length > 0 ? parts : ['根目录']
}
```

#### 路径重建
```typescript
// 面包屑: ["D:", "Users", "Documents"]
// 重建路径: "D:\Users\Documents"

const fullPath = pathArray.join('\\')
await fileSystemAPI.browse(fullPath)
```

### 数据流

#### 初始加载
```
用户访问页面
  → FilePane.onMounted()
  → fileSystemAPI.browse() (空路径)
  → 后端返回驱动器列表
  → 显示面包屑: ["所有驱动器"]
```

#### 双击文件夹
```
用户双击 "D: (Data)"
  → handleItemDoubleClick(item)
  → navigateToFolderByPath(item.path) // "D:\"
  → fileSystemAPI.browse("D:\")
  → 后端返回 D:\ 内容
  → convertPathToBreadcrumbs("D:\") → ["D:\"]
  → 更新 UI
```

#### 面包屑导航
```
用户点击面包屑 "Users"
  → navigateToIndex(1)
  → 重建路径: ["D:", "Users"] → "D:\Users"
  → fileSystemAPI.browse("D:\Users")
  → 更新 UI
```

#### 历史前进/后退
```
用户点击鼠标后退
  → goBack()
  → 从历史获取路径数组
  → navigateToPathArray(pathArray)
  → 重建路径并调用 API
  → 更新 UI
```

## 🐛 已知问题

### 1. 路径格式不统一
- **问题**: 有时面包屑显示 `["D:"]`，有时显示 `["D:\"]`
- **影响**: 路径重建时可能出错
- **状态**: 需要统一处理

### 2. 特殊路径处理
- **问题**: 网络路径 `\\server\share` 未测试
- **影响**: 可能无法访问网络共享
- **状态**: 待测试

### 3. 权限错误提示
- **问题**: 访问受保护目录时错误提示不够友好
- **影响**: 用户体验
- **状态**: 可优化

## 🚀 后续开发计划

### 短期 (1-2天)
- [ ] 集成文件操作 API (创建/删除/重命名/复制/移动)
- [ ] 集成搜索功能
- [ ] 优化路径处理，统一格式
- [ ] 改进错误提示

### 中期 (1周)
- [ ] 集成拖放操作到真实 API
- [ ] 添加面包屑下拉菜单支持 (需要新 API)
- [ ] 文件预览功能
- [ ] 批量操作支持

### 长期 (2周+)
- [ ] 性能优化 (虚拟滚动 + API 分页)
- [ ] 操作进度显示 (大文件复制)
- [ ] 操作历史和撤销
- [ ] 收藏夹功能

## 📖 开发指南

### 添加新的文件操作

1. **确认后端 API 存在**
   - 查看 `rear/tests/filesystem.http`
   - 确认端点和参数格式

2. **在 fileSystemService.ts 添加方法**
   ```typescript
   async createDirectory(path: string): Promise<OperationResult> {
     const response = await httpClient.post('/v1/filesystem/directory', { path })
     return response.data
   }
   ```

3. **在 useFileSystemAPI.ts 封装**
   ```typescript
   async function createDirectory(path: string) {
     try {
       const result = await fileSystemService.createDirectory(path)
       await browse(currentPath.value) // 刷新
       return result
     } catch (err) {
       error.value = err.message
       throw err
     }
   }
   ```

4. **在 FilePane.vue 调用**
   ```typescript
   async function handleCreateFolder() {
     if (USE_REAL_API.value) {
       const newPath = `${fileSystemAPI.currentPath.value}\\新建文件夹`
       await fileSystemAPI.createDirectory(newPath)
     } else {
       // Mock 数据模式
     }
   }
   ```

### 测试建议

1. **基础功能测试**
   - 浏览根目录
   - 进入驱动器
   - 进入多层文件夹
   - 点击面包屑返回
   - 鼠标前进/后退

2. **边界测试**
   - 访问无权限目录
   - 访问不存在路径
   - 网络断开
   - 后端服务停止

3. **性能测试**
   - 大量文件的文件夹
   - 深层嵌套目录
   - 快速连续导航

## 🔗 相关文档

- [文件系统集成说明.md](./文件系统集成说明.md) - 完整集成文档
- [快速开始.md](./快速开始.md) - 快速启动指南
- [后端 API 测试](../../rear/tests/filesystem.http) - API 测试用例

---

**维护者**: Claude Code
**项目**: Argus 照片管理系统
