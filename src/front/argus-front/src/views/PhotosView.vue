<template>
  <div class="photos-timeline">
    <!-- 左侧照片展示区域 -->
    <div class="photos-content" ref="photosContentRef" @scroll="handleScroll">
      <div
        v-for="period in photoPeriods"
        :key="period.id"
        :ref="(el) => setPeriodRef(period.id, el as HTMLElement)"
        class="photo-period"
      >
        <!-- 时期标题 -->
        <div class="period-header">
          <h2 class="period-title">{{ period.displayName }}</h2>
          <span class="photo-count">{{ period.photos.length }} 张照片</span>
        </div>

        <!-- 照片网格 -->
        <div class="photos-grid">
          <div
            v-for="photo in period.photos"
            :key="photo.id"
            class="photo-item"
            :style="{
              backgroundColor: photo.color,
              width: `${(photo.width || 1) * 160}px`,
              flexGrow: (photo.width || 1) * 0.1,
              flexShrink: 1
            }"
          >
            <div class="photo-placeholder">
              <span class="photo-id">#{{ photo.id }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 右侧时间线导航 -->
    <div 
      class="timeline-nav"
      :class="{ 
        hovered: timelineHovered,
        scrolling: isScrolling 
      }"
      @mouseenter="timelineHovered = true"
      @mouseleave="timelineHovered = false"
    >
      <div class="timeline-container">
        <!-- 时间线背景线 -->
        <div class="timeline-line"></div>
        
        <!-- 进度指示器 -->
        <div 
          class="timeline-progress"
          :style="{ height: `${scrollProgress * 100}%` }"
        ></div>
        
        <!-- 当前位置指示器 -->
        <div 
          class="timeline-indicator"
          :style="{ top: `${scrollProgress * 100}%` }"
        ></div>
        
        <!-- 当前时间显示 -->
        <div 
          class="current-time-display"
          :class="{ visible: isScrolling || timelineHovered }"
          :style="{ top: `${scrollProgress * 100}%` }"
        >
          {{ currentViewTime }}
        </div>

        <div
          v-for="(period, index) in photoPeriods"
          :key="period.id"
          class="timeline-node"
          :class="{ 
            active: activePeriodId === period.id,
            inactive: isScrolling && activePeriodId !== period.id
          }"
          :style="{ 
            top: `${(index / Math.max(photoPeriods.length - 1, 1)) * 100}%` 
          }"
          @click.stop="scrollToPeriod(period.id)"
        >
          <div class="node-dot"></div>
          <div class="node-label">{{ period.period.split('-')[1] }}月</div>
        </div>
      </div>

      <!-- 悬停提示 -->
      <div v-if="tooltipVisible" class="timeline-tooltip" :style="tooltipStyle">
        {{ tooltipContent }}
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, nextTick, reactive } from 'vue'

interface Photo {
  id: string
  color: string
  width?: number
  height?: number
}

interface PhotoPeriod {
  id: string
  period: string // "2024-08"
  displayName: string // "2024年8月"
  photos: Photo[]
}

// 生成随机颜色
const generateRandomColor = (): string => {
  const colors = [
    '#FF6B6B',
    '#4ECDC4',
    '#45B7D1',
    '#96CEB4',
    '#FFEAA7',
    '#DDA0DD',
    '#98D8C8',
    '#F7DC6F',
    '#BB8FCE',
    '#85C1E9',
    '#F8C471',
    '#82E0AA',
    '#F1948A',
    '#85C1E9',
    '#D7BDE2',
  ]
  return colors[Math.floor(Math.random() * colors.length)]
}

// 生成随机照片
const generatePhotos = (count: number): Photo[] => {
  return Array.from({ length: count }, (_, i) => ({
    id: String(i + 1).padStart(3, '0'),
    color: generateRandomColor(),
    width: Math.random() > 0.7 ? 2 : 1, // 30% 概率宽照片
    height: Math.random() > 0.8 ? 2 : 1, // 20% 概率高照片
  }))
}

