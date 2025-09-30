<template>
  <Teleport to="body">
    <div
      v-if="visible"
      ref="menuRef"
      class="context-menu"
      :style="{ left: `${position.x}px`, top: `${position.y}px` }"
      @click="handleMenuClick"
    >
      <template v-for="(item, index) in menuItems" :key="index">
        <div v-if="item.separator" class="menu-separator"></div>
        <div
          v-else
          :class="['menu-item', { disabled: item.disabled }]"
          @click="handleItemClick(item)"
        >
          <span class="menu-icon">{{ item.icon }}</span>
          <span class="menu-label">{{ item.label }}</span>
          <span v-if="item.shortcut" class="menu-shortcut">{{ item.shortcut }}</span>
        </div>
      </template>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'

interface ContextMenuItem {
  label: string
  icon: string
  action: () => void
  disabled?: boolean
  separator?: boolean
  shortcut?: string
}

interface Props {
  visible: boolean
  x: number
  y: number
  targetItem?: string | null
  selectedCount?: number
  paneId: string
}

const props = defineProps<Props>()
const emit = defineEmits<{
  close: []
  action: [action: string, params?: any]
}>()

const menuRef = ref<HTMLElement>()
const position = ref({ x: 0, y: 0 })

// 根据上下文生成菜单项
const menuItems = computed((): ContextMenuItem[] => {
  if (props.targetItem) {
    // 右键点击文件/文件夹
    return [
      { label: '打开', icon: '📂', action: () => emit('action', 'open', props.targetItem) },
      { separator: true } as any,
      {
        label: `剪切${props.selectedCount && props.selectedCount > 1 ? ` (${props.selectedCount} 项)` : ''}`,
        icon: '✂️',
        action: () => emit('action', 'cut'),
        shortcut: 'Ctrl+X'
      },
      {
        label: `复制${props.selectedCount && props.selectedCount > 1 ? ` (${props.selectedCount} 项)` : ''}`,
        icon: '📋',
        action: () => emit('action', 'copy'),
        shortcut: 'Ctrl+C'
      },
      { separator: true } as any,
      {
        label: `删除${props.selectedCount && props.selectedCount > 1 ? ` (${props.selectedCount} 项)` : ''}`,
        icon: '🗑️',
        action: () => emit('action', 'delete'),
        shortcut: 'Delete'
      },
      { label: '重命名', icon: '✏️', action: () => emit('action', 'rename', props.targetItem), shortcut: 'F2' },
      { separator: true } as any,
      { label: '属性', icon: 'ℹ️', action: () => emit('action', 'properties', props.targetItem) }
    ]
  } else {
    // 右键点击空白区域
    return [
      { label: '粘贴', icon: '📋', action: () => emit('action', 'paste'), shortcut: 'Ctrl+V', disabled: true },
      { separator: true } as any,
      { label: '新建文件夹', icon: '📁', action: () => emit('action', 'newFolder') },
      { separator: true } as any,
      { label: '全选', icon: '☑️', action: () => emit('action', 'selectAll'), shortcut: 'Ctrl+A' },
      { label: '刷新', icon: '🔄', action: () => emit('action', 'refresh'), shortcut: 'F5' }
    ]
  }
})

// 调整菜单位置，防止超出屏幕
function adjustPosition() {
  nextTick(() => {
    if (!menuRef.value) return

    const menuWidth = menuRef.value.offsetWidth || 200
    const menuHeight = menuRef.value.offsetHeight || 300
    let x = props.x
    let y = props.y

    // 防止超出右边界
    if (x + menuWidth > window.innerWidth) {
      x = window.innerWidth - menuWidth - 10
    }

    // 防止超出下边界
    if (y + menuHeight > window.innerHeight) {
      y = window.innerHeight - menuHeight - 10
    }

    position.value = { x, y }
  })
}

watch(() => props.visible, (visible) => {
  if (visible) {
    adjustPosition()
    // 添加全局点击监听，点击外部关闭菜单
    setTimeout(() => {
      document.addEventListener('click', handleClickOutside)
      document.addEventListener('contextmenu', handleClickOutside)
    }, 0)
  } else {
    document.removeEventListener('click', handleClickOutside)
    document.removeEventListener('contextmenu', handleClickOutside)
  }
})

function handleClickOutside() {
  emit('close')
}

function handleMenuClick(event: MouseEvent) {
  event.stopPropagation()
}

function handleItemClick(item: ContextMenuItem) {
  if (item.disabled) return
  item.action()
  emit('close')
}
</script>

<style scoped>
.context-menu {
  position: fixed;
  background: white;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  box-shadow: 0 10px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04);
  min-width: 200px;
  padding: 4px;
  z-index: 1000;
  user-select: none;
}

.menu-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border-radius: 4px;
  cursor: pointer;
  transition: background-color 0.15s;
  font-size: 14px;
  color: #374151;
}

.menu-item:hover {
  background-color: #f3f4f6;
}

.menu-item.disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.menu-item.disabled:hover {
  background-color: transparent;
}

.menu-icon {
  font-size: 16px;
  width: 20px;
  text-align: center;
}

.menu-label {
  flex: 1;
}

.menu-shortcut {
  font-size: 12px;
  color: #9ca3af;
}

.menu-separator {
  height: 1px;
  background-color: #e5e7eb;
  margin: 4px 0;
}
</style>