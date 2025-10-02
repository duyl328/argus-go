<template>
  <div :class="['file-pane', { active: isActive, inactive: !isActive }]" @click="$emit('activate')" @mousedown="handlePaneMouseDown">
    <!-- Top indicator -->
    <div v-if="isActive" class="active-indicator"></div>

    <!-- Breadcrumb and Search Bar -->
    <div class="breadcrumb-container" @click.stop="$emit('activate')">
      <!-- Breadcrumb Navigation -->
      <div class="breadcrumb">
        <div v-if="!pathEditMode" class="breadcrumb-content">
          <span
            v-for="(segment, index) in currentPath"
            :key="index"
            :class="['breadcrumb-segment']"
          >
            <span
              :class="['breadcrumb-item', { current: index === currentPath.length - 1 }]"
              @click.stop="navigateToIndex(index)"
            >
              {{ segment }}
            </span>
            <span
              v-if="index < currentPath.length - 1"
              :class="['separator', { active: breadcrumbDropdown.index === index }]"
              @click.stop="toggleBreadcrumbDropdown(index)"
            >
              ›
            </span>
          </span>
          <button class="path-edit-btn" @click.stop="enterPathEditMode" title="编辑路径">
            ✏️
          </button>

          <!-- File Stats -->
          <div class="file-stats">
            <span v-if="fileStats.photos > 0" class="stat-item stat-photo" title="照片">
              🖼️ {{ fileStats.photos }}
            </span>
            <span v-if="fileStats.videos > 0" class="stat-item stat-video" title="视频">
              🎬 {{ fileStats.videos }}
            </span>
          </div>
        </div>
        <div v-else class="path-edit-container">
          <input
            ref="pathInputRef"
            v-model="pathEditValue"
            class="path-input"
            @blur="exitPathEditMode"
            @keydown.enter="applyPathEdit"
            @keydown.esc="cancelPathEdit"
            @click.stop
          />
          <button class="path-edit-confirm" @click.stop="applyPathEdit" title="确认">✓</button>
          <button class="path-edit-cancel" @click.stop="cancelPathEdit" title="取消">✗</button>
        </div>
      </div>

      <!-- Search Box (always visible) -->
      <div class="search-box">
        <input
          v-model="props.filterOptions.nameQuery"
          type="text"
          class="search-input"
          placeholder="搜索..."
          @click.stop
        />
        <button
          v-if="props.filterOptions.nameQuery"
          class="search-clear"
          @click.stop="props.filterOptions.nameQuery = ''"
        >
          ×
        </button>
      </div>
    </div>

    <!-- Content Area -->
    <div
      ref="contentAreaRef"
      :class="['content-area', { 'pane-drag-over': dragState.isPaneDragOver }]"
      @mousedown="handleMouseDown"
      @mousemove="handleMouseMove"
      @mouseup="handleMouseUp"
      @click="handleContentClick"
      @contextmenu="handleContextMenu"
      @dragover="handlePaneDragOver"
      @dragleave="handlePaneDragLeave"
      @drop="handlePaneDrop"
      @scroll="handleScroll"
    >
      <!-- Grid View -->
      <div v-if="viewMode === 'grid'" :style="virtualContainerStyle">
        <div :class="['file-grid', `grid-${thumbnailSize}`]" :style="{ ...virtualContentStyle, ...gridItemSize }">
          <div
            v-for="(item, index) in renderItems"
          :key="item.name"
          :class="['file-item', {
            selected: selection.isSelected(item.name),
            focused: isFocused(item.name),
            'drop-target': item.type === 'folder' && dragState.isDragging && !selection.isSelected(item.name),
            'drag-over': dragState.dropTarget === item.name
          }]"
          :data-item-name="item.name"
          :data-item-index="index"
          :data-item-type="item.type"
          :draggable="true"
          @click="handleItemClick($event, item.name, index)"
          @dblclick="handleItemDoubleClick(item)"
          @mouseenter="handleItemMouseEnter($event, item)"
          @mouseleave="handleItemMouseLeave"
          @dragstart="handleDragStart($event, item)"
          @dragend="handleDragEnd"
          @dragover="handleDragOver($event, item)"
          @dragleave="handleDragLeave($event)"
          @drop="handleDrop($event, item)"
        >
          <div class="file-icon">
            {{ item.type === 'folder' ? '📁' : '🖼️' }}
          </div>
          <div class="file-name" v-html="highlightText(item.name)"></div>
          <div v-if="item.size" class="file-size">{{ item.size }}</div>
        </div>
        </div>
      </div>

      <!-- List View -->
      <div v-else class="file-list">
        <!-- List Header -->
        <div class="list-header">
          <div class="header-cell icon-cell"></div>
          <div class="header-cell name-cell" @click="handleHeaderSort('name')">
            <span>名称</span>
            <span v-if="props.sortOptions.field === 'name'" class="sort-indicator">
              {{ props.sortOptions.order === 'asc' ? '↑' : '↓' }}
            </span>
          </div>
          <div class="header-cell size-cell" @click="handleHeaderSort('size')">
            <span>大小</span>
            <span v-if="props.sortOptions.field === 'size'" class="sort-indicator">
              {{ props.sortOptions.order === 'asc' ? '↑' : '↓' }}
            </span>
          </div>
          <div class="header-cell date-cell" @click="handleHeaderSort('date')">
            <span>修改日期</span>
            <span v-if="props.sortOptions.field === 'date'" class="sort-indicator">
              {{ props.sortOptions.order === 'asc' ? '↑' : '↓' }}
            </span>
          </div>
        </div>

        <!-- List Items Container -->
        <div :style="virtualContainerStyle">
          <div :style="virtualContentStyle">
            <div
              v-for="(item, index) in renderItems"
          :key="item.name"
          :class="['list-item', {
            selected: selection.isSelected(item.name),
            focused: isFocused(item.name),
            'drop-target': item.type === 'folder' && dragState.isDragging && !selection.isSelected(item.name),
            'drag-over': dragState.dropTarget === item.name
          }]"
          :data-item-name="item.name"
          :data-item-index="index"
          :data-item-type="item.type"
          :draggable="true"
          @click="handleItemClick($event, item.name, index)"
          @dblclick="handleItemDoubleClick(item)"
          @mouseenter="handleItemMouseEnter($event, item)"
          @mouseleave="handleItemMouseLeave"
          @dragstart="handleDragStart($event, item)"
          @dragend="handleDragEnd"
          @dragover="handleDragOver($event, item)"
          @dragleave="handleDragLeave($event)"
          @drop="handleDrop($event, item)"
        >
          <span class="list-icon">{{ item.type === 'folder' ? '📁' : '🖼️' }}</span>
          <span class="list-name" v-html="highlightText(item.name)"></span>
          <span class="list-size">{{ item.size || '-' }}</span>
          <span class="list-date">{{ item.date || '-' }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Selection Rectangle -->
      <div v-if="dragSelection.isSelecting" class="selection-rectangle" :style="selectionBoxStyle"></div>
    </div>

    <!-- Inactive Overlay -->
    <div v-if="!isActive" class="inactive-overlay"></div>

    <!-- Context Menu -->
    <ContextMenu
      :visible="contextMenu.visible"
      :x="contextMenu.x"
      :y="contextMenu.y"
      :target-item="contextMenu.targetItem"
      :selected-count="selection.selectionCount.value"
      :pane-id="paneId"
      @close="contextMenu.visible = false"
      @action="handleContextMenuAction"
    />

    <!-- Tooltip -->
    <Tooltip ref="tooltipRef" />

    <!-- Breadcrumb Dropdown -->
    <Teleport to="body">
      <div
        v-if="breadcrumbDropdown.visible"
        class="breadcrumb-dropdown"
        :style="{
          left: `${breadcrumbDropdown.x}px`,
          top: `${breadcrumbDropdown.y}px`
        }"
        @click.stop
      >
        <div
          v-for="folder in breadcrumbDropdown.folders"
          :key="folder"
          class="dropdown-item"
          @click="navigateToBreadcrumbFolder(folder)"
        >
          <span class="dropdown-icon">📁</span>
          <span class="dropdown-label">{{ folder }}</span>
        </div>
        <div v-if="breadcrumbDropdown.folders.length === 0" class="dropdown-empty">
          没有其他文件夹
        </div>
      </div>
    </Teleport>

    <!-- Quick Preview -->
    <QuickPreview
      :visible="quickPreview.visible"
      :item="quickPreview.item"
      @close="closeQuickPreview"
    />

    <!-- Debug Panel (开发环境) -->
    <DebugPanel
      v-if="isDevelopment"
      :visible="debugPanelVisible"
      :metrics="debugMetrics"
      @close="debugPanelVisible = false"
      @generate="generateTestData"
      @scrollTo="handleDebugScrollTo"
      @clear="clearTestData"
    />

    <!-- Debug Toggle Button (开发环境) -->
    <button
      v-if="isDevelopment && isActive"
      class="debug-toggle"
      @click="debugPanelVisible = !debugPanelVisible"
      title="虚拟滚动调试面板"
    >
      🔬
    </button>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted, nextTick, reactive } from 'vue'