// 生成时期数据
const generatePhotoPeriods = (): PhotoPeriod[] => {
  const periods: PhotoPeriod[] = []
  const now = new Date()

  for (let i = 0; i < 8; i++) {
    const date = new Date(now.getFullYear(), now.getMonth() - i, 1)
    const year = date.getFullYear()
    const month = String(date.getMonth() + 1).padStart(2, '0')
    const period = `${year}-${month}`

    periods.push({
      id: period,
      period,
      displayName: `${year}年${parseInt(month)}月`,
      photos: generatePhotos(Math.floor(Math.random() * 30) + 10),
    })
  }

  return periods
}

// 响应式数据
const photoPeriods = ref<PhotoPeriod[]>(generatePhotoPeriods())
const photosContentRef = ref<HTMLElement>()
const periodRefs = reactive<Record<string, HTMLElement>>({})
const activePeriodId = ref<string>(photoPeriods.value[0]?.id || '')

// 时间线提示
const tooltipVisible = ref(false)
const tooltipContent = ref('')
const tooltipStyle = ref<Record<string, string>>({})

// 时间线交互状态
const timelineHovered = ref(false)
const isScrolling = ref(false)
let scrollTimeout: number

// 连续进度条状态
const scrollProgress = ref(0)
const timelineHoverY = ref(0)
const timelineRect = ref<DOMRect | null>(null)
const hoverTime = ref('')

// 当前时间显示
const currentViewTime = ref('')

// 设置时期元素引用
const setPeriodRef = (periodId: string, el: HTMLElement) => {
  if (el) {
    periodRefs[periodId] = el
  }
}

// 滚动处理
const handleScroll = () => {
  if (!photosContentRef.value) return
  
  // 滚动状态管理
  isScrolling.value = true
  clearTimeout(scrollTimeout)
  scrollTimeout = setTimeout(() => {
    isScrolling.value = false
  }, 150)

  const scrollTop = photosContentRef.value.scrollTop
  const containerHeight = photosContentRef.value.clientHeight
  const scrollHeight = photosContentRef.value.scrollHeight
  const viewportCenter = scrollTop + containerHeight / 2
  
  // 计算滚动进度 (0-1)
  scrollProgress.value = Math.min(scrollTop / (scrollHeight - containerHeight), 1)

  // 找到当前视口中心对应的时期并计算精确时间
  let currentPeriod = null
  for (const period of photoPeriods.value) {
    const periodEl = periodRefs[period.id]
    if (periodEl) {
      const periodTop = periodEl.offsetTop
      const periodBottom = periodTop + periodEl.offsetHeight

      if (viewportCenter >= periodTop && viewportCenter < periodBottom) {
        activePeriodId.value = period.id
        currentPeriod = period
        
        // 计算期间内的进度，推算具体日期
        const periodProgress = (viewportCenter - periodTop) / periodEl.offsetHeight
        const dayInMonth = Math.floor(periodProgress * 30) + 1
        const [year, month] = period.period.split('-')
        currentViewTime.value = `${year}年${parseInt(month)}月${dayInMonth}日`
        break
      }
    }
  }
  
  // 如果没有找到精确匹配，使用第一个或最后一个时期
  if (!currentPeriod && photoPeriods.value.length > 0) {
    const firstPeriod = photoPeriods.value[0]
    const [year, month] = firstPeriod.period.split('-')
    if (scrollProgress.value < 0.1) {
      currentViewTime.value = `${year}年${parseInt(month)}月1日`
    } else {
      const lastPeriod = photoPeriods.value[photoPeriods.value.length - 1]
      const [lastYear, lastMonth] = lastPeriod.period.split('-')
      currentViewTime.value = `${lastYear}年${parseInt(lastMonth)}月30日`
    }
  }
}

// 跳转到指定时期
const scrollToPeriod = (periodId: string) => {
  const periodEl = periodRefs[periodId]
  if (periodEl && photosContentRef.value) {
    const offsetTop = periodEl.offsetTop - 20 // 留一些边距
    photosContentRef.value.scrollTo({
      top: offsetTop,
      behavior: 'smooth',
    })
    activePeriodId.value = periodId
  }
}

// 显示悬停提示
const showTooltip = (period: PhotoPeriod, event: MouseEvent) => {
  tooltipContent.value = `${period.displayName} • ${period.photos.length} 张照片`
  tooltipVisible.value = true

  const rect = (event.target as HTMLElement).getBoundingClientRect()
  tooltipStyle.value = {
    left: `${rect.left - 120}px`,
    top: `${rect.top + rect.height / 2}px`,
    transform: 'translateY(-50%)',
  }
}

