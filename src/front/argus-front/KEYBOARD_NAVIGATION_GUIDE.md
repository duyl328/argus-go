# 键盘导航完整方案

## 📋 Naive UI 键盘支持现状

### Naive UI 原生支持

Naive UI 对键盘导航的支持比较**基础**，主要问题：

1. **Tab 导航支持不完整**
   - 部分组件缺少 `tabindex` 支持
   - 焦点顺序不总是符合预期
   - 某些组件需要额外配置才能通过 Tab 访问

2. **对话框（Dialog）键盘支持**
   - ✅ `autoFocus` - 支持自动聚焦到输入框
   - ✅ `Enter` - 可以触发 positive 按钮（需要配置）
   - ✅ `Esc` - 关闭对话框（默认支持）
   - ❌ **Alt+快捷键** - 不支持
   - ❌ **默认按钮** - 需要手动实现

3. **表单组件**
   - ✅ Input/Select/DatePicker 支持 Tab 导航
   - ✅ 方向键在 Select/DatePicker 中可用
   - ❌ 缺少统一的 `accessKey` 支持

### Naive UI 键盘相关 API

```typescript
// Dialog 组件
dialog.create({
  autoFocus: true,              // ✅ 自动聚焦第一个可聚焦元素
  closable: true,               // ✅ 显示关闭按钮（可通过 Esc 关闭）
  closeOnEsc: true,             // ✅ Esc 关闭（默认 true）
  maskClosable: true,           // ✅ 点击遮罩关闭

  // ❌ 以下功能需要自己实现
  // defaultButton: 'positive',  // 不支持
  // shortcuts: { 'Alt+S': ... } // 不支持
})

// Input 组件
h(NInput, {
  autofocus: true,              // ✅ 自动聚焦
  selectOnFocus: true,          // ✅ 聚焦时选中文本
  onKeydown: handleKeydown,     // ✅ 键盘事件
})

// Button 组件
h(NButton, {
  keyboard: true,               // ✅ 键盘可访问
  // accesskey: 'S',            // ❌ 不支持原生 accesskey
})
```

---

## 🎯 推荐的键盘导航方案

### 方案 1: 自定义全局键盘管理器（推荐）

**适用场景**: 桌面应用，需要完全控制键盘行为

#### 实现思路

创建一个全局键盘管理器，统一处理所有快捷键：

```typescript
// composables/useKeyboardManager.ts
import { onMounted, onUnmounted, ref } from 'vue'

interface ShortcutHandler {
  key: string              // 'Enter', 'Escape', 'Alt+S'
  handler: (event: KeyboardEvent) => void
  priority?: number        // 优先级（数字越大越优先）
  context?: string         // 上下文（dialog, form, etc）
}

export function useKeyboardManager() {
  const shortcuts = ref<ShortcutHandler[]>([])
  const currentContext = ref<string>('global')

  function register(shortcut: ShortcutHandler) {
    shortcuts.value.push(shortcut)
    // 按优先级排序
    shortcuts.value.sort((a, b) => (b.priority ?? 0) - (a.priority ?? 0))
  }

  function unregister(key: string) {
    shortcuts.value = shortcuts.value.filter(s => s.key !== key)
  }

  function handleKeyDown(event: KeyboardEvent) {
    // 构建快捷键字符串
    const modifiers = []
    if (event.ctrlKey) modifiers.push('Ctrl')
    if (event.altKey) modifiers.push('Alt')
    if (event.shiftKey) modifiers.push('Shift')
    if (event.metaKey) modifiers.push('Meta')

    const key = event.key
    const shortcutString = modifiers.length > 0
      ? `${modifiers.join('+')}+${key}`
      : key

    // 查找匹配的处理器
    for (const shortcut of shortcuts.value) {
      // 检查上下文
      if (shortcut.context && shortcut.context !== currentContext.value) {
        continue
      }

      // 匹配快捷键
      if (shortcut.key === shortcutString) {
        event.preventDefault()
        event.stopPropagation()
        shortcut.handler(event)
        return
      }
    }
  }

  onMounted(() => {
    document.addEventListener('keydown', handleKeyDown)
  })

  onUnmounted(() => {
    document.removeEventListener('keydown', handleKeyDown)
  })

  return {
    register,
    unregister,
    setContext: (ctx: string) => { currentContext.value = ctx },
    getContext: () => currentContext.value
  }
}
```

