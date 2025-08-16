<template>
  <div class="file-explorer">
    <!-- 左侧目录树 -->
    <div class="sidebar" :class="{ collapsed: sidebarCollapsed }">
      <div class="sidebar-header">
        <n-button quaternary circle size="small" @click="toggleSidebar">
          <template #icon>
            <n-icon><menu-outline /></n-icon>
          </template>
        </n-button>
        <span v-if="!sidebarCollapsed" class="sidebar-title">文件夹</span>
      </div>
      
      <div v-if="!sidebarCollapsed" class="directory-tree">
        <n-tree
          :data="directoryTree"
          :selected-keys="selectedKeys"
          :expanded-keys="expandedKeys"
          selectable
          expand-on-click
          @update:selected-keys="handleDirectorySelect"
          @update:expanded-keys="handleDirectoryExpand"
        >
          <template #prefix="{ option }">
            <n-icon :color="option.type === 'folder' ? 'var(--color-warning)' : 'var(--text-color-secondary)'">
              <folder-outline v-if="option.type === 'folder'" />
              <image-outline v-else-if="option.type === 'images'" />
              <document-outline v-else />
            </n-icon>
          </template>
        </n-tree>
      </div>
    </div>

    <!-- 主内容区域 -->
    <div class="main-content">
      <!-- 顶部工具栏 -->
      <div class="toolbar">
        <div class="toolbar-left">
          <!-- 面包屑导航 -->
          <n-breadcrumb class="breadcrumb">
            <n-breadcrumb-item
              v-for="(item, index) in breadcrumbItems"
              :key="index"
              @click="navigateTo(item.path)"
            >
              <template #icon>
                <n-icon><home-outline v-if="index === 0" /><folder-outline v-else /></n-icon>
              </template>
              {{ item.label }}
            </n-breadcrumb-item>
          </n-breadcrumb>
        </div>

        <div class="toolbar-center">
          <!-- 搜索框 -->
          <n-input
            v-model:value="searchQuery"
            placeholder="搜索文件和文件夹..."
            class="search-input"
            clearable
          >
            <template #prefix>
              <n-icon><search-outline /></n-icon>
            </template>
          </n-input>
        </div>

        <div class="toolbar-right">
          <!-- 视图模式切换 -->
          <n-button-group>
            <n-button
              :type="viewMode === 'grid' ? 'primary' : 'default'"
              quaternary
              @click="setViewMode('grid')"
            >
              <template #icon>
                <n-icon><grid-outline /></n-icon>
              </template>
            </n-button>
            <n-button
              :type="viewMode === 'list' ? 'primary' : 'default'"
              quaternary
              @click="setViewMode('list')"
            >
              <template #icon>
                <n-icon><list-outline /></n-icon>
              </template>
            </n-button>
          </n-button-group>

          <!-- 排序方式 -->
          <n-dropdown :options="sortOptions" @select="handleSortSelect">
            <n-button quaternary>
              <template #icon>
                <n-icon><funnel-outline /></n-icon>
              </template>
              排序
            </n-button>
          </n-dropdown>

          <!-- 缩放控制 -->
          <div class="zoom-controls">
            <n-button quaternary size="small" @click="zoomOut" :disabled="zoomLevel <= 1">
              <template #icon>
                <n-icon><remove-outline /></n-icon>
              </template>
            </n-button>
            <span class="zoom-level">{{ Math.round(zoomLevel * 100) }}%</span>
            <n-button quaternary size="small" @click="zoomIn" :disabled="zoomLevel >= 3">
              <template #icon>
                <n-icon><add-outline /></n-icon>
              </template>
            </n-button>
          </div>

          <!-- 更多操作 -->
          <n-dropdown :options="moreOptions" @select="handleMoreAction">
            <n-button quaternary>
              <template #icon>
                <n-icon><ellipsis-vertical-outline /></n-icon>
              </template>
            </n-button>
          </n-dropdown>
        </div>
      </div>

      <!-- 文件显示区域 -->
      <div class="file-area" @wheel="handleWheel" @click="clearSelection">
        <!-- 网格视图 -->
        <div v-if="viewMode === 'grid'" class="file-grid" :style="{ '--zoom-level': zoomLevel }">
          <div
            v-for="file in filteredFiles"
            :key="file.id"
            class="file-item"
            :class="{ 
              selected: selectedFiles.includes(file.id),
              'is-image': file.type === 'image'
            }"
            @click="handleFileClick(file, $event)"
            @dblclick="handleFileDoubleClick(file)"
          >
            <!-- 图片预览 -->
            <div v-if="file.type === 'image'" class="file-preview image-preview">
              <img :src="file.thumbnail" :alt="file.name" @load="handleImageLoad" />
              <div class="image-overlay">
                <div class="image-info">
                  <span class="image-resolution">{{ file.resolution }}</span>
                </div>
              </div>
            </div>

            <!-- 非图片文件图标 -->
            <div v-else class="file-preview icon-preview" :class="`file-type-${file.type}`">
              <n-icon size="48" :color="getFileTypeColor(file.type)">
                <component :is="getFileIcon(file.type)" />
              </n-icon>
            </div>

            <div class="file-info">
              <div class="file-name" :title="file.name">{{ file.name }}</div>
              <div class="file-meta">
                <span class="file-size">{{ formatFileSize(file.size) }}</span>
                <span class="file-date">{{ formatDate(file.modifiedAt) }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- 列表视图 -->
        <div v-else class="file-list">
          <n-data-table
            :columns="listColumns"
            :data="filteredFiles"
            :row-key="(row) => row.id"
            :checked-row-keys="selectedFiles"
            @update:checked-row-keys="handleRowSelection"
          >
            <template #name="{ row }">
              <div class="list-file-name">
                <n-icon :color="getFileTypeColor(row.type)">
                  <component :is="getFileIcon(row.type)" />
                </n-icon>
                <span>{{ row.name }}</span>
              </div>
            </template>
            <template #size="{ row }">
              {{ formatFileSize(row.size) }}
            </template>
            <template #modifiedAt="{ row }">
              {{ formatDate(row.modifiedAt) }}
            </template>
          </n-data-table>
        </div>

        <!-- 空状态 -->
        <div v-if="filteredFiles.length === 0" class="empty-state">
          <n-empty description="此文件夹为空">
            <template #icon>
              <n-icon size="64" color="var(--text-color-tertiary)">
                <folder-outline />
              </n-icon>
            </template>
          </n-empty>
        </div>
      </div>
    </div>

    <!-- 预览模态框 -->
    <n-modal
      v-model:show="previewVisible"
      preset="card"
      :title="currentPreviewFile?.name"
      size="huge"
      :bordered="false"
      class="preview-modal"
    >
      <div v-if="currentPreviewFile" class="preview-content">
        <img 
          v-if="currentPreviewFile.type === 'image'"
          :src="currentPreviewFile.url"
          :alt="currentPreviewFile.name"
          class="preview-image"
        />
        <div v-else class="preview-placeholder">
          <n-icon size="64" :color="getFileTypeColor(currentPreviewFile.type)">
            <component :is="getFileIcon(currentPreviewFile.type)" />
          </n-icon>
          <p>无法预览此文件类型</p>
        </div>
      </div>
      
      <div class="preview-controls">
        <n-button @click="previousFile" :disabled="currentFileIndex <= 0">
          <template #icon>
            <n-icon><chevron-back-outline /></n-icon>
          </template>
          上一个
        </n-button>
        <span>{{ currentFileIndex + 1 }} / {{ filteredFiles.length }}</span>
        <n-button @click="nextFile" :disabled="currentFileIndex >= filteredFiles.length - 1">
          下一个
          <template #icon>
            <n-icon><chevron-forward-outline /></n-icon>
          </template>
        </n-button>
      </div>
    </n-modal>

    <!-- 右键菜单 -->
    <n-dropdown
      placement="bottom-start"
      trigger="manual"
      :x="contextMenuX"
      :y="contextMenuY"
      :options="contextMenuOptions"
      :show="showContextMenu"
      @select="handleContextMenuSelect"
      @clickoutside="showContextMenu = false"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import {
  NButton,
  NButtonGroup,
  NInput,
  NIcon,
  NTree,
  NBreadcrumb,
  NBreadcrumbItem,
  NDropdown,
  NDataTable,
  NEmpty,
  NModal
} from 'naive-ui'
import {
  MenuOutline,
  FolderOutline,
  ImageOutline,
  DocumentOutline,
  HomeOutline,
  SearchOutline,
  GridOutline,
  ListOutline,
  FunnelOutline,
  AddOutline,
  RemoveOutline,
  EllipsisVerticalOutline,
  ChevronBackOutline,
  ChevronForwardOutline,
  VideocamOutline,
  MusicalNoteOutline,
  ArchiveOutline,
  CodeSlashOutline
} from '@vicons/ionicons5'

// 状态管理
const sidebarCollapsed = ref(false)
const viewMode = ref<'grid' | 'list'>('grid')
const zoomLevel = ref(1.2)
const searchQuery = ref('')
const currentPath = ref('/Photos')
const selectedFiles = ref<string[]>([])
const selectedKeys = ref<string[]>(['photos'])
const expandedKeys = ref<string[]>(['root'])
const previewVisible = ref(false)
const currentPreviewFile = ref<any>(null)
const showContextMenu = ref(false)
const contextMenuX = ref(0)
const contextMenuY = ref(0)

// 假数据
const directoryTree = ref([
  {
    key: 'root',
    label: '我的照片',
    type: 'folder',
    children: [
      {
        key: 'photos',
        label: '照片',
        type: 'images',
        children: [
          { key: 'photos-2024', label: '2024', type: 'folder' },
          { key: 'photos-2023', label: '2023', type: 'folder' }
        ]
      },
      { key: 'videos', label: '视频', type: 'folder' },
      { key: 'documents', label: '文档', type: 'folder' },
      { key: 'downloads', label: '下载', type: 'folder' }
    ]
  }
])

const allFiles = ref([
  {
    id: '1',
    name: 'sunset-beach.jpg',
    type: 'image',
    size: 2048576,
    modifiedAt: '2024-01-15T10:30:00Z',
    thumbnail: 'https://picsum.photos/300/200?random=1',
    url: 'https://picsum.photos/1920/1080?random=1',
    resolution: '1920×1080'
  },
  {
    id: '2',
    name: 'mountain-view.jpg',
    type: 'image',
    size: 3145728,
    modifiedAt: '2024-01-14T15:45:00Z',
    thumbnail: 'https://picsum.photos/300/200?random=2',
    url: 'https://picsum.photos/1920/1080?random=2',
    resolution: '1920×1080'
  },
  {
    id: '3',
    name: 'presentation.pdf',
    type: 'document',
    size: 1048576,
    modifiedAt: '2024-01-13T09:15:00Z'
  },
  {
    id: '4',
    name: 'holiday-video.mp4',
    type: 'video',
    size: 52428800,
    modifiedAt: '2024-01-12T20:00:00Z'
  },
  {
    id: '5',
    name: 'project-files.zip',
    type: 'archive',
    size: 10485760,
    modifiedAt: '2024-01-11T11:30:00Z'
  },
  {
    id: '6',
    name: 'script.js',
    type: 'code',
    size: 4096,
    modifiedAt: '2024-01-10T16:20:00Z'
  }
])

// 计算属性
const filteredFiles = computed(() => {
  let files = allFiles.value
  
  if (searchQuery.value) {
    files = files.filter(file => 
      file.name.toLowerCase().includes(searchQuery.value.toLowerCase())
    )
  }
  
  return files
})

const breadcrumbItems = computed(() => {
  const parts = currentPath.value.split('/').filter(Boolean)
  const items = [{ label: '主页', path: '/' }]
  
  let currentFullPath = ''
  for (const part of parts) {
    currentFullPath += '/' + part
    items.push({ label: part, path: currentFullPath })
  }
  
  return items
})

const currentFileIndex = computed(() => {
  if (!currentPreviewFile.value) return -1
  return filteredFiles.value.findIndex(file => file.id === currentPreviewFile.value.id)
})

// 排序和更多操作选项
const sortOptions = ref([
  { label: '按名称排序', key: 'name' },
  { label: '按大小排序', key: 'size' },
  { label: '按修改时间排序', key: 'modifiedAt' },
  { label: '按类型排序', key: 'type' }
])

const moreOptions = ref([
  { label: '新建文件夹', key: 'new-folder' },
  { label: '上传文件', key: 'upload' },
  { label: '刷新', key: 'refresh' },
  { label: '显示隐藏文件', key: 'show-hidden' }
])

const contextMenuOptions = computed(() => [
  { label: '打开', key: 'open' },
  { label: '重命名', key: 'rename' },
  { label: '删除', key: 'delete', disabled: selectedFiles.value.length === 0 },
  { type: 'divider' },
  { label: '复制', key: 'copy' },
  { label: '剪切', key: 'cut' },
  { label: '粘贴', key: 'paste' },
  { type: 'divider' },
  { label: '属性', key: 'properties' }
])

// 列表视图的表格列定义
const listColumns = [
  { type: 'selection' },
  { title: '名称', key: 'name', render: 'name' },
  { title: '大小', key: 'size', render: 'size', width: 120 },
  { title: '修改时间', key: 'modifiedAt', render: 'modifiedAt', width: 180 }
]

// 方法
const toggleSidebar = () => {
  sidebarCollapsed.value = !sidebarCollapsed.value
}

const setViewMode = (mode: 'grid' | 'list') => {
  viewMode.value = mode
}

const zoomIn = () => {
  if (zoomLevel.value < 3) {
    zoomLevel.value += 0.2
  }
}

const zoomOut = () => {
  if (zoomLevel.value > 0.6) {
    zoomLevel.value -= 0.2
  }
}

const handleWheel = (e: WheelEvent) => {
  if (e.ctrlKey || e.metaKey) {
    e.preventDefault()
    if (e.deltaY < 0) {
      zoomIn()
    } else {
      zoomOut()
    }
  }
}

const handleDirectorySelect = (keys: string[]) => {
  selectedKeys.value = keys
  if (keys.length > 0) {
    // 这里可以根据选中的目录加载相应的文件
    console.log('Selected directory:', keys[0])
  }
}

const handleDirectoryExpand = (keys: string[]) => {
  expandedKeys.value = keys
}

const navigateTo = (path: string) => {
  currentPath.value = path
  // 加载对应路径的文件
}

const handleFileClick = (file: any, event: MouseEvent) => {
  event.stopPropagation()
  
  if (event.ctrlKey || event.metaKey) {
    // Ctrl+点击多选
    if (selectedFiles.value.includes(file.id)) {
      selectedFiles.value = selectedFiles.value.filter(id => id !== file.id)
    } else {
      selectedFiles.value.push(file.id)
    }
  } else if (event.shiftKey && selectedFiles.value.length > 0) {
    // Shift+点击范围选择
    const lastSelected = selectedFiles.value[selectedFiles.value.length - 1]
    const lastIndex = filteredFiles.value.findIndex(f => f.id === lastSelected)
    const currentIndex = filteredFiles.value.findIndex(f => f.id === file.id)
    
    const start = Math.min(lastIndex, currentIndex)
    const end = Math.max(lastIndex, currentIndex)
    
    selectedFiles.value = filteredFiles.value.slice(start, end + 1).map(f => f.id)
  } else {
    // 单选
    selectedFiles.value = [file.id]
  }
}

const handleFileDoubleClick = (file: any) => {
  if (file.type === 'image') {
    currentPreviewFile.value = file
    previewVisible.value = true
  } else {
    // 其他文件类型的打开逻辑
    console.log('Opening file:', file.name)
  }
}

const clearSelection = () => {
  selectedFiles.value = []
}

const handleRowSelection = (keys: string[]) => {
  selectedFiles.value = keys
}

const previousFile = () => {
  if (currentFileIndex.value > 0) {
    currentPreviewFile.value = filteredFiles.value[currentFileIndex.value - 1]
  }
}

const nextFile = () => {
  if (currentFileIndex.value < filteredFiles.value.length - 1) {
    currentPreviewFile.value = filteredFiles.value[currentFileIndex.value + 1]
  }
}

const handleSortSelect = (key: string) => {
  console.log('Sort by:', key)
}

const handleMoreAction = (key: string) => {
  console.log('More action:', key)
}

const handleContextMenuSelect = (key: string) => {
  console.log('Context menu:', key)
  showContextMenu.value = false
}

const handleImageLoad = (event: Event) => {
  // 图片加载完成的处理
}

// 工具函数
const getFileIcon = (type: string) => {
  const iconMap: { [key: string]: any } = {
    image: ImageOutline,
    video: VideocamOutline,
    audio: MusicalNoteOutline,
    document: DocumentOutline,
    archive: ArchiveOutline,
    code: CodeSlashOutline,
    folder: FolderOutline
  }
  return iconMap[type] || DocumentOutline
}

const getFileTypeColor = (type: string) => {
  const colorMap: { [key: string]: string } = {
    image: 'var(--color-success)',
    video: 'var(--color-info)',
    audio: 'var(--color-secondary)',
    document: 'var(--text-color-secondary)',
    archive: 'var(--color-warning)',
    code: 'var(--color-primary)',
    folder: 'var(--color-warning)'
  }
  return colorMap[type] || 'var(--text-color-secondary)'
}

const formatFileSize = (bytes: number) => {
  const sizes = ['B', 'KB', 'MB', 'GB']
  if (bytes === 0) return '0 B'
  const i = Math.floor(Math.log(bytes) / Math.log(1024))
  return Math.round(bytes / Math.pow(1024, i) * 100) / 100 + ' ' + sizes[i]
}

const formatDate = (dateString: string) => {
  const date = new Date(dateString)
  return date.toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}

onMounted(() => {
  // 组件挂载后的初始化
})
</script>

<style scoped>
.file-explorer {
  display: flex;
  height: 100vh;
  background: var(--bg-color);
  overflow: hidden;
}

/* 左侧边栏 */
.sidebar {
  width: 280px;
  background: var(--sidebar-bg);
  border-right: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
  transition: width var(--transition-normal);
}

.sidebar.collapsed {
  width: 60px;
}

.sidebar-header {
  display: flex;
  align-items: center;
  padding: var(--spacing-md);
  border-bottom: 1px solid var(--border-color-light);
  gap: var(--spacing-sm);
}

.sidebar-title {
  font-weight: var(--font-weight-semibold);
  color: var(--text-color);
}

.directory-tree {
  flex: 1;
  padding: var(--spacing-sm);
  overflow: auto;
}

/* 主内容区域 */
.main-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

/* 工具栏 */
.toolbar {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  padding: var(--spacing-md) var(--spacing-lg);
  background: var(--header-bg);
  border-bottom: 1px solid var(--border-color-light);
  min-height: 60px;
}

.toolbar-left {
  min-width: 200px;
}

.toolbar-center {
  flex: 1;
  max-width: 400px;
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
}

.search-input {
  width: 100%;
}

.breadcrumb :deep(.n-breadcrumb-item) {
  cursor: pointer;
}

.zoom-controls {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
  padding: 0 var(--spacing-sm);
}

.zoom-level {
  font-size: var(--font-size-sm);
  color: var(--text-color-secondary);
  min-width: 40px;
  text-align: center;
}

/* 文件区域 */
.file-area {
  flex: 1;
  padding: var(--spacing-lg);
  overflow: auto;
  background: var(--content-bg);
}

/* 网格视图 */
.file-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(calc(160px * var(--zoom-level)), 1fr));
  gap: calc(var(--spacing-md) * var(--zoom-level));
  padding: var(--spacing-sm);
}

