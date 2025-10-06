<template>
  <Teleport to="body">
    <!-- 半透明背景 -->
    <div v-if="visible" class="photo-preview-backdrop" @click.self="handleClose">
      <!-- 悬浮预览窗口 -->
      <div class="photo-preview-window" @click.stop>
        <!-- 窗口标题栏 -->
        <div class="window-header">
          <div class="window-title">
            <span class="file-name">{{ fileName }}</span>
            <span v-if="fileSize" class="file-size">{{ formatFileSize(fileSize) }}</span>
          </div>
          <button class="close-btn" @click="handleClose" title="关闭 (ESC)">
            ✕
          </button>
        </div>

        <!-- 图片容器 -->
        <div class="image-container" ref="containerRef">
        <div
          class="image-wrapper"
          :style="{
            transform: `scale(${scale}) translate(${translateX}px, ${translateY}px)`,
            transition: isAnimating ? 'transform 0.3s ease-out' : 'none',
            cursor: isDragging ? 'grabbing' : (scale > 1 ? 'grab' : 'default')
          }"
          @mousedown="handleMouseDown"
          @mousemove="handleMouseMove"
          @mouseup="handleMouseUp"
          @mouseleave="handleMouseUp"
          @wheel="handleWheel"
          @dblclick="handleDoubleClick"
        >
          <!-- 加载状态 -->
          <div v-if="loading" class="loading-indicator">
            <div class="spinner"></div>
            <p>{{ loadingText }}</p>
          </div>

          <!-- 图片显示 -->
          <img
            v-show="!loading && !error"
            :src="currentImageUrl"
            :alt="fileName"
            class="preview-image"
            @load="handleImageLoad"
            @error="handleImageError"
            draggable="false"
          />

          <!-- 错误提示 -->
          <div v-if="error" class="error-message">
            <p>❌ 图片加载失败</p>
            <p class="error-detail">{{ error }}</p>
          </div>
        </div>
      </div>

        <!-- 窗口底部工具栏 -->
        <div class="window-footer">
          <!-- 缩放控制 -->
          <div class="zoom-controls">
            <button @click="zoomOut" :disabled="scale <= 0.1" title="缩小 (-)">−</button>
            <span class="zoom-level">{{ Math.round(scale * 100) }}%</span>
            <button @click="zoomIn" :disabled="scale >= 5" title="放大 (+)">+</button>
            <button @click="resetZoom" title="重置 (0)">1:1</button>
          </div>

          <!-- 图片质量切换 -->
          <div class="quality-controls">
            <button
              :class="['quality-btn', { active: currentQuality === 'thumbnail' }]"
              @click="switchQuality('thumbnail')"
              title="缩略图模式（快速）"
            >
              缩略图
            </button>
            <button
              :class="['quality-btn', { active: currentQuality === 'highRes' }]"
              @click="switchQuality('highRes')"
              title="高清模式"
            >
              高清
            </button>
            <button
              :class="['quality-btn', { active: currentQuality === 'original' }]"
              @click="switchQuality('original')"
              title="原图模式（可能较大）"
            >
              原图
            </button>
          </div>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { photoPreviewService } from '@/services/photoPreviewService'

interface Props {
  visible: boolean
  filePath: string
  fileName?: string
  fileSize?: number
}