#### 使用示例

```typescript
// 在组件中使用
const keyboard = useKeyboardManager()

// 注册对话框快捷键
keyboard.register({
  key: 'Enter',
  handler: () => handleSave(),
  priority: 10,
  context: 'dialog'
})

keyboard.register({
  key: 'Alt+S',
  handler: () => handleSave(),
  priority: 10,
  context: 'dialog'
})

keyboard.register({
  key: 'Escape',
  handler: () => handleCancel(),
  priority: 10,
  context: 'dialog'
})

// 设置上下文
keyboard.setContext('dialog')

// 对话框关闭时清理
onUnmounted(() => {
  keyboard.unregister('Enter')
  keyboard.unregister('Alt+S')
  keyboard.setContext('global')
})
```

---

### 方案 2: 增强 Naive UI Dialog（实用）

**适用场景**: 与 Naive UI 集成，不需要完全自定义

#### 创建增强版 Dialog 工具

```typescript
// utils/enhancedDialog.ts
import { useDialog, useMessage } from 'naive-ui'
import { h, ref } from 'vue'

interface DialogButton {
  text: string
  onClick: () => void | Promise<void>
  type?: 'primary' | 'default' | 'error'
  shortcut?: string  // 'Enter', 'Alt+S', etc
  autoFocus?: boolean
}

interface EnhancedDialogOptions {
  title: string
  content: string | (() => VNode)
  buttons: DialogButton[]
  onClose?: () => void
}

export function useEnhancedDialog() {
  const dialog = useDialog()
  const message = useMessage()

  function show(options: EnhancedDialogOptions) {
    const shortcuts = new Map<string, () => void>()
    let dialogInstance: any = null

    // 收集快捷键
    options.buttons.forEach(btn => {
      if (btn.shortcut) {
        shortcuts.set(btn.shortcut, btn.onClick)
      }
    })

    // 键盘事件处理
    function handleKeyDown(event: KeyboardEvent) {
      // 构建快捷键字符串
      const key = event.key
      let shortcutKey = key

      if (event.altKey) shortcutKey = `Alt+${key}`
      else if (event.ctrlKey) shortcutKey = `Ctrl+${key}`
      else if (event.shiftKey) shortcutKey = `Shift+${key}`

      // 查找匹配的快捷键
      const handler = shortcuts.get(shortcutKey)
      if (handler) {
        event.preventDefault()
        event.stopPropagation()
        handler()
        dialogInstance?.destroy()
      }
    }

    // 注册键盘监听
    document.addEventListener('keydown', handleKeyDown)

    // 创建对话框
    dialogInstance = dialog.create({
      title: options.title,
      content: options.content,
      closable: true,
      closeOnEsc: true,
      autoFocus: true,

      // 在对话框关闭时清理
      onAfterLeave: () => {
        document.removeEventListener('keydown', handleKeyDown)
        options.onClose?.()
      },

      // 自定义 action 按钮
      action: () => {
        return h('div',
          { style: 'display: flex; gap: 8px; justify-content: flex-end;' },
          options.buttons.map(btn =>
            h(NButton, {
              type: btn.type || 'default',
              onClick: async () => {
                await btn.onClick()
                dialogInstance?.destroy()
              },
              // 显示快捷键提示
              ...(btn.shortcut ? {
                // 在按钮文本中显示快捷键
                default: () => `${btn.text} (${btn.shortcut})`
              } : {
                default: () => btn.text
              })
            })
          )
        )
      }
    })

    return dialogInstance
  }

  return { show }
}
```

#### 使用示例

```typescript
const enhancedDialog = useEnhancedDialog()

enhancedDialog.show({
  title: '新建文件夹',
  content: () => h(NInput, {
    defaultValue: '新建文件夹',
    autofocus: true,
    selectOnFocus: true
  }),
  buttons: [
    {
      text: '取消',
      onClick: () => {},
      shortcut: 'Escape'
    },
    {
      text: '创建',
      type: 'primary',
      onClick: async () => {
        await createFolder()
      },
      shortcut: 'Enter',
      autoFocus: true
    },
    {
      text: '保存',
      type: 'primary',
      onClick: async () => {
        await saveFolder()
      },
      shortcut: 'Alt+S'
    }
  ]
})
```