.file-item {
  background: var(--card-bg);
  border: 2px solid transparent;
  border-radius: var(--radius-md);
  padding: var(--spacing-sm);
  cursor: pointer;
  transition: all var(--transition-fast);
  user-select: none;
}

.file-item:hover {
  border-color: var(--border-color);
  transform: translateY(-2px);
  box-shadow: var(--shadow-md);
}

.file-item.selected {
  border-color: var(--color-primary);
  background: rgba(59, 130, 246, 0.08);
}

.file-item.is-image .file-preview {
  aspect-ratio: 4/3;
}

.file-preview {
  position: relative;
  border-radius: var(--radius-sm);
  overflow: hidden;
  margin-bottom: var(--spacing-sm);
  background: var(--bg-color-secondary);
}

.image-preview img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.image-overlay {
  position: absolute;
  inset: 0;
  background: linear-gradient(to bottom, transparent 60%, rgba(0,0,0,0.7));
  opacity: 0;
  transition: opacity var(--transition-fast);
  display: flex;
  align-items: flex-end;
}

.file-item:hover .image-overlay {
  opacity: 1;
}

.image-info {
  padding: var(--spacing-sm);
  color: white;
  font-size: var(--font-size-xs);
}

.icon-preview {
  display: flex;
  align-items: center;
  justify-content: center;
  aspect-ratio: 4/3;
  background: var(--bg-color-tertiary);
}

