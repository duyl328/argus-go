# 文件管理器快捷键说明

> **设计准则**: 本项目采用桌面应用优先的设计理念，所有应用快捷键优先级高于系统/浏览器快捷键。

## 🎯 桌面应用快捷键策略

### 快捷键优先级
1. **应用快捷键 > 系统快捷键 > 浏览器快捷键**
2. 所有快捷键都会阻止默认行为和事件冒泡
3. 与系统冲突的快捷键，以应用为准（适配未来的 Electron/Tauri 桌面应用）

### 已接管的系统快捷键
| 快捷键 | 系统默认行为 | 应用行为 |
|--------|-------------|---------|
| `F5` | 浏览器刷新 | 刷新当前文件列表 |
| `Backspace` | 浏览器后退 | 返回上一级目录 |
| `Ctrl+C` | 系统复制 | 复制选中的文件/文件夹 |
| `Ctrl+X` | 系统剪切 | 剪切选中的文件/文件夹 |
| `Ctrl+V` | 系统粘贴 | 粘贴文件/文件夹 |
| `Delete` | 浏览器行为 | 删除文件/文件夹 |

## 📋 快捷键列表

### 导航快捷键

| 快捷键 | 功能 | 说明 |
|--------|------|------|
| `↑` `↓` | 上下导航 | 列表视图：上下移动<br>网格视图：按列移动 |
| `←` `→` | 左右导航 | 仅在网格视图中有效 |
| `Enter` | 进入文件夹 | 选中文件夹时，按 Enter 进入该文件夹 |
| `Backspace` | 返回上一级 | 返回到父目录（无选中项时） |
| `Space` | 快速预览 | 打开/关闭快速预览窗口 |
| `Esc` | 取消选择 | 清除所有选中项 |

### 文件操作快捷键

| 快捷键 | 功能 | 说明 |
|--------|------|------|
| `Delete` | 删除 | 删除选中的文件/文件夹 |
| `Ctrl+C` | 复制 | 复制选中的文件/文件夹到剪贴板 |
| `Ctrl+X` | 剪切 | 剪切选中的文件/文件夹到剪贴板 |
| `Ctrl+V` | 粘贴 | 粘贴剪贴板中的文件/文件夹到当前目录 |
| `F5` | 刷新 | 刷新当前目录 |

### 选择快捷键

| 快捷键 | 功能 | 说明 |
|--------|------|------|
| `Ctrl+A` | 全选 | 选中当前目录下的所有项目 |
| `Shift+↑/↓` | 范围选择 | 从当前焦点项扩展选择 |
| `Ctrl+↑/↓` | 移动焦点 | 移动焦点但不改变选择 |
| `Ctrl+Click` | 多选 | 点击添加/移除选择项 |
| `Shift+Click` | 范围选择 | 从上次点击到当前位置范围选择 |

## 🎯 使用场景示例

### 场景 1: 快速删除文件
1. 使用方向键选中要删除的文件
2. 按 `Delete` 键
3. 在确认对话框中点击"删除"

### 场景 2: 复制文件到另一个位置
1. 选中要复制的文件（可以用 `Ctrl+Click` 多选）
2. 按 `Ctrl+C` 复制
3. 使用面板切换或导航到目标文件夹
4. 按 `Ctrl+V` 粘贴

### 场景 3: 快速浏览文件夹
1. 使用方向键选中文件夹
2. 按 `Enter` 进入
3. 浏览完成后按 `Backspace` 返回上一级

### 场景 4: 快速预览文件
1. 使用方向键选中文件
2. 按 `Space` 打开快速预览
3. 再次按 `Space` 关闭预览

## 🔧 快捷键配置

所有快捷键在 `src/composables/fileManager/useKeyboardNav.ts` 中定义。

### 当前支持的快捷键：

```typescript
// 导航
- ArrowUp/Down/Left/Right  // 方向键导航
- Enter                     // 进入文件夹
- Space                     // 快速预览
- Backspace                 // 返回上一级
- Escape                    // 取消选择

// 文件操作
- Delete                    // 删除
- Ctrl+C                    // 复制
- Ctrl+X                    // 剪切
- Ctrl+V                    // 粘贴
- F5                        // 刷新

// 选择
- Ctrl+A                    // 全选
- Shift + 方向键            // 范围选择
- Ctrl + 方向键             // 移动焦点
```