---

### 方案 3: Tab 导航增强

**适用场景**: 表单、设置页面等需要 Tab 导航的场景

#### 自动 Tab 索引管理

```typescript
// composables/useTabNavigation.ts
import { onMounted, onUnmounted, ref } from 'vue'

export function useTabNavigation() {
  const focusableElements = ref<HTMLElement[]>([])
  const currentIndex = ref(0)

  function scanFocusableElements(container?: HTMLElement) {
    const root = container || document.body

    // 查找所有可聚焦元素
    const selector = [
      'button:not([disabled])',
      'input:not([disabled])',
      'select:not([disabled])',
      'textarea:not([disabled])',
      '[tabindex]:not([tabindex="-1"])',
      'a[href]'
    ].join(',')

    focusableElements.value = Array.from(
      root.querySelectorAll<HTMLElement>(selector)
    )
  }

  function focusNext() {
    if (focusableElements.value.length === 0) return

    currentIndex.value = (currentIndex.value + 1) % focusableElements.value.length
    focusableElements.value[currentIndex.value]?.focus()
  }

  function focusPrevious() {
    if (focusableElements.value.length === 0) return

    currentIndex.value = currentIndex.value - 1
    if (currentIndex.value < 0) {
      currentIndex.value = focusableElements.value.length - 1
    }
    focusableElements.value[currentIndex.value]?.focus()
  }

  function handleKeyDown(event: KeyboardEvent) {
    if (event.key === 'Tab') {
      event.preventDefault()

      if (event.shiftKey) {
        focusPrevious()
      } else {
        focusNext()
      }
    }
  }

  onMounted(() => {
    document.addEventListener('keydown', handleKeyDown)
  })

  onUnmounted(() => {
    document.removeEventListener('keydown', handleKeyDown)
  })

  return {
    scanFocusableElements,
    focusNext,
    focusPrevious,
    focusableElements
  }
}
```

---

## 🎨 桌面应用级键盘导航标准

### 标准快捷键映射

| 场景 | 快捷键 | 功能 |
|------|--------|------|
| **对话框** | | |
| | `Enter` | 确认（触发默认按钮） |
| | `Esc` | 取消/关闭 |
| | `Alt+字母` | 触发对应按钮 |
| | `Tab` | 在控件间切换 |
| **表单** | | |
| | `Tab` | 下一个控件 |
| | `Shift+Tab` | 上一个控件 |
| | `Enter` | 提交表单 |
| | `Ctrl+Enter` | 快速提交 |
| **列表/网格** | | |
| | `↑↓←→` | 导航 |
| | `Enter` | 打开/选择 |
| | `Space` | 预览 |
| | `Ctrl+A` | 全选 |
| | `Delete` | 删除 |
| **菜单** | | |
| | `Alt` | 激活菜单栏 |
| | `Alt+字母` | 打开对应菜单 |
| | `↑↓` | 菜单项导航 |
| | `Enter` | 选择菜单项 |

### 实现示例：完整的对话框键盘支持