import { useFileSelection } from '@/composables/fileManager/useFileSelection'
import { useKeyboardNav } from '@/composables/fileManager/useKeyboardNav'
import { useDragSelection } from '@/composables/fileManager/useDragSelection'
import { useDragAndDrop } from '@/composables/fileManager/useDragAndDrop'
import { useVirtualScroll, useVirtualGrid } from '@/composables/fileManager/useVirtualScroll'
import { moveItems, getFolderByPath, searchItems } from '@/utils/fileManager/fileOperations'
import { mockFolderStructure as originalMockData } from './mockData'
import type { FileItem, ViewMode, ThumbnailSize, PaneId } from './types'
import ContextMenu from './ContextMenu.vue'
import Tooltip from './Tooltip.vue'
import QuickPreview from './QuickPreview.vue'
import DebugPanel from './DebugPanel.vue'

// 将 mockData 转换为响应式对象（全局共享）
const mockFolderStructure = reactive(originalMockData)

// 开发环境标识
const isDevelopment = import.meta.env.DEV

const props = defineProps<{
  paneId: PaneId
  viewMode: ViewMode
  thumbnailSize: ThumbnailSize
  isActive: boolean
  sortOptions: {
    field: string
    order: string
  }
  filterOptions: {
    nameQuery: string
    fileType: string
  }
}>()

const emit = defineEmits<{
  activate: []
}>()

// State
const currentPath = ref<string[]>(['Home'])
const contentAreaRef = ref<HTMLElement>()
const tooltipRef = ref<InstanceType<typeof Tooltip>>()
const pathInputRef = ref<HTMLInputElement>()
const pathEditMode = ref(false)
const pathEditValue = ref('')
const zoomLevel = ref(100) // 缩放级别：50-200%

// History management
const history = ref<string[][]>([['Home']]) // 存储路径历史
const historyIndex = ref(0) // 当前历史位置

// Context Menu state
const contextMenu = ref({
  visible: false,
  x: 0,
  y: 0,
  targetItem: null as string | null
})

// Breadcrumb dropdown state
const breadcrumbDropdown = ref({
  visible: false,
  x: 0,
  y: 0,
  index: -1,
  folders: [] as string[]
})

// Quick Preview state
const quickPreview = ref({
  visible: false,
  item: null as FileItem | null
})

// Debug Panel (开发环境)
const debugPanelVisible = ref(false)

const debugMetrics = computed(() => {
  const total = visibleItems.value.length
  const rendered = renderItems.value.length
  const savings = total > 0 ? Math.round((1 - rendered / total) * 100) : 0

  const virtual = props.viewMode === 'list' ? listVirtual : gridVirtual

  return {
    totalItems: total,
    renderedItems: rendered,
    savings: savings,
    virtualScrollEnabled: shouldUseVirtualScroll.value,
    viewMode: props.viewMode,
    zoomLevel: zoomLevel.value,
    scrollTop: virtual.scrollTop.value,
    startIndex: virtual.startIndex.value,
    endIndex: virtual.endIndex.value,
    totalHeight: virtual.totalHeight.value,
    columns: props.viewMode === 'grid' ? gridVirtual.columns.value : undefined
  }
})

// Container size for virtual scrolling
const containerWidth = ref(800)
const containerHeight = ref(600)

// Composables
const selection = useFileSelection()
const dragSelectionLogic = useDragSelection()
const dragDropLogic = useDragAndDrop()

// 解构composables以便使用
const { dragSelection, autoScroll, selectionBox, selectionBoxStyle, startDragSelection, updateDragSelection, checkAutoScroll, startAutoScroll, finishDragSelection, cancelDragSelection, isIntersecting } = dragSelectionLogic
const { dragState, startDrag, setDropTarget, setPaneDragOver, createDragPreview, cleanupDragPreview, endDrag, resetDragState } = dragDropLogic

// 虚拟滚动配置
const ENABLE_VIRTUAL_SCROLL = ref(true) // 可以通过这个开关控制是否启用
const VIRTUAL_SCROLL_THRESHOLD = 100 // 超过100个项目时启用虚拟滚动

// Computed
const currentFolder = computed(() => {
  return getFolderByPath(mockFolderStructure, currentPath.value)
})

// 总项目数（未过滤）
const totalItemsCount = computed(() => {
  if (!currentFolder.value) return 0
  return Object.keys(currentFolder.value).length
})

// 文件类型统计
const fileStats = computed(() => {
  if (!currentFolder.value) {
    return { photos: 0, videos: 0, folders: 0, files: 0 }
  }

  const items = Object.values(currentFolder.value)
  return {
    photos: items.filter(item => item.type === 'photo').length,
    videos: items.filter(item => item.type === 'video').length,
    folders: items.filter(item => item.type === 'folder').length,
    files: items.filter(item => item.type === 'file').length
  }
})

// 表头排序点击处理
function handleHeaderSort(field: string) {
  if (props.sortOptions.field === field) {
    // 如果点击的是当前排序字段，切换顺序
    const current = props.sortOptions.order
    props.sortOptions.order = current === 'asc' ? 'desc' : 'asc'
  } else {
    // 如果点击的是新字段，设置为该字段并默认升序
    props.sortOptions.field = field as any
    props.sortOptions.order = 'asc'
  }
}

