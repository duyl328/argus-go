/**
 * SSE 订阅管理使用示例
 *
 * 这个文件展示了如何使用新的订阅系统
 */

import { sseService } from '@/services/sseService'

export async function subscriptionExample() {
  // 1. 连接到 SSE 服务器
  console.log('正在连接 SSE 服务器...')
  await sseService.connect()

  // 等待客户端 ID 初始化（connected 事件）
  await new Promise(resolve => setTimeout(resolve, 1000))

  // 2. 获取客户端 ID
  const clientID = sseService.getClientID()
  console.log('客户端 ID:', clientID)

  // 3. 订阅文件夹路径
  const pathToWatch = 'D:\\EdgeDownload'
  const subscribed = await sseService.subscribe(pathToWatch)

  if (subscribed) {
    console.log(`✅ 已订阅: ${pathToWatch}`)

    // 4. 监听文件系统变化事件
    const unsubscribeCallback = sseService.onFileSystemChange((event) => {
      console.log('📁 文件系统变化:', event)
      console.log(`  类型: ${event.type}`)
      console.log(`  路径: ${event.path}`)
      console.log(`  名称: ${event.name}`)
      console.log(`  是否文件夹: ${event.is_dir}`)
    })

    // 5. 获取当前所有订阅
    const subscriptions = await sseService.getSubscriptions()
    console.log('当前订阅列表:', subscriptions)

    // 6. 取消订阅（例如在组件卸载时）
    // await sseService.unsubscribe(pathToWatch)
    // unsubscribeCallback()
  }
}

// Vue 组件中的使用示例
export const useFileSystemSubscription = () => {
  const currentPath = ref('D:\\EdgeDownload')
  const events = ref<any[]>([])

  onMounted(async () => {
    // 连接 SSE
    await sseService.connect()

    // 等待客户端 ID 初始化
    await new Promise(resolve => setTimeout(resolve, 1000))

    // 订阅当前路径
    await sseService.subscribe(currentPath.value)

    // 监听文件系统变化
    const unsubscribe = sseService.onFileSystemChange((event) => {
      events.value.unshift(event)
      // 保留最近 100 个事件
      if (events.value.length > 100) {
        events.value = events.value.slice(0, 100)
      }
    })

    // 组件卸载时清理
    onUnmounted(() => {
      sseService.unsubscribe(currentPath.value)
      unsubscribe()
    })
  })

  // 切换监听的文件夹
  const watchFolder = async (newPath: string) => {
    // 取消旧订阅
    await sseService.unsubscribe(currentPath.value)

    // 添加新订阅
    await sseService.subscribe(newPath)

    currentPath.value = newPath
    events.value = []
  }

  return {
    currentPath,
    events,
    watchFolder
  }
}