```vue
<template>
  <n-dialog
    v-model:show="visible"
    title="新建文件夹"
    :closable="true"
    :close-on-esc="true"
    :auto-focus="true"
    @after-enter="handleDialogEnter"
    @after-leave="handleDialogLeave"
  >
    <n-form ref="formRef" :model="form">
      <n-form-item label="文件夹名称" path="name">
        <n-input
          ref="inputRef"
          v-model:value="form.name"
          placeholder="请输入文件夹名称"
          @keydown.enter="handleEnter"
        />
      </n-form-item>
    </n-form>

    <template #action>
      <div class="dialog-actions">
        <n-button @click="handleCancel" @keydown.alt.c="handleCancel">
          取消 <span class="shortcut-hint">Alt+C</span>
        </n-button>
        <n-button
          type="primary"
          :loading="loading"
          @click="handleConfirm"
          @keydown.alt.s="handleConfirm"
        >
          创建 <span class="shortcut-hint">Alt+S / Enter</span>
        </n-button>
      </div>
    </template>
  </n-dialog>
</template>

<script setup lang="ts">
import { ref, nextTick } from 'vue'

const visible = ref(false)
const inputRef = ref()
const formRef = ref()
const loading = ref(false)
const form = ref({ name: '新建文件夹' })

// 对话框进入后聚焦并选中文本
function handleDialogEnter() {
  nextTick(() => {
    inputRef.value?.focus()
    inputRef.value?.select()
  })

  // 注册全局快捷键
  document.addEventListener('keydown', handleGlobalKeyDown)
}

// 对话框离开时清理
function handleDialogLeave() {
  document.removeEventListener('keydown', handleGlobalKeyDown)
}

// 全局快捷键处理
function handleGlobalKeyDown(event: KeyboardEvent) {
  // Alt+S - 保存
  if (event.altKey && event.key === 's') {
    event.preventDefault()
    handleConfirm()
  }
  // Alt+C - 取消
  else if (event.altKey && event.key === 'c') {
    event.preventDefault()
    handleCancel()
  }
}

// Enter 键处理
function handleEnter(event: KeyboardEvent) {
  event.preventDefault()
  handleConfirm()
}

async function handleConfirm() {
  // 验证表单
  await formRef.value?.validate()

  loading.value = true
  try {
    // 执行创建逻辑
    await createFolder(form.value.name)
    visible.value = false
  } finally {
    loading.value = false
  }
}

function handleCancel() {
  visible.value = false
}
</script>

<style scoped>
.dialog-actions {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
}

.shortcut-hint {
  font-size: 12px;
  opacity: 0.6;
  margin-left: 4px;
}
</style>
```

---

## 🔧 推荐的整体方案

### For 桌面应用（Electron/Tauri）

**推荐**: 方案 1（全局键盘管理器） + 方案 2（增强 Dialog）

**理由**:
1. 完全控制所有快捷键
2. 可以实现 Alt+快捷键
3. 可以统一管理上下文
4. 符合桌面应用标准

### 实施步骤

1. **创建全局键盘管理器** (`composables/useKeyboardManager.ts`)
2. **创建增强 Dialog 工具** (`utils/enhancedDialog.ts`)
3. **在 App.vue 中初始化全局键盘管理器**
4. **逐步迁移现有对话框到增强版本**
5. **添加快捷键提示 UI**（在按钮旁显示 Alt+S 等）

---

## 📝 最佳实践

### 1. 快捷键一致性

```typescript
// 定义全局快捷键常量
export const SHORTCUTS = {
  CONFIRM: 'Enter',
  CANCEL: 'Escape',
  SAVE: 'Alt+S',
  DELETE: 'Delete',
  REFRESH: 'F5',
  // ...
}
```

### 2. 快捷键提示

```vue
<!-- 在按钮中显示快捷键 -->
<n-button type="primary">
  保存 <kbd>Alt+S</kbd>
</n-button>

<style>
kbd {
  font-size: 11px;
  padding: 2px 4px;
  background: rgba(0,0,0,0.1);
  border-radius: 3px;
  margin-left: 8px;
}
</style>
```

### 3. 焦点管理

```typescript
// 对话框打开时保存当前焦点
let previousFocus: HTMLElement | null = null

function openDialog() {
  previousFocus = document.activeElement as HTMLElement
  // 打开对话框...
}

function closeDialog() {
  // 恢复之前的焦点
  previousFocus?.focus()
}
```

### 4. 键盘可访问性

```vue
<!-- 所有交互元素都应该可以通过键盘访问 -->
<div
  role="button"
  tabindex="0"
  @click="handleClick"
  @keydown.enter="handleClick"
  @keydown.space="handleClick"
>
  可点击的区域
</div>
```

---

## 🎯 总结

### Naive UI 的限制
- ❌ 原生不支持 Alt+快捷键
- ❌ 没有内置的快捷键管理
- ❌ Tab 导航支持不完整
- ✅ 基础的 Enter/Esc 支持
- ✅ 自动聚焦支持

### 推荐方案
1. **全局键盘管理器** - 统一管理所有快捷键
2. **增强 Dialog** - 扩展 Naive UI Dialog 功能
3. **自定义焦点管理** - 实现完整的 Tab 导航
4. **标准化快捷键** - 遵循桌面应用标准

### 优先级
1. ⭐⭐⭐ 实现全局键盘管理器（核心）
2. ⭐⭐⭐ 增强对话框快捷键支持
3. ⭐⭐ 添加快捷键提示 UI
4. ⭐ 实现完整的 Tab 导航