.file-info {
  text-align: center;
}

.file-name {
  font-weight: var(--font-weight-medium);
  color: var(--text-color);
  font-size: var(--font-size-sm);
  margin-bottom: var(--spacing-xs);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-meta {
  display: flex;
  justify-content: space-between;
  font-size: var(--font-size-xs);
  color: var(--text-color-secondary);
  gap: var(--spacing-xs);
}

/* 列表视图 */
.file-list {
  background: var(--card-bg);
  border-radius: var(--radius-md);
  overflow: hidden;
}

.list-file-name {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
}

/* 空状态 */
.empty-state {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 300px;
}

/* 预览模态框 */
.preview-modal {
  max-width: 90vw;
  max-height: 90vh;
}

.preview-content {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 400px;
  background: var(--bg-color-secondary);
  border-radius: var(--radius-md);
  overflow: hidden;
}

.preview-image {
  max-width: 100%;
  max-height: 70vh;
  object-fit: contain;
}

.preview-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--spacing-md);
  color: var(--text-color-secondary);
}

.preview-controls {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: var(--spacing-lg);
  padding-top: var(--spacing-lg);
  border-top: 1px solid var(--border-color-light);
}

/* 文件类型颜色 */
.file-type-document {
  background: linear-gradient(135deg, #f8f9fa 0%, #e9ecef 100%);
}

.file-type-video {
  background: linear-gradient(135deg, #e3f2fd 0%, #bbdefb 100%);
}

.file-type-audio {
  background: linear-gradient(135deg, #f3e5f5 0%, #e1bee7 100%);
}

.file-type-archive {
  background: linear-gradient(135deg, #fff8e1 0%, #ffecb3 100%);
}

.file-type-code {
  background: linear-gradient(135deg, #e8f5e8 0%, #c8e6c9 100%);
}

/* 响应式设计 */
@media (max-width: 768px) {
  .file-explorer {
    flex-direction: column;
  }
  
  .sidebar {
    width: 100%;
    height: auto;
    max-height: 200px;
  }
  
  .toolbar {
    flex-wrap: wrap;
    gap: var(--spacing-sm);
  }
  
  .toolbar-left,
  .toolbar-center {
    min-width: auto;
    flex: 1;
  }
  
  .file-grid {
    grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
  }
}

@media (max-width: 480px) {
  .file-grid {
    grid-template-columns: repeat(auto-fill, minmax(100px, 1fr));
  }
}
</style>