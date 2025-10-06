# Argus Front - 照片管理系统前端架构文档

> 本文档专为 Claude Code 设计，提供完整的项目架构、功能模块、工具集和开发规范

## 📋 项目概述

**Argus Front** 是一个基于 Vue 3 + TypeScript 的现代化照片管理系统前端，采用模块化架构设计，通过 HTTP API 与后端服务通信。

### 🎯 核心设计准则

**本项目采用桌面应用优先的设计理念**

1. **桌面应用视角**
   - 项目未来将打包为 Electron/Tauri 桌面应用
   - 所有设计决策以桌面软件的用户体验为准
   - 不需要考虑浏览器兼容性限制

2. **快捷键优先级**
   - **应用快捷键 > 系统/浏览器快捷键**
   - 所有应用内快捷键必须阻止默认行为 (`event.preventDefault()`)
   - 所有应用内快捷键必须阻止事件冒泡 (`event.stopPropagation()`)
   - 与系统冲突的快捷键，以我们的应用为准

3. **常用快捷键示例**
   - `F5` - 刷新当前视图（禁用浏览器刷新）
   - `Backspace` - 返回上一级（禁用浏览器后退）
   - `Ctrl+C/X/V` - 复制/剪切/粘贴（接管系统剪贴板）
   - `Delete` - 删除文件（不触发浏览器行为）
   - `Enter` - 进入文件夹/确认操作
   - `Esc` - 取消选择/关闭对话框

4. **UI/UX 设计原则**
   - 使用原生桌面应用的交互模式（如 Windows 资源管理器）
   - 提供丰富的右键菜单
   - 支持拖拽操作
   - 快捷键提示和帮助文档

5. **性能优化方向**
   - 本地文件系统优先（减少网络请求）
   - 虚拟滚动处理大量文件
   - 懒加载和缓存策略
   - 原生 API 集成（文件系统、系统托盘等）

### 技术栈

| 技术 | 版本 | 用途 |
|------|------|------|
| Vue | 3.5.13 | 响应式前端框架 |
| TypeScript | 5.8 | 类型安全 |
| Vite | 6.2.4 | 构建工具 |
| Naive UI | 2.41.1 | UI 组件库（自动导入） |
| Vue Router | 4.5.0 | 路由管理 |
| Pinia | 3.0.1 | 状态管理 |
| Axios | 1.9.0 | HTTP 客户端 |
| @vicons/ionicons5 | ^0.13.0 | 图标库 |
| vue-virtual-scroller | 2.0.0-beta.8 | 虚拟滚动 |
| justified-layout | ^4.1.0 | 瀑布流布局 |

---

## 📁 项目结构

```
src/
├── assets/              # 静态资源
│   ├── var.css         # CSS 变量（设计系统 tokens）
│   ├── colors.css      # 主题颜色定义
│   ├── base.css        # 基础样式
│   ├── main.css        # 主样式入口
│   └── icon/ali/       # 阿里图标库
├── components/          # 可复用组件
│   ├── Layout/
│   │   └── AppLayout.vue          # 主布局（头部+侧边栏+内容区）
│   ├── FileManager/               # 文件管理器（核心功能）
│   │   ├── FileManager.vue        # 主组件（工具栏+双面板）
│   │   ├── FilePane.vue           # 单面板（文件展示+交互）
│   │   ├── ContextMenu.vue        # 右键菜单
│   │   ├── Tooltip.vue            # 工具提示
│   │   ├── QuickPreview.vue       # 快速预览
│   │   ├── DebugPanel.vue         # 调试面板
│   │   ├── types.ts               # 文件管理器类型定义
│   │   ├── mockData.ts            # 模拟数据
│   │   └── index.ts               # 导出配置
│   ├── ThemeToggle.vue            # 主题切换开关
│   └── icons/                     # 自定义图标组件
├── composables/         # 可组合函数（Vue Composition API）
│   ├── useLayout.ts               # 布局相关（响应式断点）
│   └── fileManager/               # 文件管理器专用 composables
│       ├── index.ts               # 统一导出
│       ├── useFileSelection.ts    # 文件选择逻辑
│       ├── useKeyboardNav.ts      # 键盘导航
│       ├── useDragSelection.ts    # 拖拽框选
│       ├── useDragAndDrop.ts      # 拖放操作
│       └── useVirtualScroll.ts    # 虚拟滚动
├── config/              # 配置管理
│   └── httpConfig.ts              # HTTP 配置类（HttpConfigManager）
├── constants/           # 常量定义
│   ├── app.ts                     # 应用常量（环境、模式）
│   └── logUtil.ts                 # 日志颜色映射
├── model/               # 数据模型
│   └── libraryTable.ts            # 资料库表格模型
├── plugins/             # Vue 插件
│   └── http.ts                    # HTTP 插件（全局注入）
├── router/              # 路由配置
│   └── index.ts                   # 路由定义
├── services/            # 业务服务层
│   └── timelineService.ts         # 照片时间线服务
├── stores/              # Pinia 状态管理
│   ├── theme.ts                   # 主题状态
│   └── counter.ts                 # 示例计数器
├── types/               # TypeScript 类型定义
│   ├── http.ts                    # HTTP 相关类型
│   ├── api.ts                     # API 响应类型
│   ├── menu.ts                    # 菜单配置类型
│   ├── vue-virtual-scroller.d.ts  # 虚拟滚动类型声明
│   └── justified-layout.d.ts      # 布局库类型声明
├── utils/               # 工具函数
│   ├── http.ts                    # HTTP 客户端（HttpClient 类）
│   ├── logHelper/                 # 日志工具
│   │   ├── logUtils.ts            # 彩色日志输出
│   │   └── logEnum.ts             # 日志等级枚举
│   └── fileManager/               # 文件操作工具
│       └── fileOperations.ts      # 文件增删改查
├── views/               # 页面组件
│   ├── PhotosView.vue             # 照片视图（主页）
│   ├── PhotosView1.vue            # 照片视图变体1
│   ├── PhotosView2.vue            # 照片视图变体2（当前使用）
│   ├── SearchView.vue             # 搜索
│   ├── PeopleView.vue             # 人物
│   ├── FieldView.vue              # 文件夹
│   ├── LocalFilesView.vue         # 本地文件
│   ├── LibraryView.vue            # 资料库
│   └── SettingsView.vue           # 设置
├── App.vue              # 根组件
└── main.ts              # 应用入口
```

