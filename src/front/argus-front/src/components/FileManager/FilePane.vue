<template>
  <div :class="['file-pane', { active: isActive, inactive: !isActive }]" @click="$emit('activate')">
    <!-- Top indicator -->
    <div v-if="isActive" class="active-indicator"></div>

    <!-- Breadcrumb -->
    <div class="breadcrumb" @click.stop="$emit('activate')">
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

      <!-- Search Input -->
      <div class="search-box">
        <input
          v-model="searchQuery"
          type="text"
          class="search-input"
          placeholder="搜索..."
          @click.stop
        />
        <button v-if="searchQuery" class="search-clear" @click.stop="searchQuery = ''">×</button>
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
    >
      <!-- Grid View -->
      <div v-if="viewMode === 'grid'" :class="['file-grid', `grid-${thumbnailSize}`]" :style="gridItemSize">
        <div
          v-for="(item, index) in visibleItems"
          :key="item.name"
          :class="['file-item', {
            selected: selection.isSelected(item.name),
            focused: isFocused(item.name),
            'drop-target': item.type === 'folder' && dragState.isDragging && !selection.isSelected(item.name),
            'drag-over': dragState.dropTarget === item.name
          }]"
          :data-item-name="item.name"
          :data-item-index="index"
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

      <!-- List View -->
      <div v-else class="file-list">
        <div
          v-for="(item, index) in visibleItems"
          :key="item.name"
          :class="['list-item', {
            selected: selection.isSelected(item.name),
            focused: isFocused(item.name),
            'drop-target': item.type === 'folder' && dragState.isDragging && !selection.isSelected(item.name),
            'drag-over': dragState.dropTarget === item.name
          }]"
          :data-item-name="item.name"
          :data-item-index="index"
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

      <!-- Selection Rectangle -->
      <div v-if="dragSelection.isSelecting" class="selection-rectangle" :style="selectionRectStyle"></div>
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
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted, nextTick } from 'vue'
import { useFileSelection } from '@/composables/fileManager/useFileSelection'
import { useKeyboardNav } from '@/composables/fileManager/useKeyboardNav'
import { mockFolderStructure } from './mockData'
import type { FileItem, ViewMode, ThumbnailSize, PaneId } from './types'
import ContextMenu from './ContextMenu.vue'
import Tooltip from './Tooltip.vue'

const props = defineProps<{
  paneId: PaneId
  viewMode: ViewMode
  thumbnailSize: ThumbnailSize
  isActive: boolean
}>()

const emit = defineEmits<{
  activate: []
}>()

// State
const currentPath = ref<string[]>(['Home'])
const contentAreaRef = ref<HTMLElement>()
const tooltipRef = ref<InstanceType<typeof Tooltip>>()
const pathInputRef = ref<HTMLInputElement>()
const selection = useFileSelection()
const searchQuery = ref('')
const pathEditMode = ref(false)
const pathEditValue = ref('')
const zoomLevel = ref(100) // 缩放级别：50-200%

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

// Drag selection state
const dragSelection = ref({
  isSelecting: false,
  startX: 0,
  startY: 0,
  currentX: 0,
  currentY: 0,
  initialSelections: new Set<string>(),
  ctrlKey: false,
  justFinished: false
})

// Auto scroll state
const autoScroll = ref({
  isScrolling: false,
  direction: { x: 0, y: 0 }
})

// Drag and drop state
const dragState = ref({
  isDragging: false,
  draggedItems: [] as string[],
  dropTarget: null as string | null,
  sourcePane: null as string | null,
  isPaneDragOver: false
})

// Computed
const currentFolder = computed(() => {
  let folder = mockFolderStructure
  for (const segment of currentPath.value) {
    if (folder[segment] && folder[segment].children) {
      folder = folder[segment].children as any
    } else {
      return null
    }
  }
  return folder
})

