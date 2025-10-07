import { ref, Ref } from 'vue'
import { httpClient } from '@/utils/http'

interface ThumbnailTask {
  path: string
  size: number
  index: number
}

interface ThumbnailResult {
  index: number
  path: string
  url: string
  loaded: boolean
  error?: string
}

/**
 * 批量缩略图加载Hook（并发控制）
 *
 * 策略：使用并发队列，同时请求多个缩略图，控制并发数避免过载
 */
export function useBatchThumbnails() {
  const maxConcurrency = ref(6) // 最大并发数
  const activeRequests = ref(0)
  const pendingQueue: ThumbnailTask[] = []
  const results = ref<Map<string, ThumbnailResult>>(new Map())

  /**
   * 批量加载缩略图
   * @param paths 文件路径列表
   * @param size 缩略图尺寸
   * @param onProgress 进度回调 (current, total)
   * @param onItemLoaded 单个项目加载完成回调
   */
  async function loadBatchThumbnails(
    paths: string[],
    size: number = 512,
    onProgress?: (current: number, total: number) => void,
    onItemLoaded?: (result: ThumbnailResult) => void
  ): Promise<Map<string, ThumbnailResult>> {
    if (!paths || paths.length === 0) {
      return new Map()
    }

    // 清空之前的结果
    results.value.clear()

    // 创建任务队列
    const tasks: ThumbnailTask[] = paths.map((path, index) => ({
      path,
      size,
      index
    }))

    let completedCount = 0
    const totalCount = tasks.length

    // 执行并发队列
    const promises: Promise<void>[] = []

    for (let i = 0; i < Math.min(maxConcurrency.value, tasks.length); i++) {
      promises.push(processQueue(tasks, totalCount, (result) => {
        completedCount++

        // 存储结果
        results.value.set(result.path, result)

        // 触发回调
        if (onItemLoaded) {
          onItemLoaded(result)
        }

        if (onProgress) {
          onProgress(completedCount, totalCount)
        }
      }))
    }

    await Promise.all(promises)

    return results.value
  }

  /**
   * 处理任务队列（工作线程）
   */
  async function processQueue(
    queue: ThumbnailTask[],
    total: number,
    onComplete: (result: ThumbnailResult) => void
  ): Promise<void> {
    while (queue.length > 0) {
      const task = queue.shift()
      if (!task) break

      activeRequests.value++

      try {
        const result = await loadSingleThumbnail(task)
        onComplete(result)
      } catch (error) {
        const errorResult: ThumbnailResult = {
          index: task.index,
          path: task.path,
          url: '',
          loaded: false,
          error: error instanceof Error ? error.message : '加载失败'
        }
        onComplete(errorResult)
      } finally {
        activeRequests.value--
      }
    }
  }

  /**
   * 加载单个缩略图
   */
  async function loadSingleThumbnail(task: ThumbnailTask): Promise<ThumbnailResult> {
    const baseURL = httpClient.getAxiosInstance().defaults.baseURL || ''
    const encodedPath = encodeURIComponent(task.path)
    const url = `${baseURL}/v1/photo/preview?path=${encodedPath}&size=${task.size}`

    // 使用fetch获取图片
    const response = await fetch(url)

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}`)
    }

    // 转换为Blob
    const blob = await response.blob()

    // 创建Blob URL
    const blobUrl = URL.createObjectURL(blob)

    return {
      index: task.index,
      path: task.path,
      url: blobUrl,
      loaded: true
    }
  }

  /**
   * 取消所有pending的请求
   */
  function cancelAll() {
    pendingQueue.length = 0
  }

  /**
   * 清理Blob URLs
   */
  function cleanup() {
    results.value.forEach(result => {
      if (result.url && result.url.startsWith('blob:')) {
        URL.revokeObjectURL(result.url)
      }
    })
    results.value.clear()
  }

  /**
   * 设置最大并发数
   */
  function setMaxConcurrency(value: number) {
    maxConcurrency.value = Math.max(1, Math.min(value, 20))
  }

  return {
    loadBatchThumbnails,
    cancelAll,
    cleanup,
    setMaxConcurrency,
    results: results as Ref<Map<string, ThumbnailResult>>,
    activeRequests: activeRequests as Ref<number>
  }
}