## ⚙️ 技术实现

### 事件流程
1. **useKeyboardNav** composable 监听全局键盘事件
2. 检查当前面板是否激活（`isActive`）
3. 检查事件来源是否为输入框（跳过输入框事件）
4. 根据按键类型触发相应的回调函数
5. FilePane emit 事件到 FileManager
6. FileManager 执行相应的操作函数

### 文件结构
```
src/
├── composables/
│   └── fileManager/
│       └── useKeyboardNav.ts          # 快捷键逻辑
├── components/
│   └── FileManager/
│       ├── FilePane.vue               # 面板组件（emit 事件）
│       └── FileManager.vue            # 管理器（处理操作）
```

### 添加新快捷键

如果需要添加新的快捷键，按以下步骤：

1. **修改 `useKeyboardNav.ts`**:
```typescript
interface UseKeyboardNavOptions {
  // ... 现有选项
  onYourAction?: () => void  // 添加新回调
}

// 在 handleKeyDown 中添加
case 'YourKey':
  if (onYourAction) {
    event.preventDefault()
    onYourAction()
  }
  break
```

2. **修改 `FilePane.vue`**:
```typescript
// 添加 emit 定义
const emit = defineEmits<{
  yourAction: []
}>()

// 在 useKeyboardNav 调用中添加
useKeyboardNav({
  // ... 现有选项
  onYourAction: () => {
    emit('yourAction')
  }
})
```

3. **修改 `FileManager.vue`**:
```vue
<!-- 模板中添加事件监听 -->
<FilePane
  @your-action="handleYourAction"
/>
```

```typescript
// 添加处理函数
function handleYourAction() {
  // 实现你的功能
}
```

## 📝 注意事项

### 快捷键行为
1. **输入框例外**: 在输入框中输入时，大部分快捷键会被忽略（`Esc` 除外）
2. **面板激活**: 只有激活的面板才会响应快捷键
3. **路径验证**: 某些操作（如粘贴、新建）会验证当前路径是否有效
4. **Mac 兼容**: `Ctrl` 键在 Mac 上自动映射为 `Cmd` 键（`event.metaKey`）

### 系统快捷键禁用
以下快捷键会完全接管系统/浏览器的默认行为：

- **F5**: 不会刷新浏览器页面，只刷新文件列表
- **Backspace**: 不会后退浏览器历史，只返回上一级目录
- **Ctrl+C/X/V**: 不会操作系统剪贴板，只操作应用内剪贴板
- **Delete**: 不会触发浏览器的删除行为
- **方向键**: 不会滚动页面，只在文件列表中导航

### 桌面应用特性
本应用设计为未来打包成桌面应用（Electron/Tauri），因此：
- 快捷键体验类似 Windows 资源管理器
- 不受浏览器安全限制
- 可以访问本地文件系统
- 可以集成系统托盘和原生 API

## 🐛 常见问题

### Q: 快捷键不生效？
A: 检查以下几点：
1. 当前面板是否已激活（点击面板激活）
2. 焦点是否在输入框中
3. 浏览器控制台是否有错误

### Q: Delete 键删除了浏览器历史？
A: 这是正常的，我们已经使用 `event.preventDefault()` 阻止了默认行为。如果仍然发生，请检查是否有其他扩展程序干扰。

### Q: 如何禁用某个快捷键？
A: 在 `useKeyboardNav.ts` 中注释掉对应的 case 分支即可。

## 📊 性能优化

- 使用 `onMounted/onUnmounted` 正确管理事件监听器
- 事件处理中添加了早期返回检查，避免不必要的计算
- 使用 `event.preventDefault()` 避免浏览器默认行为触发
- 所有快捷键操作都是异步的，不会阻塞 UI

## 🔄 更新历史

- **2025-10-05**: 添加完整的快捷键支持
  - 新增 Delete、Ctrl+C/X/V、F5、Backspace 快捷键
  - 优化键盘导航体验
  - 添加 Mac Cmd 键支持