interface Emits {
  (e: 'update:visible', value: boolean): void
  (e: 'close'): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

// 状态
const loading = ref(true)
const loadingText = ref('加载中...')
const error = ref('')
const currentQuality = ref<'thumbnail' | 'highRes' | 'original'>('thumbnail')
const currentImageUrl = ref('')

// 缩放和平移
const scale = ref(1)
const translateX = ref(0)
const translateY = ref(0)
const isAnimating = ref(false)

// 拖拽
const isDragging = ref(false)
const dragStartX = ref(0)
const dragStartY = ref(0)
const dragStartTranslateX = ref(0)
const dragStartTranslateY = ref(0)

// 引用
const containerRef = ref<HTMLElement | null>(null)

// 计算属性
const fileName = computed(() => {
  if (props.fileName) return props.fileName
  const parts = props.filePath.split(/[/\\]/)
  return parts[parts.length - 1] || '未知文件'
})

// 监听 visible 变化
watch(() => props.visible, (newVal) => {
  if (newVal) {
    loadPreview()
  } else {
    resetState()
  }
})

// 监听文件路径变化
watch(() => props.filePath, () => {
  if (props.visible) {
    loadPreview()
  }
})

// 加载预览
async function loadPreview() {
  loading.value = true
  error.value = ''
  currentQuality.value = 'thumbnail'

  try {
    const urls = photoPreviewService.getAllPreviewUrls(props.filePath)
    currentImageUrl.value = urls.thumbnail
    loadingText.value = '加载缩略图...'
  } catch (err) {
    error.value = err instanceof Error ? err.message : '未知错误'
    loading.value = false
  }
}

// 图片加载完成
function handleImageLoad() {
  loading.value = false
  isAnimating.value = false
}

// 图片加载失败
function handleImageError() {
  loading.value = false
  error.value = '无法加载图片，请检查文件格式或路径'
}

// 切换图片质量
async function switchQuality(quality: 'thumbnail' | 'highRes' | 'original') {
  if (currentQuality.value === quality) return

  loading.value = true
  currentQuality.value = quality

  const urls = photoPreviewService.getAllPreviewUrls(props.filePath)

  switch (quality) {
    case 'thumbnail':
      currentImageUrl.value = urls.thumbnail
      loadingText.value = '加载缩略图...'
      break
    case 'highRes':
      currentImageUrl.value = urls.highRes
      loadingText.value = '加载高清图片...'
      break
    case 'original':
      currentImageUrl.value = urls.original
      loadingText.value = '加载原图（可能较大）...'
      break
  }
}

// 缩放控制
function zoomIn() {
  isAnimating.value = true
  scale.value = Math.min(scale.value + 0.2, 5)
}

function zoomOut() {
  isAnimating.value = true
  scale.value = Math.max(scale.value - 0.2, 0.1)
}

function resetZoom() {
  isAnimating.value = true
  scale.value = 1
  translateX.value = 0
  translateY.value = 0
}

// 鼠标滚轮缩放
function handleWheel(event: WheelEvent) {
  event.preventDefault()
  event.stopPropagation()

  const delta = event.deltaY > 0 ? -0.1 : 0.1
  scale.value = Math.max(0.1, Math.min(5, scale.value + delta))
}

// 双击放大
function handleDoubleClick() {
  if (scale.value === 1) {
    isAnimating.value = true
    scale.value = 2

    // 自动切换到高清模式
    if (currentQuality.value === 'thumbnail') {
      switchQuality('highRes')
    }
  } else {
    resetZoom()
  }
}

// 拖拽平移
function handleMouseDown(event: MouseEvent) {
  if (scale.value <= 1) return // 只有放大时才能拖拽

  event.preventDefault()
  isDragging.value = true
  dragStartX.value = event.clientX
  dragStartY.value = event.clientY
  dragStartTranslateX.value = translateX.value
  dragStartTranslateY.value = translateY.value
}

function handleMouseMove(event: MouseEvent) {
  if (!isDragging.value) return

  const deltaX = event.clientX - dragStartX.value
  const deltaY = event.clientY - dragStartY.value

  translateX.value = dragStartTranslateX.value + deltaX / scale.value
  translateY.value = dragStartTranslateY.value + deltaY / scale.value
}

function handleMouseUp() {
  isDragging.value = false
}

// 关闭预览
function handleClose() {
  emit('update:visible', false)
  emit('close')
}

// 重置状态
function resetState() {
  loading.value = true
  error.value = ''
  scale.value = 1
  translateX.value = 0
  translateY.value = 0
  isDragging.value = false
  isAnimating.value = false
  currentQuality.value = 'thumbnail'
}

// 格式化文件大小
function formatFileSize(bytes: number): string {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  if (bytes < 1024 * 1024 * 1024) return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
  return (bytes / (1024 * 1024 * 1024)).toFixed(1) + ' GB'
}

// 键盘事件处理
function handleKeyDown(event: KeyboardEvent) {
  if (!props.visible) return

  switch (event.key) {
    case 'Escape':
      event.preventDefault()
      event.stopPropagation()
      handleClose()
      break
    case '+':
    case '=':
      event.preventDefault()
      event.stopPropagation()
      zoomIn()
      break
    case '-':
    case '_':
      event.preventDefault()
      event.stopPropagation()
      zoomOut()
      break
    case '0':
      event.preventDefault()
      event.stopPropagation()
      resetZoom()
      break
  }
}

// 生命周期
onMounted(() => {
  document.addEventListener('keydown', handleKeyDown, true)
})

onUnmounted(() => {
  document.removeEventListener('keydown', handleKeyDown, true)
})
</script>

<style scoped>
/* macOS Quick Look 风格 */
.photo-preview-backdrop {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 9999;
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  display: flex;
  align-items: center;
  justify-content: center;
  animation: fadeIn 0.2s ease-out;
}

@keyframes fadeIn {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

.photo-preview-window {
  width: 85vw;
  max-width: 1400px;
  height: 85vh;
  max-height: 900px;
  background: #1e1e1e;
  border-radius: 12px;
  box-shadow: 0 25px 50px rgba(0, 0, 0, 0.5);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  animation: scaleIn 0.25s cubic-bezier(0.34, 1.56, 0.64, 1);
}

@keyframes scaleIn {
  from {
    transform: scale(0.9);
    opacity: 0;
  }
  to {
    transform: scale(1);
    opacity: 1;
  }
}

/* 窗口标题栏 */
.window-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  background: rgba(0, 0, 0, 0.3);
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.window-title {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.window-title .file-name {
  color: white;
  font-size: 14px;
  font-weight: 500;
}

.window-title .file-size {
  color: rgba(255, 255, 255, 0.6);
  font-size: 12px;
}

.close-btn {
  width: 32px;
  height: 32px;
  border: none;
  border-radius: 6px;
  background: rgba(255, 255, 255, 0.1);
  color: white;
  font-size: 18px;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
}

.close-btn:hover {
  background: rgba(255, 59, 48, 0.8);
  transform: scale(1.05);
}

.image-container {
  flex: 1;
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  position: relative;
}

.image-wrapper {
  display: flex;
  align-items: center;
  justify-content: center;
  transform-origin: center;
  will-change: transform;
}

.preview-image {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
  user-select: none;
  -webkit-user-drag: none;
}

.loading-indicator {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
  color: white;
}

.spinner {
  width: 50px;
  height: 50px;
  border: 4px solid rgba(255, 255, 255, 0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.error-message {
  color: white;
  text-align: center;
}

.error-detail {
  margin-top: 8px;
  font-size: 14px;
  color: #ff6b6b;
}

/* 窗口底部工具栏 */
.window-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  background: rgba(0, 0, 0, 0.3);
  border-top: 1px solid rgba(255, 255, 255, 0.1);
}

.zoom-controls {
  display: flex;
  align-items: center;
  gap: 8px;
}

.zoom-controls button {
  width: 32px;
  height: 32px;
  border: none;
  border-radius: 6px;
  background: rgba(255, 255, 255, 0.1);
  color: white;
  font-size: 18px;
  cursor: pointer;
  transition: all 0.2s;
}

.zoom-controls button:hover:not(:disabled) {
  background: rgba(255, 255, 255, 0.2);
}

.zoom-controls button:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}

.zoom-level {
  color: white;
  font-size: 14px;
  min-width: 50px;
  text-align: center;
}

.quality-controls {
  display: flex;
  gap: 8px;
}

.quality-btn {
  padding: 6px 12px;
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 6px;
  background: rgba(255, 255, 255, 0.1);
  color: white;
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s;
}

.quality-btn:hover {
  background: rgba(255, 255, 255, 0.2);
}

.quality-btn.active {
  background: rgba(59, 130, 246, 0.6);
  border-color: rgba(59, 130, 246, 1);
  font-weight: 500;
}
</style>
