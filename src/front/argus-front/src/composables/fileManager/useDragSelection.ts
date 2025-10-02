import { ref, computed } from 'vue'

/**
 * 拖拽框选功能
 * 处理鼠标拖拽选择文件/文件夹的逻辑
 */
export function useDragSelection() {
  // 拖拽框选状态
  const dragSelection = ref({
    isSelecting: false,
    startX: 0,
    startY: 0,
    currentX: 0,
    currentY: 0,
    initialSelections: new Set<string>(),
    ctrlKey: false,
    justFinished: false
  })

  // 自动滚动状态
  const autoScroll = ref({
    isScrolling: false,
    direction: { x: 0, y: 0 }
  })

  // 计算选择框边界（用于碰撞检测）
  const selectionBox = computed(() => {
    if (!dragSelection.value.isSelecting) return null

    const left = Math.min(dragSelection.value.startX, dragSelection.value.currentX)
    const top = Math.min(dragSelection.value.startY, dragSelection.value.currentY)
    const right = Math.max(dragSelection.value.startX, dragSelection.value.currentX)
    const bottom = Math.max(dragSelection.value.startY, dragSelection.value.currentY)

    return { top, left, bottom, right }
  })

  // 计算选择框样式
  const selectionBoxStyle = computed(() => {
    if (!dragSelection.value.isSelecting) return { display: 'none' }

    const left = Math.min(dragSelection.value.startX, dragSelection.value.currentX)
    const top = Math.min(dragSelection.value.startY, dragSelection.value.currentY)
    const width = Math.abs(dragSelection.value.currentX - dragSelection.value.startX)
    const height = Math.abs(dragSelection.value.currentY - dragSelection.value.startY)

    return {
      display: 'block',
      left: `${left}px`,
      top: `${top}px`,
      width: `${width}px`,
      height: `${height}px`
    }
  })

  /**
   * 开始拖拽选择
   */
  function startDragSelection(
    event: MouseEvent,
    containerRef: HTMLElement,
    ctrlKey: boolean,
    initialSelections: Set<string>
  ) {
    const rect = containerRef.getBoundingClientRect()
    const scrollLeft = containerRef.scrollLeft
    const scrollTop = containerRef.scrollTop

    dragSelection.value = {
      isSelecting: true,
      startX: event.clientX - rect.left + scrollLeft,
      startY: event.clientY - rect.top + scrollTop,
      currentX: event.clientX - rect.left + scrollLeft,
      currentY: event.clientY - rect.top + scrollTop,
      initialSelections: new Set(initialSelections),
      ctrlKey,
      justFinished: false
    }
  }

  /**
   * 更新拖拽选择位置
   */
  function updateDragSelection(event: MouseEvent, containerRef: HTMLElement) {
    if (!dragSelection.value.isSelecting) return

    const rect = containerRef.getBoundingClientRect()
    const scrollEl = containerRef

    const rawX = event.clientX - rect.left + scrollEl.scrollLeft
    const rawY = event.clientY - rect.top + scrollEl.scrollTop

    const maxScrollLeft = Math.max(0, scrollEl.scrollWidth - scrollEl.clientWidth)
    const maxScrollTop = Math.max(0, scrollEl.scrollHeight - scrollEl.clientHeight)

    const maxX = scrollEl.clientWidth + maxScrollLeft
    const maxY = scrollEl.clientHeight + maxScrollTop

    dragSelection.value.currentX = Math.max(0, Math.min(rawX, maxX))
    dragSelection.value.currentY = Math.max(0, Math.min(rawY, maxY))
  }

  /**
   * 检查是否需要自动滚动
   */
  function checkAutoScroll(event: MouseEvent, containerRef: HTMLElement) {
    const rect = containerRef.getBoundingClientRect()
    const scrollEl = containerRef
    const threshold = 50
    const baseScrollSpeed = 10

    let dx = 0
    let dy = 0

    // 垂直滚动检测
    if (event.clientY < rect.top) {
      dy = -baseScrollSpeed
    } else if (event.clientY < rect.top + threshold) {
      const distance = event.clientY - rect.top
      const ratio = 1 - distance / threshold
      dy = -Math.max(3, baseScrollSpeed * ratio)
    } else if (event.clientY > rect.bottom) {
      dy = baseScrollSpeed
    } else if (event.clientY > rect.bottom - threshold) {
      const distance = rect.bottom - event.clientY
      const ratio = 1 - distance / threshold
      dy = Math.max(3, baseScrollSpeed * ratio)
    }

    // 水平滚动检测（仅当存在水平滚动空间时）
    const hasHorizontalScroll = scrollEl.scrollWidth > scrollEl.clientWidth
    if (hasHorizontalScroll) {
      if (event.clientX < rect.left) {
        dx = -baseScrollSpeed
      } else if (event.clientX < rect.left + threshold) {
        const distance = event.clientX - rect.left
        const ratio = 1 - distance / threshold
        dx = -Math.max(3, baseScrollSpeed * ratio)
      } else if (event.clientX > rect.right) {
        dx = baseScrollSpeed
      } else if (event.clientX > rect.right - threshold) {
        const distance = rect.right - event.clientX
        const ratio = 1 - distance / threshold
        dx = Math.max(3, baseScrollSpeed * ratio)
      }
    }

    // 更新自动滚动状态
    if (dx !== 0 || dy !== 0) {
      autoScroll.value.isScrolling = true
      autoScroll.value.direction = { x: dx, y: dy }
    } else {
      autoScroll.value.isScrolling = false
      autoScroll.value.direction = { x: 0, y: 0 }
    }
  }

  /**
   * 启动自动滚动循环
   */
  function startAutoScroll(containerRef: HTMLElement, onScroll: () => void) {
    let frameId: number | null = null

    function scroll() {
      if (!autoScroll.value.isScrolling || !containerRef) {
        if (frameId !== null) {
          cancelAnimationFrame(frameId)
          frameId = null
        }
        return
      }

      const scrollEl = containerRef
      const maxScrollLeft = Math.max(0, scrollEl.scrollWidth - scrollEl.clientWidth)
      const maxScrollTop = Math.max(0, scrollEl.scrollHeight - scrollEl.clientHeight)

      let newScrollLeft = scrollEl.scrollLeft + autoScroll.value.direction.x
      let newScrollTop = scrollEl.scrollTop + autoScroll.value.direction.y

      newScrollLeft = Math.max(0, Math.min(newScrollLeft, maxScrollLeft))
      newScrollTop = Math.max(0, Math.min(newScrollTop, maxScrollTop))

      const actualDx = newScrollLeft - scrollEl.scrollLeft
      const actualDy = newScrollTop - scrollEl.scrollTop

      scrollEl.scrollLeft = newScrollLeft
      scrollEl.scrollTop = newScrollTop

      if (actualDx !== 0 || actualDy !== 0) {
        dragSelection.value.currentX += actualDx
        dragSelection.value.currentY += actualDy

        const maxX = scrollEl.clientWidth + maxScrollLeft
        const maxY = scrollEl.clientHeight + maxScrollTop
        dragSelection.value.currentX = Math.max(0, Math.min(dragSelection.value.currentX, maxX))
        dragSelection.value.currentY = Math.max(0, Math.min(dragSelection.value.currentY, maxY))

        onScroll()
      }

      frameId = requestAnimationFrame(scroll)
    }

    frameId = requestAnimationFrame(scroll)
  }

  /**
   * 完成拖拽选择
   */
  function finishDragSelection() {
    dragSelection.value.isSelecting = false
    dragSelection.value.justFinished = true
    autoScroll.value.isScrolling = false

    setTimeout(() => {
      dragSelection.value.justFinished = false
    }, 10)
  }

  /**
   * 取消拖拽选择
   */
  function cancelDragSelection() {
    dragSelection.value.isSelecting = false
    autoScroll.value.isScrolling = false
  }

  /**
   * 检测元素是否与选择框相交
   */
  function isIntersecting(element: HTMLElement, containerRef: HTMLElement): boolean {
    const rect = element.getBoundingClientRect()
    const containerRect = containerRef.getBoundingClientRect()

    const scrollLeft = containerRef.scrollLeft
    const scrollTop = containerRef.scrollTop

    const elemLeft = rect.left - containerRect.left + scrollLeft
    const elemTop = rect.top - containerRect.top + scrollTop
    const elemRight = elemLeft + rect.width
    const elemBottom = elemTop + rect.height

    const selLeft = Math.min(dragSelection.value.startX, dragSelection.value.currentX)
    const selTop = Math.min(dragSelection.value.startY, dragSelection.value.currentY)
    const selRight = Math.max(dragSelection.value.startX, dragSelection.value.currentX)
    const selBottom = Math.max(dragSelection.value.startY, dragSelection.value.currentY)

    return !(elemRight < selLeft || elemLeft > selRight || elemBottom < selTop || elemTop > selBottom)
  }

  return {
    dragSelection,
    autoScroll,
    selectionBox,
    selectionBoxStyle,
    startDragSelection,
    updateDragSelection,
    checkAutoScroll,
    startAutoScroll,
    finishDragSelection,
    cancelDragSelection,
    isIntersecting
  }
}