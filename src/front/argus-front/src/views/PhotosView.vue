<template>
  <div class="photos-timeline">
    <!-- 左侧照片展示区域（虚拟滚动） -->
    <div class="photos-content" ref="photosContentRef">
      <RecycleScroller
        class="virtual-scroller"
        :items="virtualItems"
        :item-size="null"
        key-field="id"
        size-field="size"
        v-slot="{ item }"
        @scroll="handleVirtualScroll"
        :buffer="400"
        :prerender="10"
      >
        <!-- 时期标题 -->
        <div
          v-if="item.type === 'header'"
          class="period-header"
          :ref="(el) => setPeriodRef(item.periodId, el as HTMLElement)"
        >
          <h2 class="period-title">{{ item.period?.displayName }}</h2>
          <span class="photo-count">{{ item.period?.photos.length }} 张照片</span>
        </div>

        <!-- 照片行或间距 -->
        <div
          v-else-if="item.type === 'row'"
          class="photos-row"
          :class="{ 'fill-width': (item.photos?.length || 0) > 2 }"
        >
          <div
            v-for="photo in item.photos"
            :key="photo.id"
            class="photo-item"
            :style="{
              backgroundColor: photo.color,
              width: `${(photo.width || 1) * 160}px`,
              flexGrow: (item.photos?.length || 0) > 2 ? (photo.width || 1) * 0.3 : 0,
              flexShrink: (item.photos?.length || 0) > 2 ? 1 : 0
            }"
          >
            <div class="photo-placeholder">
              <span class="photo-id">#{{ photo.id }}</span>
            </div>
          </div>
        </div>
      </RecycleScroller>
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
      @mousemove="handleTimelineHover"
      @click="handleTimelineClick"
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

        <!-- 悬停时间显示 -->
        <div
          v-if="timelineHovered"
          class="hover-time-display"
          :style="{ top: `${hoverTimePosition}%` }"
        >
          {{ hoverTimeText }}
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
import { ref, onMounted, onUnmounted, nextTick, reactive, computed } from 'vue'
import { RecycleScroller } from 'vue-virtual-scroller'
import 'vue-virtual-scroller/dist/vue-virtual-scroller.css'

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

  // 生成过去36个月的数据，模拟大量照片，总计约1万张以上
  for (let i = 0; i < 36; i++) {
    const date = new Date(now.getFullYear(), now.getMonth() - i, 1)
    const year = date.getFullYear()
    const month = String(date.getMonth() + 1).padStart(2, '0')
    const period = `${year}-${month}`

    periods.push({
      id: period,
      period,
      displayName: `${year}年${parseInt(month)}月`,
      photos: generatePhotos(Math.floor(Math.random() * 400) + 200), // 200-600张照片每月
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

// 悬停时间显示
const hoverTimePosition = ref(0)
const hoverTimeText = ref('')

// 连续进度条状态
const scrollProgress = ref(0)
const timelineHoverY = ref(0)
const timelineRect = ref<DOMRect | null>(null)
const hoverTime = ref('')

// 当前时间显示
const currentViewTime = ref('')

// 虚拟滚动数据结构：以行为单位，包含时期标题和照片行
interface VirtualItem {
  type: 'header' | 'row'
  periodId: string
  period?: PhotoPeriod
  photos?: Photo[] // 当前行的照片
  id: string
  size: number
}

// 响应式容器宽度
const containerWidth = ref(1200)

const virtualItems = computed<VirtualItem[]>(() => {
  const items: VirtualItem[] = []
  // 响应式调整每行照片数量
  let photosPerRow: number
  if (containerWidth.value >= 1400) {
    photosPerRow = Math.floor(containerWidth.value / 172) // 大屏幕
  } else if (containerWidth.value >= 1200) {
    photosPerRow = Math.floor(containerWidth.value / 180) // 中屏幕，稍微紧凑
  } else if (containerWidth.value >= 900) {
    photosPerRow = Math.floor(containerWidth.value / 190) // 小屏幕，更紧凑
  } else {
    photosPerRow = Math.max(2, Math.floor(containerWidth.value / 200)) // 移动端最少2列
  }
  photosPerRow = Math.max(1, photosPerRow)

  photoPeriods.value.forEach(period => {
    // 添加时期标题
    items.push({
      type: 'header',
      periodId: period.id,
      period,
      id: `header-${period.id}`,
      size: 80 // 标题固定高度
    })

    // 将照片分割成行
    for (let i = 0; i < period.photos.length; i += photosPerRow) {
      const rowPhotos = period.photos.slice(i, i + photosPerRow)
      items.push({
        type: 'row',
        periodId: period.id,
        photos: rowPhotos,
        id: `row-${period.id}-${Math.floor(i / photosPerRow)}`,
        size: isMobile.value ? 152 : isTablet.value ? 192 : 216 // 响应式行高度
      })
    }

    // 添加时期底部间距
    items.push({
      type: 'row',
      periodId: period.id,
      photos: [],
      id: `spacer-${period.id}`,
      size: 32 // 底部间距
    })
  })

  return items
})

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

// 虚拟滚动事件处理
const handleVirtualScroll = (event: Event) => {
  const target = event.target as HTMLElement
  if (!target) return

  // 滚动状态管理
  isScrolling.value = true
  clearTimeout(scrollTimeout)
  scrollTimeout = setTimeout(() => {
    isScrolling.value = false
  }, 150)

  const scrollTop = target.scrollTop
  const containerHeight = target.clientHeight
  const scrollHeight = target.scrollHeight

  // 计算滚动进度 (0-1)
  scrollProgress.value = Math.min(scrollTop / (scrollHeight - containerHeight), 1)

  // 更精确地查找当前可见的时期
  let accumulatedHeight = 0
  let currentPeriodId = ''
  const viewportCenter = scrollTop + containerHeight / 2

  for (const item of virtualItems.value) {
    if (accumulatedHeight + item.size > viewportCenter) {
      currentPeriodId = item.periodId
      break
    }
    accumulatedHeight += item.size
  }

  if (currentPeriodId && currentPeriodId !== activePeriodId.value) {
    activePeriodId.value = currentPeriodId
    // 根据当前时期更新时间显示
    const period = photoPeriods.value.find(p => p.id === currentPeriodId)
    if (period) {
      const [year, month] = period.period.split('-')
      const dayInMonth = Math.floor(Math.random() * 28) + 1
      currentViewTime.value = `${year}年${parseInt(month)}月${dayInMonth}日`
    }
  }
}

// 跳转到指定时期
const scrollToPeriod = (periodId: string) => {
  // 在虚拟列表中查找目标时期的标题索引
  const targetIndex = virtualItems.value.findIndex(item =>
    item.type === 'header' && item.periodId === periodId
  )

  if (targetIndex >= 0) {
    const scroller = document.querySelector('.virtual-scroller') as HTMLElement
    if (scroller) {
      // 计算精确的滚动位置：累计前面所有项的高度
      let accumulatedHeight = 0
      for (let i = 0; i < targetIndex; i++) {
        accumulatedHeight += virtualItems.value[i].size
      }

      scroller.scrollTo({
        top: accumulatedHeight,
        behavior: 'smooth'
      })

      // 同步更新状态
      activePeriodId.value = periodId

      // 立即更新进度条位置
      const scrollHeight = scroller.scrollHeight - scroller.clientHeight
      const newProgress = Math.min(accumulatedHeight / scrollHeight, 1)
      scrollProgress.value = newProgress
    }
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

// 时间线悬停处理
const handleTimelineHover = (event: MouseEvent) => {
  const timelineNav = event.currentTarget as HTMLElement
  const rect = timelineNav.getBoundingClientRect()
  const relativeY = event.clientY - rect.top
  const hoverProgress = Math.max(0, Math.min(1, relativeY / rect.height))

  hoverTimePosition.value = hoverProgress * 100

  // 根据悬停位置计算对应的时间
  const totalPeriods = photoPeriods.value.length
  if (totalPeriods > 0) {
    const periodIndex = Math.floor(hoverProgress * totalPeriods)
    const safeIndex = Math.max(0, Math.min(totalPeriods - 1, periodIndex))
    const period = photoPeriods.value[safeIndex]

    // 在时期内计算具体日期
    const periodProgress = (hoverProgress * totalPeriods) - periodIndex
    const dayInMonth = Math.floor(periodProgress * 30) + 1
    const [year, month] = period.period.split('-')
    hoverTimeText.value = `${year}年${parseInt(month)}月${Math.min(dayInMonth, 30)}日`
  }
}

// 时间线点击跳转处理
const handleTimelineClick = (event: MouseEvent) => {
  const timelineNav = event.currentTarget as HTMLElement
  const rect = timelineNav.getBoundingClientRect()
  const relativeY = event.clientY - rect.top
  const clickProgress = Math.max(0, Math.min(1, relativeY / rect.height))

  // 计算目标滚动位置
  const scroller = document.querySelector('.virtual-scroller') as HTMLElement
  if (scroller && virtualItems.value.length > 0) {
    // 计算总内容高度
    const totalHeight = virtualItems.value.reduce((sum, item) => sum + item.size, 0)
    const scrollHeight = scroller.scrollHeight - scroller.clientHeight
    const targetScrollTop = Math.min(clickProgress * scrollHeight, scrollHeight)

    // 平滑滚动到目标位置
    scroller.scrollTo({
      top: targetScrollTop,
      behavior: 'smooth'
    })

    // 更新进度条位置
    scrollProgress.value = clickProgress

    // 计算并更新当前时间显示
    const totalPeriods = photoPeriods.value.length
    if (totalPeriods > 0) {
      const periodIndex = Math.floor(clickProgress * totalPeriods)
      const safeIndex = Math.max(0, Math.min(totalPeriods - 1, periodIndex))
      const period = photoPeriods.value[safeIndex]

      const periodProgress = (clickProgress * totalPeriods) - periodIndex
      const dayInMonth = Math.floor(periodProgress * 30) + 1
      const [year, month] = period.period.split('-')
      currentViewTime.value = `${year}年${parseInt(month)}月${Math.min(dayInMonth, 30)}日`
    }
  }
}

// 窗口大小监听和响应式适配
const updateContainerWidth = () => {
  const timeline = document.querySelector('.photos-timeline') as HTMLElement
  if (timeline) {
    const timelineWidth = window.innerWidth <= 900 ? 0 : 24 // 小屏幕隐藏时间线，新的更窄宽度
    containerWidth.value = timeline.clientWidth - timelineWidth - 48 // 减去时间线宽度和内边距
  }
}

// 响应式断点检测
const isMobile = computed(() => containerWidth.value <= 768)
const isTablet = computed(() => containerWidth.value > 768 && containerWidth.value <= 1024)
const isDesktop = computed(() => containerWidth.value > 1024)

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

    // 初始化容器宽度
    updateContainerWidth()
    window.addEventListener('resize', updateContainerWidth)
  })
})

// 清理
onUnmounted(() => {
  window.removeEventListener('resize', updateContainerWidth)
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

/* 虚拟滚动容器 */
.virtual-scroller {
  height: 100%;
  width: 100%;
  overflow-y: auto !important;
  overflow-x: hidden !important;
}

.virtual-scroller::-webkit-scrollbar {
  width: 0px !important;
  background: transparent !important;
}

.virtual-scroller::-webkit-scrollbar-thumb {
  background: transparent !important;
}

/* 照片行 */
.photos-row {
  display: flex;
  gap: 8px;
  align-items: flex-start;
  margin-bottom: 8px;
  padding: 0 4px;
}

/* 填充宽度的照片行 */
.photos-row.fill-width {
  justify-content: stretch;
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
  width: 24px;
  background: transparent;
  border-left: none;
  position: relative;
  padding: 20px 4px;
  overflow: visible;
  flex-shrink: 0;
  transition: opacity 0.2s ease, background-color 0.2s ease;
  opacity: 0.3;
  cursor: row-resize;
}

@media (max-width: 1200px) {
  .timeline-nav {
    width: 20px;
    padding: 20px 2px;
  }
}

@media (max-width: 900px) {
  .timeline-nav {
    display: none;
  }

  .photos-content {
    padding: 16px;
  }

  .photo-item {
    height: 160px;
  }

  .period-title {
    font-size: 24px;
  }
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
  margin-right: 12px;
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

.hover-time-display {
  position: absolute;
  right: 100%;
  margin-right: 12px;
  background: rgba(0, 0, 0, 0.8);
  color: white;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 500;
  white-space: nowrap;
  pointer-events: none;
  z-index: 11;
  transform: translateY(-50%);
  opacity: 0.9;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
  backdrop-filter: blur(4px);
}

.hover-time-display::after {
  content: '';
  position: absolute;
  left: 100%;
  top: 50%;
  width: 0;
  height: 0;
  border-left: 6px solid rgba(0, 0, 0, 0.8);
  border-top: 4px solid transparent;
  border-bottom: 4px solid transparent;
  transform: translateY(-50%);
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
    width: 18px;
  }

  .photos-content {
    padding: 20px;
  }

  .photo-item {
    height: 180px;
  }

  .photos-row {
    gap: 6px;
    padding: 0 2px;
  }
}

@media (max-width: 768px) {
  .photos-content {
    padding: 12px;
  }

  .photo-item {
    height: 140px;
  }

  .period-title {
    font-size: 22px;
  }

  .photos-row {
    gap: 4px;
    padding: 0 1px;
  }

  .period-header {
    margin-bottom: 16px;
  }
}

@media (max-width: 480px) {
  .photos-content {
    padding: 8px;
  }

  .photo-item {
    height: 120px;
  }

  .period-title {
    font-size: 20px;
  }

  .photos-row {
    gap: 2px;
    padding: 0;
  }

  .period-header {
    margin-bottom: 12px;
    flex-direction: column;
    align-items: flex-start;
    gap: 4px;
  }
}
</style>