---

## 🎯 核心功能模块

### 1. 布局系统（AppLayout.vue）

**位置**: `src/components/Layout/AppLayout.vue`

#### 功能特性
- **响应式头部** (64px 高度)
  - Logo + 站点标题
  - 中央搜索框
  - 主题切换按钮
- **可折叠侧边栏** (290px ↔ 64px)
  - 图标菜单系统（支持分组）
  - 路由自动激活状态
  - 图标映射机制
- **主内容区域**
  - 动态 `<router-view />` 渲染
  - 自适应高度 (calc(100vh - 64px))

#### 关键依赖
```typescript
import { useThemeStore } from '@/stores/theme'
import { menuOptions } from '@/types/menu'
import { ImageOutline, SearchOutline, ... } from '@vicons/ionicons5'
```

#### 图标映射
```typescript
const iconMap = {
  'image-outline': ImageOutline,
  'search-outline': SearchOutline,
  // ... 其他图标
}
```

---

### 2. 文件管理器（FileManager）

**位置**: `src/components/FileManager/`

#### 核心组件
1. **FileManager.vue** - 主容器
   - 双面板布局（左/右、上/下、单面板）
   - 工具栏（新建、复制、粘贴、视图切换）
   - 面板配置管理（独立的 viewMode、thumbnailSize、sortOptions）
   - 拖拽分隔器调整面板大小

2. **FilePane.vue** - 单文件面板
   - 网格/列表视图切换
   - 文件拖拽框选
   - 键盘导航
   - 虚拟滚动（性能优化）
   - 历史记录（前进/后退）

3. **辅助组件**
   - `ContextMenu.vue` - 右键菜单
   - `Tooltip.vue` - 自定义工具提示
   - `QuickPreview.vue` - 快速预览
   - `DebugPanel.vue` - 调试信息

#### Composables（可组合逻辑）

| Composable | 文件 | 功能 |
|-----------|------|------|
| `useFileSelection` | useFileSelection.ts | 文件选择、多选、范围选择 |
| `useKeyboardNav` | useKeyboardNav.ts | 方向键、Home/End、Enter 导航 |
| `useDragSelection` | useDragSelection.ts | 鼠标拖拽框选、边缘自动滚动 |
| `useDragAndDrop` | useDragAndDrop.ts | 拖放文件/文件夹、预览反馈 |
| `useVirtualScroll` | useVirtualScroll.ts | 虚拟滚动列表优化 |

#### 文件操作工具（fileOperations.ts）

**位置**: `src/utils/fileManager/fileOperations.ts`

**提供函数**:
```typescript
// 路径操作
getFolderByPath(structure, path) // 根据路径获取文件夹

// CRUD 操作
moveItems(structure, itemNames, sourcePath, targetPath) // 移动
copyItems(structure, itemNames, sourcePath, targetPath) // 复制
deleteItems(structure, itemNames, path)                 // 删除
renameItem(structure, oldName, newName, path)           // 重命名
createFolder(structure, folderName, path)               // 新建文件夹

// 查询操作
getItems(structure, path)                               // 获取项目列表
searchItems(structure, path, query)                     // 搜索文件
```

