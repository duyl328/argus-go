import { ref, computed, onMounted, onUnmounted } from 'vue'

export function useLayout() {
  // 侧边栏是否展开
  const sidebarCollapsed = ref(false)

  // 移动端断点
  const MOBILE_BREAKPOINT = 768

  // 当前屏幕宽度
  const screenWidth = ref(window.innerWidth)

  // 是否是移动端
  const isMobile = computed(() => screenWidth.value < MOBILE_BREAKPOINT)

  // 移动端菜单抽屉是否显示
  const showMobileMenu = ref(false)

  // 侧边栏宽度
  const sidebarWidth = computed(() => {
    if (isMobile.value) return 0
    return sidebarCollapsed.value ? 64 : 240
  })

  // 监听窗口大小变化
  const handleResize = () => {
    screenWidth.value = window.innerWidth
    // 如果从移动端切换到桌面端，关闭移动端菜单
    if (!isMobile.value) {
      showMobileMenu.value = false
    }
  }

  // 切换侧边栏展开/折叠
  const toggleSidebar = () => {
    if (isMobile.value) {
      showMobileMenu.value = !showMobileMenu.value
    } else {
      sidebarCollapsed.value = !sidebarCollapsed.value
    }
  }

  // 关闭移动端菜单
  const closeMobileMenu = () => {
    showMobileMenu.value = false
  }

  onMounted(() => {
    window.addEventListener('resize', handleResize)
  })

  onUnmounted(() => {
    window.removeEventListener('resize', handleResize)
  })

  return {
    sidebarCollapsed,
    isMobile,
    showMobileMenu,
    sidebarWidth,
    toggleSidebar,
    closeMobileMenu
  }
}
