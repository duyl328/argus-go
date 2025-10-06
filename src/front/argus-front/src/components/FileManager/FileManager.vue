<template>
  <div class="file-manager">
    <!-- Toolbar -->
    <div class="toolbar">
      <div class="toolbar-left">
        <!-- 新建按钮（带下拉菜单） -->
        <div class="dropdown-wrapper">
          <button class="btn btn-primary" @click="toggleNewMenu">
            <span class="icon">+</span>
            <span>新建</span>
            <span class="arrow">▼</span>
          </button>
          <div v-if="showNewMenu" class="dropdown-menu" @click.stop>
            <div class="dropdown-item" @click="handleNewFolder">
              <span class="item-icon">📁</span>
              <span>新建文件夹</span>
            </div>
            <div class="dropdown-item disabled" title="即将推出">
              <span class="item-icon">💭</span>
              <span>新建回忆</span>
            </div>
          </div>
        </div>
        <div class="divider"></div>
        <button
          class="btn"
          :disabled="!hasSelection"
          @click="handleCopy"
        >
          复制
        </button>
        <button
          class="btn"
          :disabled="!hasSelection"
          @click="handleCut"
        >
          剪切
        </button>
        <button
          class="btn"
          :disabled="!hasClipboard"
          @click="handlePaste"
        >
          粘贴
        </button>
        <button
          class="btn"
          :disabled="!hasSelection"
          @click="handleDelete"
        >
          删除
        </button>
      </div>
      <div class="toolbar-right">
        <!-- View Controls -->
        <div class="view-controls">
          <button
            :class="['btn-view', { active: activeConfig.viewMode === 'grid' }]"
            @click="updateActiveViewMode('grid')"
          >
            网格
          </button>
          <button
            :class="['btn-view', { active: activeConfig.viewMode === 'list' }]"
            @click="updateActiveViewMode('list')"
          >
            列表
          </button>
        </div>

        <!-- Thumbnail Size -->
        <div v-if="activeConfig.viewMode === 'grid'" class="thumbnail-controls">
          <span class="label">大小:</span>
          <button
            v-for="size in (['small', 'medium', 'large'] as ThumbnailSize[])"
            :key="size"
            :class="['btn-size', { active: activeConfig.thumbnailSize === size }]"
            @click="updateActiveThumbnailSize(size)"
          >
            {{ size === 'small' ? '小' : size === 'medium' ? '中' : '大' }}
          </button>
        </div>

        <!-- Layout Controls -->
        <div class="layout-controls">
          <button
            :class="['btn-layout', { active: layoutMode === 'single' }]"
            @click="layoutMode = 'single'"
            title="单面板"
          >
            ▢
          </button>
          <button
            :class="['btn-layout', { active: layoutMode === 'horizontal' }]"
            @click="layoutMode = 'horizontal'"
            title="左右双面板"
          >
            ▭
          </button>
          <button
            :class="['btn-layout', { active: layoutMode === 'vertical' }]"
            @click="layoutMode = 'vertical'"
            title="上下双面板"
          >
            ▯
          </button>
        </div>
      </div>
    </div>

    <!-- Main Content -->
    <div :class="['main-content', `layout-${layoutMode}`]">
      <FilePane
        ref="leftPaneRef"
        pane-id="left"
        :view-mode="paneConfigs.left.viewMode"
        :thumbnail-size="paneConfigs.left.thumbnailSize"
        :sort-options="paneConfigs.left.sortOptions"
        :filter-options="paneConfigs.left.filterOptions"
        :is-active="activePane === 'left'"
        :is-dialog-open="isDialogOpen"
        :use-real-api="props.useRealApi"
        @activate="activePane = 'left'"
        @delete="handleDelete"
        @copy="handleCopy"
        @cut="handleCut"
        @paste="handlePaste"
        @refresh="handleRefresh"
        @go-back="handleGoBack"
      />

      <div
        v-if="layoutMode === 'horizontal' || layoutMode === 'vertical'"
        :class="['splitter', layoutMode === 'vertical' ? 'splitter-vertical' : '']"
        @mousedown="handleSplitterMouseDown"
      ></div>

      <FilePane
        v-if="layoutMode === 'horizontal' || layoutMode === 'vertical'"
        ref="rightPaneRef"
        pane-id="right"
        :view-mode="paneConfigs.right.viewMode"
        :thumbnail-size="paneConfigs.right.thumbnailSize"
        :sort-options="paneConfigs.right.sortOptions"
        :filter-options="paneConfigs.right.filterOptions"
        :is-active="activePane === 'right'"
        :is-dialog-open="isDialogOpen"
        :use-real-api="props.useRealApi"
        @activate="activePane = 'right'"
        @delete="handleDelete"
        @copy="handleCopy"
        @cut="handleCut"
        @paste="handlePaste"
        @refresh="handleRefresh"
        @go-back="handleGoBack"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, h, nextTick } from 'vue'
