<template>
  <div class="file-manager">
    <!-- Toolbar -->
    <div class="toolbar">
      <div class="toolbar-left">
        <button class="btn btn-primary">
          <span class="icon">+</span>
          <span>新建</span>
        </button>
        <div class="divider"></div>
        <button class="btn">复制</button>
        <button class="btn">剪切</button>
        <button class="btn">粘贴</button>
      </div>
      <div class="toolbar-right">
        <!-- View Controls -->
        <div class="view-controls">
          <button
            :class="['btn-view', { active: viewMode === 'grid' }]"
            @click="viewMode = 'grid'"
          >
            网格
          </button>
          <button
            :class="['btn-view', { active: viewMode === 'list' }]"
            @click="viewMode = 'list'"
          >
            列表
          </button>
        </div>

        <!-- Thumbnail Size -->
        <div v-if="viewMode === 'grid'" class="thumbnail-controls">
          <span class="label">大小:</span>
          <button
            v-for="size in (['small', 'medium', 'large'] as ThumbnailSize[])"
            :key="size"
            :class="['btn-size', { active: thumbnailSize === size }]"
            @click="thumbnailSize = size"
          >
            {{ size === 'small' ? '小' : size === 'medium' ? '中' : '大' }}
          </button>
        </div>

        <!-- Layout Controls -->
        <div class="layout-controls">
          <button
            :class="['btn-layout', { active: layoutMode === 'single' }]"
            @click="layoutMode = 'single'"
            title="单面板"
          >
            ▢
          </button>
          <button
            :class="['btn-layout', { active: layoutMode === 'horizontal' }]"
            @click="layoutMode = 'horizontal'"
            title="左右双面板"
          >
            ▭
          </button>
          <button
            :class="['btn-layout', { active: layoutMode === 'vertical' }]"
            @click="layoutMode = 'vertical'"
            title="上下双面板"
          >
            ▯
          </button>
        </div>
      </div>
    </div>

    <!-- Main Content -->
    <div :class="['main-content', `layout-${layoutMode}`]">
      <FilePane
        pane-id="left"
        :view-mode="viewMode"
        :thumbnail-size="thumbnailSize"
        :is-active="activePane === 'left'"
        @activate="activePane = 'left'"
      />

      <div
        v-if="layoutMode === 'horizontal' || layoutMode === 'vertical'"
        :class="['splitter', layoutMode === 'vertical' ? 'splitter-vertical' : '']"
        @mousedown="handleSplitterMouseDown"
      ></div>

      <FilePane
        v-if="layoutMode === 'horizontal' || layoutMode === 'vertical'"
        pane-id="right"
        :view-mode="viewMode"
        :thumbnail-size="thumbnailSize"
        :is-active="activePane === 'right'"
        @activate="activePane = 'right'"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import FilePane from './FilePane.vue'
import type { ViewMode, ThumbnailSize, LayoutMode, PaneId } from './types'

const viewMode = ref<ViewMode>('grid')
const thumbnailSize = ref<ThumbnailSize>('medium')
const layoutMode = ref<LayoutMode>('single')
const activePane = ref<PaneId>('left')

// 分隔器拖拽
const isDraggingSplitter = ref(false)

function handleSplitterMouseDown(event: MouseEvent) {
  isDraggingSplitter.value = true
  const isVertical = layoutMode.value === 'vertical'
  document.body.style.cursor = isVertical ? 'row-resize' : 'col-resize'
  document.body.style.userSelect = 'none'

  const mainContent = document.querySelector('.main-content') as HTMLElement
  if (!mainContent) return

  const onMouseMove = (e: MouseEvent) => {
    if (!isDraggingSplitter.value) return

    const rect = mainContent.getBoundingClientRect()

    if (isVertical) {
      const y = e.clientY - rect.top
      const percentage = (y / rect.height) * 100

      if (percentage > 20 && percentage < 80) {
        mainContent.style.gridTemplateRows = `${percentage}% 4px 1fr`
      }
    } else {
      const x = e.clientX - rect.left
      const percentage = (x / rect.width) * 100

      if (percentage > 20 && percentage < 80) {
        mainContent.style.gridTemplateColumns = `${percentage}% 4px 1fr`
      }
    }
  }

  const onMouseUp = () => {
    isDraggingSplitter.value = false
    document.body.style.cursor = ''
    document.body.style.userSelect = ''
    document.removeEventListener('mousemove', onMouseMove)
    document.removeEventListener('mouseup', onMouseUp)
  }

  document.addEventListener('mousemove', onMouseMove)
  document.addEventListener('mouseup', onMouseUp)
}
</script>

<style scoped>
.file-manager {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: #f9fafb;
}

/* Toolbar */
.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background: white;
  border-bottom: 1px solid #e5e7eb;
}

.toolbar-left, .toolbar-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.btn {
  padding: 8px 12px;
  background: #f3f4f6;
  border: none;
  border-radius: 6px;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s;
}

.btn:hover {
  background: #e5e7eb;
}

.btn-primary {
  background: #3b82f6;
  color: white;
}

.btn-primary:hover {
  background: #2563eb;
}

.divider {
  width: 1px;
  height: 24px;
  background: #d1d5db;
}

.view-controls, .thumbnail-controls, .layout-controls {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px;
  background: #f3f4f6;
  border-radius: 6px;
}

.btn-view, .btn-size, .btn-layout {
  padding: 6px 12px;
  background: transparent;
  border: none;
  border-radius: 4px;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.15s;
}

.btn-view.active, .btn-size.active, .btn-layout.active {
  background: white;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
}

.label {
  font-size: 13px;
  color: #6b7280;
  margin-right: 4px;
}

/* Main Content */
.main-content {
  flex: 1;
  display: flex;
  overflow: hidden;
}

.main-content.layout-single {
  display: block;
}

.main-content.layout-horizontal {
  display: grid;
  grid-template-columns: 1fr 4px 1fr;
}

.main-content.layout-vertical {
  display: grid;
  grid-template-rows: 1fr 4px 1fr;
}

.splitter {
  background: #d1d5db;
  cursor: col-resize;
}

.splitter.splitter-vertical {
  cursor: row-resize;
}

.splitter:hover {
  background: #9ca3af;
}
</style>