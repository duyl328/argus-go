<template>
  <n-button
    quaternary
    circle
    size="medium"
    @click="toggleTheme"
    :title="isDark ? '切换到浅色主题' : '切换到深色主题'"
    class="theme-toggle"
  >
    <template #icon>
      <n-icon size="18">
        <SunnyOutline v-if="isDark" />
        <MoonOutline v-else />
      </n-icon>
    </template>
  </n-button>
</template>

<script lang="ts" setup>
import { computed } from 'vue'
import { NButton, NIcon } from 'naive-ui'
import { SunnyOutline, MoonOutline } from '@vicons/ionicons5'
import { useThemeStore } from '@/stores/theme'

const themeStore = useThemeStore()

const isDark = computed(() => themeStore.isDark)

const toggleTheme = () => {
  themeStore.toggleTheme()
}
</script>

<style scoped>
.theme-toggle {
  transition: all 0.3s ease;
}

.theme-toggle:hover {
  transform: scale(1.1);
}
</style>