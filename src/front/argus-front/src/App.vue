<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { NConfigProvider, NMessageProvider, NDialogProvider, NGlobalStyle, lightTheme, darkTheme } from 'naive-ui'
import AppLayout from '@/components/Layout/AppLayout.vue'
import { useThemeStore } from '@/stores/theme'

const themeStore = useThemeStore()

// Naive UI 主题跟随我们的主题状态
const naiveTheme = computed(() => {
  return themeStore.isDark ? darkTheme : lightTheme
})

// 初始化主题
onMounted(() => {
  themeStore.initTheme()
})
</script>

<template>
  <n-config-provider :theme="naiveTheme">
    <n-global-style />
    <n-message-provider>
      <n-dialog-provider>
        <div class="app-container">
          <AppLayout />
        </div>
      </n-dialog-provider>
    </n-message-provider>
  </n-config-provider>
</template>

<style scoped>
.app-container {
  min-height: 100vh;
  background-color: var(--bg-color);
  color: var(--text-color);
  transition: background-color 0.3s ease, color 0.3s ease;
}
</style>