**注意**: 所有操作需要传入响应式的 `structure` 对象，才能触发 Vue 响应式更新。

---

### 3. HTTP 请求系统

**架构**: 插件化 + 拦截器 + 配置管理

#### HttpClient 类（utils/http.ts）

**核心特性**:
- ✅ 请求拦截器
  - 自动添加 `Authorization` Token
  - GET 请求添加时间戳防缓存
  - 重复请求自动取消
- ✅ 响应拦截器
  - 统一业务状态码处理
  - 错误分类处理（401/403/404/500）
  - 日志输出（开发环境）
- ✅ 请求方法封装
  - `get()`, `post()`, `put()`, `delete()`, `patch()`
  - 泛型支持完整类型推导

**使用示例**:
```typescript
import { httpClient } from '@/utils/http'

// 带类型推导的请求
const response = await httpClient.get<UserInfo>('/api/user', { id: 1 })
// response.data 自动推导为 UserInfo 类型
```

#### HttpConfigManager（config/httpConfig.ts）

**配置管理器**:
```typescript
import { httpConfig } from '@/config/httpConfig'

// 设置基础 URL
httpConfig.setBaseURL('http://127.0.0.1:9484/api')

// 设置端口
httpConfig.setPort('127.0.0.1', 9484)

// 设置请求头
httpConfig.setHeaders({ 'X-Custom': 'value' })
```

#### HTTP 插件（plugins/http.ts）

**全局注入**:
```typescript
// main.ts 中已注册
app.use(httpPlugin)

// 组件中使用
import { inject } from 'vue'
const http = inject('$http')
```

---

### 4. 照片时间线服务（timelineService.ts）

**位置**: `src/services/timelineService.ts`

#### API 接口

| 函数 | 端点 | 功能 |
|-----|------|------|
| `getTimeline()` | `GET /v1/photos/timeline` | 获取时间线统计 |
| `getPhotos()` | `GET /v1/photos` | 获取照片列表（列式存储） |
| `getPhotosByDate()` | - | 分页获取某天所有照片 |
| `getFullTimeline()` | - | 获取完整时间线（含照片） |

#### 数据处理流程

1. **获取时间线** → 列式存储转行式存储
```typescript
// API 返回格式（列式）
{
  hash: ['abc', 'def'],
  isImage: [true, false],
  takenAt: ['2024-01-01', '2024-01-02'],
  ratio: [1.5, 1.0]
}

// 转换为行式存储
[
  { hash: 'abc', isImage: true, takenAt: '2024-01-01', ratio: 1.5 },
  { hash: 'def', isImage: false, takenAt: '2024-01-02', ratio: 1.0 }
]
```

2. **按月分组** → `groupTimelineByMonth()`
```typescript
interface MonthGroup {
  year: number
  month: number
  title: string           // "2024年1月"
  subtitle: string        // "128张照片"
  days: DayGroup[]
  totalCount: number
}
```

3. **预渲染优化字段**
```typescript
interface Photo {
  // ... 其他字段
  loaded?: boolean        // 是否已加载真实图片
  inViewport?: boolean    // 是否在视口中
  placeholder?: string    // 占位符颜色
}
```

#### 工具函数

```typescript
// 获取缩略图 URL
getThumbnailUrl(hash, size = '400x400')

// 获取原图 URL
getOriginalUrl(hash)
```

---

### 5. 主题系统

**位置**: `src/stores/theme.ts`

#### Pinia Store

```typescript
import { useThemeStore } from '@/stores/theme'

const themeStore = useThemeStore()

// 初始化主题（读取 localStorage 或系统偏好）
themeStore.initTheme()

// 切换主题
themeStore.toggleTheme()

// 设置主题
themeStore.setTheme('dark' | 'light')
```

#### CSS 变量系统（var.css）

**设计 Tokens**:
```css
:root {
  /* 布局尺寸 */
  --header-height: 64px;
  --sidebar-width: 290px;

  /* 间距系统 */
  --spacing-xs: 4px;
  --spacing-sm: 8px;
  --spacing-md: 16px;
  --spacing-lg: 24px;
  --spacing-xl: 32px;
  --spacing-2xl: 48px;

  /* 字体系统 */
  --font-size-xs: 0.75rem;   /* 12px */
  --font-size-sm: 0.875rem;  /* 14px */
  --font-size-base: 1rem;    /* 16px */
  --font-weight-normal: 400;
  --font-weight-bold: 700;

  /* z-index 层级 */
  --z-dropdown: 1000;
  --z-modal: 1050;
  --z-tooltip: 1070;
}
```

