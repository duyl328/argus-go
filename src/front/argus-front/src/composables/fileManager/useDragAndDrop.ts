import { ref } from 'vue'

/**
 * 拖放功能
 * 处理文件/文件夹的拖放操作
 */
export function useDragAndDrop() {
  // 拖放状态
  const dragState = ref({
    isDragging: false,
    draggedItems: [] as string[],
    dropTarget: null as string | null,
    sourcePane: null as string | null,
    isPaneDragOver: false
  })

  /**
   * 开始拖拽
   */
  function startDrag(items: string[], paneId: string) {
    dragState.value.isDragging = true
    dragState.value.draggedItems = items
    dragState.value.sourcePane = paneId
  }

  /**
   * 设置放置目标
   */
  function setDropTarget(targetName: string | null) {
    dragState.value.dropTarget = targetName
  }

  /**
   * 设置面板拖放状态
   */
  function setPaneDragOver(isOver: boolean) {
    dragState.value.isPaneDragOver = isOver
  }

  /**
   * 创建自定义拖拽预览
   */
  function createDragPreview(itemCount: number, firstItemName: string): HTMLElement {
    const preview = document.createElement('div')
    preview.style.cssText = `
      position: absolute;
      top: -9999px;
      padding: 12px 16px;
      background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
      color: white;
      border-radius: 8px;
      font-size: 14px;
      font-weight: 500;
      box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
      pointer-events: none;
      white-space: nowrap;
      display: flex;
      align-items: center;
      gap: 8px;
    `

    const icon = document.createElement('span')
    icon.textContent = '📋'
    icon.style.fontSize = '20px'

    const text = document.createElement('span')
    if (itemCount === 1) {
      text.textContent = firstItemName
    } else {
      text.textContent = `${itemCount} 个项目`
    }

    preview.appendChild(icon)
    preview.appendChild(text)
    document.body.appendChild(preview)

    return preview
  }

  /**
   * 清理拖拽预览
   */
  function cleanupDragPreview(preview: HTMLElement | null) {
    if (preview && preview.parentNode) {
      preview.parentNode.removeChild(preview)
    }
  }

  /**
   * 结束拖拽
   */
  function endDrag() {
    dragState.value.isDragging = false
    dragState.value.dropTarget = null
    dragState.value.sourcePane = null
  }

  /**
   * 重置所有拖拽状态
   */
  function resetDragState() {
    dragState.value.isDragging = false
    dragState.value.draggedItems = []
    dragState.value.dropTarget = null
    dragState.value.sourcePane = null
    dragState.value.isPaneDragOver = false
  }

  return {
    dragState,
    startDrag,
    setDropTarget,
    setPaneDragOver,
    createDragPreview,
    cleanupDragPreview,
    endDrag,
    resetDragState
  }
}