const visibleItems = computed(() => {
  if (!currentFolder.value) return []

  let items = Object.values(currentFolder.value)

  // 文件名搜索过滤
  if (props.filterOptions.nameQuery.trim()) {
    const query = props.filterOptions.nameQuery.toLowerCase()
    items = items.filter(item => item.name.toLowerCase().includes(query))
  }

  // 排序
  return sortItems(items, props.sortOptions.field, props.sortOptions.order)
})

// 判断是否需要启用虚拟滚动
const shouldUseVirtualScroll = computed(() => {
  return ENABLE_VIRTUAL_SCROLL.value && visibleItems.value.length > VIRTUAL_SCROLL_THRESHOLD
})

// 列表模式虚拟滚动
const LIST_ITEM_HEIGHT = 40
const listVirtual = useVirtualScroll({
  items: visibleItems,
  itemHeight: LIST_ITEM_HEIGHT,
  containerHeight,
  overscan: 5
})

// 计算网格项目尺寸
const gridItemWidth = computed(() => {
  const baseSize = props.thumbnailSize === 'small' ? 80 : props.thumbnailSize === 'medium' ? 100 : 130
  return Math.round(baseSize * (zoomLevel.value / 100))
})

const gridItemHeightValue = computed(() => {
  const baseSize = props.thumbnailSize === 'small' ? 80 : props.thumbnailSize === 'medium' ? 100 : 130
  return Math.round(baseSize * (zoomLevel.value / 100)) + 24 // +24 for filename
})

// 网格模式虚拟滚动（传递响应式引用）
const gridVirtual = useVirtualGrid({
  items: visibleItems,
  itemWidth: gridItemWidth,
  itemHeight: gridItemHeightValue,
  containerWidth,
  containerHeight,
  gap: 8,
  overscan: 2
})

// 根据视图模式和是否启用虚拟滚动选择要渲染的项目
const renderItems = computed(() => {
  if (!shouldUseVirtualScroll.value) {
    return visibleItems.value
  }

  return props.viewMode === 'list'
    ? listVirtual.visibleItems.value
    : gridVirtual.visibleItems.value
})

// 虚拟滚动容器样式
const virtualContainerStyle = computed(() => {
  if (!shouldUseVirtualScroll.value) {
    return {}
  }

  const virtual = props.viewMode === 'list' ? listVirtual : gridVirtual
  return {
    height: `${virtual.totalHeight.value}px`,
    position: 'relative' as const
  }
})

// 虚拟滚动内容偏移样式
const virtualContentStyle = computed(() => {
  if (!shouldUseVirtualScroll.value) {
    return {}
  }

  const virtual = props.viewMode === 'list' ? listVirtual : gridVirtual
  return {
    transform: `translateY(${virtual.offsetY.value}px)`
  }
})

// 排序函数
function sortItems(items: FileItem[], field: string, order: string): FileItem[] {
  return [...items].sort((a, b) => {
    // 文件夹始终优先（除非按类型排序）
    if (field !== 'type') {
      if (a.type === 'folder' && b.type !== 'folder') return -1
      if (a.type !== 'folder' && b.type === 'folder') return 1
    }

    let comparison = 0

    switch (field) {
      case 'name':
        comparison = a.name.localeCompare(b.name, 'zh-CN')
        break
      case 'extension': {
        const extA = getFileExtension(a.name)
        const extB = getFileExtension(b.name)
        comparison = extA.localeCompare(extB)
        // 如果扩展名相同，按文件名排序
        if (comparison === 0) {
          comparison = a.name.localeCompare(b.name, 'zh-CN')
        }
        break
      }
      case 'date':
        comparison = (a.date || '').localeCompare(b.date || '')
        break
      case 'size': {
        const sizeA = parseSizeToBytes(a.size || '0')
        const sizeB = parseSizeToBytes(b.size || '0')
        comparison = sizeA - sizeB
        break
      }
      case 'type':
        comparison = a.type.localeCompare(b.type)
        // 如果类型相同，按文件名排序
        if (comparison === 0) {
          comparison = a.name.localeCompare(b.name, 'zh-CN')
        }
        break
    }

    return order === 'desc' ? -comparison : comparison
  })
}

// 获取文件扩展名
function getFileExtension(filename: string): string {
  const lastDot = filename.lastIndexOf('.')
  if (lastDot === -1 || lastDot === 0) return ''
  return filename.substring(lastDot + 1).toLowerCase()
}

// 解析大小字符串为字节数
function parseSizeToBytes(sizeStr: string): number {
  const match = sizeStr.match(/^([\d.]+)\s*(B|KB|MB|GB)?$/i)
  if (!match) return 0

  const value = parseFloat(match[1])
  const unit = (match[2] || 'B').toUpperCase()

  switch (unit) {
    case 'GB': return value * 1024 * 1024 * 1024
    case 'MB': return value * 1024 * 1024
    case 'KB': return value * 1024
    case 'B': return value
    default: return value
  }
}

// Methods
function highlightText(text: string): string {
  if (!props.filterOptions.nameQuery.trim()) {
    return text
  }

  const query = props.filterOptions.nameQuery.trim()
  const regex = new RegExp(`(${query})`, 'gi')
  return text.replace(regex, '<mark>$1</mark>')
}

function navigateToIndex(index: number) {
  currentPath.value = currentPath.value.slice(0, index + 1)
  addToHistory(currentPath.value)
  selection.clearSelection()
  breadcrumbDropdown.value.visible = false
}

// History navigation functions
function addToHistory(path: string[]) {
  // 检查新路径是否与当前历史位置的路径相同
  const currentHistoryPath = history.value[historyIndex.value]
  if (currentHistoryPath && JSON.stringify(currentHistoryPath) === JSON.stringify(path)) {
    return // 路径相同，不添加重复历史
  }

  // 删除当前位置之后的所有历史（用户在历史中间位置进行了新的导航）
  history.value = history.value.slice(0, historyIndex.value + 1)
  // 添加新路径
  history.value.push([...path])
  historyIndex.value = history.value.length - 1
}

function goBack() {
  if (historyIndex.value > 0) {
    historyIndex.value--
    currentPath.value = [...history.value[historyIndex.value]]
    selection.clearSelection()
  }
}

function goForward() {
  if (historyIndex.value < history.value.length - 1) {
    historyIndex.value++
    currentPath.value = [...history.value[historyIndex.value]]
    selection.clearSelection()
  }
}

