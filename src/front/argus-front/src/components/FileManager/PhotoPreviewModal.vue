<template>
  <Teleport to="body">
    <!-- 半透明背景 -->
    <div v-if="visible" class="photo-preview-backdrop" @click.self="handleClose">
      <!-- 悬浮预览窗口（支持全屏） -->
      <div class="photo-preview-window" :class="{ fullscreen: isFullscreen }" @click.stop>
        <!-- 信息面板 (Tab 键切换) - HoneyView 风格 -->
        <div v-if="showInfo" class="info-panel">
          <div class="info-item">文件名: {{ fileName }}</div>
          <div class="info-item" v-if="fileSize">大小: {{ formatFileSize(fileSize) }}</div>
          <div class="info-item">缩放: {{ Math.round(scale * 100) }}%</div>
          <div class="info-item">质量: {{ qualityText }}</div>

          <!-- EXIF 信息（如果有） -->
          <template v-if="exifData">
            <div class="info-divider"></div>
            <div class="info-item" v-if="exifData.make">相机: {{ exifData.make }} {{ exifData.model }}</div>
            <div class="info-item" v-if="exifData.lensModel">镜头: {{ exifData.lensModel }}</div>
            <div class="info-item" v-if="exifData.iso">ISO: {{ exifData.iso }}</div>
            <div class="info-item" v-if="exifData.fNumber">光圈: f/{{ exifData.fNumber }}</div>
            <div class="info-item" v-if="exifData.exposureTime">快门: {{ exifData.exposureTime }}s</div>
            <div class="info-item" v-if="exifData.focalLength">焦距: {{ exifData.focalLength }}mm</div>
            <div class="info-item" v-if="exifData.dateTimeOriginal">拍摄时间: {{ formatDateTime(exifData.dateTimeOriginal) }}</div>
            <div class="info-item" v-if="exifData.imageWidth && exifData.imageHeight">
              尺寸: {{ exifData.imageWidth }} × {{ exifData.imageHeight }}
            </div>
          </template>
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
            <!-- 多层渐进式图片加载 -->
            <!-- 缩略图层（720p，最先加载） -->
            <img
              v-if="imageUrls.thumbnail"
              :src="imageUrls.thumbnail"
              :alt="fileName"
              class="preview-image preview-layer"
              :class="{ 'layer-visible': currentQuality === 'thumbnail' }"
              @load="onLayerLoad('thumbnail')"
              @error="handleImageError"
              draggable="false"
            />

            <!-- 高清层（2K） -->
            <img
              v-if="imageUrls.highRes && (currentQuality === 'highRes' || currentQuality === 'original')"
              :src="imageUrls.highRes"
              :alt="fileName"
              class="preview-image preview-layer"
              :class="{ 'layer-visible': currentQuality === 'highRes' && layersLoaded.highRes }"
              @load="onLayerLoad('highRes')"
              @error="handleImageError"
              draggable="false"
            />

            <!-- 原图层 -->
            <img
              v-if="imageUrls.original && currentQuality === 'original'"
              :src="imageUrls.original"
              :alt="fileName"
              class="preview-image preview-layer"
              :class="{ 'layer-visible': currentQuality === 'original' && layersLoaded.original }"
              @load="onLayerLoad('original')"
              @error="handleImageError"
              draggable="false"
            />

            <!-- 错误提示 -->
            <div v-if="error" class="error-message">
              <p>❌ 无法加载图片</p>
            </div>
          </div>
        </div>

        <!-- 导航提示 (鼠标悬停时显示，全屏模式下不显示) -->
        <div v-if="!isFullscreen" class="navigation-hint">
          <span>← → 切换</span>
          <span>Space 关闭</span>
          <span>Tab 信息</span>
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
  (e: 'navigate', key: 'ArrowLeft' | 'ArrowRight' | 'ArrowUp' | 'ArrowDown'): void  // 导航键传递
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

// 状态
const error = ref('')
const currentQuality = ref<'thumbnail' | 'highRes' | 'original'>('thumbnail')

// 多层图片 URL
const imageUrls = ref({
  thumbnail: '',
  highRes: '',
  original: ''
})

// 各层加载状态
const layersLoaded = ref({
  thumbnail: false,
  highRes: false,
  original: false
})

// EXIF 数据
interface ExifData {
  make?: string
  model?: string
  lensModel?: string
  iso?: number
  fNumber?: number
  exposureTime?: string
  focalLength?: number
  dateTimeOriginal?: string
  imageWidth?: number
  imageHeight?: number
}
const exifData = ref<ExifData | null>(null)

// UI 状态
const showInfo = ref(false)  // 是否显示信息面板
const isFullscreen = ref(false)  // 是否全屏模式

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

