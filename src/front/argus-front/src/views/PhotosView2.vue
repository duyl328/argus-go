<template>
  <div class="photos-timeline">
    <!-- 加载状态 -->
    <div v-if="loading" class="loading-state">
      <div class="loading-spinner"></div>
      <div class="loading-text">正在加载照片时间线...</div>
    </div>

    <!-- 空状态 -->
    <div v-else-if="timelineItems.length === 0" class="empty-state">
      <div class="empty-icon">📷</div>
      <div class="empty-text">暂无照片</div>
      <div class="empty-subtitle">开始上传一些照片来创建你的时间线吧</div>
    </div>

    <!-- 时间线内容 -->
    <DynamicScroller
      v-else
      ref="scrollerRef"
      :items="timelineItems"
      :min-item-size="100"
      key-field="id"
      class="scroller"
      :buffer="200"
    >
      <template #default="{ item, index, active }">
        <DynamicScrollerItem
          :item="item"
          :index="index"
          :active="active"
          :size-dependencies="[containerWidth]"
          :key="item.id"
        >
          <!-- 月份标题 -->
          <div v-if="item.type === 'month'" class="month-header">
            <h2 class="month-title">{{ item.title }}</h2>
            <div class="month-subtitle">{{ item.subtitle }}</div>
          </div>

          <!-- 回忆卡片 -->
          <div v-else-if="item.type === 'memory'" class="memory-card">
            <div class="memory-icon">📸</div>
            <div class="memory-content">
              <div class="memory-title">{{ item.title }}</div>
              <div class="memory-subtitle">{{ item.subtitle }}</div>
            </div>
          </div>

          <!-- 时间卡片 -->
          <div v-else-if="item.type === 'date'" class="date-card">
            <div class="date-label">{{ item.title }}</div>
          </div>

          <!-- 照片网格 -->
          <div v-else-if="item.type === 'photos'" class="photos-grid" :style="{ height: item.height + 'px' }">
            <div
              v-for="photo in item.photos"
              :key="photo.id"
              class="photo-item"
              :style="{
                position: 'absolute',
                top: photo.box.top + 'px',
                left: photo.box.left + 'px',
                width: photo.box.width + 'px',
                height: photo.box.height + 'px'
              }"
              @click="handlePhotoClick(photo)"
            >
              <!-- 延迟加载的图片 -->
              <img
                v-if="photo.loaded"
                :src="photo.thumbnailUrl"
                :alt="`照片 ${photo.hash.slice(0, 8)}`"
                class="photo-image"
                @error="handleImageError"
                @load="handleImageLoad(photo)"
              />
              <!-- 占位符背景 -->
              <div
                v-else
                class="photo-placeholder"
                :style="{
                  backgroundColor: photo.color,
                  width: '100%',
                  height: '100%'
                }"
                :data-hash="photo.hash"
              ></div>
              <!-- 开发时显示比例信息 -->
              <div v-if="false" class="photo-debug-info">
                <span class="photo-ratio">{{ photo.ratio.toFixed(2) }}</span>
              </div>
            </div>
          </div>
        </DynamicScrollerItem>
      </template>
    </DynamicScroller>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, nextTick } from 'vue'
import { DynamicScroller, DynamicScrollerItem } from 'vue-virtual-scroller'
import justifiedLayout from 'justified-layout'
import 'vue-virtual-scroller/dist/vue-virtual-scroller.css'
import { getFullTimeline, getThumbnailUrl } from '@/services/timelineService'
import type { Photo as ApiPhoto, MonthGroup, DayGroup } from '@/types/api'

interface Photo extends ApiPhoto {
  id: string
  color?: string  // 占位颜色，用于加载时显示
  thumbnailUrl: string
  box: {
    top: number
    left: number
    width: number
    height: number
  }
}

interface TimelineItem {
  id: string
  type: 'month' | 'memory' | 'date' | 'photos'
  title?: string
  subtitle?: string
  height?: number
  photos?: Photo[]
  loading?: boolean
}

const scrollerRef = ref()
const containerWidth = ref(1200)
const loading = ref(true)
const error = ref('')