// Breadcrumb dropdown
function toggleBreadcrumbDropdown(index: number) {
  if (breadcrumbDropdown.value.visible && breadcrumbDropdown.value.index === index) {
    breadcrumbDropdown.value.visible = false
    return
  }

  // 获取父文件夹中的所有文件夹
  const pathToParent = currentPath.value.slice(0, index + 1)
  let parentFolder: any = mockFolderStructure

  for (const segment of pathToParent) {
    if (parentFolder[segment] && parentFolder[segment].children) {
      parentFolder = parentFolder[segment].children
    } else {
      return
    }
  }

  // 获取当前路径的下一个segment（当前选中的文件夹）
  const currentSegment = currentPath.value[index + 1]

  // 收集所有文件夹（排除当前选中的）
  const folders = Object.values(parentFolder)
    .filter((item: any) => item.type === 'folder' && item.name !== currentSegment)
    .map((item: any) => item.name)
    .sort()

  // 计算位置（相对于separator）
  const separatorElements = document.querySelectorAll('.breadcrumb-segment .separator')
  const separatorEl = separatorElements[index] as HTMLElement
  if (!separatorEl) return

  const rect = separatorEl.getBoundingClientRect()

  breadcrumbDropdown.value = {
    visible: true,
    x: rect.left,
    y: rect.bottom + 4,
    index,
    folders
  }
}

function navigateToBreadcrumbFolder(folderName: string) {
  const index = breadcrumbDropdown.value.index
  if (index < 0) return

  // 导航到选中的文件夹
  const newPath = [...currentPath.value.slice(0, index + 1), folderName]
  currentPath.value = newPath
  addToHistory(currentPath.value)
  selection.clearSelection()
  breadcrumbDropdown.value.visible = false
}

// Path editing
function enterPathEditMode() {
  pathEditMode.value = true
  pathEditValue.value = currentPath.value.join('/')
  nextTick(() => {
    pathInputRef.value?.focus()
    pathInputRef.value?.select()
  })
}

function exitPathEditMode() {
  // 延迟一点，让按钮点击事件先触发
  setTimeout(() => {
    pathEditMode.value = false
  }, 150)
}

function applyPathEdit() {
  const path = pathEditValue.value.trim()
  if (!path) {
    cancelPathEdit()
    return
  }

  // 解析路径
  const segments = path.split('/').filter(s => s.trim())
  if (segments.length === 0) {
    cancelPathEdit()
    return
  }

  // 验证路径是否存在
  let folder: any = mockFolderStructure
  for (const segment of segments) {
    if (folder[segment] && folder[segment].children) {
      folder = folder[segment].children
    } else {
      alert(`路径不存在: ${path}`)
      return
    }
  }

  // 应用路径
  currentPath.value = segments
  addToHistory(currentPath.value)
  selection.clearSelection()
  pathEditMode.value = false
}

function cancelPathEdit() {
  pathEditMode.value = false
  pathEditValue.value = ''
}

function navigateToFolder(folderName: string) {
  currentPath.value.push(folderName)
  addToHistory(currentPath.value)
  selection.clearSelection()
}

function isFocused(itemName: string): boolean {
  return selection.focusedItem.value?.name === itemName
}

function handleItemClick(event: MouseEvent, itemName: string, index: number) {
  event.stopPropagation()

  // 激活当前面板
  emit('activate')

  if (event.shiftKey && selection.anchorItem.value) {
    // Range selection
    selection.clearSelection()
    selection.selectRange(selection.anchorItem.value.index, index, visibleItems.value)
    selection.setFocusedItem(itemName, index)
  } else if (event.ctrlKey || event.metaKey) {
    // Toggle selection
    selection.toggleItemSelection(itemName, index)
  } else {
    // Single selection
    selection.clearSelection()
    selection.selectItem(itemName, index)
  }
}

function handleItemDoubleClick(item: FileItem) {
  if (item.type === 'folder') {
    navigateToFolder(item.name)
  }
}

// 面板级别的鼠标按下处理（已由 FieldView 统一处理，此处仅作为备用）
function handlePaneMouseDown(event: MouseEvent) {
  // 鼠标侧键由 FieldView 全局处理，这里不再需要处理
}

function handleMouseDown(event: MouseEvent) {
  const target = event.target as HTMLElement
  if (target.closest('.file-item') || target.closest('.list-item')) {
    return
  }

  // 如果正在拖动文件，不要开始框选
  if (dragState.value.isDragging) {
    return
  }

  emit('activate')

  if (!contentAreaRef.value) return

  // 使用composable开始拖拽选择
  const ctrlKey = event.ctrlKey || event.metaKey
  startDragSelection(event, contentAreaRef.value, ctrlKey, selection.selectedItems.value)

  if (!ctrlKey) {
    selection.clearSelection()
  }

  // 启动自动滚动（带回调更新选择）
  startAutoScroll(contentAreaRef.value, updateSelectionFromDragBox)

  // 在 document 级别监听 mousemove 和 mouseup
  const onDocumentMouseMove = (e: MouseEvent) => {
    handleMouseMove(e)
  }

  const onDocumentMouseUp = (e: MouseEvent) => {
    handleMouseUp()
    document.removeEventListener('mousemove', onDocumentMouseMove)
    document.removeEventListener('mouseup', onDocumentMouseUp)
  }

  document.addEventListener('mousemove', onDocumentMouseMove)
  document.addEventListener('mouseup', onDocumentMouseUp)

  event.preventDefault()
}

function handleMouseMove(event: MouseEvent) {
  // 如果不在拖拽选择状态，隐藏tooltip（防止在空白区域显示）
  if (!dragSelection.value.isSelecting) {
    const target = event.target as HTMLElement
    if (!target.closest('.file-item') && !target.closest('.list-item')) {
      tooltipRef.value?.hide()
    }
    return
  }

  if (!contentAreaRef.value) return

  // 使用composable更新拖拽选择
  updateDragSelection(event, contentAreaRef.value)

  // 自动滚动检测
  checkAutoScroll(event, contentAreaRef.value)

  // 更新框选的项
  updateSelectionFromDragBox()
}

function handleMouseUp() {
  if (dragSelection.value.isSelecting) {
    // 最后更新一次选择
    updateSelectionFromDragBox()

    // 使用composable完成拖拽选择
    finishDragSelection()

    // 设置焦点到最后选中的项
    if (selection.selectedItems.value.size > 0) {
      const lastSelected = Array.from(selection.selectedItems.value).pop()!
      const index = visibleItems.value.findIndex(item => item.name === lastSelected)
      if (index !== -1) {
        selection.setFocusedItem(lastSelected, index)
      }
    }
  }
}