const qualityText = computed(() => {
  switch (currentQuality.value) {
    case 'thumbnail': return '缩略图 (720p)'
    case 'highRes': return '高清 (2K)'
    case 'original': return '原图'
    default: return '未知'
  }
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
  console.log('🖼️ [PhotoPreviewModal.loadPreview] 开始加载预览:', {
    filePath: props.filePath,
    fileName: props.fileName,
    fileSize: props.fileSize
  })

  error.value = ''
  currentQuality.value = 'thumbnail'

  // 重置加载状态
  layersLoaded.value = {
    thumbnail: false,
    highRes: false,
    original: false
  }

  try {
    // 传递 fileSize 以启用智能预览策略
    const urls = photoPreviewService.getAllPreviewUrls(props.filePath, props.fileSize)

    console.log('🔗 [PhotoPreviewModal.loadPreview] 生成的 URLs:', {
      thumbnail: urls.thumbnail,
      highRes: urls.highRes,
      original: urls.original
    })

    imageUrls.value = {
      thumbnail: urls.thumbnail,
      highRes: urls.highRes,
      original: urls.original
    }

    // TODO: 获取 EXIF 信息
    // loadExifData(props.filePath)
  } catch (err) {
    error.value = err instanceof Error ? err.message : '未知错误'
    console.error('❌ [PhotoPreviewModal.loadPreview] 加载失败:', err)
  }
}

// 各层图片加载完成
function onLayerLoad(layer: 'thumbnail' | 'highRes' | 'original') {
  layersLoaded.value[layer] = true
  isAnimating.value = false
  console.log(`✅ [PhotoPreviewModal.onLayerLoad] ${layer} 层加载成功`)
}

// 图片加载失败
function handleImageError(event: Event) {
  error.value = '无法加载图片，请检查文件格式或路径'
  const img = event.target as HTMLImageElement
  console.error('❌ [PhotoPreviewModal.handleImageError] 图片加载失败:', {
    src: img.src,
    error: error.value
  })
}

// 切换图片质量（多层渲染，只需改变 currentQuality）
function switchQuality(quality: 'thumbnail' | 'highRes' | 'original') {
  if (currentQuality.value === quality) return
  currentQuality.value = quality
}

// 缩放控制
function zoomIn() {
  isAnimating.value = true
  const newScale = Math.min(scale.value + 0.2, 5)

  // 如果放大超过 1.5 倍，自动切换到原图
  if (newScale > 1.5 && currentQuality.value !== 'original') {
    switchQuality('original')
  }

  scale.value = newScale
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

  // 重置到缩略图模式
  if (currentQuality.value !== 'thumbnail') {
    switchQuality('thumbnail')
  }
}

// 鼠标滚轮缩放
function handleWheel(event: WheelEvent) {
  event.preventDefault()
  event.stopPropagation()

  const delta = event.deltaY > 0 ? -0.1 : 0.1
  const newScale = Math.max(0.1, Math.min(5, scale.value + delta))

  // 放大时自动进入全屏 + 切换原图
  if (newScale > 1.2 && !isFullscreen.value) {
    isFullscreen.value = true
  }

  // 如果放大超过 1.5 倍，自动切换到原图
  if (newScale > 1.5 && currentQuality.value !== 'original') {
    switchQuality('original')
  }

  // 缩小到接近 1.0 时退出全屏
  if (newScale <= 1.1 && isFullscreen.value) {
    isFullscreen.value = false
  }

  scale.value = newScale
}

// 双击切换全屏
function handleDoubleClick() {
  isFullscreen.value = !isFullscreen.value

  // 全屏时自动切换到原图
  if (isFullscreen.value && currentQuality.value !== 'original') {
    switchQuality('original')
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
  error.value = ''
  scale.value = 1
  translateX.value = 0
  translateY.value = 0
  isDragging.value = false
  isAnimating.value = false
  currentQuality.value = 'thumbnail'
  showInfo.value = false
  isFullscreen.value = false
  exifData.value = null
  imageUrls.value = { thumbnail: '', highRes: '', original: '' }
  layersLoaded.value = { thumbnail: false, highRes: false, original: false }
}

// 格式化文件大小
function formatFileSize(bytes: number): string {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  if (bytes < 1024 * 1024 * 1024) return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
  return (bytes / (1024 * 1024 * 1024)).toFixed(1) + ' GB'
}

// 格式化日期时间
function formatDateTime(dateStr: string): string {
  if (!dateStr) return ''
  // EXIF 时间格式通常是 "2024:10:06 15:30:45"
  const cleanDate = dateStr.replace(/^(\d{4}):(\d{2}):(\d{2})/, '$1-$2-$3')
  const date = new Date(cleanDate)
  if (isNaN(date.getTime())) return dateStr

  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')

  return `${year}-${month}-${day} ${hours}:${minutes}`
}

// 键盘事件处理
function handleKeyDown(event: KeyboardEvent) {
  if (!props.visible) return

  switch (event.key) {
    case ' ':  // 空格键：关闭预览
      event.preventDefault()
      event.stopPropagation()
      handleClose()
      break
    case 'Escape':  // ESC：关闭预览或退出全屏
      event.preventDefault()
      event.stopPropagation()
      if (isFullscreen.value) {
        isFullscreen.value = false
      } else {
        handleClose()
      }
      break
    case 'Tab':  // Tab：切换信息面板
      event.preventDefault()
      event.stopPropagation()
      showInfo.value = !showInfo.value
      break
    case 'f':  // F：切换全屏
    case 'F':
      event.preventDefault()
      event.stopPropagation()
      isFullscreen.value = !isFullscreen.value
      if (isFullscreen.value && currentQuality.value !== 'original') {
        switchQuality('original')
      }
      break
    case 'ArrowLeft':  // 左键：导航
    case 'ArrowRight':  // 右键：导航
    case 'ArrowUp':  // 上键：导航
    case 'ArrowDown':  // 下键：导航
      event.preventDefault()
      event.stopPropagation()
      resetZoomForNavigation()
      emit('navigate', event.key as 'ArrowLeft' | 'ArrowRight' | 'ArrowUp' | 'ArrowDown')
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

// 切换文件时重置缩放（快速切换体验）
function resetZoomForNavigation() {
  scale.value = 1
  translateX.value = 0
  translateY.value = 0
  isAnimating.value = false
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
  position: relative;
  width: 85vw;
  max-width: 1400px;
  height: 85vh;
  max-height: 900px;
  display: flex;
  flex-direction: column;
  animation: scaleIn 0.25s cubic-bezier(0.34, 1.56, 0.64, 1);
  transition: all 0.3s ease;
}

/* 全屏模式 */
.photo-preview-window.fullscreen {
  width: 100vw;
  height: 100vh;
  max-width: none;
  max-height: none;
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

/* 信息面板 (Tab 键切换) - HoneyView 风格 */
.info-panel {
  position: absolute;
  top: 20px;
  left: 20px;
  z-index: 10;
  padding: 12px 16px;
  background: rgba(0, 0, 0, 0.5);
  border-radius: 6px;
  backdrop-filter: blur(10px);
  -webkit-backdrop-filter: blur(10px);
  animation: slideIn 0.2s ease-out;
  font-size: 13px;
  line-height: 1.6;
  color: rgba(255, 255, 255, 0.95);
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.3);
}

@keyframes slideIn {
  from {
    opacity: 0;
    transform: translateY(-10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* 信息项 - 简单文本列表 */
.info-item {
  margin: 4px 0;
  white-space: nowrap;
}

/* 分隔线 */
.info-divider {
  height: 1px;
  background: rgba(255, 255, 255, 0.2);
  margin: 8px 0;
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
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  transform-origin: center;
  will-change: transform;
}

/* 多层图片渐进式加载 */
.preview-layer {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  max-width: 90%;
  max-height: 90%;
  width: auto;
  height: auto;
  object-fit: contain;
  user-select: none;
  -webkit-user-drag: none;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.4);
  border-radius: 4px;
  opacity: 0;
  transition: opacity 0.4s ease-in-out;
  pointer-events: none;
}

/* 当前显示的图层 */
.preview-layer.layer-visible {
  opacity: 1;
  pointer-events: auto;
}

/* 兼容旧的 class 名（如果有地方还在用） */
.preview-image {
  max-width: 100%;
  max-height: 100%;
  width: auto;
  height: auto;
  object-fit: contain;
  user-select: none;
  -webkit-user-drag: none;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.4);
  border-radius: 4px;
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

/* 导航提示（悬停时显示） */
.navigation-hint {
  position: absolute;
  bottom: 20px;
  left: 50%;
  transform: translateX(-50%);
  display: flex;
  gap: 24px;
  padding: 10px 20px;
  background: rgba(0, 0, 0, 0.7);
  border-radius: 20px;
  backdrop-filter: blur(10px);
  opacity: 0;
  transition: opacity 0.3s ease;
  pointer-events: none;
  z-index: 10;
}

.photo-preview-window:hover .navigation-hint {
  opacity: 1;
}

.navigation-hint span {
  color: rgba(255, 255, 255, 0.9);
  font-size: 12px;
  white-space: nowrap;
}
</style>
