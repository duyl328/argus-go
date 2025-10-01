import { ref, computed, watch, type Ref } from 'vue'

interface UseVirtualScrollOptions {
  items: Ref<any[]>
  itemHeight: number // 固定高度（列表模式）
  containerHeight: Ref<number>
  overscan?: number // 上下额外渲染的项目数
}

export function useVirtualScroll(options: UseVirtualScrollOptions) {
  const { items, itemHeight, containerHeight, overscan = 5 } = options

  const scrollTop = ref(0)

  // 可视区域可以显示的项目数
  const visibleCount = computed(() => {
    return Math.ceil(containerHeight.value / itemHeight) + overscan * 2
  })

  // 开始索引
  const startIndex = computed(() => {
    const index = Math.floor(scrollTop.value / itemHeight) - overscan
    return Math.max(0, index)
  })

  // 结束索引
  const endIndex = computed(() => {
    return Math.min(items.value.length, startIndex.value + visibleCount.value)
  })

  // 可见项目
  const visibleItems = computed(() => {
    return items.value.slice(startIndex.value, endIndex.value)
  })

  // 总高度
  const totalHeight = computed(() => {
    return items.value.length * itemHeight
  })

  // 偏移量
  const offsetY = computed(() => {
    return startIndex.value * itemHeight
  })

  // 滚动处理
  function onScroll(event: Event) {
    const target = event.target as HTMLElement
    scrollTop.value = target.scrollTop
  }

  return {
    visibleItems,
    totalHeight,
    offsetY,
    startIndex,
    endIndex,
    onScroll,
    scrollTop
  }
}

// 网格虚拟滚动（更复杂，需要计算列数）
interface UseVirtualGridOptions {
  items: Ref<any[]>
  itemWidth: number
  itemHeight: number
  containerWidth: Ref<number>
  containerHeight: Ref<number>
  gap?: number
  overscan?: number
}

export function useVirtualGrid(options: UseVirtualGridOptions) {
  const {
    items,
    itemWidth,
    itemHeight,
    containerWidth,
    containerHeight,
    gap = 8,
    overscan = 2
  } = options

  const scrollTop = ref(0)

  // 每行的列数
  const columns = computed(() => {
    const availableWidth = containerWidth.value - gap
    return Math.max(1, Math.floor(availableWidth / (itemWidth + gap)))
  })

  // 总行数
  const rowCount = computed(() => {
    return Math.ceil(items.value.length / columns.value)
  })

  // 行高（包含间距）
  const rowHeight = computed(() => {
    return itemHeight + gap
  })

  // 可见行数
  const visibleRowCount = computed(() => {
    return Math.ceil(containerHeight.value / rowHeight.value) + overscan * 2
  })

  // 开始行索引
  const startRow = computed(() => {
    const row = Math.floor(scrollTop.value / rowHeight.value) - overscan
    return Math.max(0, row)
  })

  // 结束行索引
  const endRow = computed(() => {
    return Math.min(rowCount.value, startRow.value + visibleRowCount.value)
  })

  // 开始项目索引
  const startIndex = computed(() => {
    return startRow.value * columns.value
  })

  // 结束项目索引
  const endIndex = computed(() => {
    return Math.min(items.value.length, endRow.value * columns.value)
  })

  // 可见项目
  const visibleItems = computed(() => {
    return items.value.slice(startIndex.value, endIndex.value)
  })

  // 总高度
  const totalHeight = computed(() => {
    return rowCount.value * rowHeight.value
  })

  // 偏移量
  const offsetY = computed(() => {
    return startRow.value * rowHeight.value
  })

  // 滚动处理
  function onScroll(event: Event) {
    const target = event.target as HTMLElement
    scrollTop.value = target.scrollTop
  }

  return {
    visibleItems,
    totalHeight,
    offsetY,
    startIndex,
    endIndex,
    columns,
    rowCount,
    onScroll,
    scrollTop
  }
}