// 更新框选的项（优化：基于数据遍历而非DOM查询，支持虚拟滚动）
function updateSelectionFromDragBox() {
  if (!contentAreaRef.value || !selectionBox.value) return

  const box = selectionBox.value
  const scrollTop = contentAreaRef.value.scrollTop
  const itemsInSelection = new Set<string>()

  // 遍历所有数据项而非DOM节点
  visibleItems.value.forEach((item, index) => {
    // 根据视图模式计算项目边界
    let itemBounds

    if (props.viewMode === 'list') {
      // 列表模式：使用虚拟滚动的边界计算
      if (shouldUseVirtualScroll.value && listVirtual.getItemBounds) {
        itemBounds = listVirtual.getItemBounds(index)
      } else {
        itemBounds = {
          top: index * LIST_ITEM_HEIGHT,
          left: 0,
          bottom: (index + 1) * LIST_ITEM_HEIGHT,
          right: containerWidth.value
        }
      }
    } else {
      // 网格模式：使用虚拟滚动的边界计算
      if (shouldUseVirtualScroll.value && gridVirtual.getItemBounds) {
        itemBounds = gridVirtual.getItemBounds(index)
      } else {
        const col = index % Math.floor(containerWidth.value / (gridItemWidth.value + 8))
        const row = Math.floor(index / Math.floor(containerWidth.value / (gridItemWidth.value + 8)))
        itemBounds = {
          top: row * (gridItemHeightValue.value + 8),
          left: col * (gridItemWidth.value + 8),
          bottom: (row + 1) * (gridItemHeightValue.value + 8),
          right: (col + 1) * (gridItemWidth.value + 8)
        }
      }
    }

    // 调整为相对于可见区域的坐标
    const adjustedBounds = {
      top: itemBounds.top - scrollTop,
      left: itemBounds.left,
      bottom: itemBounds.bottom - scrollTop,
      right: itemBounds.right
    }

    // 检查是否与选择框相交
    if (isBoxIntersecting(box, adjustedBounds)) {
      itemsInSelection.add(item.name)
    }
  })

  // 更新选择
  if (dragSelection.value.ctrlKey) {
    // Ctrl拖拽：保留初始选择 + 新框选
    selection.clearSelection()
    dragSelection.value.initialSelections.forEach(name => {
      selection.selectedItems.value.add(name)
    })
    itemsInSelection.forEach(name => {
      selection.selectedItems.value.add(name)
    })
  } else {
    // 普通拖拽：只选择框选的
    selection.clearSelection()
    itemsInSelection.forEach(name => {
      selection.selectedItems.value.add(name)
    })
  }
}

// 辅助函数：判断两个矩形是否相交
function isBoxIntersecting(box1: { top: number; left: number; bottom: number; right: number }, box2: { top: number; left: number; bottom: number; right: number }) {
  return !(
    box1.right < box2.left ||
    box1.left > box2.right ||
    box1.bottom < box2.top ||
    box1.top > box2.bottom
  )
}

function handleContentClick(event: MouseEvent) {
  // 如果刚完成拖拽选择，不要清除选择
  if (dragSelection.value.justFinished) {
    return
  }

  const target = event.target as HTMLElement
  if (!target.closest('.file-item') && !target.closest('.list-item')) {
    if (!event.ctrlKey) {
      selection.clearSelection()
    }
  }
}

function handleContextMenu(event: MouseEvent) {
  event.preventDefault()

  const target = event.target as HTMLElement
  const item = target.closest('.file-item, .list-item') as HTMLElement

  contextMenu.value = {
    visible: true,
    x: event.clientX,
    y: event.clientY,
    targetItem: item?.dataset.itemName || null
  }
}

function handleContextMenuAction(action: string, params?: any) {
  console.log('Context menu action:', action, params)

  switch (action) {
    case 'open':
      if (params) {
        const item = visibleItems.value.find(i => i.name === params)
        if (item && item.type === 'folder') {
          navigateToFolder(item.name)
        }
      }
      break

    case 'cut':
    case 'copy':
    case 'paste':
    case 'delete':
    case 'rename':
    case 'properties':
      // TODO: 实现这些操作
      alert(`执行操作: ${action}`)
      break

    case 'newFolder':
      alert('创建新文件夹')
      break

    case 'selectAll':
      selection.selectAll(visibleItems.value)
      break

    case 'refresh':
      // 重新加载当前文件夹
      const temp = [...currentPath.value]
      currentPath.value = []
      nextTick(() => {
        currentPath.value = temp
      })
      break
  }
}

// Tooltip 处理
function handleItemMouseEnter(event: MouseEvent, item: FileItem) {
  if (dragState.value.isDragging) return

  const lines: string[] = []
  lines.push(`类型: ${item.type === 'folder' ? '文件夹' : '文件'}`)
  if (item.size) lines.push(`大小: ${item.size}`)
  if (item.date) lines.push(`修改日期: ${item.date}`)

  const tooltip = lines.join('\n')
  tooltipRef.value?.show(tooltip, event.clientX, event.clientY)
}

function handleItemMouseLeave() {
  tooltipRef.value?.hide()
}

// 拖放处理
function handleDragStart(event: DragEvent, item: FileItem) {
  // 如果拖动的项未选中，选中它
  if (!selection.isSelected(item.name)) {
    selection.clearSelection()
    const index = visibleItems.value.findIndex(i => i.name === item.name)
    selection.selectItem(item.name, index)
  }

  const draggedItems = Array.from(selection.selectedItems.value)

  // 使用composable设置拖动状态
  startDrag(draggedItems, props.paneId)

  // 设置拖动数据
  const dragData = {
    items: draggedItems,
    sourcePane: props.paneId,
    sourcePath: currentPath.value
  }
  event.dataTransfer!.effectAllowed = 'move'
  event.dataTransfer!.setData('application/json', JSON.stringify(dragData))

  // 使用composable创建自定义拖动预览
  const preview = createDragPreview(draggedItems.length, draggedItems[0])
  event.dataTransfer!.setDragImage(preview, 0, 0)
  setTimeout(() => cleanupDragPreview(preview), 0)

  // 隐藏tooltip
  tooltipRef.value?.hide()
}

function handleDragEnd() {
  resetDragState()
}

function handleDragOver(event: DragEvent, item: FileItem) {
  // 只有文件夹可以作为drop target
  if (item.type !== 'folder') return

  // 不能拖到自己身上
  if (selection.isSelected(item.name)) return

  event.preventDefault()
  event.dataTransfer!.dropEffect = 'move'
  setDropTarget(item.name)
}

function handleDragLeave(event: DragEvent) {
  const target = event.target as HTMLElement
  const relatedTarget = event.relatedTarget as HTMLElement

  // 只有真正离开元素时才清除
  if (!target.contains(relatedTarget)) {
    setDropTarget(null)
  }
}

function handleDrop(event: DragEvent, targetItem: FileItem) {
  event.preventDefault()

  if (targetItem.type !== 'folder') return
  if (selection.isSelected(targetItem.name)) return

  try {
    const dragData = JSON.parse(event.dataTransfer!.getData('application/json'))

    // 显示移动操作日志
    const itemCount = dragData.items.length
    const itemLabel = itemCount === 1 ? dragData.items[0] : `${itemCount} 个项目`

    console.log(`[文件操作] 移动 ${itemLabel} 到文件夹 "${targetItem.name}"`, {
      source: dragData.sourcePane,
      target: props.paneId,
      items: dragData.items,
      targetFolder: targetItem.name
    })

    // 执行文件移动（使用utils）
    const targetPath = [...currentPath.value, targetItem.name]
    const success = moveItems(mockFolderStructure, dragData.items, dragData.sourcePath, targetPath)

    if (success) {
      // 清除选择
      selection.clearSelection()

      // 触发视图更新
      nextTick()
    }

  } catch (e) {
    console.error('[文件操作] 拖放失败:', e)
  }

  // 清除所有拖拽状态（包括绿框）
  resetDragState()
}

// Pane-level drop handlers (for dropping into empty space or other pane)
function handlePaneDragOver(event: DragEvent) {
  // 检查是否在拖动文件项
  const target = event.target as HTMLElement
  if (target.closest('.file-item') || target.closest('.list-item')) {
    return
  }

  try {
    const data = event.dataTransfer?.types.includes('application/json')
    if (data) {
      event.preventDefault()
      event.dataTransfer!.dropEffect = 'move'
      setPaneDragOver(true)
    }
  } catch (e) {
    // Ignore
  }
}