import { useDialog, useMessage, NInput } from 'naive-ui'
import FilePane from './FilePane.vue'
import type { ViewMode, ThumbnailSize, LayoutMode, PaneId, SortOptions, FilterOptions, FileItem } from './types'
import { fileSystemService } from '@/services/fileSystemService'

// Naive UI
const dialog = useDialog()
const message = useMessage()

// Props
const props = withDefaults(defineProps<{
  useRealApi?: boolean  // 是否使用真实 API
}>(), {
  useRealApi: false  // 默认使用 mock 数据
})

// FilePane refs
const leftPaneRef = ref<InstanceType<typeof FilePane>>()
const rightPaneRef = ref<InstanceType<typeof FilePane>>()

// 新建菜单状态
const showNewMenu = ref(false)

// 对话框状态（用于禁用背景快捷键）
const isDialogOpen = ref(false)

// 剪贴板状态
const clipboard = ref<{
  operation: 'copy' | 'cut' | null
  items: FileItem[]
  sourcePath: string
}>({
  operation: null,
  items: [],
  sourcePath: ''
})

// 每个面板的独立配置
const paneConfigs = ref({
  left: {
    viewMode: 'grid' as ViewMode,
    thumbnailSize: 'medium' as ThumbnailSize,
    sortOptions: {
      field: 'name',
      order: 'asc'
    } as SortOptions,
    filterOptions: {
      nameQuery: '',
      fileType: 'all'
    } as FilterOptions
  },
  right: {
    viewMode: 'grid' as ViewMode,
    thumbnailSize: 'medium' as ThumbnailSize,
    sortOptions: {
      field: 'name',
      order: 'asc'
    } as SortOptions,
    filterOptions: {
      nameQuery: '',
      fileType: 'all'
    } as FilterOptions
  }
})

const layoutMode = ref<LayoutMode>('single')
const activePane = ref<PaneId>('left')

// 当前活动面板的配置（用于工具栏显示）
const activeConfig = computed(() => paneConfigs.value[activePane.value])

// 当前活动面板的引用
const activePaneRef = computed(() =>
  activePane.value === 'left' ? leftPaneRef.value : rightPaneRef.value
)

// 计算属性：是否有选中项
const hasSelection = computed(() => {
  const pane = activePaneRef.value
  return pane && pane.getSelectedItems && pane.getSelectedItems().length > 0
})

// 计算属性：是否有剪贴板内容
const hasClipboard = computed(() => {
  return clipboard.value.operation !== null && clipboard.value.items.length > 0
})

// 获取当前路径
function getCurrentPath(): string {
  const pane = activePaneRef.value
  if (!pane || !pane.getCurrentPath) {
    return ''
  }
  return pane.getCurrentPath()
}

// 更新当前活动面板的配置
function updateActiveViewMode(mode: ViewMode) {
  paneConfigs.value[activePane.value].viewMode = mode
}

function updateActiveThumbnailSize(size: ThumbnailSize) {
  paneConfigs.value[activePane.value].thumbnailSize = size
}

// ==================== 新建菜单 ====================

// 切换新建菜单
function toggleNewMenu() {
  showNewMenu.value = !showNewMenu.value
}

// 关闭新建菜单
function closeNewMenu() {
  showNewMenu.value = false
}