const visibleItems = computed(() => {
  if (!currentFolder.value) return []

  let items = Object.values(currentFolder.value).sort((a, b) => {
    if (a.type === 'folder' && b.type !== 'folder') return -1
    if (a.type !== 'folder' && b.type === 'folder') return 1
    return a.name.localeCompare(b.name)
  })

  // 搜索过滤
  if (searchQuery.value.trim()) {
    const query = searchQuery.value.toLowerCase()
    items = items.filter(item => item.name.toLowerCase().includes(query))
  }

  return items
})

const selectionRectStyle = computed(() => {
  const left = Math.min(dragSelection.value.startX, dragSelection.value.currentX)
  const top = Math.min(dragSelection.value.startY, dragSelection.value.currentY)
  const width = Math.abs(dragSelection.value.currentX - dragSelection.value.startX)
  const height = Math.abs(dragSelection.value.currentY - dragSelection.value.startY)

  return {
    left: `${left}px`,
    top: `${top}px`,
    width: `${width}px`,
    height: `${height}px`
  }
})

// Methods
function highlightText(text: string): string {
  if (!searchQuery.value.trim()) {
    return text
  }

  const query = searchQuery.value.trim()
  const regex = new RegExp(`(${query})`, 'gi')
  return text.replace(regex, '<mark>$1</mark>')
}

function navigateToIndex(index: number) {
  currentPath.value = currentPath.value.slice(0, index + 1)
  selection.clearSelection()
  searchQuery.value = ''
  breadcrumbDropdown.value.visible = false
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
  selection.clearSelection()
  searchQuery.value = ''
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
  selection.clearSelection()
  searchQuery.value = ''
  pathEditMode.value = false
}

function cancelPathEdit() {
  pathEditMode.value = false
  pathEditValue.value = ''
}

function navigateToFolder(folderName: string) {
  currentPath.value.push(folderName)
  selection.clearSelection()
  searchQuery.value = ''
}

function isFocused(itemName: string): boolean {
  return selection.focusedItem.value?.name === itemName
}

function handleItemClick(event: MouseEvent, itemName: string, index: number) {
  event.stopPropagation()

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

  const rect = contentAreaRef.value?.getBoundingClientRect()
  if (!rect) return

  const scrollEl = contentAreaRef.value!
  dragSelection.value = {
    isSelecting: true,
    startX: event.clientX - rect.left + scrollEl.scrollLeft,
    startY: event.clientY - rect.top + scrollEl.scrollTop,
    currentX: event.clientX - rect.left + scrollEl.scrollLeft,
    currentY: event.clientY - rect.top + scrollEl.scrollTop,
    initialSelections: new Set(selection.selectedItems.value),
    ctrlKey: event.ctrlKey || event.metaKey,
    justFinished: false
  }

  if (!dragSelection.value.ctrlKey) {
    selection.clearSelection()
  }

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
  if (!dragSelection.value.isSelecting) return

  const rect = contentAreaRef.value?.getBoundingClientRect()
  if (!rect) return

  const scrollEl = contentAreaRef.value!

  // 计算鼠标相对于内容区域的位置（包含滚动偏移）
  // 鼠标在视口中的位置 + 滚动偏移 = 在文档中的绝对位置
  const rawX = event.clientX - rect.left + scrollEl.scrollLeft
  const rawY = event.clientY - rect.top + scrollEl.scrollTop

  // 限制坐标在合理范围内
  // 当鼠标超出边界时，坐标应该停留在边界上，而不是继续增长
  const maxScrollLeft = Math.max(0, scrollEl.scrollWidth - scrollEl.clientWidth)
  const maxScrollTop = Math.max(0, scrollEl.scrollHeight - scrollEl.clientHeight)

  // 最大坐标 = 可见区域大小 + 最大滚动距离
  const maxX = scrollEl.clientWidth + maxScrollLeft
  const maxY = scrollEl.clientHeight + maxScrollTop

  const relativeX = Math.max(0, Math.min(rawX, maxX))
  const relativeY = Math.max(0, Math.min(rawY, maxY))

  // 存储坐标
  dragSelection.value.currentX = relativeX
  dragSelection.value.currentY = relativeY

  // 自动滚动检测
  checkAutoScroll(event)

  // 更新框选
  updateDragSelection()
}