#### 主题颜色（colors.css）

```css
/* 亮色主题 */
.light {
  --bg-color: #ffffff;
  --text-color: #18181b;
}

/* 暗色主题 */
.dark {
  --bg-color: #18181b;
  --text-color: #fafafa;
}
```

---

### 6. 日志系统

**位置**: `src/utils/logHelper/logUtils.ts`

#### 彩色日志工具

```typescript
import { logB, logN, logS } from '@/utils/logHelper/logUtils'

// 带背景色日志
logB.primary('主要信息', data)
logB.success('成功', result)
logB.error('错误', error)

// 命名空间日志（带边框）
logN.info('HTTP', '请求发送', config)
logN.warning('Store', '状态已更新')

// 简单彩色文本
logS.success('操作成功')
logS.danger('危险操作')
```

#### 日志颜色

| 类型 | 颜色 |
|------|------|
| primary | 蓝色 |
| success | 绿色 |
| info | 浅蓝 |
| warning | 橙色 |
| danger | 红色 |
| error | 深红 |

**环境控制**:
```typescript
// 仅在开发环境输出日志
const isPrint = import.meta.env.VITE_MODE !== 'production'
```

---

## 🛠️ 工具函数清单

### HTTP 工具（utils/http.ts）

```typescript
import { httpClient } from '@/utils/http'

// HTTP 方法
httpClient.get<T>(url, params?, config?)
httpClient.post<T>(url, data?, config?)
httpClient.put<T>(url, data?, config?)
httpClient.delete<T>(url, params?, config?)
httpClient.patch<T>(url, data?, config?)

// 工具方法
httpClient.cancelAllRequests()        // 取消所有请求
httpClient.getAxiosInstance()         // 获取 Axios 实例
httpClient.updateConfig(config)       // 更新配置
```

### 文件操作（utils/fileManager/fileOperations.ts）

```typescript
import {
  getFolderByPath,
  moveItems,
  copyItems,
  deleteItems,
  renameItem,
  createFolder,
  getItems,
  searchItems
} from '@/utils/fileManager/fileOperations'
```

**⚠️ 重要**: 所有函数需要传入响应式 `structure` 对象。

### 日志工具（utils/logHelper/logUtils.ts）

```typescript
import { logB, logN, logS, disLog } from '@/utils/logHelper/logUtils'

// 三种日志样式
logB.{type}(msg, ...args)  // 背景色日志
logN.{type}(ns, msg, ...args)  // 命名空间日志
logS.{type}(msg, ...args)  // 简单彩色日志

disLog()  // 禁用所有 console.log
```

### 布局工具（composables/useLayout.ts）

```typescript
import { useLayout } from '@/composables/useLayout'

const {
  isMobile,        // ref<boolean>
  isTablet,        // ref<boolean>
  isDesktop,       // ref<boolean>
  screenWidth,     // ref<number>
  screenHeight     // ref<number>
} = useLayout()
```

---

## 📦 类型系统

### HTTP 类型（types/http.ts）

```typescript
// HTTP 配置
interface HttpConfig {
  baseURL: string
  timeout?: number
  headers?: Record<string, string>
}

// API 响应格式
interface ApiResponse<T = unknown> {
  code: number
  data: T
  message: string
  success: boolean
}

// 请求配置
interface RequestConfig {
  url: string
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH'
  params?: unknown
  data?: unknown
}
```

### API 类型（types/api.ts）

```typescript
// 时间线
interface TimelineItem {
  date: string      // YYYY-MM-DD
  count: number
}

// 照片（列式存储）
interface PhotosData {
  hash: string[]
  isImage: boolean[]
  takenAt: string[]
  ratio: number[]
}

// 照片（行式存储）
interface Photo {
  hash: string
  isImage: boolean
  takenAt: string
  ratio: number
  width: number
  height: number
  loaded?: boolean
  inViewport?: boolean
  placeholder?: string
}

// 月份分组
interface MonthGroup {
  year: number
  month: number
  title: string
  subtitle: string
  days: DayGroup[]
  totalCount: number
}
```

### 文件管理器类型（components/FileManager/types.ts）

```typescript
interface FileItem {
  name: string
  type: 'folder' | 'photo' | 'video' | 'file'
  size?: string
  date?: string
  children?: Record<string, FileItem>
}

type ViewMode = 'grid' | 'list'
type ThumbnailSize = 'small' | 'medium' | 'large'
type LayoutMode = 'single' | 'horizontal' | 'vertical'
type PaneId = 'left' | 'right'

interface SortOptions {
  field: 'name' | 'extension' | 'date' | 'size' | 'type'
  order: 'asc' | 'desc'
}

interface FilterOptions {
  nameQuery: string
  fileType: 'all' | 'photo' | 'video' | 'folder' | 'file'
}
```