// 新建文件夹
function handleNewFolder() {
  closeNewMenu()

  const currentPath = getCurrentPath()
  if (!currentPath) {
    message.error('无法获取当前路径')
    return
  }

  // 检查路径是否有效（不能在根级别或驱动器列表创建文件夹）
  if (currentPath === '/' || currentPath === '所有驱动器' || !currentPath.includes('\\') && !currentPath.includes('/')) {
    message.warning('请先选择一个具体的文件夹或驱动器')
    return
  }

  let inputValue = '新建文件夹'

  // 标记对话框已打开
  isDialogOpen.value = true

  // 使用 Naive UI 的 Dialog
  const d = dialog.create({
    title: '新建文件夹',
    style: {
      width: '460px'
    },
    onAfterLeave: () => {
      // 对话框关闭后恢复快捷键
      isDialogOpen.value = false
    },
    content: () => {
      // 创建输入框组件
      const input = h(NInput, {
        defaultValue: inputValue,
        placeholder: '请输入文件夹名称',
        autofocus: true,
        selectOnFocus: true,
        size: 'large',
        onUpdateValue: (v: string) => {
          inputValue = v
        },
        onKeydown: (e: KeyboardEvent) => {
          if (e.key === 'Enter') {
            e.preventDefault()
            // 模拟点击确定按钮
            const okBtn = document.querySelector('.n-dialog__action .n-button--primary') as HTMLButtonElement
            okBtn?.click()
          }
        },
        onVnodeMounted: (vnode: any) => {
          // 挂载后自动选中所有文本
          nextTick(() => {
            const inputElement = vnode.el?.querySelector('input')
            if (inputElement) {
              inputElement.focus()
              inputElement.select()
            }
          })
        }
      })

      return h('div', { style: 'padding: 8px 0;' }, input)
    },
    positiveText: '创建',
    negativeText: '取消',
    onPositiveClick: async () => {
      const folderName = inputValue.trim()

      if (!folderName) {
        message.warning('请输入文件夹名称')
        return false // 阻止关闭对话框
      }

      try {
        const newFolderPath = `${currentPath}${currentPath.endsWith('\\') || currentPath.endsWith('/') ? '' : '\\'}${folderName}`

        if (props.useRealApi) {
          await fileSystemService.createDirectory(newFolderPath)
          message.success(`文件夹创建成功: ${folderName}`)
          console.log('✅ 文件夹创建成功:', newFolderPath)
        } else {
          // Mock 模式：需要手动刷新
          const pane = activePaneRef.value
          if (pane && pane.refresh) {
            pane.refresh()
          }
          message.success(`文件夹创建成功: ${folderName}`)
        }
        return true // 允许关闭对话框
      } catch (error: any) {
        console.error('❌ 创建文件夹失败:', error)
        // 提取详细错误信息
        const errorMsg = error.message || error.toString()
        message.error(`创建失败: ${errorMsg}`)
        return false // 阻止关闭对话框
      }
    }
  })
}

// ==================== 复制/剪切/粘贴 ====================

// 复制
function handleCopy() {
  const pane = activePaneRef.value
  if (!pane || !pane.getSelectedItems) {
    return
  }

  const selectedItems = pane.getSelectedItems()
  if (selectedItems.length === 0) {
    message.warning('请先选择要复制的项目')
    return
  }

  const currentPath = getCurrentPath()
  if (!currentPath || currentPath === '/' || currentPath === '所有驱动器') {
    message.warning('无法在此位置进行复制操作')
    return
  }

  clipboard.value = {
    operation: 'copy',
    items: selectedItems,
    sourcePath: currentPath
  }

  message.success(`已复制 ${selectedItems.length} 个项目`)
  console.log('📋 已复制到剪贴板:', selectedItems.map(item => item.name))
}

// 剪切
function handleCut() {
  const pane = activePaneRef.value
  if (!pane || !pane.getSelectedItems) {
    return
  }

  const selectedItems = pane.getSelectedItems()
  if (selectedItems.length === 0) {
    message.warning('请先选择要剪切的项目')
    return
  }

  const currentPath = getCurrentPath()
  if (!currentPath || currentPath === '/' || currentPath === '所有驱动器') {
    message.warning('无法在此位置进行剪切操作')
    return
  }

  clipboard.value = {
    operation: 'cut',
    items: selectedItems,
    sourcePath: currentPath
  }

  message.success(`已剪切 ${selectedItems.length} 个项目`)
  console.log('✂️ 已剪切到剪贴板:', selectedItems.map(item => item.name))
}