// 生成随机颜色作为占位符
function generateRandomColor(): string {
  const colors = [
    '#FF6B6B', '#4ECDC4', '#45B7D1', '#96CEB4', '#FFEAA7',
    '#DDA0DD', '#98D8C8', '#F7DC6F', '#BB8FCE', '#85C1E9',
    '#F8C471', '#82E0AA', '#F1948A', '#85C1E9', '#D7BDE2'
  ]
  return colors[Math.floor(Math.random() * colors.length)]
}

// 转换API照片数据为组件用的照片数据
function convertApiPhotosToPhotos(apiPhotos: ApiPhoto[]): Photo[] {
  return apiPhotos.map(apiPhoto => ({
    ...apiPhoto,
    id: apiPhoto.hash, // 使用hash作为id
    color: apiPhoto.placeholder || generateRandomColor(), // 使用预设占位颜色
    thumbnailUrl: getThumbnailUrl(apiPhoto.hash, '400'),
    box: { top: 0, left: 0, width: 0, height: 0 }, // 稍后计算
    // 预渲染优化字段
    loaded: apiPhoto.loaded || false,
    inViewport: apiPhoto.inViewport || false
  }))
}

// 使用 justified-layout 计算照片布局
function calculatePhotosLayout(photos: Photo[]): { photos: Photo[], height: number } {
  if (photos.length === 0) {
    return { photos: [], height: 0 }
  }

  const photoInputs = photos.map(photo => ({
    width: photo.width,
    height: photo.height
  }))

  const geometry = justifiedLayout(photoInputs, {
    containerWidth: containerWidth.value - 40, // 减去左右边距
    targetRowHeight: 220,
    targetRowHeightTolerance: 0.25,
    boxSpacing: 4,
    containerPadding: 0
  })

  const layoutPhotos = photos.map((photo, index) => ({
    ...photo,
    box: geometry.boxes[index] || { top: 0, left: 0, width: 0, height: 0 }
  }))

  return {
    photos: layoutPhotos,
    height: geometry.containerHeight
  }
}

// 转换API数据为时间线展示数据
function convertMonthGroupsToTimelineItems(monthGroups: MonthGroup[]): TimelineItem[] {
  const items: TimelineItem[] = []

  monthGroups.forEach((monthGroup, monthIndex) => {
    // 添加月份标题
    items.push({
      id: `month-${monthGroup.year}-${monthGroup.month}`,
      type: 'month',
      title: monthGroup.title,
      subtitle: monthGroup.subtitle
    })

    // 添加回忆卡片 (随机，约40%概率出现)
    if (Math.random() > 0.6) {
      items.push({
        id: `memory-${monthGroup.year}-${monthGroup.month}`,
        type: 'memory',
        title: '回忆',
        subtitle: '一年前的今天'
      })
    }

    // 添加每天的照片数据
    monthGroup.days.forEach((dayGroup, dayIndex) => {
      if (dayGroup.photos.length > 0) {
        const dayId = `day-${monthGroup.year}-${monthGroup.month}-${dayIndex}`

        // 添加日期卡片
        const date = new Date(dayGroup.date)
        const dateTitle = `${date.getMonth() + 1}月${date.getDate()}日`

        items.push({
          id: `date-${dayId}`,
          type: 'date',
          title: dateTitle
        })

        // 转换并布局照片
        const photos = convertApiPhotosToPhotos(dayGroup.photos)
        const layout = calculatePhotosLayout(photos)

        items.push({
          id: `photos-${dayId}`,
          type: 'photos',
          height: layout.height,
          photos: layout.photos
        })
      }
    })
  })

  return items
}

// 响应式数据
const timelineItems = ref<TimelineItem[]>([])

// 加载时间线数据
async function loadTimelineData() {
  try {
    loading.value = true
    error.value = ''

    // 获取最近2年的数据
    const endDate = new Date().toISOString().split('T')[0]
    const startDate = new Date(Date.now() - 2 * 365 * 24 * 60 * 60 * 1000).toISOString().split('T')[0]

    const monthGroups = await getFullTimeline({
      start_date: startDate,
      end_date: endDate
    })

    // 转换为时间线展示数据
    timelineItems.value = convertMonthGroupsToTimelineItems(monthGroups)
  } catch (err) {
    console.error('Failed to load timeline data:', err)
    error.value = err instanceof Error ? err.message : '加载时间线数据失败'

    // 降级到空状态，不显示错误给用户，在控制台记录即可
    timelineItems.value = []
  } finally {
    loading.value = false
  }
}