### 菜单类型（types/menu.ts）

```typescript
interface MenuItem {
  key: string
  label: string
  icon?: string
  path: string
}

interface MenuGroup {
  key: string
  label: string
  children: MenuItem[]
}

type MenuOption = MenuItem | MenuGroup

export const menuOptions: MenuOption[] = [...]
```

---

## 🚀 开发规范

### 1. 代码组织规范

#### 导入顺序
```typescript
// 1. Vue 核心库
import { ref, computed, onMounted } from 'vue'

// 2. 第三方库
import axios from 'axios'
import { NButton, NInput } from 'naive-ui'

// 3. 项目内部模块（按类型分组）
import type { User } from '@/types/api'
import { httpClient } from '@/utils/http'
import { useThemeStore } from '@/stores/theme'
import MyComponent from '@/components/MyComponent.vue'
```

#### 文件命名
- **组件**: PascalCase - `AppLayout.vue`, `FileManager.vue`
- **工具函数**: camelCase - `fileOperations.ts`, `logUtils.ts`
- **类型文件**: camelCase - `http.ts`, `api.ts`
- **常量文件**: camelCase - `app.ts`, `logUtil.ts`

### 2. 快捷键实现规范

**⚠️ 重要：所有快捷键必须遵循桌面应用优先原则**

#### 快捷键处理模板
```typescript
function handleKeyDown(event: KeyboardEvent) {
  // 1. 检查面板是否激活
  if (!isActive.value) return

  // 2. 跳过输入框（除了特定快捷键如 Esc）
  if (event.target instanceof HTMLInputElement && event.key !== 'Escape') return

  switch (event.key) {
    case 'F5':
      // ✅ 正确：始终阻止默认行为和事件冒泡
      event.preventDefault()
      event.stopPropagation()
      onRefresh()
      break

    case 'Delete':
      // ✅ 正确：阻止默认行为
      event.preventDefault()
      event.stopPropagation()
      onDelete()
      break

    case 'c':
    case 'C':
      if (event.ctrlKey || event.metaKey) {
        // ✅ 正确：Ctrl+C 优先于系统复制
        event.preventDefault()
        event.stopPropagation()
        onCopy()
      }
      break
  }
}
```

#### 必须阻止默认行为的快捷键
- `F5` - 防止浏览器刷新
- `Backspace` - 防止浏览器后退
- `Ctrl+C/X/V` - 接管系统剪贴板
- `Delete` - 防止浏览器行为
- 所有方向键 - 防止页面滚动

#### 快捷键注册位置
- **全局快捷键**: 在 `composables/fileManager/useKeyboardNav.ts`
- **组件快捷键**: 在组件的 `onMounted` 中注册
- **对话框快捷键**: 在对话框组件内部处理

#### 快捷键冲突解决
```typescript
// ❌ 错误：没有阻止默认行为
case 'F5':
  onRefresh()
  break

// ✅ 正确：完全接管快捷键
case 'F5':
  event.preventDefault()      // 阻止浏览器刷新
  event.stopPropagation()    // 阻止事件冒泡
  onRefresh()
  break
```

### 3. TypeScript 规范

#### 类型定义
```typescript
// ✅ 优先使用 interface（可扩展）
interface User {
  id: number
  name: string
}

// ✅ 联合类型使用 type
type Status = 'pending' | 'success' | 'error'

// ✅ 泛型函数
async function fetchData<T>(url: string): Promise<ApiResponse<T>> {
  return httpClient.get<T>(url)
}
```

#### 类型导入
```typescript
// ✅ 使用 type 关键字导入类型
import type { ApiResponse } from '@/types/http'

// ❌ 避免混合导入
import { ApiResponse } from '@/types/http'  // 可能导致编译问题
```

### 3. 工具函数复用规范

#### ⚠️ 检查已有工具
**在添加新工具函数前，必须先检查以下位置**:

1. **HTTP 工具** → `utils/http.ts` (HttpClient 类)
2. **文件操作** → `utils/fileManager/fileOperations.ts`
3. **日志工具** → `utils/logHelper/logUtils.ts`
4. **布局工具** → `composables/useLayout.ts`
5. **文件管理器** → `composables/fileManager/index.ts`

#### 工具函数注册流程
```typescript
// 1. 定义工具函数（utils/myUtils.ts）
export function formatDate(date: Date): string {
  return date.toISOString()
}

// 2. 如果是 composable，统一在 index.ts 导出
// composables/index.ts
export { formatDate } from './myUtils'

// 3. 使用时导入
import { formatDate } from '@/utils/myUtils'
// 或
import { formatDate } from '@/composables'
```

