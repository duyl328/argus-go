<template>
  <div class="page-container">
    <!-- 页面标题 -->
    <div class="page-header">
      <h1 class="page-title">
        <n-icon class="title-icon">
          <archive-outline />
        </n-icon>
        本地文件
      </h1>
      <div class="page-actions">
        <n-button type="primary">
          <template #icon>
            <n-icon><cloud-upload-outline /></n-icon>
          </template>
          上传文件
        </n-button>
        <n-button>
          <template #icon>
            <n-icon><folder-open-outline /></n-icon>
          </template>
          选择文件夹
        </n-button>
      </div>
    </div>

    <!-- 搜索和筛选 -->
    <div class="toolbar">
      <div class="toolbar-left">
        <n-input
          v-model:value="searchQuery"
          placeholder="搜索本地文件..."
          clearable
          style="width: 300px"
        >
          <template #prefix>
            <n-icon><search-outline /></n-icon>
          </template>
        </n-input>
      </div>
      <div class="toolbar-right">
        <n-select
          v-model:value="fileTypeFilter"
          placeholder="文件类型"
          :options="fileTypeOptions"
          style="width: 150px"
        />
        <n-button-group>
          <n-button
            :type="viewMode === 'grid' ? 'primary' : 'default'"
            @click="setViewMode('grid')"
          >
            <template #icon>
              <n-icon><grid-outline /></n-icon>
            </template>
          </n-button>
          <n-button
            :type="viewMode === 'list' ? 'primary' : 'default'"
            @click="setViewMode('list')"
          >
            <template #icon>
              <n-icon><list-outline /></n-icon>
            </template>
          </n-button>
        </n-button-group>
      </div>
    </div>

    <!-- 文件显示区域 -->
    <div class="content-area">
      <!-- 网格视图 -->
      <div v-if="viewMode === 'grid'" class="file-grid">
        <div
          v-for="file in filteredFiles"
          :key="file.id"
          class="file-card"
          @click="handleFileClick(file)"
        >
          <div class="file-preview">
            <n-icon size="48" :color="getFileTypeColor(file.type)">
              <component :is="getFileIcon(file.type)" />
            </n-icon>
          </div>
          <div class="file-info">
            <div class="file-name" :title="file.name">{{ file.name }}</div>
            <div class="file-meta">
              <span class="file-size">{{ formatFileSize(file.size) }}</span>
              <span class="file-date">{{ formatDate(file.lastModified) }}</span>
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
          :pagination="pagination"
        />
      </div>

      <!-- 空状态 -->
      <div v-if="filteredFiles.length === 0" class="empty-state">
        <n-empty description="暂无本地文件">
          <template #icon>
            <n-icon size="64" color="var(--text-color-tertiary)">
              <archive-outline />
            </n-icon>
          </template>
          <template #extra>
            <n-button type="primary" @click="handleSelectFiles">
              选择文件
            </n-button>
          </template>
        </n-empty>
      </div>
    </div>

    <!-- 文件详情模态框 -->
    <n-modal
      v-model:show="detailVisible"
      preset="card"
      title="文件详情"
      size="medium"
    >
      <div v-if="selectedFile" class="file-details">
        <div class="detail-item">
          <strong>文件名：</strong>{{ selectedFile.name }}
        </div>
        <div class="detail-item">
          <strong>文件大小：</strong>{{ formatFileSize(selectedFile.size) }}
        </div>
        <div class="detail-item">
          <strong>文件类型：</strong>{{ selectedFile.type }}
        </div>
        <div class="detail-item">
          <strong>最后修改：</strong>{{ formatDate(selectedFile.lastModified) }}
        </div>
        <div class="detail-item">
          <strong>文件路径：</strong>{{ selectedFile.path }}
        </div>
      </div>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import {
  NButton,
  NButtonGroup,
  NInput,
  NIcon,
  NSelect,
  NDataTable,
  NEmpty,
  NModal
} from 'naive-ui'
import {
  ArchiveOutline,
  CloudUploadOutline,
  FolderOpenOutline,
  SearchOutline,
  GridOutline,
  ListOutline,
  DocumentOutline,
  ImageOutline,
  VideocamOutline,
  MusicalNoteOutline
} from '@vicons/ionicons5'

// 响应式数据
const searchQuery = ref('')
const fileTypeFilter = ref('')
const viewMode = ref<'grid' | 'list'>('grid')
const detailVisible = ref(false)
const selectedFile = ref<any>(null)

// 文件类型选项
const fileTypeOptions = [
  { label: '全部类型', value: '' },
  { label: '图片', value: 'image' },
  { label: '视频', value: 'video' },
  { label: '音频', value: 'audio' },
  { label: '文档', value: 'document' }
]