// 粘贴
async function handlePaste() {
  if (!hasClipboard.value) {
    message.warning('剪贴板为空')
    return
  }

  const currentPath = getCurrentPath()
  if (!currentPath) {
    message.error('无法获取当前路径')
    return
  }

  // 检查路径是否有效（不能粘贴到根级别或驱动器列表）
  if (currentPath === '/' || currentPath === '所有驱动器' || !currentPath.includes('\\') && !currentPath.includes('/')) {
    message.warning('请先选择一个具体的文件夹')
    return
  }

  const { operation, items, sourcePath } = clipboard.value

  try {
    for (const item of items) {
      const sourceFull = item.path || `${sourcePath}\\${item.name}`
      const destFull = `${currentPath}${currentPath.endsWith('\\') || currentPath.endsWith('/') ? '' : '\\'}${item.name}`

      if (props.useRealApi) {
        if (operation === 'copy') {
          await fileSystemService.copyItem(sourceFull, destFull, undefined, false)
          console.log('✅ 复制成功:', item.name)
        } else if (operation === 'cut') {
          await fileSystemService.moveItem(sourceFull, destFull, undefined, false)
          console.log('✅ 移动成功:', item.name)
        }
      } else {
        // Mock 模式
        console.log(`Mock ${operation}:`, sourceFull, '->', destFull)
      }
    }

    // 剪切后清空剪贴板
    if (operation === 'cut') {
      clipboard.value = {
        operation: null,
        items: [],
        sourcePath: ''
      }
    }

    // Mock 模式需要手动刷新
    if (!props.useRealApi) {
      const pane = activePaneRef.value
      if (pane && pane.refresh) {
        pane.refresh()
      }
    }

    message.success(`${operation === 'copy' ? '复制' : '移动'}完成`)
  } catch (error: any) {
    console.error(`❌ ${operation === 'copy' ? '复制' : '移动'}失败:`, error)
    // 提取详细错误信息
    const errorMsg = error.message || error.toString()
    message.error(`${operation === 'copy' ? '复制' : '移动'}失败: ${errorMsg}`)
  }
}

// ==================== 删除 ====================

// 删除
function handleDelete() {
  const pane = activePaneRef.value
  if (!pane || !pane.getSelectedItems) {
    return
  }

  const selectedItems = pane.getSelectedItems()
  if (selectedItems.length === 0) {
    message.warning('请先选择要删除的项目')
    return
  }

  const currentPath = getCurrentPath()
  if (!currentPath || currentPath === '/' || currentPath === '所有驱动器') {
    message.warning('无法在此位置进行删除操作')
    return
  }

  const itemNames = selectedItems.map(item => item.name).join(', ')

  // 标记对话框已打开
  isDialogOpen.value = true

  // 使用 Naive UI 的确认对话框
  dialog.warning({
    onAfterLeave: () => {
      // 对话框关闭后恢复快捷键
      isDialogOpen.value = false
    },
    title: '确认删除',
    content: `确定要删除以下 ${selectedItems.length} 个项目吗？\n\n${itemNames}`,
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        for (const item of selectedItems) {
          const itemPath = item.path || `${currentPath}\\${item.name}`

          if (props.useRealApi) {
            await fileSystemService.deleteItem(itemPath, undefined, true)
            console.log('✅ 删除成功:', item.name)
          } else {
            // Mock 模式
            console.log('Mock delete:', itemPath)
          }
        }

        // Mock 模式需要手动刷新
        if (!props.useRealApi) {
          if (pane && pane.refresh) {
            pane.refresh()
          }
        }

        message.success(`成功删除 ${selectedItems.length} 个项目`)
      } catch (error: any) {
        console.error('❌ 删除失败:', error)
        // 提取详细错误信息
        const errorMsg = error.message || error.toString()
        message.error(`删除失败: ${errorMsg}`)
      }
    }
  })
}

// ==================== 刷新和导航快捷键 ====================

// 刷新当前面板
function handleRefresh() {
  const pane = activePaneRef.value
  if (pane && pane.refresh) {
    pane.refresh()
  }
}

// 后退到上一级目录
function handleGoBack() {
  const pane = activePaneRef.value
  if (pane && pane.goBack) {
    pane.goBack()
  }
}

// ==================== 点击外部关闭菜单 ====================

function handleClickOutside(event: MouseEvent) {
  const target = event.target as HTMLElement
  if (!target.closest('.dropdown-wrapper')) {
    closeNewMenu()
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})

// 分隔器拖拽
const isDraggingSplitter = ref(false)