// 隐藏悬停提示
const hideTooltip = () => {
  tooltipVisible.value = false
}

// 初始化
onMounted(() => {
  nextTick(() => {
    if (photoPeriods.value.length > 0) {
      activePeriodId.value = photoPeriods.value[0].id
      // 设置初始时间显示
      const firstPeriod = photoPeriods.value[0]
      const [year, month] = firstPeriod.period.split('-')
      currentViewTime.value = `${year}年${parseInt(month)}月1日`
    }
  })
})
</script>

<style scoped>
.photos-timeline {
  display: flex;
  height: 100%;
  background: var(--bg-color);
  overflow: hidden;
}

/* 左侧照片内容区域 */
.photos-content {
  flex: 1;
  padding: 24px;
  overflow-y: auto;
  overflow-x: hidden;
  scroll-behavior: smooth;
  
  /* 隐藏滚动条 */
  scrollbar-width: none; /* Firefox */
  -ms-overflow-style: none; /* IE/Edge */
}

.photos-content::-webkit-scrollbar {
  display: none; /* Webkit browsers */
}

.photo-period {
  margin-bottom: 48px;
}

.period-header {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  margin-bottom: 20px;
  padding-bottom: 8px;
  border-bottom: 2px solid var(--border-color);
}

.period-title {
  font-size: 28px;
  font-weight: 600;
  color: var(--text-color);
  margin: 0;
}

.photo-count {
  font-size: 14px;
  color: var(--text-color-secondary);
}

/* 照片网格 */
.photos-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  margin-bottom: 32px;
  align-items: flex-start;
}

.photo-item {
  height: 200px;
  border-radius: 4px;
  overflow: hidden;
  transition:
    transform 0.15s ease,
    box-shadow 0.15s ease;
  cursor: pointer;
  position: relative;
}

.photo-item:hover {
  transform: scale(1.02);
  box-shadow: 0 8px 25px rgba(0, 0, 0, 0.15);
  z-index: 2;
}

.photo-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
}

.photo-id {
  color: rgba(255, 255, 255, 0.8);
  font-size: 12px;
  font-weight: 600;
  background: rgba(0, 0, 0, 0.2);
  padding: 4px 8px;
  border-radius: 4px;
  backdrop-filter: blur(4px);
}

/* 右侧时间线导航 */
.timeline-nav {
  width: 80px;
  background: transparent;
  border-left: none;
  position: relative;
  padding: 20px 8px;
  overflow: visible;
  flex-shrink: 0;
  transition: opacity 0.2s ease, background-color 0.2s ease;
  opacity: 0.3;
  cursor: crosshair;
}

.timeline-nav:hover,
.timeline-nav.hovered {
  background: rgba(0, 0, 0, 0.02);
  opacity: 1;
}

.timeline-nav.scrolling {
  opacity: 0.9;
  background: rgba(0, 0, 0, 0.01);
}

.timeline-container {
  position: relative;
  height: 100%;
  display: flex;
  flex-direction: column;
  justify-content: space-around;
  min-height: 400px;
}

.timeline-line {
  position: absolute;
  left: 50%;
  top: 0;
  bottom: 0;
  width: 2px;
  background: rgba(0, 0, 0, 0.1);
  transform: translateX(-50%);
  border-radius: 1px;
}

.timeline-progress {
  position: absolute;
  left: 50%;
  top: 0;
  width: 2px;
  background: #1a73e8;
  transform: translateX(-50%);
  border-radius: 1px;
  transition: all 0.1s ease;
  z-index: 2;
}

.timeline-indicator {
  position: absolute;
  left: 50%;
  width: 8px;
  height: 8px;
  background: #1a73e8;
  border: 2px solid white;
  border-radius: 50%;
  transform: translate(-50%, -50%);
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
  z-index: 3;
  transition: all 0.1s ease;
}

.timeline-nav:hover .timeline-indicator {
  width: 12px;
  height: 12px;
  box-shadow: 0 3px 8px rgba(26, 115, 232, 0.3);
}

