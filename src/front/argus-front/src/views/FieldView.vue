<template>
  <div class="field-view">
    <FileManager ref="fileManagerRef" :use-real-api="true" />
  </div>
</template>

<script setup lang="ts">
import FileManager from '@/components/FileManager/FileManager.vue'
import { ref, onMounted, onUnmounted } from 'vue'

const fileManagerRef = ref<InstanceType<typeof FileManager>>()

// 使用全局事件监听来阻止浏览器的前进/后退
// 使用 mouseup 事件，确保只在按键抬起时执行一次
function handleGlobalMouseUp(event: MouseEvent) {
  // 检查是否在 field-view 区域内
  const target = event.target as HTMLElement
  if (!target.closest('.field-view')) {
    return
  }

  // 鼠标侧键：button 3 = 后退, button 4 = 前进
  if (event.button === 3 || event.button === 4) {
    event.preventDefault()
    event.stopPropagation()

    // 找到鼠标所在的面板并激活它
    const filePane = target.closest('.file-pane') as HTMLElement
    if (filePane) {
      filePane.click()

      // 确保面板激活后再执行导航
      setTimeout(() => {
        if (event.button === 3) {
          fileManagerRef.value?.goBack()
        } else if (event.button === 4) {
          fileManagerRef.value?.goForward()
        }
      }, 0)
    } else {
      // 如果不在面板内，直接对当前激活的面板执行操作
      if (event.button === 3) {
        fileManagerRef.value?.goBack()
      } else if (event.button === 4) {
        fileManagerRef.value?.goForward()
      }
    }
  }
}

// 在 mousedown 阶段也需要阻止默认行为，但不执行导航
function handleGlobalMouseDown(event: MouseEvent) {
  const target = event.target as HTMLElement
  if (!target.closest('.field-view')) {
    return
  }

  // 鼠标侧键：只阻止默认行为，不执行导航
  if (event.button === 3 || event.button === 4) {
    event.preventDefault()
    event.stopPropagation()
  }
}

onMounted(() => {
  // 在捕获阶段监听，确保最早拦截
  document.addEventListener('mousedown', handleGlobalMouseDown, true)
  document.addEventListener('mouseup', handleGlobalMouseUp, true)
})

onUnmounted(() => {
  document.removeEventListener('mousedown', handleGlobalMouseDown, true)
  document.removeEventListener('mouseup', handleGlobalMouseUp, true)
})
</script>

<style scoped>
.field-view {
  width: 100%;
  height: 100%;
  overflow: hidden;
}
</style>