function handleSplitterMouseDown(event: MouseEvent) {
  isDraggingSplitter.value = true
  const isVertical = layoutMode.value === 'vertical'
  document.body.style.cursor = isVertical ? 'row-resize' : 'col-resize'
  document.body.style.userSelect = 'none'

  const mainContent = document.querySelector('.main-content') as HTMLElement
  if (!mainContent) return

  const onMouseMove = (e: MouseEvent) => {
    if (!isDraggingSplitter.value) return

    const rect = mainContent.getBoundingClientRect()

    if (isVertical) {
      const y = e.clientY - rect.top
      const percentage = (y / rect.height) * 100

      if (percentage > 20 && percentage < 80) {
        mainContent.style.gridTemplateRows = `${percentage}% 4px 1fr`
      }
    } else {
      const x = e.clientX - rect.left
      const percentage = (x / rect.width) * 100

      if (percentage > 20 && percentage < 80) {
        mainContent.style.gridTemplateColumns = `${percentage}% 4px 1fr`
      }
    }
  }

  const onMouseUp = () => {
    isDraggingSplitter.value = false
    document.body.style.cursor = ''
    document.body.style.userSelect = ''
    document.removeEventListener('mousemove', onMouseMove)
    document.removeEventListener('mouseup', onMouseUp)
  }

  document.addEventListener('mousemove', onMouseMove)
  document.addEventListener('mouseup', onMouseUp)
}

// 暴露方法给父组件调用
function goBack() {
  const activePaneRef = activePane.value === 'left' ? leftPaneRef.value : rightPaneRef.value
  if (activePaneRef && typeof activePaneRef.goBack === 'function') {
    activePaneRef.goBack()
  }
}

function goForward() {
  const activePaneRef = activePane.value === 'left' ? leftPaneRef.value : rightPaneRef.value
  if (activePaneRef && typeof activePaneRef.goForward === 'function') {
    activePaneRef.goForward()
  }
}

defineExpose({
  goBack,
  goForward
})
</script>

<style scoped>
.file-manager {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: #f9fafb;
}

/* Toolbar */
.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background: white;
  border-bottom: 1px solid #e5e7eb;
}

.toolbar-left, .toolbar-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.btn {
  padding: 8px 12px;
  background: #f3f4f6;
  border: none;
  border-radius: 6px;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  gap: 4px;
}

.btn:hover:not(:disabled) {
  background: #e5e7eb;
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-primary {
  background: #3b82f6;
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background: #2563eb;
}

.btn .icon {
  font-size: 16px;
  font-weight: bold;
}

.btn .arrow {
  font-size: 10px;
  margin-left: 2px;
}

/* 下拉菜单 */
.dropdown-wrapper {
  position: relative;
}

.dropdown-menu {
  position: absolute;
  top: 100%;
  left: 0;
  margin-top: 4px;
  background: white;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06);
  min-width: 180px;
  z-index: 1000;
  overflow: hidden;
}

.dropdown-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 16px;
  cursor: pointer;
  transition: background 0.15s;
  font-size: 14px;
}

.dropdown-item:hover:not(.disabled) {
  background: #f3f4f6;
}

.dropdown-item.disabled {
  opacity: 0.5;
  cursor: not-allowed;
  color: #9ca3af;
}

.dropdown-item .item-icon {
  font-size: 18px;
}

.dropdown-item + .dropdown-item {
  border-top: 1px solid #f3f4f6;
}

.divider {
  width: 1px;
  height: 24px;
  background: #d1d5db;
}

.view-controls, .thumbnail-controls, .layout-controls {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px;
  background: #f3f4f6;
  border-radius: 6px;
}

.btn-view, .btn-size, .btn-layout {
  padding: 6px 12px;
  background: transparent;
  border: none;
  border-radius: 4px;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.15s;
}

.btn-view.active, .btn-size.active, .btn-layout.active {
  background: white;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
}

.label {
  font-size: 13px;
  color: #6b7280;
  margin-right: 4px;
}

/* Main Content */
.main-content {
  flex: 1;
  display: flex;
  overflow: hidden;
}

.main-content.layout-single {
  display: block;
}

.main-content.layout-horizontal {
  display: grid;
  grid-template-columns: 1fr 4px 1fr;
}

.main-content.layout-vertical {
  display: grid;
  grid-template-rows: 1fr 4px 1fr;
}

.splitter {
  background: #d1d5db;
  cursor: col-resize;
}

.splitter.splitter-vertical {
  cursor: row-resize;
}

.splitter:hover {
  background: #9ca3af;
}
</style>