// 模拟本地文件数据
const localFiles = ref([
  {
    id: '1',
    name: 'vacation-photos.zip',
    type: 'archive',
    size: 25600000,
    lastModified: new Date('2024-01-15'),
    path: '/Users/Downloads/vacation-photos.zip'
  },
  {
    id: '2',
    name: 'presentation.pdf',
    type: 'document',
    size: 2048576,
    lastModified: new Date('2024-01-14'),
    path: '/Users/Documents/presentation.pdf'
  },
  {
    id: '3',
    name: 'background-music.mp3',
    type: 'audio',
    size: 5242880,
    lastModified: new Date('2024-01-13'),
    path: '/Users/Music/background-music.mp3'
  },
  {
    id: '4',
    name: 'family-video.mp4',
    type: 'video',
    size: 104857600,
    lastModified: new Date('2024-01-12'),
    path: '/Users/Videos/family-video.mp4'
  },
  {
    id: '5',
    name: 'screenshot.png',
    type: 'image',
    size: 1024000,
    lastModified: new Date('2024-01-11'),
    path: '/Users/Pictures/screenshot.png'
  }
])

// 计算属性
const filteredFiles = computed(() => {
  let files = localFiles.value

  // 按文件类型筛选
  if (fileTypeFilter.value) {
    files = files.filter(file => file.type === fileTypeFilter.value)
  }

  // 按搜索词筛选
  if (searchQuery.value) {
    files = files.filter(file =>
      file.name.toLowerCase().includes(searchQuery.value.toLowerCase())
    )
  }

  return files
})

// 分页配置
const pagination = {
  pageSize: 10
}

// 列表视图列定义
const listColumns = [
  {
    title: '文件名',
    key: 'name',
    render: (row: any) => {
      return `${row.name}`
    }
  },
  {
    title: '类型',
    key: 'type',
    width: 100
  },
  {
    title: '大小',
    key: 'size',
    width: 120,
    render: (row: any) => formatFileSize(row.size)
  },
  {
    title: '修改时间',
    key: 'lastModified',
    width: 180,
    render: (row: any) => formatDate(row.lastModified)
  }
]

// 方法
const setViewMode = (mode: 'grid' | 'list') => {
  viewMode.value = mode
}

const handleFileClick = (file: any) => {
  selectedFile.value = file
  detailVisible.value = true
}

const handleSelectFiles = () => {
  // 处理文件选择逻辑
  console.log('选择文件')
}

const getFileIcon = (type: string) => {
  const iconMap: { [key: string]: any } = {
    image: ImageOutline,
    video: VideocamOutline,
    audio: MusicalNoteOutline,
    document: DocumentOutline,
    archive: ArchiveOutline
  }
  return iconMap[type] || DocumentOutline
}

const getFileTypeColor = (type: string) => {
  const colorMap: { [key: string]: string } = {
    image: 'var(--color-success)',
    video: 'var(--color-info)',
    audio: 'var(--color-secondary)',
    document: 'var(--text-color-secondary)',
    archive: 'var(--color-warning)'
  }
  return colorMap[type] || 'var(--text-color-secondary)'
}

const formatFileSize = (bytes: number) => {
  const sizes = ['B', 'KB', 'MB', 'GB']
  if (bytes === 0) return '0 B'
  const i = Math.floor(Math.log(bytes) / Math.log(1024))
  return Math.round(bytes / Math.pow(1024, i) * 100) / 100 + ' ' + sizes[i]
}

const formatDate = (date: Date) => {
  return date.toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: 'short',
    day: 'numeric'
  })
}
</script>

<style scoped>
.page-container {
  padding: var(--spacing-lg);
  max-width: 1200px;
  margin: 0 auto;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-lg);
}

.page-title {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  font-size: var(--font-size-2xl);
  font-weight: var(--font-weight-semibold);
  color: var(--text-color);
  margin: 0;
}

.title-icon {
  color: var(--color-primary);
}

.page-actions {
  display: flex;
  gap: var(--spacing-sm);
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--spacing-md);
  background: var(--card-bg);
  border-radius: var(--radius-md);
  margin-bottom: var(--spacing-lg);
  border: 1px solid var(--border-color-light);
}

.toolbar-left {
  display: flex;
  gap: var(--spacing-md);
}

.toolbar-right {
  display: flex;
  gap: var(--spacing-sm);
  align-items: center;
}

.content-area {
  min-height: 400px;
}

/* 网格视图 */
.file-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: var(--spacing-md);
}

.file-card {
  background: var(--card-bg);
  border: 1px solid var(--border-color-light);
  border-radius: var(--radius-md);
  padding: var(--spacing-md);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.file-card:hover {
  border-color: var(--color-primary);
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.file-preview {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 80px;
  background: var(--bg-color-secondary);
  border-radius: var(--radius-sm);
  margin-bottom: var(--spacing-sm);
}

.file-info {
  text-align: center;
}

.file-name {
  font-weight: var(--font-weight-medium);
  color: var(--text-color);
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
  border: 1px solid var(--border-color-light);
  overflow: hidden;
}

/* 空状态 */
.empty-state {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 300px;
}

/* 文件详情 */
.file-details {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
}

.detail-item {
  padding: var(--spacing-sm) 0;
  border-bottom: 1px solid var(--border-color-light);
}

.detail-item:last-child {
  border-bottom: none;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .page-container {
    padding: var(--spacing-md);
  }

  .page-header {
    flex-direction: column;
    align-items: flex-start;
    gap: var(--spacing-md);
  }

  .toolbar {
    flex-direction: column;
    gap: var(--spacing-md);
  }

  .toolbar-left,
  .toolbar-right {
    width: 100%;
    justify-content: space-between;
  }

  .file-grid {
    grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
  }
}
</style>
