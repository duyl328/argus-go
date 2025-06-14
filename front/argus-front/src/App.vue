<script setup lang="ts">
import { ref } from 'vue'
import { NConfigProvider, NMessageProvider, NGlobalStyle, lightTheme } from 'naive-ui'
import AppLayout from '@/components/Layout/AppLayout.vue'

const theme = ref(lightTheme)
</script>

<template>
  <n-config-provider :theme="theme">
    <n-global-style />
    <n-message-provider>
      <AppLayout />
    </n-message-provider>
  </n-config-provider>
</template>

