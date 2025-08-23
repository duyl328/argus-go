<template>
  <div class="photos-timeline">
    <DynamicScroller
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
                height: photo.box.height + 'px',
                backgroundColor: photo.color
              }"
            >
              <div class="photo-placeholder">
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

interface Photo {
  id: string
  width: number
  height: number
  ratio: number
  color: string
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
}

const scrollerRef = ref()
const containerWidth = ref(1200)

// 生成随机颜色
function generateRandomColor(): string {
  const colors = [
    '#FF6B6B', '#4ECDC4', '#45B7D1', '#96CEB4', '#FFEAA7',
    '#DDA0DD', '#98D8C8', '#F7DC6F', '#BB8FCE', '#85C1E9',
    '#F8C471', '#82E0AA', '#F1948A', '#85C1E9', '#D7BDE2'
  ]
  return colors[Math.floor(Math.random() * colors.length)]
}

// 生成随机照片数据
function generateRandomPhotos(count: number): Photo[] {
  const photos: Photo[] = []
  for (let i = 0; i < count; i++) {
    const width = Math.floor(Math.random() * 400) + 200 // 200-600px
    const height = Math.floor(Math.random() * 400) + 200 // 200-600px
    const ratio = width / height
    
    photos.push({
      id: `photo-${i}`,
      width,
      height,
      ratio,
      color: generateRandomColor(),
      box: { top: 0, left: 0, width: 0, height: 0 } // 先设置默认值，后面会计算
    })
  }
  return photos
}

// 使用 justified-layout 计算照片布局
function calculatePhotosLayout(photos: Photo[]): { photos: Photo[], height: number } {
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
    box: geometry.boxes[index]
  }))

  return {
    photos: layoutPhotos,
    height: geometry.containerHeight
  }
}

// 生成时间线数据
function generateTimelineData(): TimelineItem[] {
  const items: TimelineItem[] = []
  const months = [
    { title: '2024年1月', subtitle: '128张照片' },
    { title: '2023年12月', subtitle: '95张照片' },
    { title: '2023年11月', subtitle: '156张照片' }
  ]

  months.forEach((month, monthIndex) => {
    // 添加月份标题
    items.push({
      id: `month-${monthIndex}`,
      type: 'month',
      title: month.title,
      subtitle: month.subtitle
    })

    // 添加回忆卡片 (随机)
    if (Math.random() > 0.6) {
      items.push({
        id: `memory-${monthIndex}`,
        type: 'memory',
        title: '回忆',
        subtitle: '一年前的今天'
      })
    }

    // 添加日期和照片组
    const daysInMonth = Math.floor(Math.random() * 8) + 3 // 3-10天
    for (let day = 1; day <= daysInMonth; day++) {
      const dayId = `day-${monthIndex}-${day}`
      
      // 添加日期卡片
      items.push({
        id: `date-${dayId}`,
        type: 'date',
        title: `${month.title.slice(0, 7)}${day}日`
      })

      // 添加照片组
      const photoCount = Math.floor(Math.random() * 20) + 5 // 5-24张照片
      const photos = generateRandomPhotos(photoCount)
      const layout = calculatePhotosLayout(photos)

      items.push({
        id: `photos-${dayId}`,
        type: 'photos',
        height: layout.height,
        photos: layout.photos
      })
    }
  })

  return items
}

// 响应式数据
const timelineItems = ref<TimelineItem[]>([])

// 初始化数据
onMounted(async () => {
  await nextTick()
  
  // 获取容器宽度
  const container = document.querySelector('.photos-timeline')
  if (container) {
    containerWidth.value = container.clientWidth
  }
  
  // 生成时间线数据
  timelineItems.value = generateTimelineData()
  
  // 监听窗口大小变化
  const handleResize = () => {
    if (container) {
      const newWidth = container.clientWidth
      if (Math.abs(newWidth - containerWidth.value) > 50) { // 只有明显变化时才重新计算
        containerWidth.value = newWidth
        // 重新计算所有照片布局
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
    }
  }
  
  window.addEventListener('resize', handleResize)
  
  // 清理事件监听器
  return () => {
    window.removeEventListener('resize', handleResize)
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
  transition: transform 0.2s ease;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.photo-item:hover {
  transform: scale(1.02);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
  z-index: 1;
}

.photo-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  color: white;
  font-weight: bold;
  text-shadow: 0 1px 3px rgba(0, 0, 0, 0.3);
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