<template>
  <Teleport to="body">
    <Transition name="preview-fade">
      <div v-if="visible" class="quick-preview-overlay" @click="handleClose" @keydown.space.prevent>
        <div class="preview-container" @click.stop>
          <!-- Header -->
          <div class="preview-header">
            <div class="preview-title">
              <span class="file-icon">{{ getFileIcon(item?.type) }}</span>
              <span class="file-name">{{ item?.name || '预览' }}</span>
            </div>
            <button class="btn-close" @click="handleClose" title="关闭 (ESC)">✕</button>
          </div>

          <!-- Content -->
          <div class="preview-content">
            <div v-if="item?.type === 'photo'" class="preview-photo">
              <div class="placeholder-image">📷 照片预览</div>
              <p class="placeholder-text">这里将显示照片内容</p>
            </div>

            <div v-else-if="item?.type === 'video'" class="preview-video">
              <div class="placeholder-video">🎬 视频预览</div>
              <p class="placeholder-text">这里将显示视频播放器</p>
            </div>

            <div v-else-if="item?.type === 'folder'" class="preview-folder">
              <div class="folder-icon-large">📁</div>
              <p class="folder-name">{{ item?.name }}</p>
              <div class="folder-info">
                <div class="info-item">
                  <span class="label">包含:</span>
                  <span class="value">{{ mockFolderStats.files }} 个文件</span>
                </div>
                <div class="info-item">
                  <span class="label">大小:</span>
                  <span class="value">{{ mockFolderStats.size }}</span>
                </div>
                <div class="info-item">
                  <span class="label">修改时间:</span>
                  <span class="value">{{ item?.date || '2024-01-01' }}</span>
                </div>
              </div>
            </div>

            <div v-else class="preview-file">
              <div class="placeholder-file">📄 文件预览</div>
              <p class="placeholder-text">暂不支持此类型文件预览</p>
            </div>
          </div>

          <!-- Footer -->
          <div class="preview-footer">
            <div class="file-info">
              <span class="info-label">大小:</span>
              <span class="info-value">{{ item?.size || '-' }}</span>
              <span class="info-separator">|</span>
              <span class="info-label">修改日期:</span>
              <span class="info-value">{{ item?.date || '-' }}</span>
            </div>
            <div class="keyboard-hint">
              <span>按 <kbd>空格</kbd> 或 <kbd>ESC</kbd> 关闭</span>
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, onUnmounted } from 'vue'
import type { FileItem } from './types'

interface Props {
  visible: boolean
  item: FileItem | null
}

const props = defineProps<Props>()
const emit = defineEmits<{
  close: []
}>()

// 模拟文件夹统计数据
const mockFolderStats = ref({
  files: 42,
  size: '2.3 GB'
})

function getFileIcon(type?: string) {
  switch (type) {
    case 'folder':
      return '📁'
    case 'photo':
      return '🖼️'
    case 'video':
      return '🎬'
    default:
      return '📄'
  }
}

function handleClose() {
  emit('close')
}

// 键盘事件处理
function handleKeyDown(event: KeyboardEvent) {
  if (!props.visible) return

  // 只处理 ESC 键，空格键由父组件处理
  if (event.key === 'Escape') {
    event.preventDefault()
    handleClose()
  }
}

onMounted(() => {
  document.addEventListener('keydown', handleKeyDown)
})

onUnmounted(() => {
  document.removeEventListener('keydown', handleKeyDown)
})

// 当预览打开时，阻止body滚动
watch(() => props.visible, (visible) => {
  if (visible) {
    document.body.style.overflow = 'hidden'
  } else {
    document.body.style.overflow = ''
  }
})
</script>

<style scoped>
.quick-preview-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.75);
  backdrop-filter: blur(8px);
  z-index: 2000;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 40px;
}

.preview-container {
  background: white;
  border-radius: 12px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
  max-width: 900px;
  max-height: 90vh;
  width: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

/* Header */
.preview-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  border-bottom: 1px solid #e5e7eb;
  background: #f9fafb;
}

.preview-title {
  display: flex;
  align-items: center;
  gap: 12px;
  flex: 1;
  min-width: 0;
}

.file-icon {
  font-size: 24px;
  flex-shrink: 0;
}

.file-name {
  font-size: 16px;
  font-weight: 600;
  color: #1f2937;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.btn-close {
  width: 32px;
  height: 32px;
  border: none;
  background: transparent;
  border-radius: 6px;
  font-size: 20px;
  color: #6b7280;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.btn-close:hover {
  background: #f3f4f6;
  color: #1f2937;
}

/* Content */
.preview-content {
  flex: 1;
  overflow: auto;
  padding: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.preview-photo,
.preview-video,
.preview-folder,
.preview-file {
  text-align: center;
  width: 100%;
}

.placeholder-image,
.placeholder-video,
.placeholder-file {
  font-size: 80px;
  margin-bottom: 16px;
  opacity: 0.6;
}

.placeholder-text {
  color: #9ca3af;
  font-size: 14px;
}

/* Folder preview */
.folder-icon-large {
  font-size: 120px;
  margin-bottom: 20px;
}

.folder-name {
  font-size: 24px;
  font-weight: 600;
  color: #1f2937;
  margin-bottom: 24px;
}

.folder-info {
  display: flex;
  flex-direction: column;
  gap: 12px;
  background: #f9fafb;
  border-radius: 8px;
  padding: 20px;
  max-width: 400px;
  margin: 0 auto;
}

.info-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 14px;
}

.info-item .label {
  color: #6b7280;
  font-weight: 500;
}

.info-item .value {
  color: #1f2937;
  font-weight: 600;
}

/* Footer */
.preview-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 20px;
  border-top: 1px solid #e5e7eb;
  background: #f9fafb;
}

.file-info {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
}

.info-label {
  color: #6b7280;
}

.info-value {
  color: #1f2937;
  font-weight: 500;
}

.info-separator {
  color: #d1d5db;
}

.keyboard-hint {
  font-size: 12px;
  color: #6b7280;
}

kbd {
  padding: 2px 6px;
  background: white;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-family: monospace;
  font-size: 11px;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
}

/* Transitions */
.preview-fade-enter-active,
.preview-fade-leave-active {
  transition: opacity 0.2s ease;
}

.preview-fade-enter-active .preview-container,
.preview-fade-leave-active .preview-container {
  transition: transform 0.2s ease, opacity 0.2s ease;
}

.preview-fade-enter-from,
.preview-fade-leave-to {
  opacity: 0;
}

.preview-fade-enter-from .preview-container,
.preview-fade-leave-to .preview-container {
  transform: scale(0.95);
  opacity: 0;
}
</style>