function handlePaneDragLeave(event: DragEvent) {
  const target = event.target as HTMLElement
  const relatedTarget = event.relatedTarget as HTMLElement

  // 只有真正离开pane时才清除
  if (!contentAreaRef.value?.contains(relatedTarget)) {
    setPaneDragOver(false)
  }
}

function handlePaneDrop(event: DragEvent) {
  // 如果拖到具体的item上，不处理（由item的drop处理）
  const target = event.target as HTMLElement
  if (target.closest('.file-item') || target.closest('.list-item')) {
    return
  }

  event.preventDefault()

  try {
    const dragData = JSON.parse(event.dataTransfer!.getData('application/json'))

    // 拖到当前面板的空白区域 = 移动到当前文件夹
    const currentFolderName = currentPath.value[currentPath.value.length - 1]
    const itemCount = dragData.items.length
    const itemLabel = itemCount === 1 ? dragData.items[0] : `${itemCount} 个项目`

    if (dragData.sourcePane !== props.paneId) {
      // 跨面板移动
      const sourceLabel = dragData.sourcePane === 'left' ? '左' : '右'
      const targetLabel = props.paneId === 'left' ? '左' : '右'

      console.log(`[文件操作] 从${sourceLabel}面板移动 ${itemLabel} 到${targetLabel}面板的 "${currentFolderName}"`, {
        source: dragData.sourcePane,
        target: props.paneId,
        items: dragData.items,
        targetPath: currentPath.value
      })

      // 执行跨面板移动（使用utils）
      const success = moveItems(mockFolderStructure, dragData.items, dragData.sourcePath, currentPath.value)
      if (success) {
        selection.clearSelection()
        nextTick()
      }
    } else {
      // 同面板内移动（相当于什么都不做）
      console.log(`[文件操作] 在当前文件夹内操作 ${itemLabel}（无变化）`)
      selection.clearSelection()
    }

  } catch (e) {
    console.error('[文件操作] 面板拖放失败:', e)
  }

  // 清除所有拖拽状态（包括绿框）
  resetDragState()
}

// 计算Grid列数
function getGridColumns(): number {
  if (!contentAreaRef.value) return 4

  const items = contentAreaRef.value.querySelectorAll('.file-item')
  if (items.length < 2) return 1

  const firstRect = items[0].getBoundingClientRect()
  const secondRect = items[1].getBoundingClientRect()

  if (Math.abs(secondRect.top - firstRect.top) < 10) {
    let cols = 1
    for (let i = 1; i < items.length; i++) {
      const rect = items[i].getBoundingClientRect()
      if (Math.abs(rect.top - firstRect.top) < 10) {
        cols++
      } else {
        break
      }
    }
    return cols
  }

  return 1
}

// 处理滚动事件（虚拟滚动）
function handleScroll(event: Event) {
  if (!shouldUseVirtualScroll.value) return

  const virtual = props.viewMode === 'list' ? listVirtual : gridVirtual
  virtual.onScroll(event)
}

// 处理Ctrl+滚轮缩放
function handleWheel(event: WheelEvent) {
  if (!event.ctrlKey) return

  event.preventDefault()

  const delta = event.deltaY > 0 ? -10 : 10
  const newZoom = Math.max(50, Math.min(200, zoomLevel.value + delta))
  zoomLevel.value = newZoom
}

// 计算动态网格大小
const gridItemSize = computed(() => {
  if (props.viewMode !== 'grid') return {}

  const baseSize = props.thumbnailSize === 'small' ? 80 :
                   props.thumbnailSize === 'medium' ? 100 : 130

  const size = Math.round(baseSize * (zoomLevel.value / 100))

  return {
    '--grid-item-size': `${size}px`
  }
})

// 监听搜索变化清除选择
watch(() => props.filterOptions.nameQuery, () => {
  selection.clearSelection()
})

// 容器尺寸监听
function updateContainerSize() {
  if (contentAreaRef.value) {
    const rect = contentAreaRef.value.getBoundingClientRect()
    containerWidth.value = rect.width
    containerHeight.value = rect.height
  }
}

// 关闭breadcrumb dropdown 和 添加滚轮监听
onMounted(() => {
  const closeDropdown = () => {
    breadcrumbDropdown.value.visible = false
  }
  document.addEventListener('click', closeDropdown)

  if (contentAreaRef.value) {
    contentAreaRef.value.addEventListener('wheel', handleWheel, { passive: false })

    // 初始化容器尺寸
    updateContainerSize()

    // 监听容器尺寸变化
    const resizeObserver = new ResizeObserver(() => {
      updateContainerSize()
    })
    resizeObserver.observe(contentAreaRef.value)

    // 存储observer以便清理
    ;(contentAreaRef.value as any)._resizeObserver = resizeObserver
  }
})

onUnmounted(() => {
  document.removeEventListener('click', () => {
    breadcrumbDropdown.value.visible = false
  })

  if (contentAreaRef.value) {
    contentAreaRef.value.removeEventListener('wheel', handleWheel)

    // 清理 ResizeObserver
    const observer = (contentAreaRef.value as any)._resizeObserver
    if (observer) {
      observer.disconnect()
    }
  }
})

// 键盘导航
useKeyboardNav({
  items: computed(() => visibleItems.value),
  viewMode: computed(() => props.viewMode),
  isActive: computed(() => props.isActive),
  focusedItem: selection.focusedItem,
  anchorItem: selection.anchorItem,
  selectedItems: selection.selectedItems,
  onNavigate: (newIndex: number, event: KeyboardEvent) => {
    const newItem = visibleItems.value[newIndex]
    if (!newItem) return

    if (event.shiftKey) {
      // Shift+方向键：扩展选择
      const anchor = selection.anchorItem.value
      if (anchor) {
        selection.clearSelection()
        selection.selectRange(anchor.index, newIndex, visibleItems.value)
      } else {
        selection.clearSelection()
        selection.selectItem(newItem.name, newIndex)
      }
      selection.setFocusedItem(newItem.name, newIndex)
    } else if (event.ctrlKey || event.metaKey) {
      // Ctrl+方向键：只移动焦点
      selection.setFocusedItem(newItem.name, newIndex)
    } else {
      // 普通方向键：单选并移动焦点
      selection.clearSelection()
      selection.selectItem(newItem.name, newIndex)
    }
  },
  onEnter: () => {
    const focused = selection.focusedItem.value
    if (focused) {
      const item = visibleItems.value[focused.index]
      if (item && item.type === 'folder') {
        navigateToFolder(item.name)
      }
    }
  },
  onSpace: (event: KeyboardEvent) => {
    event.preventDefault()

    // 如果预览已打开，关闭预览
    if (quickPreview.value.visible) {
      quickPreview.value.visible = false
      quickPreview.value.item = null
      return
    }

    // 打开预览
    const focused = selection.focusedItem.value
    if (focused) {
      const item = visibleItems.value[focused.index]
      if (item) {
        quickPreview.value.visible = true
        quickPreview.value.item = item
      }
    }
  },
  onSelectAll: () => {
    selection.selectAll(visibleItems.value)
  },
  getGridColumns
})