// 重新计算照片布局
function recalculatePhotosLayout() {
  timelineItems.value = timelineItems.value.map(item => {
    if (item.type === 'photos' && item.photos) {
      const layout = calculatePhotosLayout(item.photos)
      return {
        ...item,
        height: layout.height,
        photos: layout.photos
      }
    }
    return item
  })
}

// 视口观察器
let viewportObserver: IntersectionObserver | null = null

// 处理照片点击事件
function handlePhotoClick(photo: Photo) {
  // TODO: 实现照片详情查看或大图预览功能
  console.log('Photo clicked:', photo)
}

// 处理图片加载完成
function handleImageLoad(photo: Photo) {
  console.log(`图片加载完成: ${photo.hash.slice(0, 8)}`)
}

// 处理图片加载失败
function handleImageError(event: Event) {
  const img = event.target as HTMLImageElement
  const hash = img.alt?.match(/照片 (\w+)/)?.[1] || 'unknown'
  console.warn(`图片加载失败: ${hash}`)
  
  // 恢复占位符显示
  img.style.display = 'none'
  const parent = img.parentElement
  if (parent) {
    // 找到对应的照片对象并重置loaded状态
    timelineItems.value = timelineItems.value.map(item => {
      if (item.type === 'photos' && item.photos) {
        return {
          ...item,
          photos: item.photos.map(p => 
            p.hash.startsWith(hash) ? { ...p, loaded: false } : p
          )
        }
      }
      return item
    })
  }
}

// 初始化视口观察器
function initViewportObserver() {
  if (viewportObserver) {
    viewportObserver.disconnect()
  }

  viewportObserver = new IntersectionObserver(
    (entries) => {
      entries.forEach((entry) => {
        const hash = (entry.target as HTMLElement).dataset.hash
        if (hash && entry.isIntersecting) {
          // 照片进入视口，开始加载
          loadPhotoImage(hash)
        }
      })
    },
    {
      // 提前200px开始加载
      rootMargin: '200px',
      threshold: 0
    }
  )
}

// 加载具体照片图片
function loadPhotoImage(hash: string) {
  timelineItems.value = timelineItems.value.map(item => {
    if (item.type === 'photos' && item.photos) {
      return {
        ...item,
        photos: item.photos.map(photo => {
          if (photo.hash === hash && !photo.loaded) {
            console.log(`开始加载图片: ${hash.slice(0, 8)}`)
            return { ...photo, loaded: true, inViewport: true }
          }
          return photo
        })
      }
    }
    return item
  })
}

// 观察所有占位符元素
function observePhotos() {
  if (!viewportObserver) return

  // 观察当前所有的占位符元素
  nextTick(() => {
    const placeholders = document.querySelectorAll('.photo-placeholder[data-hash]')
    placeholders.forEach(placeholder => {
      viewportObserver?.observe(placeholder)
    })
    console.log(`开始观察 ${placeholders.length} 个照片占位符`)
  })
}

// 初始化数据
onMounted(async () => {
  await nextTick()

  // 获取容器宽度
  const container = document.querySelector('.photos-timeline')
  if (container) {
    containerWidth.value = container.clientWidth
  }

  // 初始化视口观察器
  initViewportObserver()
  
  // 加载时间线数据
  await loadTimelineData()

  // 数据加载完成后开始观察照片
  await nextTick()
  observePhotos()

  // 监听窗口大小变化
  const handleResize = () => {
    if (container) {
      const newWidth = container.clientWidth
      if (Math.abs(newWidth - containerWidth.value) > 50) { // 只有明显变化时才重新计算
        containerWidth.value = newWidth
        recalculatePhotosLayout()
        // 重新计算后重新观察
        nextTick(() => observePhotos())
      }
    }
  }

  window.addEventListener('resize', handleResize)

  // 清理事件监听器和观察器
  return () => {
    window.removeEventListener('resize', handleResize)
    if (viewportObserver) {
      viewportObserver.disconnect()
      viewportObserver = null
    }
  }
})
</script>

