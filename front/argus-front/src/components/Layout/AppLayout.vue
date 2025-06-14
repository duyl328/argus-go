<template>
  <n-layout class="layout-container">
    <!-- 头部 -->
    <n-layout-header bordered class="layout-header">
      <div class="header-content">
        <!-- 左侧 Logo 和标题 -->
        <div class="header-left">
          <n-icon size="32" class="logo-icon">
            <img src="@/assets/logo.svg" alt="logo" />
          </n-icon>
          <span class="site-title">标题</span>
        </div>

        <!-- 中间搜索框 -->
        <div class="header-center">
          <n-input placeholder="搜索..." class="search-input" clearable>
            <template #prefix>
              <n-icon>
                <svg viewBox="0 0 24 24">
                  <path
                    fill="currentColor"
                    d="M15.5 14h-.79l-.28-.27C15.41 12.59 16 11.11 16 9.5 16 5.91 13.09 3 9.5 3S3 5.91 3 9.5 5.91 16 9.5 16c1.61 0 3.09-.59 4.23-1.57l.27.28v.79l5 4.99L20.49 19l-4.99-5zm-6 0C7.01 14 5 11.99 5 9.5S7.01 5 9.5 5 14 7.01 14 9.5 11.99 14 9.5 14z"
                  />
                </svg>
              </n-icon>
            </template>
          </n-input>
        </div>

        <!-- 右侧预留空间（可放用户信息等） -->
        <div class="header-right">
          <!-- 可以添加用户头像、设置按钮等 -->
        </div>
      </div>
    </n-layout-header>

    <!-- 主体布局 -->
    <n-layout has-sider class="layout-body">
      <!-- 左侧导航栏 -->
      <n-layout-sider
        bordered
        show-trigger
        collapse-mode="width"
        :collapsed-width="64"
        :width="240"
        :native-scrollbar="false"
        class="layout-sider"
      >
        <n-menu
          v-model:value="activeKey"
          :collapsed-width="64"
          :collapsed-icon-size="22"
          :options="menuOptionsWithIcons"
          :root-indent="24"
          :indent="24"
          @update:value="handleMenuSelect"
        />
      </n-layout-sider>

      <!-- 主内容区域 -->
      <n-layout-content class="layout-content">
        <div class="content-wrapper">
          <router-view />
        </div>
      </n-layout-content>
    </n-layout>

    <!-- 底部（暂时隐藏） -->
    <n-layout-footer v-if="false" bordered> Footer Footer Footer</n-layout-footer>
  </n-layout>
</template>

<script lang="ts" setup>
import { computed, h } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { NMenu, NIcon } from 'naive-ui'
import {
  ImageOutline,
  SearchOutline,
  PeopleOutline,
  HeartOutline,
  LocationOutline,
  CalendarOutline,
  PricetagOutline,
  LibraryOutline,
  SettingsOutline,
} from '@vicons/ionicons5'
import { menuOptions, type MenuOption, type MenuItem, type MenuGroup } from '@/types/menu'

const router = useRouter()
const route = useRoute()

// 图标映射
const iconMap = {
  'image-outline': ImageOutline,
  'search-outline': SearchOutline,
  'people-outline': PeopleOutline,
  'heart-outline': HeartOutline,
  'location-outline': LocationOutline,
  'calendar-outline': CalendarOutline,
  'pricetag-outline': PricetagOutline,
  'library-outline': LibraryOutline,
  'settings-outline': SettingsOutline,
}

// 渲染图标
const renderIcon = (iconName: string) => {
  const IconComponent = iconMap[iconName as keyof typeof iconMap]
  return IconComponent ? () => h(NIcon, null, { default: () => h(IconComponent) }) : undefined
}

// 转换菜单选项，添加图标渲染函数
const menuOptionsWithIcons = computed(() => {
  return menuOptions.map((option: MenuOption) => {
    if ('children' in option) {
      // 分组菜单
      const group: MenuGroup = option
      return {
        key: group.key,
        label: group.label,
        type: 'group',
        children: group.children.map((child: MenuItem) => ({
          key: child.key,
          label: child.label,
          icon: child.icon ? renderIcon(child.icon) : undefined,
        })),
      }
    } else {
      // 普通菜单项
      const item: MenuItem = option
      return {
        key: item.key,
        label: item.label,
        icon: item.icon ? renderIcon(item.icon) : undefined,
      }
    }
  })
})

// 当前激活的菜单项
const activeKey = computed(() => {
  const currentPath = route.path
  // 从菜单选项中找到匹配当前路径的key
  for (const option of menuOptions) {
    if ('children' in option) {
      for (const child of option.children) {
        if (child.path === currentPath) {
          return child.key
        }
      }
    } else {
      if (option.path === currentPath) {
        return option.key
      }
    }
  }
  return ''
})

// 处理菜单选择
const handleMenuSelect = (key: string) => {
  // 根据key找到对应的路径
  let targetPath = ''

  for (const option of menuOptions) {
    if ('children' in option) {
      for (const child of option.children) {
        if (child.key === key) {
          targetPath = child.path
          break
        }
      }
    } else {
      if (option.key === key) {
        targetPath = option.path
        break
      }
    }
    if (targetPath) break
  }

  if (targetPath && targetPath !== route.path) {
    router.push(targetPath)
  }
}
</script>
<style>
:root {
  --header-height: 64px;
}
</style>
<style scoped>
.layout-container {
  height: 100vh;
  overflow: hidden;
}

.layout-header {
  height: var(--header-height);
  padding: 0;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.08);
  z-index: 10;
}

.header-content {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  max-width: 100%;
}

.header-left {
  display: flex;
  align-items: center;
  flex-shrink: 0;
  min-width: 200px;
}

.logo-icon {
  margin-right: 12px;
}

.logo-icon img {
  width: 32px;
  height: 32px;
  object-fit: contain;
}

.site-title {
  font-size: 20px;
  font-weight: 600;
  color: #333;
  white-space: nowrap;
}

.header-center {
  flex: 1;
  display: flex;
  justify-content: center;
  padding: 0 24px;
  max-width: 600px;
}

.search-input {
  width: 100%;
  max-width: 400px;
}

.header-right {
  flex-shrink: 0;
  min-width: 120px;
  display: flex;
  justify-content: flex-end;
  align-items: center;
}

.layout-body {
  height: calc(100vh - var(--header-height));
  overflow: hidden;
}

.layout-sider {
  height: 100%;
  box-shadow: 2px 0 8px rgba(0, 0, 0, 0.04);
}

.sidebar-menu {
  height: 100%;
  padding: 8px 0;
}

.layout-content {
  height: 100%;
  overflow: auto;
  flex: 1;
  background-color: #f5f5f5;
}

.content-wrapper {
  padding: 24px;
  height: 100%;
  min-height: calc(100vh - 64px);
  box-sizing: border-box;
}

.content-wrapper h1 {
  margin: 0 0 16px 0;
  color: #333;
}

.content-wrapper p {
  color: #666;
  line-height: 1.6;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .header-content {
    padding: 0 16px;
  }

  .header-center {
    padding: 0 16px;
  }

  .content-wrapper {
    padding: 16px;
  }

  .site-title {
    display: none;
  }

  .header-left {
    min-width: auto;
  }
}

@media (max-width: 480px) {
  .search-input {
    max-width: 200px;
  }
}
</style>
