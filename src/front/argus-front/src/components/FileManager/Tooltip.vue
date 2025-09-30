<template>
  <Teleport to="body">
    <div
      v-if="visible && content"
      class="tooltip"
      :style="{ left: `${position.x}px`, top: `${position.y}px` }"
    >
      {{ content }}
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const visible = ref(false)
const content = ref('')
const position = ref({ x: 0, y: 0 })
const timeout = ref<number>()

function show(text: string, x: number, y: number, delay = 500) {
  clearTimeout(timeout.value)

  timeout.value = window.setTimeout(() => {
    content.value = text
    position.value = { x: x + 10, y: y + 10 }
    visible.value = true
  }, delay)
}

function hide() {
  clearTimeout(timeout.value)
  visible.value = false
}

defineExpose({
  show,
  hide
})
</script>

<style scoped>
.tooltip {
  position: fixed;
  background: rgba(0, 0, 0, 0.85);
  color: white;
  padding: 6px 10px;
  border-radius: 6px;
  font-size: 12px;
  pointer-events: none;
  z-index: 100;
  white-space: nowrap;
  max-width: 300px;
  word-wrap: break-word;
  white-space: normal;
}
</style>