#### ❌ 禁止重复定义
```typescript
// ❌ 错误：重复定义已有功能
function moveFile() { /* ... */ }  // fileOperations.ts 已有 moveItems()

// ✅ 正确：复用已有工具
import { moveItems } from '@/utils/fileManager/fileOperations'
```

### 4. HTTP 请求规范

#### 统一使用 httpClient
```typescript
// ✅ 正确：使用封装的 httpClient
import { httpClient } from '@/utils/http'
const data = await httpClient.get('/api/users')

// ❌ 错误：直接使用 axios
import axios from 'axios'
const data = await axios.get('http://...')  // 跳过了拦截器
```

#### 类型安全请求
```typescript
// ✅ 定义响应类型
interface UserResponse {
  id: number
  name: string
}

// ✅ 使用泛型
const response = await httpClient.get<UserResponse>('/api/user/1')
// response.data 自动推导为 UserResponse
```

#### Service 层封装
```typescript
// services/userService.ts
import { httpClient } from '@/utils/http'
import type { User } from '@/types/api'

export async function getUserById(id: number): Promise<User> {
  const response = await httpClient.get<User>(`/api/user/${id}`)
  return response.data
}
```

### 5. 样式开发规范

#### CSS 变量优先
```vue
<style scoped>
.my-component {
  /* ✅ 使用 CSS 变量 */
  padding: var(--spacing-md);
  font-size: var(--font-size-base);
  color: var(--text-color);

  /* ❌ 避免硬编码 */
  padding: 16px;
  font-size: 16px;
  color: #18181b;
}
</style>
```

#### 响应式断点
```typescript
import { useLayout } from '@/composables/useLayout'

const { isMobile, isTablet, isDesktop } = useLayout()
```

```vue
<template>
  <div :class="{ 'mobile-layout': isMobile }">
    <!-- ... -->
  </div>
</template>
```

### 6. 组件开发规范

#### Composition API 标准结构
```vue
<script setup lang="ts">
// 1. 导入
import { ref, computed, onMounted } from 'vue'
import type { User } from '@/types/api'

// 2. Props 定义
interface Props {
  userId: number
}
const props = defineProps<Props>()

// 3. Emits 定义
interface Emits {
  update: [user: User]
}
const emit = defineEmits<Emits>()

// 4. 响应式状态
const user = ref<User | null>(null)

// 5. 计算属性
const userName = computed(() => user.value?.name ?? '')

// 6. 方法
async function fetchUser() {
  // ...
}

// 7. 生命周期
onMounted(() => {
  fetchUser()
})

// 8. 暴露给父组件
defineExpose({
  fetchUser
})
</script>
```

#### 组件注册
```typescript
// ✅ Naive UI 组件自动导入（无需注册）
<template>
  <n-button>按钮</n-button>
</template>

// ✅ 自定义组件按需导入
<script setup lang="ts">
import MyComponent from '@/components/MyComponent.vue'
</script>
```

### 7. 状态管理规范

#### Pinia Store 结构
```typescript
// stores/user.ts
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useUserStore = defineStore('user', () => {
  // State
  const user = ref<User | null>(null)

  // Getters
  const isLoggedIn = computed(() => !!user.value)

  // Actions
  async function login(credentials: LoginParams) {
    const response = await httpClient.post('/api/login', credentials)
    user.value = response.data
  }

  return {
    user,
    isLoggedIn,
    login
  }
})
```

#### Store 使用
```typescript
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()
userStore.login({ username: 'admin', password: '123' })
```

### 8. 错误处理规范

#### HTTP 错误处理
```typescript
try {
  const data = await httpClient.get('/api/data')
} catch (error) {
  if (error instanceof Error) {
    logB.error('请求失败', error.message)
    // 业务错误处理
  }
}
```

#### 组件错误边界
```vue
<script setup lang="ts">
import { onErrorCaptured } from 'vue'

onErrorCaptured((err, instance, info) => {
  logB.error('组件错误', { err, info })
  return false  // 阻止错误继续传播
})
</script>
```

---

## 📍 路由配置

**位置**: `src/router/index.ts`

```typescript
const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', redirect: '/photos' },

    // 主要页面
    { path: '/photos', component: () => import('@/views/PhotosView2.vue') },
    { path: '/search', component: () => import('@/views/SearchView.vue') },
    { path: '/people', component: () => import('@/views/PeopleView.vue') },
    { path: '/field', component: () => import('@/views/FieldView.vue') },
    { path: '/local-files', component: () => import('@/views/LocalFilesView.vue') },
    { path: '/library', component: () => import('@/views/LibraryView.vue') },
    { path: '/settings', component: () => import('@/views/SettingsView.vue') }
  ]
})
```