<style scoped>
.photos-timeline {
  height: 100vh;
  width: 100%;
  background: var(--bg-color);
}

.scroller {
  height: 100%;
}

/* 加载状态样式 */
.loading-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100vh;
  gap: 16px;
}

.loading-spinner {
  width: 40px;
  height: 40px;
  border: 3px solid var(--border-color);
  border-top: 3px solid var(--primary-color, #007bff);
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.loading-text {
  font-size: 16px;
  color: var(--text-color-secondary);
}

/* 空状态样式 */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100vh;
  gap: 12px;
}

.empty-icon {
  font-size: 64px;
  opacity: 0.5;
}

.empty-text {
  font-size: 20px;
  font-weight: 500;
  color: var(--text-color);
}

.empty-subtitle {
  font-size: 14px;
  color: var(--text-color-secondary);
  opacity: 0.7;
}

/* 月份标题样式 */
.month-header {
  padding: 40px 20px 20px 20px;
  background: var(--bg-color);
}

.month-title {
  font-size: 28px;
  font-weight: 600;
  color: var(--text-color);
  margin: 0 0 4px 0;
}

.month-subtitle {
  font-size: 14px;
  color: var(--text-color-secondary);
  opacity: 0.7;
}

/* 回忆卡片样式 */
.memory-card {
  margin: 16px 20px;
  padding: 16px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 12px;
  display: flex;
  align-items: center;
  gap: 12px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.memory-icon {
  font-size: 24px;
}

.memory-content {
  flex: 1;
}

.memory-title {
  font-size: 16px;
  font-weight: 600;
  color: white;
  margin-bottom: 4px;
}

.memory-subtitle {
  font-size: 14px;
  color: rgba(255, 255, 255, 0.8);
}

/* 日期卡片样式 */
.date-card {
  padding: 16px 20px 12px 20px;
  background: var(--bg-color);
}

.date-label {
  font-size: 16px;
  font-weight: 500;
  color: var(--text-color);
}

/* 照片网格样式 */
.photos-grid {
  position: relative;
  margin: 0 20px 24px 20px;
  background: var(--bg-color);
}

.photo-item {
  border-radius: 4px;
  overflow: hidden;
  cursor: pointer;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  position: relative;
}

.photo-item:hover {
  transform: scale(1.02);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
  z-index: 1;
}

.photo-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: opacity 0.3s ease;
}

.photo-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  transition: background-color 0.2s ease;
}

.photo-placeholder::after {
  content: '';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 24px;
  height: 24px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top: 2px solid rgba(255, 255, 255, 0.8);
  border-radius: 50%;
  animation: placeholder-spin 1s linear infinite;
  opacity: 0;
  transition: opacity 0.3s ease;
}

.photo-placeholder:hover::after {
  opacity: 1;
}

@keyframes placeholder-spin {
  0% { transform: translate(-50%, -50%) rotate(0deg); }
  100% { transform: translate(-50%, -50%) rotate(360deg); }
}

/* 图片加载失败显示 */
.photo-error {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  color: var(--text-color-secondary);
  gap: 4px;
}

.photo-error span:first-child {
  font-size: 20px;
  opacity: 0.5;
}

/* 开发调试信息 */
.photo-debug-info {
  position: absolute;
  top: 4px;
  right: 4px;
  background: rgba(0, 0, 0, 0.7);
  color: white;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 10px;
  pointer-events: none;
}

.photo-ratio {
  background: rgba(0, 0, 0, 0.3);
  padding: 2px 6px;
  border-radius: 4px;
}

/* 深色主题适配 */
.dark .month-header {
  background: var(--bg-color);
}

.dark .date-card {
  background: var(--bg-color);
}

.dark .photos-grid {
  background: var(--bg-color);
}

/* 响应式设计 */
@media (max-width: 768px) {
  .month-header {
    padding: 32px 16px 16px 16px;
  }

  .month-title {
    font-size: 24px;
  }

  .memory-card {
    margin: 12px 16px;
  }

  .date-card {
    padding: 12px 16px 8px 16px;
  }

  .photos-grid {
    margin: 0 16px 20px 16px;
  }
}
</style>