// 关闭快速预览
function closeQuickPreview() {
  quickPreview.value.visible = false
  quickPreview.value.item = null
}

// ==================== 调试工具函数 ====================

// 生成测试数据
function generateTestData(count: number) {
  const testItems: Record<string, FileItem> = {}

  // 生成文件夹（10%）
  const folderCount = Math.floor(count * 0.1)
  for (let i = 0; i < folderCount; i++) {
    const name = `folder_${String(i + 1).padStart(5, '0')}`
    testItems[name] = {
      name,
      type: 'folder',
      size: '',
      date: new Date().toLocaleString()
    }
  }

  // 生成照片文件（70%）
  const photoCount = Math.floor(count * 0.7)
  for (let i = 0; i < photoCount; i++) {
    const name = `photo_${String(i + 1).padStart(5, '0')}.jpg`
    const sizeMB = (Math.random() * 10 + 0.5).toFixed(2)
    testItems[name] = {
      name,
      type: 'photo',
      size: `${sizeMB} MB`,
      date: new Date(Date.now() - Math.random() * 365 * 24 * 60 * 60 * 1000).toLocaleString()
    }
  }

  // 生成视频文件（20%）
  const videoCount = count - folderCount - photoCount
  for (let i = 0; i < videoCount; i++) {
    const name = `video_${String(i + 1).padStart(5, '0')}.mp4`
    const sizeMB = (Math.random() * 500 + 10).toFixed(2)
    testItems[name] = {
      name,
      type: 'video',
      size: `${sizeMB} MB`,
      date: new Date(Date.now() - Math.random() * 365 * 24 * 60 * 60 * 1000).toLocaleString()
    }
  }

  // 更新mock数据结构中当前路径的文件夹
  // 导航到根目录
  currentPath.value = []

  // 直接修改mock数据
  Object.keys(mockFolderStructure).forEach(key => delete mockFolderStructure[key])
  Object.assign(mockFolderStructure, testItems)

  console.log(`✅ 已生成 ${count} 个测试文件`)
  console.log(`📁 文件夹: ${folderCount}`)
  console.log(`🖼️ 照片: ${photoCount}`)
  console.log(`🎬 视频: ${videoCount}`)
}

// 清空测试数据
function clearTestData() {
  // 导航到根目录
  currentPath.value = []

  // 清空mock数据
  Object.keys(mockFolderStructure).forEach(key => delete mockFolderStructure[key])

  console.log('🗑️ 已清空测试数据')
}

// 调试滚动操作
function handleDebugScrollTo(position: 'top' | 'bottom' | 'middle') {
  if (!contentAreaRef.value) return

  const container = contentAreaRef.value
  const virtual = props.viewMode === 'list' ? listVirtual : gridVirtual

  switch (position) {
    case 'top':
      container.scrollTop = 0
      break
    case 'bottom':
      container.scrollTop = virtual.totalHeight.value
      break
    case 'middle':
      container.scrollTop = virtual.totalHeight.value / 2
      break
  }

  console.log(`🎯 滚动到${position === 'top' ? '顶部' : position === 'bottom' ? '底部' : '中间'}`)
}

// 暴露方法给父组件
defineExpose({
  goBack,
  goForward,
  generateTestData,  // 暴露给外部调试使用
  clearTestData
})
</script>

<style scoped>
.file-pane {
  position: relative;
  display: flex;
  flex-direction: column;
  background: white;
  transition: opacity 0.15s;
  height: 100%;
  overflow: hidden;
}

.file-pane.inactive {
  opacity: 0.88;
}

.active-indicator {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 2px;
  background: linear-gradient(to right, #3b82f6, #60a5fa);
  z-index: 10;
}

.inactive-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.015);
  pointer-events: none;
  z-index: 5;
}

/* Breadcrumb Container */
.breadcrumb-container {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 12px 16px;
  border-bottom: 1px solid #f3f4f6;
  background: #f9fafb;
  cursor: pointer;
  transition: background-color 0.15s;
}

.breadcrumb-container:hover {
  background: rgba(59, 130, 246, 0.03);
}

/* Breadcrumb */
.breadcrumb {
  flex: 1;
  min-width: 0;
  display: flex;
  align-items: center;
}

.breadcrumb-content {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 14px;
  flex: 1;
  min-width: 0;
  position: relative;
}

.path-edit-btn {
  margin-left: 8px;
  padding: 4px 8px;
  background: transparent;
  border: none;
  border-radius: 4px;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s;
  opacity: 0.6;
}

.path-edit-btn:hover {
  background: #f3f4f6;
  opacity: 1;
}

/* File Stats */
.file-stats {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-left: 16px;
  padding-left: 16px;
  border-left: 1px solid #e5e7eb;
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
  font-weight: 500;
  padding: 4px 8px;
  border-radius: 4px;
  background: white;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
}

.stat-photo {
  color: #3b82f6;
}

.stat-video {
  color: #ec4899;
}

.path-edit-container {
  display: flex;
  align-items: center;
  gap: 6px;
  flex: 1;
}

.path-input {
  flex: 1;
  padding: 6px 10px;
  border: 1px solid #3b82f6;
  border-radius: 6px;
  font-size: 13px;
  outline: none;
  background: white;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.path-edit-confirm,
.path-edit-cancel {
  padding: 6px 10px;
  border: none;
  border-radius: 4px;
  font-size: 16px;
  cursor: pointer;
  transition: all 0.2s;
}

.path-edit-confirm {
  background: #10b981;
  color: white;
}

.path-edit-confirm:hover {
  background: #059669;
}

.path-edit-cancel {
  background: #ef4444;
  color: white;
}

.path-edit-cancel:hover {
  background: #dc2626;
}

.breadcrumb-item {
  padding: 4px 8px;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.2s;
  color: #6b7280;
}

.breadcrumb-item:hover {
  background: #f3f4f6;
  color: #3b82f6;
}

.breadcrumb-item.current {
  background: rgba(59, 130, 246, 0.1);
  color: #3b82f6;
  font-weight: 500;
}

.breadcrumb-segment {
  display: inline-flex;
  align-items: center;
}

.separator {
  margin: 0 4px;
  color: #9ca3af;
  padding: 2px 6px;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.2s;
  user-select: none;
}

.separator:hover {
  background: #f3f4f6;
  color: #3b82f6;
}

.separator.active {
  background: #dbeafe;
  color: #3b82f6;
}

/* Breadcrumb Dropdown */
.breadcrumb-dropdown {
  position: fixed;
  background: white;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1), 0 10px 15px rgba(0, 0, 0, 0.1);
  max-height: 300px;
  overflow-y: auto;
  min-width: 200px;
  z-index: 1000;
  padding: 4px;
}

.dropdown-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.15s;
  font-size: 14px;
}

.dropdown-item:hover {
  background: #f3f4f6;
}

.dropdown-icon {
  font-size: 16px;
}

.dropdown-label {
  color: #374151;
}

.dropdown-empty {
  padding: 12px;
  text-align: center;
  color: #9ca3af;
  font-size: 13px;
}

