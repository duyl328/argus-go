import { onMounted, onUnmounted, type Ref } from 'vue'
import type { FileItem, ViewMode } from '@/components/FileManager/types'

interface UseKeyboardNavOptions {
  items: Ref<FileItem[]>
  viewMode: Ref<ViewMode>
  isActive: Ref<boolean>
  focusedItem: Ref<{ name: string; index: number } | null>
  anchorItem: Ref<{ name: string; index: number } | null>
  selectedItems: Ref<Set<string>>
  onNavigate: (newIndex: number, event: KeyboardEvent) => void
  onEnter: () => void
  onSpace: () => void
  onSelectAll: () => void
  getGridColumns: () => number
}

export function useKeyboardNav(options: UseKeyboardNavOptions) {
  const {
    items,
    viewMode,
    isActive,
    focusedItem,
    anchorItem,
    selectedItems,
    onNavigate,
    onEnter,
    onSpace,
    onSelectAll,
    getGridColumns
  } = options

  function handleKeyDown(event: KeyboardEvent) {
    // 只处理激活面板的键盘事件
    if (!isActive.value) return

    // 不处理输入框的事件
    if (event.target instanceof HTMLInputElement) return

    const itemList = items.value
    if (itemList.length === 0) return

    const focused = focusedItem.value
    if (!focused && itemList.length > 0) {
      // 如果没有焦点项，设置第一个
      onNavigate(0, event)
      return
    }

    if (!focused) return

    let newIndex = focused.index
    let shouldNavigate = false

    switch (event.key) {
      case 'ArrowUp':
        event.preventDefault()
        if (viewMode.value === 'list') {
          newIndex = Math.max(0, focused.index - 1)
        } else {
          // Grid视图 - 向上移动一行
          const cols = getGridColumns()
          newIndex = Math.max(0, focused.index - cols)
        }
        shouldNavigate = true
        break

      case 'ArrowDown':
        event.preventDefault()
        if (viewMode.value === 'list') {
          newIndex = Math.min(itemList.length - 1, focused.index + 1)
        } else {
          // Grid视图 - 向下移动一行
          const cols = getGridColumns()
          newIndex = Math.min(itemList.length - 1, focused.index + cols)
        }
        shouldNavigate = true
        break

      case 'ArrowLeft':
        if (viewMode.value === 'grid') {
          event.preventDefault()
          newIndex = Math.max(0, focused.index - 1)
          shouldNavigate = true
        }
        break

      case 'ArrowRight':
        if (viewMode.value === 'grid') {
          event.preventDefault()
          newIndex = Math.min(itemList.length - 1, focused.index + 1)
          shouldNavigate = true
        }
        break

      case 'Enter':
        event.preventDefault()
        onEnter()
        break

      case ' ':
        event.preventDefault()
        onSpace()
        break

      case 'a':
        if (event.ctrlKey || event.metaKey) {
          event.preventDefault()
          onSelectAll()
        }
        break

      case 'Escape':
        event.preventDefault()
        selectedItems.value.clear()
        break
    }

    if (shouldNavigate && newIndex !== focused.index) {
      onNavigate(newIndex, event)
    }
  }

  onMounted(() => {
    document.addEventListener('keydown', handleKeyDown)
  })

  onUnmounted(() => {
    document.removeEventListener('keydown', handleKeyDown)
  })

  return {
    handleKeyDown
  }
}