.current-time-display {
  position: absolute;
  right: 100%;
  margin-right: 16px;
  background: #1a73e8;
  color: white;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 500;
  white-space: nowrap;
  pointer-events: none;
  z-index: 10;
  transform: translateY(-50%);
  opacity: 0;
  transition: opacity 0.2s ease;
  box-shadow: 0 2px 8px rgba(26, 115, 232, 0.3);
}

.current-time-display.visible {
  opacity: 1;
}

.current-time-display::after {
  content: '';
  position: absolute;
  left: 100%;
  top: 50%;
  width: 0;
  height: 0;
  border-left: 6px solid #1a73e8;
  border-top: 4px solid transparent;
  border-bottom: 4px solid transparent;
  transform: translateY(-50%);
}

.timeline-node {
  position: absolute;
  display: flex;
  flex-direction: column;
  align-items: center;
  cursor: pointer;
  transition: all 0.3s ease;
  padding: 8px 0;
  opacity: 0.6;
  z-index: 4;
  left: 50%;
  transform: translateX(-50%);
}

.timeline-node.inactive {
  opacity: 0.1;
  transform: translateX(-50%) scale(0.6);
}

.node-dot {
  width: 4px;
  height: 4px;
  border-radius: 50%;
  background: rgba(0, 0, 0, 0.3);
  border: none;
  transition: all 0.3s ease;
  position: relative;
  margin-bottom: 2px;
}

.timeline-nav:hover .node-dot {
  width: 6px;
  height: 6px;
  background: rgba(0, 0, 0, 0.5);
}

.timeline-node.active .node-dot {
  width: 6px;
  height: 6px;
  background: #1a73e8;
}

.timeline-node:hover .node-dot {
  width: 8px;
  height: 8px;
  background: #1a73e8;
}

.node-label {
  font-size: 9px;
  color: rgba(0, 0, 0, 0.4);
  font-weight: 400;
  text-align: center;
  transition: all 0.3s ease;
  opacity: 0;
  white-space: nowrap;
}

.timeline-nav:hover .node-label {
  opacity: 0.8;
  font-size: 10px;
}

.timeline-node:hover .node-label {
  color: #1a73e8;
  opacity: 1;
}

.timeline-node.active .node-label {
  color: #1a73e8;
  font-weight: 500;
  opacity: 0.9;
}

/* 悬停时间提示 */
.timeline-hover-tooltip {
  position: absolute;
  right: 100%;
  margin-right: 12px;
  background: rgba(0, 0, 0, 0.8);
  color: white;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 11px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
  white-space: nowrap;
  pointer-events: none;
  backdrop-filter: blur(4px);
  z-index: 1000;
  transform: translateY(-50%);
}

/* 旧的节点悬停提示（保留用于兼容） */
.timeline-tooltip {
  position: fixed;
  z-index: 1000;
  background: rgba(0, 0, 0, 0.8);
  color: white;
  padding: 6px 10px;
  border-radius: 4px;
  font-size: 11px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
  white-space: nowrap;
  pointer-events: none;
  backdrop-filter: blur(4px);
}

/* 响应式适配 */
@media (max-width: 1024px) {
  .timeline-nav {
    width: 60px;
  }

  .photos-content {
    padding: 20px;
  }

  .photo-item {
    height: 160px;
  }
}

@media (max-width: 768px) {
  .photos-timeline {
    flex-direction: column;
  }

  .timeline-nav {
    width: 100%;
    height: 60px;
    border-left: none;
    border-bottom: 1px solid rgba(0, 0, 0, 0.1);
    padding: 8px 16px;
  }

  .timeline-container {
    flex-direction: row;
    justify-content: space-around;
    min-height: auto;
    height: 100%;
    align-items: center;
  }

  .timeline-line {
    left: 16px;
    right: 16px;
    top: 50%;
    bottom: auto;
    height: 1px;
    width: auto;
    transform: translateY(-50%);
  }

  .timeline-node {
    padding: 4px;
    margin-bottom: 0;
  }

  .node-label {
    font-size: 10px;
  }

  .photos-content {
    padding: 16px;
    height: calc(100% - 60px);
  }
  
  .current-time-display {
    display: none;
  }

  .photo-item {
    height: 140px;
  }

  .period-title {
    font-size: 24px;
  }
}
</style>