function handleMouseUp() {
  if (dragSelection.value.isSelecting) {
    // 最后更新一次选择
    updateDragSelection()

    dragSelection.value.isSelecting = false
    autoScroll.value.isScrolling = false

    // 标记刚完成拖拽，防止 handleContentClick 清除选择
    dragSelection.value.justFinished = true
    setTimeout(() => {
      dragSelection.value.justFinished = false
    }, 100)

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

// 检查是否需要自动滚动
function checkAutoScroll(event: MouseEvent) {
  if (!contentAreaRef.value) return

  const rect = contentAreaRef.value.getBoundingClientRect()
  const scrollEl = contentAreaRef.value
  const threshold = 50
  const baseScrollSpeed = 10

  let dx = 0
  let dy = 0

  // 检查垂直滚动
  if (event.clientY < rect.top) {
    dy = -baseScrollSpeed
  } else if (event.clientY < rect.top + threshold) {
    const distance = event.clientY - rect.top
    const ratio = 1 - distance / threshold
    dy = -Math.max(3, baseScrollSpeed * ratio)
  } else if (event.clientY > rect.bottom) {
    dy = baseScrollSpeed
  } else if (event.clientY > rect.bottom - threshold) {
    const distance = rect.bottom - event.clientY
    const ratio = 1 - distance / threshold
    dy = Math.max(3, baseScrollSpeed * ratio)
  }

  // 检查水平滚动 - 只在确实有横向滚动空间时才启用
  const hasHorizontalScroll = scrollEl.scrollWidth > scrollEl.clientWidth

  if (hasHorizontalScroll) {
    if (event.clientX < rect.left) {
      dx = -baseScrollSpeed
    } else if (event.clientX < rect.left + threshold) {
      const distance = event.clientX - rect.left
      const ratio = 1 - distance / threshold
      dx = -Math.max(3, baseScrollSpeed * ratio)
    } else if (event.clientX > rect.right) {
      dx = baseScrollSpeed
    } else if (event.clientX > rect.right - threshold) {
      const distance = rect.right - event.clientX
      const ratio = 1 - distance / threshold
      dx = Math.max(3, baseScrollSpeed * ratio)
    }
  }

  if (dx !== 0 || dy !== 0) {
    if (!autoScroll.value.isScrolling) {
      autoScroll.value.isScrolling = true
      autoScroll.value.direction = { x: dx, y: dy }
      startAutoScroll()
    } else {
      autoScroll.value.direction = { x: dx, y: dy }
    }
  } else {
    autoScroll.value.isScrolling = false
  }
}

// 开始自动滚动
function startAutoScroll() {
  let frameId: number | null = null

  function scroll() {
    if (!autoScroll.value.isScrolling || !contentAreaRef.value) {
      if (frameId !== null) {
        cancelAnimationFrame(frameId)
        frameId = null
      }
      return
    }

    const scrollEl = contentAreaRef.value

    // 计算新的滚动位置，并限制在有效范围内
    const maxScrollLeft = Math.max(0, scrollEl.scrollWidth - scrollEl.clientWidth)
    const maxScrollTop = Math.max(0, scrollEl.scrollHeight - scrollEl.clientHeight)

    let newScrollLeft = scrollEl.scrollLeft + autoScroll.value.direction.x
    let newScrollTop = scrollEl.scrollTop + autoScroll.value.direction.y

    // 限制滚动范围
    newScrollLeft = Math.max(0, Math.min(newScrollLeft, maxScrollLeft))
    newScrollTop = Math.max(0, Math.min(newScrollTop, maxScrollTop))

    // 计算实际滚动距离
    const actualDx = newScrollLeft - scrollEl.scrollLeft
    const actualDy = newScrollTop - scrollEl.scrollTop

    // 应用滚动
    scrollEl.scrollLeft = newScrollLeft
    scrollEl.scrollTop = newScrollTop

    // 更新拖拽坐标（只在实际发生滚动时更新）
    if (actualDx !== 0 || actualDy !== 0) {
      dragSelection.value.currentX += actualDx
      dragSelection.value.currentY += actualDy

      // 确保坐标不超出内容范围
      const maxX = scrollEl.clientWidth + maxScrollLeft
      const maxY = scrollEl.clientHeight + maxScrollTop
      dragSelection.value.currentX = Math.max(0, Math.min(dragSelection.value.currentX, maxX))
      dragSelection.value.currentY = Math.max(0, Math.min(dragSelection.value.currentY, maxY))

      updateDragSelection()
    }

    frameId = requestAnimationFrame(scroll)
  }

  frameId = requestAnimationFrame(scroll)
}

// 更新框选的项
function updateDragSelection() {
  if (!contentAreaRef.value) return

  const items = contentAreaRef.value.querySelectorAll('.file-item, .list-item')
  const selectionLeft = Math.min(dragSelection.value.startX, dragSelection.value.currentX)
  const selectionTop = Math.min(dragSelection.value.startY, dragSelection.value.currentY)
  const selectionRight = Math.max(dragSelection.value.startX, dragSelection.value.currentX)
  const selectionBottom = Math.max(dragSelection.value.startY, dragSelection.value.currentY)

  const scrollEl = contentAreaRef.value
  const contentRect = scrollEl.getBoundingClientRect()

  const itemsInSelection = new Set<string>()

  items.forEach(item => {
    const itemRect = item.getBoundingClientRect()
    const itemLeft = itemRect.left - contentRect.left + scrollEl.scrollLeft
    const itemTop = itemRect.top - contentRect.top + scrollEl.scrollTop
    const itemRight = itemLeft + itemRect.width
    const itemBottom = itemTop + itemRect.height

    const intersects = !(
      itemRight < selectionLeft ||
      itemLeft > selectionRight ||
      itemBottom < selectionTop ||
      itemTop > selectionBottom
    )

    if (intersects) {
      const itemName = (item as HTMLElement).dataset.itemName
      if (itemName) {
        itemsInSelection.add(itemName)
      }
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

  // 设置拖动状态
  dragState.value.isDragging = true
  dragState.value.draggedItems = Array.from(selection.selectedItems.value)
  dragState.value.sourcePane = props.paneId

  // 设置拖动数据
  const dragData = {
    items: dragState.value.draggedItems,
    sourcePane: props.paneId,
    sourcePath: currentPath.value
  }
  event.dataTransfer!.effectAllowed = 'move'
  event.dataTransfer!.setData('application/json', JSON.stringify(dragData))

  // 隐藏tooltip
  tooltipRef.value?.hide()
}

function handleDragEnd() {
  dragState.value.isDragging = false
  dragState.value.dropTarget = null
  dragState.value.sourcePane = null
  dragState.value.isPaneDragOver = false
}

function handleDragOver(event: DragEvent, item: FileItem) {
  // 只有文件夹可以作为drop target
  if (item.type !== 'folder') return

  // 不能拖到自己身上
  if (selection.isSelected(item.name)) return

  event.preventDefault()
  event.dataTransfer!.dropEffect = 'move'
  dragState.value.dropTarget = item.name
}

function handleDragLeave(event: DragEvent) {
  const target = event.target as HTMLElement
  const relatedTarget = event.relatedTarget as HTMLElement

  // 只有真正离开元素时才清除
  if (!target.contains(relatedTarget)) {
    dragState.value.dropTarget = null
  }
}

function handleDrop(event: DragEvent, targetItem: FileItem) {
  event.preventDefault()

  if (targetItem.type !== 'folder') return
  if (selection.isSelected(targetItem.name)) return

  try {
    const dragData = JSON.parse(event.dataTransfer!.getData('application/json'))

    // 显示移动操作（实际实现需要状态管理）
    const itemNames = dragData.items.join(', ')
    console.log(`移动文件: ${itemNames} 到 ${targetItem.name}`)
    alert(`已移动 ${dragData.items.length} 个项目到 "${targetItem.name}"`)

    // TODO: 实现实际的文件移动逻辑
    // 这需要:
    // 1. 更新mockData或使用真实的文件系统API
    // 2. 从源位置移除文件
    // 3. 添加到目标位置
    // 4. 刷新视图

  } catch (e) {
    console.error('Drop failed:', e)
  }

  dragState.value.isDragging = false
  dragState.value.dropTarget = null
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
      dragState.value.isPaneDragOver = true
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
    dragState.value.isPaneDragOver = false
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
    console.log(`移动文件: ${dragData.items.join(', ')} 到当前文件夹 ${currentFolderName}`)

    if (dragData.sourcePane !== props.paneId) {
      alert(`已从${dragData.sourcePane === 'left' ? '左' : '右'}面板移动 ${dragData.items.length} 个项目到${props.paneId === 'left' ? '左' : '右'}面板的 "${currentFolderName}"`)
    } else {
      alert(`在当前文件夹内操作`)
    }

    // TODO: 实现实际的文件移动逻辑

  } catch (e) {
    console.error('Pane drop failed:', e)
  }

  dragState.value.isPaneDragOver = false
  dragState.value.isDragging = false
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
watch(searchQuery, () => {
  selection.clearSelection()
})

// 关闭breadcrumb dropdown 和 添加滚轮监听
onMounted(() => {
  const closeDropdown = () => {
    breadcrumbDropdown.value.visible = false
  }
  document.addEventListener('click', closeDropdown)

  if (contentAreaRef.value) {
    contentAreaRef.value.addEventListener('wheel', handleWheel, { passive: false })
  }
})

onUnmounted(() => {
  document.removeEventListener('click', () => {
    breadcrumbDropdown.value.visible = false
  })

  if (contentAreaRef.value) {
    contentAreaRef.value.removeEventListener('wheel', handleWheel)
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
  onSpace: () => {
    const focused = selection.focusedItem.value
    if (focused) {
      selection.toggleItemSelection(focused.name, focused.index)
    }
  },
  onSelectAll: () => {
    selection.selectAll(visibleItems.value)
  },
  getGridColumns
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

/* Breadcrumb */
.breadcrumb {
  padding: 12px 16px;
  border-bottom: 1px solid #f3f4f6;
  background: #f9fafb;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  cursor: pointer;
  transition: background-color 0.15s;
}

.breadcrumb:hover {
  background: rgba(59, 130, 246, 0.03);
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

/* Search Box */
.search-box {
  position: relative;
  flex-shrink: 0;
  width: 200px;
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
  transition: background-color 0.2s;
}

.content-area.pane-drag-over {
  background: #f0fdf4;
  outline: 2px dashed #10b981;
  outline-offset: -8px;
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
}

.file-item.drag-over {
  border-color: #10b981;
  background: #d1fae5;
  transform: scale(1.02);
  box-shadow: 0 0 0 2px #10b981;
}

.file-icon {
  font-size: 40px;
  margin-bottom: 4px;
  flex-shrink: 0;
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

.list-item {
  display: grid;
  grid-template-columns: 32px 1fr 100px 140px;
  align-items: center;
  padding: 4px 8px;
  background: white;
  border: 1px solid transparent;
  cursor: pointer;
  transition: background 0.1s;
  user-select: none;
  min-height: 32px;
}

.list-item:hover {
  background: #f3f4f6;
  border-color: #e5e7eb;
}

.list-item.selected {
  background: #cce8ff;
  border-color: #99d1ff;
}

.list-item.drop-target {
  background: #f0fdf4;
  border-left: 3px solid #10b981;
}

.list-item.drag-over {
  background: #d1fae5;
  border-left: 3px solid #10b981;
  box-shadow: 0 0 0 1px #10b981;
}

.list-icon {
  font-size: 18px;
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
</style>