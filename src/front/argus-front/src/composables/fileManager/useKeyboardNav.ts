import { onMounted, onUnmounted, type Ref } from 'vue'
import type { FileItem, ViewMode } from '@/components/FileManager/types'

interface UseKeyboardNavOptions {
  items: Ref<FileItem[]>
  viewMode: Ref<ViewMode>
  isActive: Ref<boolean>
  isDialogOpen?: Ref<boolean>  // 是否有对话框打开
  focusedItem: Ref<{ name: string; index: number } | null>
  anchorItem: Ref<{ name: string; index: number } | null>
  selectedItems: Ref<Set<string>>
  onNavigate: (newIndex: number, event: KeyboardEvent) => void
  onEnter: () => void
  onSpace: (event: KeyboardEvent) => void
  onSelectAll: () => void
  onDelete?: () => void  // 新增：删除快捷键
  onCopy?: () => void    // 新增：复制快捷键 (Ctrl+C)
  onCut?: () => void     // 新增：剪切快捷键 (Ctrl+X)
  onPaste?: () => void   // 新增：粘贴快捷键 (Ctrl+V)
  onRefresh?: () => void // 新增：刷新快捷键 (F5)
  onBack?: () => void    // 新增：后退快捷键 (Backspace)
  getGridColumns: () => number
}

export function useKeyboardNav(options: UseKeyboardNavOptions) {
  const {
    items,
    viewMode,
    isActive,
    isDialogOpen,
    focusedItem,
    anchorItem,
    selectedItems,
    onNavigate,
    onEnter,
    onSpace,
    onSelectAll,
    onDelete,
    onCopy,
    onCut,
    onPaste,
    onRefresh,
    onBack,
    getGridColumns
  } = options

  function handleKeyDown(event: KeyboardEvent) {
    // 如果有对话框打开，禁用所有快捷键（除了对话框内部的）
    if (isDialogOpen?.value) return

    // 只处理激活面板的键盘事件
    if (!isActive.value) return

    // 不处理输入框的事件（除了 Escape）
    if (event.target instanceof HTMLInputElement && event.key !== 'Escape') return

    const itemList = items.value
    if (itemList.length === 0 && event.key !== 'F5' && event.key !== 'Backspace') return

    const focused = focusedItem.value
    if (!focused && itemList.length > 0) {
      // 如果没有焦点项，设置第一个
      onNavigate(0, event)
      return
    }

    // 对于某些全局快捷键（F5, Ctrl+C/X/V等），即使没有焦点也要处理
    const isGlobalShortcut =
      event.key === 'F5' ||
      event.key === 'Backspace' ||
      ((event.ctrlKey || event.metaKey) && ['c', 'C', 'x', 'X', 'v', 'V', 'a', 'A'].includes(event.key))

    if (!focused && !isGlobalShortcut) return

    let newIndex = focused?.index ?? 0
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
        event.preventDefault()  // 始终阻止默认行为
        if (viewMode.value === 'grid' && focused) {
          newIndex = Math.max(0, focused.index - 1)
          shouldNavigate = true
        }
        break

      case 'ArrowRight':
        event.preventDefault()  // 始终阻止默认行为
        if (viewMode.value === 'grid' && focused) {
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
        onSpace(event)
        break

      case 'a':
      case 'A':
        if (event.ctrlKey || event.metaKey) {
          event.preventDefault()
          onSelectAll()
        }
        break

      case 'c':
      case 'C':
        if ((event.ctrlKey || event.metaKey) && onCopy) {
          event.preventDefault()
          event.stopPropagation()  // 阻止事件冒泡
          onCopy()
        }
        break

      case 'x':
      case 'X':
        if ((event.ctrlKey || event.metaKey) && onCut) {
          event.preventDefault()
          event.stopPropagation()  // 阻止事件冒泡
          onCut()
        }
        break

      case 'v':
      case 'V':
        if ((event.ctrlKey || event.metaKey) && onPaste) {
          event.preventDefault()
          event.stopPropagation()  // 阻止事件冒泡
          onPaste()
        }
        break

      case 'Delete':
        if (onDelete) {
          event.preventDefault()
          event.stopPropagation()  // 阻止事件冒泡
          onDelete()
        }
        break

      case 'F5':
        // F5 刷新 - 始终阻止浏览器刷新
        event.preventDefault()
        event.stopPropagation()
        if (onRefresh) {
          onRefresh()
        }
        break

      case 'Backspace':
        // Backspace 后退 - 阻止浏览器后退行为
        if (onBack && selectedItems.value.size === 0) {
          event.preventDefault()
          event.stopPropagation()
          onBack()
        } else if (selectedItems.value.size === 0) {
          // 即使没有 onBack 回调，也要阻止浏览器后退
          event.preventDefault()
        }
        break

      case 'Escape':
        event.preventDefault()
        event.stopPropagation()
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