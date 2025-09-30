import { ref, computed } from 'vue'
import type { SelectionState } from '@/components/FileManager/types'

export function useFileSelection() {
  const selectedItems = ref<Set<string>>(new Set())
  const focusedItem = ref<{ name: string; index: number } | null>(null)
  const anchorItem = ref<{ name: string; index: number } | null>(null)

  const selectionCount = computed(() => selectedItems.value.size)
  const hasSelection = computed(() => selectedItems.value.size > 0)

  function selectItem(itemName: string, index: number) {
    selectedItems.value.add(itemName)
    focusedItem.value = { name: itemName, index }
    anchorItem.value = { name: itemName, index }
  }

  function deselectItem(itemName: string) {
    selectedItems.value.delete(itemName)
  }

  function toggleItemSelection(itemName: string, index: number) {
    if (selectedItems.value.has(itemName)) {
      deselectItem(itemName)
    } else {
      selectedItems.value.add(itemName)
    }
    focusedItem.value = { name: itemName, index }
    anchorItem.value = { name: itemName, index }
  }

  function clearSelection() {
    selectedItems.value.clear()
  }

  function selectRange(startIndex: number, endIndex: number, items: Array<{ name: string }>) {
    const start = Math.min(startIndex, endIndex)
    const end = Math.max(startIndex, endIndex)

    for (let i = start; i <= end; i++) {
      if (items[i]) {
        selectedItems.value.add(items[i].name)
      }
    }
  }

  function selectAll(items: Array<{ name: string }>) {
    clearSelection()
    items.forEach(item => {
      selectedItems.value.add(item.name)
    })
  }

  function isSelected(itemName: string): boolean {
    return selectedItems.value.has(itemName)
  }

  function setFocusedItem(itemName: string, index: number) {
    focusedItem.value = { name: itemName, index }
  }

  return {
    selectedItems,
    focusedItem,
    anchorItem,
    selectionCount,
    hasSelection,
    selectItem,
    deselectItem,
    toggleItemSelection,
    clearSelection,
    selectRange,
    selectAll,
    isSelected,
    setFocusedItem
  }
}