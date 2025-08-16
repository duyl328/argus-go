# Argus Front - 照片管理项目前端

## 项目概述
这是一个Vue3项目，使用TypeScript和NaiveUI开发的照片管理项目前端，通过HTTP请求与后端服务通信来展示照片内容。

## 技术栈
- **Vue 3.5.13** - 响应式前端框架
- **TypeScript 5.8** - 类型安全的JavaScript超集
- **Naive UI 2.41.1** - 组件库（支持自动导入）
- **Vue Router 4.5.0** - 前端路由管理
- **Pinia 3.0.1** - 状态管理
- **Axios 1.9.0** - HTTP请求库
- **Vite 6.2.4** - 构建工具
- **@vicons/ionicons5** - 图标库

## 项目结构

### 目录组织
```
src/
├── components/          # 可复用组件
│   ├── Layout/         # 布局组件
│   │   └── AppLayout.vue  # 主布局组件（包含导航栏、侧边栏）
│   ├── ThemeToggle.vue  # 主题切换组件
│   └── icons/          # 图标组件
├── views/              # 页面组件
│   ├── PhotosView.vue   # 照片页面（主页）
│   ├── SearchView.vue   # 搜索页面
│   ├── PeopleView.vue   # 人物页面
│   ├── FieldView.vue    # 文件夹页面
│   ├── LibraryView.vue  # 资料库页面
│   └── SettingsView.vue # 设置页面
├── stores/             # Pinia状态管理
│   ├── theme.ts        # 主题状态管理
│   └── counter.ts      # 计数器状态（示例）
├── router/             # 路由配置
│   └── index.ts        # 路由定义
├── utils/              # 工具函数
│   ├── http.ts         # HTTP请求客户端类
│   └── logHelper/      # 日志工具
├── types/              # TypeScript类型定义
│   ├── http.ts         # HTTP相关类型
│   └── menu.ts         # 菜单相关类型
├── config/             # 配置文件
│   └── httpConfig.ts   # HTTP配置管理
├── constants/          # 常量定义
│   ├── app.ts          # 应用常量
│   └── logUtil.ts      # 日志常量
├── assets/             # 静态资源
│   ├── var.css         # CSS变量定义
│   ├── colors.css      # 颜色主题
│   ├── base.css        # 基础样式
│   └── main.css        # 主样式文件
└── composables/        # 可组合函数
    └── useLayout.ts    # 布局相关组合函数
```

## 路由结构
- `/` → 重定向到 `/photos`
- `/photos` - 照片页面（首页）
- `/search` - 搜索页面
- `/people` - 人物页面
- `/field` - 文件夹页面
- `/library` - 资料库页面
- `/settings` - 设置页面

## 核心功能模块

### 1. 布局系统 (AppLayout.vue)
- **响应式侧边导航栏**：可折叠/展开，宽度290px（展开）/64px（折叠）
- **顶部导航栏**：包含Logo、搜索框、主题切换按钮
- **主内容区域**：动态路由视图
- **菜单系统**：基于配置的图标菜单，支持分组

### 2. 主题系统
- **双主题支持**：亮色/暗色主题
- **主题切换**：通过ThemeToggle组件切换
- **主题持久化**：localStorage存储用户偏好
- **系统主题检测**：自动检测系统偏好色彩模式
- **CSS变量系统**：完整的设计token系统

### 3. HTTP请求系统
- **HttpClient类**：封装的Axios客户端
- **请求拦截器**：自动添加Token、时间戳防缓存
- **响应拦截器**：统一错误处理、业务状态码处理
- **请求重复处理**：自动取消重复请求
- **配置管理**：HttpConfigManager动态配置管理

### 4. 状态管理 (Pinia)
- **主题状态**：`useThemeStore` - 主题切换和持久化
- **组合式API风格**：使用setup语法

### 5. 工具系统
- **日志工具**：多级别彩色日志输出（开发环境）
- **类型系统**：完整的TypeScript类型定义
- **样式变量**：CSS自定义属性系统

## 开发工具配置

### NPM Scripts
```json
{
  "dev": "vite",                    // 开发服务器
  "build": "run-p type-check \"build-only {@}\" --", // 构建项目
  "type-check": "vue-tsc --build", // 类型检查
  "lint": "eslint . --fix",        // 代码检查和修复
  "format": "prettier --write src/", // 代码格式化
  "test:unit": "vitest"            // 单元测试
}
```

### Vite配置特性
- **自动导入**：Vue APIs和Naive UI组件自动导入
- **路径别名**：`@` 指向 `src` 目录
- **开发服务器**：端口动态分配
- **Vue DevTools**：开发环境调试工具

## CSS架构

### 设计系统变量
- **布局尺寸**：header-height(64px), sidebar-width(290px)
- **间距系统**：xs(4px) ~ 2xl(48px)
- **圆角系统**：sm(4px) ~ full(9999px)
- **字体系统**：size(12px-30px), weight(300-700), line-height
- **动画系统**：fast(0.15s), normal(0.3s), slow(0.5s)
- **层级管理**：z-index(1000-1080)

### 主题颜色系统
- **动态主题**：通过CSS类切换(.light/.dark)
- **语义化颜色**：bg-color, text-color, border-color等
- **组件适配**：深度样式覆盖Naive UI组件主题

## API配置
- **默认基础URL**：`http://localhost:8726`
- **环境变量支持**：`VITE_APP_API_URL`
- **请求超时**：10秒
- **自动Token处理**：从localStorage读取

## 开发工具提示

### 常用命令
- `npm run dev` - 启动开发服务器
- `npm run type-check` - TypeScript类型检查
- `npm run lint` - ESLint代码检查

### 日志系统使用
```typescript
import { logB, logN, logS } from '@/utils/logHelper/logUtils'

// 带背景色日志
logB.success('请求成功', data)
logB.error('请求失败', error)

// 命名空间日志
logN.info('HTTP', '请求发送')

// 简单彩色日志
logS.warning('警告信息')
```

### 主题切换
项目支持完整的亮色/暗色主题系统，主题状态通过Pinia管理，支持：
- 手动切换（ThemeToggle组件）
- 系统偏好检测
- 本地存储持久化

### 组件开发规范
- 使用Vue 3 Composition API
- TypeScript类型安全
- Naive UI组件自动导入
- 遵循项目CSS变量系统
- 响应式设计适配

## 注意事项
- 项目使用模块化TypeScript，确保所有导入使用`.ts`扩展名
- 图标使用@vicons/ionicons5，在菜单配置中通过字符串引用
- HTTP请求统一通过httpClient实例，支持请求/响应拦截器
- 样式开发优先使用CSS变量，保持主题一致性
- 开发环境日志系统自动启用，生产环境自动禁用