**路由元信息**:
```typescript
{
  path: '/photos',
  meta: { title: '照片' }
}
```

---

## 🔧 开发工具

### NPM Scripts

```json
{
  "dev": "vite",                    // 启动开发服务器（端口 3000）
  "build": "run-p type-check build-only",  // 类型检查 + 构建
  "type-check": "vue-tsc --build",  // TypeScript 类型检查
  "lint": "eslint . --fix",         // 代码检查并自动修复
  "format": "prettier --write src/", // 代码格式化
  "test:unit": "vitest"             // 单元测试
}
```

### Vite 配置

**位置**: `vite.config.ts`

```typescript
export default defineConfig({
  plugins: [
    vue(),
    vueDevTools(),

    // Vue APIs 自动导入
    AutoImport({
      imports: [
        'vue',
        { 'naive-ui': ['useDialog', 'useMessage', 'useNotification', 'useLoadingBar'] }
      ]
    }),

    // Naive UI 组件自动导入
    Components({
      resolvers: [NaiveUiResolver()]
    })
  ],

  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    }
  },

  server: {
    port: 3000
  }
})
```

### 环境变量

**文件**: `.env`, `.env.production`

```bash
# API 基础 URL
VITE_APP_API_URL=http://127.0.0.1:9484/api

# 运行模式
VITE_MODE=development
```

**使用**:
```typescript
const apiUrl = import.meta.env.VITE_APP_API_URL
const isDev = import.meta.env.DEV
const isProd = import.meta.env.PROD
```

---

## 🧩 依赖关系图

### 核心依赖流
```
main.ts
  ├─> App.vue
  │     └─> router-view
  │           └─> views/*
  │                 └─> components/*
  │
  ├─> Pinia (stores/*)
  ├─> Router (router/index.ts)
  └─> httpPlugin (plugins/http.ts)
        └─> httpClient (utils/http.ts)
              └─> httpConfig (config/httpConfig.ts)
```

### 工具函数依赖
```
utils/http.ts
  ├─> types/http.ts
  └─> utils/logHelper/logUtils.ts
        └─> constants/logUtil.ts

utils/fileManager/fileOperations.ts
  └─> components/FileManager/types.ts

services/timelineService.ts
  ├─> utils/http.ts
  └─> types/api.ts
```

### Composables 依赖
```
composables/fileManager/index.ts
  ├─> useFileSelection.ts
  ├─> useKeyboardNav.ts
  ├─> useDragSelection.ts
  ├─> useDragAndDrop.ts
  └─> useVirtualScroll.ts
```

---

## ⚙️ 配置管理

### HTTP 配置

**默认配置**:
```typescript
{
  baseURL: import.meta.env.VITE_APP_API_URL || 'http://127.0.0.1:9484/api',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  }
}
```

**动态修改**:
```typescript
import { httpConfig } from '@/config/httpConfig'
import { httpClient } from '@/utils/http'

// 修改配置
httpConfig.setBaseURL('http://new-api.com')
httpConfig.setTimeout(15000)

// 应用配置
httpClient.updateConfig(httpConfig.getConfig())
```

### 主题配置

**初始化**:
```typescript
// AppLayout.vue
import { useThemeStore } from '@/stores/theme'

const themeStore = useThemeStore()
onMounted(() => {
  themeStore.initTheme()  // 读取 localStorage 或系统偏好
})
```

**优先级**:
1. localStorage 存储的用户偏好
2. 系统偏好 `prefers-color-scheme`
3. 默认亮色主题

---

## 🐛 调试工具

### 开发环境日志

```typescript
import { logB } from '@/utils/logHelper/logUtils'

// HTTP 请求自动记录
// utils/http.ts 中已配置拦截器输出
```

### 文件管理器调试面板

**位置**: `src/components/FileManager/DebugPanel.vue`

**显示信息**:
- 当前路径
- 选中文件数量
- 视口信息
- 虚拟滚动状态

**使用**:
```vue
<DebugPanel
  :current-path="currentPath"
  :selected-count="selectedItems.size"
/>
```

### Vue DevTools

**已启用**: `vite.config.ts` 中配置 `vueDevTools()`

---

## 📚 常见任务示例

### 1. 添加新页面

```typescript
// 1. 创建页面组件
// src/views/MyNewView.vue
<script setup lang="ts">
import { ref } from 'vue'
const data = ref(null)
</script>

// 2. 添加路由
// src/router/index.ts
{
  path: '/my-new-page',
  name: 'my-new-page',
  component: () => import('@/views/MyNewView.vue'),
  meta: { title: '新页面' }
}

// 3. 添加菜单项
// src/types/menu.ts
{
  key: 'my-new-page',
  label: '新页面',
  icon: 'document-outline',
  path: '/my-new-page'
}
```