/* Search Box in Breadcrumb */
.search-box {
  position: relative;
  flex-shrink: 0;
  width: 200px;
  margin-left: auto;
}

.search-input {
  width: 100%;
  padding: 6px 30px 6px 10px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 13px;
  outline: none;
  transition: all 0.2s;
  background: white;
}

.search-input:focus {
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.search-input::placeholder {
  color: #9ca3af;
}

.search-clear {
  position: absolute;
  right: 8px;
  top: 50%;
  transform: translateY(-50%);
  background: none;
  border: none;
  font-size: 18px;
  color: #9ca3af;
  cursor: pointer;
  padding: 0;
  width: 20px;
  height: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  transition: all 0.2s;
}

.search-clear:hover {
  background: #f3f4f6;
  color: #374151;
}

/* Content Area */
.content-area {
  flex: 1;
  min-height: 0;
  padding: 16px;
  overflow: auto;
  position: relative;
  transition: all 0.2s;
}

.content-area.pane-drag-over {
  background: rgba(240, 253, 244, 0.3);
  padding-bottom: 24px;
  border: 2px dashed #10b981;
  border-radius: 4px;
}

/* Grid View */
.file-grid {
  display: grid;
  gap: 8px;
  --grid-item-size: 100px; /* 默认值，会被动态覆盖 */
}

.file-grid.grid-small {
  grid-template-columns: repeat(auto-fill, minmax(var(--grid-item-size, 80px), 1fr));
}

.file-grid.grid-medium {
  grid-template-columns: repeat(auto-fill, minmax(var(--grid-item-size, 100px), 1fr));
}

.file-grid.grid-large {
  grid-template-columns: repeat(auto-fill, minmax(var(--grid-item-size, 130px), 1fr));
}

.file-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 8px;
  background: white;
  border: 1px solid #e5e7eb;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.2s;
  user-select: none;
  aspect-ratio: 1 / 1;
  position: relative;
}

.file-item:hover {
  background: #f3f4f6;
  border-color: #d1d5db;
}

.file-item.selected {
  background: #dbeafe;
  border-color: #3b82f6;
}

.file-item.focused {
  outline: 2px solid #3b82f6;
  outline-offset: 2px;
}

.file-item.drop-target {
  border-color: #10b981;
  background: #f0fdf4;
  transition: all 0.15s;
}

.file-item.drag-over {
  border-color: #10b981;
  background: rgba(209, 250, 229, 0.5);
  transform: scale(1.02);
  box-shadow: 0 0 0 2px #10b981;
  z-index: 1;
}

.file-icon {
  font-size: 40px;
  margin-bottom: 4px;
  flex-shrink: 0;
}

/* 网格视图文件类型图标颜色 */
.file-item[data-item-type="folder"] .file-icon {
  color: #f59e0b;
  font-size: 44px;
}

.file-item[data-item-type="photo"] .file-icon {
  color: #3b82f6;
}

.file-item[data-item-type="video"] .file-icon {
  color: #ec4899;
}

.file-item[data-item-type="file"] .file-icon {
  color: #6b7280;
}

.file-name {
  font-size: 12px;
  text-align: center;
  word-break: break-word;
  color: #374151;
  line-height: 1.3;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.file-name :deep(mark) {
  background: #fef08a;
  color: #854d0e;
  padding: 1px 2px;
  border-radius: 2px;
  font-weight: 600;
}

.file-size {
  display: none;
}

/* List View */
.file-list {
  display: flex;
  flex-direction: column;
  gap: 0;
}

/* List Header */
.list-header {
  display: grid;
  grid-template-columns: 32px 1fr 100px 140px;
  align-items: center;
  padding: 10px 8px;
  background: #f9fafb;
  border-bottom: 2px solid #e5e7eb;
  font-weight: 600;
  font-size: 13px;
  color: #374151;
  position: sticky;
  top: 0;
  z-index: 10;
}

.header-cell {
  display: flex;
  align-items: center;
  gap: 6px;
  cursor: pointer;
  user-select: none;
  padding: 4px 8px;
  border-radius: 4px;
  transition: all 0.2s;
}

.header-cell:hover {
  background: #f3f4f6;
  color: #1f2937;
}

.header-cell.icon-cell {
  cursor: default;
  pointer-events: none;
}

.sort-indicator {
  color: #3b82f6;
  font-weight: bold;
  font-size: 14px;
}

.list-item {
  display: grid;
  grid-template-columns: 32px 1fr 100px 140px;
  align-items: center;
  padding: 8px;
  background: white;
  border-bottom: 1px solid #f3f4f6;
  cursor: pointer;
  transition: all 0.15s;
  user-select: none;
  min-height: 40px;
}

/* 斑马条纹 */
.list-item:nth-child(even) {
  background: #fafafa;
}

.list-item:hover {
  background: #f3f4f6;
  border-bottom-color: #e5e7eb;
}

.list-item.selected {
  background: #dbeafe;
  border-bottom-color: #93c5fd;
}

.list-item.selected:nth-child(even) {
  background: #dbeafe;
}

.list-item.drop-target {
  background: #f0fdf4;
  border-left: 4px solid #10b981;
  transition: all 0.15s;
}

.list-item.drag-over {
  background: rgba(209, 250, 229, 0.5);
  border-left: 3px solid #10b981;
  box-shadow: 0 0 0 1px #10b981;
  transform: translateX(2px);
}

.list-icon {
  font-size: 18px;
}

/* 文件类型图标颜色 */
.list-item[data-item-type="folder"] .list-icon {
  color: #f59e0b;
  font-size: 20px;
}

.list-item[data-item-type="photo"] .list-icon {
  color: #3b82f6;
}

.list-item[data-item-type="video"] .list-icon {
  color: #ec4899;
}

.list-item[data-item-type="file"] .list-icon {
  color: #6b7280;
}

.list-name {
  font-size: 13px;
  color: #1f2937;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.list-name :deep(mark) {
  background: #fef08a;
  color: #854d0e;
  padding: 1px 2px;
  border-radius: 2px;
  font-weight: 600;
}

.list-size, .list-date {
  font-size: 12px;
  color: #6b7280;
}

/* Selection Rectangle */
.selection-rectangle {
  position: absolute;
  border: 2px solid #3b82f6;
  background: rgba(59, 130, 246, 0.1);
  pointer-events: none;
  z-index: 10;
}

/* Debug Toggle Button */
.debug-toggle {
  position: fixed;
  bottom: 20px;
  right: 20px;
  width: 48px;
  height: 48px;
  border-radius: 50%;
  background: linear-gradient(135deg, #00ff00, #00cc00);
  border: 2px solid #00ff00;
  color: #000;
  font-size: 24px;
  cursor: pointer;
  z-index: 9999;
  box-shadow: 0 4px 16px rgba(0, 255, 0, 0.4);
  transition: all 0.3s;
  display: flex;
  align-items: center;
  justify-content: center;
}

.debug-toggle:hover {
  transform: scale(1.1) rotate(10deg);
  box-shadow: 0 6px 24px rgba(0, 255, 0, 0.6);
}

.debug-toggle:active {
  transform: scale(0.95);
}
</style>