import { ref, computed, watch, type Ref, unref, isRef } from 'vue'

interface UseVirtualScrollOptions {
  items: Ref<any[]>
  itemHeight: number | Ref<number> // 支持响应式高度
  containerHeight: Ref<number>
  containerWidth?: Ref<number> // 可选的容器宽度（用于边界计算）
  overscan?: number // 上下额外渲染的项目数
}

export function useVirtualScroll(options: UseVirtualScrollOptions) {
  const { items, itemHeight: _itemHeight, containerHeight, containerWidth, overscan = 5 } = options

  const scrollTop = ref(0)
  const rafId = ref<number | null>(null)

  // 统一处理为 Ref
  const itemHeight = computed(() => unref(_itemHeight))

  // 可视区域可以显示的项目数
  const visibleCount = computed(() => {
    return Math.ceil(containerHeight.value / itemHeight.value) + overscan * 2
  })

  // 开始索引
  const startIndex = computed(() => {
    const index = Math.floor(scrollTop.value / itemHeight.value) - overscan
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
    return items.value.length * itemHeight.value
  })

  // 偏移量
  const offsetY = computed(() => {
    return startIndex.value * itemHeight.value
  })

  // 滚动处理（使用 requestAnimationFrame 优化）
  function onScroll(event: Event) {
    const target = event.target as HTMLElement

    if (rafId.value !== null) {
      cancelAnimationFrame(rafId.value)
    }

    rafId.value = requestAnimationFrame(() => {
      scrollTop.value = target.scrollTop
      rafId.value = null
    })
  }

  // 根据索引计算项目位置（用于框选等功能）
  function getItemBounds(index: number) {
    return {
      top: index * itemHeight.value,
      left: 0,
      bottom: (index + 1) * itemHeight.value,
      right: containerWidth?.value ?? 0
    }
  }

  return {
    visibleItems,
    totalHeight,
    offsetY,
    startIndex,
    endIndex,
    onScroll,
    scrollTop,
    getItemBounds
  }
}

// 网格虚拟滚动（更复杂，需要计算列数）
interface UseVirtualGridOptions {
  items: Ref<any[]>
  itemWidth: number | Ref<number>  // 支持响应式宽度
  itemHeight: number | Ref<number> // 支持响应式高度
  containerWidth: Ref<number>
  containerHeight: Ref<number>
  gap?: number
  overscan?: number
}

export function useVirtualGrid(options: UseVirtualGridOptions) {
  const {
    items,
    itemWidth: _itemWidth,
    itemHeight: _itemHeight,
    containerWidth,
    containerHeight,
    gap = 8,
    overscan = 2
  } = options

  const scrollTop = ref(0)
  const rafId = ref<number | null>(null)

  // 统一处理为 computed
  const itemWidth = computed(() => unref(_itemWidth))
  const itemHeight = computed(() => unref(_itemHeight))

  // 每行的列数
  const columns = computed(() => {
    const availableWidth = containerWidth.value - gap
    return Math.max(1, Math.floor(availableWidth / (itemWidth.value + gap)))
  })

  // 总行数
  const rowCount = computed(() => {
    return Math.ceil(items.value.length / columns.value)
  })

  // 行高（包含间距）
  const rowHeight = computed(() => {
    return itemHeight.value + gap
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

  // 滚动处理（使用 requestAnimationFrame 优化）
  function onScroll(event: Event) {
    const target = event.target as HTMLElement

    if (rafId.value !== null) {
      cancelAnimationFrame(rafId.value)
    }

    rafId.value = requestAnimationFrame(() => {
      scrollTop.value = target.scrollTop
      rafId.value = null
    })
  }

  // 根据索引计算项目位置（用于框选等功能）
  function getItemBounds(index: number) {
    const row = Math.floor(index / columns.value)
    const col = index % columns.value

    return {
      top: row * rowHeight.value,
      left: col * (itemWidth.value + gap),
      bottom: (row + 1) * rowHeight.value,
      right: (col + 1) * (itemWidth.value + gap)
    }
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
    scrollTop,
    getItemBounds
  }
}