### 2. 添加 API 接口

```typescript
// 1. 定义类型
// src/types/api.ts
export interface MyData {
  id: number
  name: string
}

// 2. 创建 service
// src/services/myService.ts
import { httpClient } from '@/utils/http'
import type { MyData } from '@/types/api'

export async function getMyData(): Promise<MyData[]> {
  const response = await httpClient.get<MyData[]>('/api/my-data')
  return response.data
}

// 3. 在组件中使用
import { getMyData } from '@/services/myService'

const data = ref<MyData[]>([])
onMounted(async () => {
  data.value = await getMyData()
})
```

### 3. 添加工具函数

```typescript
// 1. 创建工具函数文件
// src/utils/myHelper.ts
export function formatCurrency(amount: number): string {
  return `¥${amount.toFixed(2)}`
}

// 2. 如果是 composable，创建在 composables 目录
// src/composables/useMyLogic.ts
import { ref } from 'vue'

export function useMyLogic() {
  const count = ref(0)

  function increment() {
    count.value++
  }

  return { count, increment }
}

// 3. 统一导出（可选）
// src/composables/index.ts
export { useMyLogic } from './useMyLogic'
```

### 4. 添加全局组件

```typescript
// 1. 创建组件
// src/components/GlobalComponent.vue
<script setup lang="ts">
defineProps<{ text: string }>()
</script>

// 2. 注册为全局组件（如果需要）
// main.ts
import GlobalComponent from '@/components/GlobalComponent.vue'
app.component('GlobalComponent', GlobalComponent)

// 或使用 Naive UI 自动导入（无需注册）
```

---

## 🔒 注意事项

### 安全性
- ✅ Token 存储在 localStorage，自动添加到请求头
- ✅ 401 响应自动清除 Token 并跳转登录
- ⚠️ 生产环境需配置 CORS 和 CSP

### 性能优化
- ✅ 路由懒加载 `() => import('@/views/...')`
- ✅ 虚拟滚动 (vue-virtual-scroller)
- ✅ 图片懒加载（placeholder + loaded 标记）
- ✅ HTTP 请求去重（重复请求自动取消）

### 兼容性
- ✅ 支持现代浏览器（Chrome, Firefox, Safari, Edge）
- ⚠️ 不支持 IE11（使用了 ES2020 特性）

### 类型安全
- ✅ 严格模式 TypeScript
- ✅ 所有 API 响应定义类型
- ✅ 组件 Props 类型检查
- ⚠️ 避免使用 `any`，优先 `unknown`

---

## 📖 快速参考

### 常用导入路径

```typescript
// 工具类
import { httpClient } from '@/utils/http'
import { logB, logN, logS } from '@/utils/logHelper/logUtils'
import { moveItems, copyItems } from '@/utils/fileManager/fileOperations'

// 类型
import type { ApiResponse } from '@/types/http'
import type { Photo, MonthGroup } from '@/types/api'
import type { FileItem } from '@/components/FileManager/types'

// 状态管理
import { useThemeStore } from '@/stores/theme'

// 服务
import { getTimeline, getPhotos } from '@/services/timelineService'

// Composables
import { useLayout } from '@/composables/useLayout'
import { useFileSelection } from '@/composables/fileManager'

// 配置
import { httpConfig } from '@/config/httpConfig'
```

### 常用 CSS 变量

```css
/* 布局 */
var(--header-height)
var(--sidebar-width)

/* 间距 */
var(--spacing-xs)  /* 4px */
var(--spacing-sm)  /* 8px */
var(--spacing-md)  /* 16px */
var(--spacing-lg)  /* 24px */

/* 字体 */
var(--font-size-base)  /* 16px */
var(--font-weight-normal)  /* 400 */

/* 主题颜色 */
var(--bg-color)
var(--text-color)
var(--border-color)
```

---

## 🎓 最佳实践总结

1. **优先复用** - 使用前检查已有工具函数
2. **类型安全** - 所有函数、API 都定义类型
3. **统一日志** - 使用 logB/logN/logS，禁用原生 console.log
4. **统一请求** - 通过 httpClient，不直接使用 axios
5. **CSS 变量** - 优先使用设计系统 tokens
6. **响应式数据** - 文件操作需传入响应式对象
7. **错误处理** - try-catch + 日志记录
8. **按需导入** - 利用 Vite 自动导入特性
9. **命名规范** - 组件 PascalCase，文件 camelCase
10. **文档更新** - 新增功能同步更新本文档

---

**文档版本**: 1.0
**最后更新**: 2025-10-03
**维护者**: Claude Code
