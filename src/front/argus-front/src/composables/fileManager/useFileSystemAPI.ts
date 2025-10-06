/**
 * 文件系统 API 集成 Composable
 * 用于在 FileManager 中集成后端文件系统 API
 */
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { fileSystemService } from '@/services/fileSystemService'
import { sseService } from '@/services/sseService'
import type {
  FileSystemItem,
  BrowseResult,
  OperationResult
} from '@/services/fileSystemService'
import type { FileItem } from '@/components/FileManager/types'
import { getFileTypeByMime, getFileTypeByExtension, getExtension } from '@/config/fileTypes'
import type { FileSystemChangeEvent } from '@/services/sseService'

// 防抖工具函数
function debounce<T extends (...args: any[]) => any>(
  func: T,
  wait: number
): (...args: Parameters<T>) => void {
  let timeout: ReturnType<typeof setTimeout> | null = null
  return function(this: any, ...args: Parameters<T>) {
    if (timeout) clearTimeout(timeout)
    timeout = setTimeout(() => {
      func.apply(this, args)
      timeout = null
    }, wait)
  }
}

export function useFileSystemAPI() {
  // 状态
  const currentPath = ref<string>('')  // 当前路径
  const items = ref<FileSystemItem[]>([])  // 当前目录下的项目
  const loading = ref(false)
  const error = ref<string | null>(null)
  const parentPath = ref<string>('')  // 父路径
  const previousWatchedPath = ref<string>('')  // 上一个监听的路径

  /**
   * 将后端 FileSystemItem 转换为前端 FileItem 格式
   */
  function convertToFileItem(item: FileSystemItem): FileItem {
    // 判断文件类型
    let itemType: FileItem['type'] = 'file'

    if (item.type === 'directory') {
      itemType = 'folder'
    } else if (item.type === 'drive') {
      itemType = 'folder'  // 驱动器视为特殊文件夹
    } else if (item.file_info) {
      // 优先根据 MIME 类型判断
      if (item.file_info.mime_type) {
        itemType = getFileTypeByMime(item.file_info.mime_type)
      }

      // 如果 MIME 类型无法识别，尝试使用扩展名
      if (itemType === 'file' && item.file_info.extension) {
        const ext = item.file_info.extension.replace('.', '')
        itemType = getFileTypeByExtension(ext)
      }
    } else {
      // 没有 file_info，尝试从文件名提取扩展名
      const ext = getExtension(item.name)
      if (ext) {
        itemType = getFileTypeByExtension(ext)
      }
    }

    // 格式化文件大小
    let sizeStr = ''
    if (item.size > 0) {
      const kb = item.size / 1024
      const mb = kb / 1024
      const gb = mb / 1024

      if (gb >= 1) {
        sizeStr = `${gb.toFixed(2)} GB`
      } else if (mb >= 1) {
        sizeStr = `${mb.toFixed(2)} MB`
      } else if (kb >= 1) {
        sizeStr = `${kb.toFixed(2)} KB`
      } else {
        sizeStr = `${item.size} B`
      }
    }

    // 格式化修改时间
    let dateStr = ''
    if (item.mod_time) {
      try {
        const date = new Date(item.mod_time)
        dateStr = date.toLocaleString('zh-CN', {
          year: 'numeric',
          month: '2-digit',
          day: '2-digit',
          hour: '2-digit',
          minute: '2-digit'
        })
      } catch (e) {
        dateStr = item.mod_time
      }
    }

    // 获取扩展名 (用于不支持的文件显示)
    let extension = ''
    if (item.file_info?.extension) {
      extension = item.file_info.extension.replace('.', '').toLowerCase()
    } else {
      extension = getExtension(item.name)
    }

    return {
      name: item.name,
      type: itemType,
      size: sizeStr,
      date: dateStr,
      path: item.path,  // 保存完整路径
      extension: extension || undefined,
      children: undefined  // 后端按需加载，不预加载子项
    }
  }

  /**
   * 浏览指定路径
   */
  async function browse(path?: string) {
    loading.value = true
    error.value = null

    console.log('🔄 [useFileSystemAPI.browse] 开始刷新:', path)

    // 先取消之前路径的监听
    if (currentPath.value && currentPath.value !== path) {
      try {
        await unwatchCurrentPath()
      } catch (err) {
        console.warn('⚠️ [useFileSystemAPI.browse] 取消监听失败:', err)
      }
    }

    try {
      const result: BrowseResult = await fileSystemService.browse(path)

      console.log('✅ [useFileSystemAPI.browse] API 返回:', {
        current_path: result.current_path,
        parent_path: result.parent_path,
        items_count: result.items.length,
        items_preview: result.items.slice(0, 5).map(i => i.name)
      })

      currentPath.value = result.current_path
      parentPath.value = result.parent_path

      // 转换项目列表 - 强制触发响应式更新
      items.value = [...result.items]

      console.log('✅ [useFileSystemAPI.browse] items.value 已更新:', items.value.length, '个项目')

      // 订阅新路径的监听
      console.log('🎯 [useFileSystemAPI.browse] 准备调用 watchCurrentPath()')
      try {
        await watchCurrentPath()
        console.log('✅ [useFileSystemAPI.browse] watchCurrentPath() 调用完成')
      } catch (watchErr) {
        console.error('❌ [useFileSystemAPI.browse] 订阅监听失败:', watchErr)
      }

      return result
    } catch (err: any) {
      console.error('❌ [useFileSystemAPI.browse] 刷新失败:', err)
      error.value = err.message || '浏览文件系统失败'
      items.value = []
      throw err
    } finally {
      loading.value = false
    }
  }

  /**
   * 创建目录
   */
  async function createDirectory(path: string): Promise<OperationResult> {
    loading.value = true
    error.value = null

    try {
      const result = await fileSystemService.createDirectory(path)

      console.log('✅ [useFileSystemAPI.createDirectory] 创建成功,延迟刷新依赖 SSE')

      // 不立即刷新,依赖 SSE 自动刷新 (避免双重刷新)
      // await browse(currentPath.value)

      return result
    } catch (err: any) {
      error.value = err.message || '创建目录失败'
      throw err
    } finally {
      loading.value = false
    }
  }

  /**
   * 删除文件或目录
   */
  async function deleteItem(path: string, recursive: boolean = true): Promise<OperationResult> {
    loading.value = true
    error.value = null

    try {
      const result = await fileSystemService.deleteItem(path, undefined, recursive)

      console.log('✅ [useFileSystemAPI.deleteItem] 删除成功,延迟刷新依赖 SSE')

      // 不立即刷新,依赖 SSE 自动刷新 (避免双重刷新)
      // await browse(currentPath.value)

      return result
    } catch (err: any) {
      error.value = err.message || '删除失败'
      throw err
    } finally {
      loading.value = false
    }
  }

  /**
   * 移动/重命名文件或目录
   */
  async function moveItem(
    source: string,
    destination: string,
    overwrite: boolean = false
  ): Promise<OperationResult> {
    loading.value = true
    error.value = null

    try {
      const result = await fileSystemService.moveItem(source, destination, undefined, overwrite)

      console.log('✅ [useFileSystemAPI.moveItem] 移动成功,延迟刷新依赖 SSE')

      // 不立即刷新,依赖 SSE 自动刷新 (避免双重刷新)
      // await browse(currentPath.value)

      return result
    } catch (err: any) {
      error.value = err.message || '移动失败'
      throw err
    } finally {
      loading.value = false
    }
  }

  /**
   * 复制文件或目录
   */
  async function copyItem(
    source: string,
    destination: string,
    overwrite: boolean = false
  ): Promise<OperationResult> {
    loading.value = true
    error.value = null

    try {
      const result = await fileSystemService.copyItem(source, destination, undefined, overwrite)

      console.log('✅ [useFileSystemAPI.copyItem] 复制成功,延迟刷新依赖 SSE')

      // 不立即刷新,依赖 SSE 自动刷新 (避免双重刷新)
      // await browse(currentPath.value)

      return result
    } catch (err: any) {
      error.value = err.message || '复制失败'
      throw err
    } finally {
      loading.value = false
    }
  }

  /**
   * 搜索文件
   */
  async function searchFiles(
    path: string,
    pattern: string = '*',
    type?: string,
    recursive: boolean = false
  ) {
    loading.value = true
    error.value = null

    try {
      const result = await fileSystemService.searchFiles(path, pattern, type, recursive)
      return result
    } catch (err: any) {
      error.value = err.message || '搜索失败'
      throw err
    } finally {
      loading.value = false
    }
  }

  /**
   * 转换后的 FileItem 列表（computed）
   */
  const fileItems = computed<FileItem[]>(() => {
    const result = items.value.map(convertToFileItem)
    console.log('🔄 [useFileSystemAPI.fileItems] computed 重新计算:', result.length, '个项目')
    return result
  })

  /**
   * 获取磁盘使用情况
   */
  async function getDiskUsage(path: string) {
    try {
      return await fileSystemService.getDiskUsage(path)
    } catch (err: any) {
      error.value = err.message || '获取磁盘使用情况失败'
      throw err
    }
  }

  /**
   * 订阅当前路径的文件系统监听（智能切换）
   */
  async function watchCurrentPath() {
    console.log('🔍 [watchCurrentPath] 调用', {
      currentPath: currentPath.value,
      previousWatchedPath: previousWatchedPath.value
    })

    // 如果当前路径为空或是根目录，不监听
    if (!currentPath.value || currentPath.value === '') {
      console.log('⏭️ [watchCurrentPath] 根目录无需监听')
      return
    }

    // 如果新路径和旧路径相同，不需要重新订阅
    if (currentPath.value === previousWatchedPath.value) {
      console.log('⏭️ [watchCurrentPath] 路径未变化，跳过订阅', currentPath.value)
      return
    }

    try {
      // 先取消旧路径的监听和 SSE 订阅
      if (previousWatchedPath.value) {
        console.log(`🚫 [watchCurrentPath] 取消旧路径监听和 SSE 订阅: ${previousWatchedPath.value}`)

        // 1. 取消 SSE 订阅
        await sseService.unsubscribe(previousWatchedPath.value)

        // 2. 取消文件系统监听
        await fileSystemService.unwatchPath(previousWatchedPath.value)

        console.log(`✅ [watchCurrentPath] 已停止监听: ${previousWatchedPath.value}`)
      }

      // 订阅新路径
      console.log(`👀 [watchCurrentPath] 开始订阅新路径: ${currentPath.value}`)

      // 1. 添加文件系统监听
      await fileSystemService.watchPath(currentPath.value)
      console.log(`✅ [watchCurrentPath] 文件系统监听已添加`)

      // 2. 添加 SSE 订阅
      const subscribeSuccess = await sseService.subscribe(currentPath.value)
      if (subscribeSuccess) {
        console.log(`✅ [watchCurrentPath] SSE 订阅成功: ${currentPath.value}`)
      } else {
        console.warn(`⚠️ [watchCurrentPath] SSE 订阅失败，但文件系统监听已添加`)
      }

      // 更新记录
      previousWatchedPath.value = currentPath.value
    } catch (err: any) {
      console.error('❌ [watchCurrentPath] 文件夹监听切换失败:', err)
    }
  }

  /**
   * 取消当前路径的文件系统监听
   */
  async function unwatchCurrentPath() {
    if (previousWatchedPath.value) {
      try {
        // 1. 取消 SSE 订阅
        await sseService.unsubscribe(previousWatchedPath.value)

        // 2. 取消文件系统监听
        await fileSystemService.unwatchPath(previousWatchedPath.value)

        console.log(`✅ [unwatchCurrentPath] 已停止监听: ${previousWatchedPath.value}`)
        previousWatchedPath.value = ''
      } catch (err: any) {
        console.error('❌ [unwatchCurrentPath] 取消文件夹监听失败:', err)
      }
    }
  }

  /**
   * 处理文件系统变化事件（带防抖）
   */
  const debouncedRefresh = debounce((path: string) => {
    browse(path).catch(err => {
      // 忽略取消错误
      if (err.name !== 'CanceledError') {
        console.error('自动刷新失败:', err)
      }
    })
  }, 300) // 300ms 防抖

  /**
   * 处理文件系统变化事件
   */
  function handleFileSystemChange(event: FileSystemChangeEvent) {
    console.log('📨 [useFileSystemAPI.handleFileSystemChange] 收到 SSE 事件:', {
      type: event.type,
      path: event.path,
      name: event.name,
      is_dir: event.is_dir,
      timestamp: event.timestamp
    })

    // 统一路径分隔符处理 (支持 Windows/Linux/macOS)
    const normalize = (p: string) => p.replace(/[\/\\]+$/, '').replace(/\\/g, '/')

    const eventPath = normalize(event.path)
    const eventDir = eventPath.substring(0, eventPath.lastIndexOf('/'))
    const currentDir = normalize(currentPath.value)

    console.log('🔍 [useFileSystemAPI.handleFileSystemChange] 路径匹配检查:', {
      eventDir,
      currentDir,
      match: eventDir === currentDir || eventPath.startsWith(currentDir + '/')
    })

    // 检查变化是否发生在当前路径
    if (eventDir === currentDir || eventPath.startsWith(currentDir + '/')) {
      console.log(`✅ [useFileSystemAPI.handleFileSystemChange] 当前文件夹内容发生变化 (${event.type}): ${event.name}`)

      // 使用防抖的自动刷新
      debouncedRefresh(currentPath.value)
    } else {
      console.log(`⏭️ [useFileSystemAPI.handleFileSystemChange] 事件不在当前路径,忽略`)
    }
  }

  // 组件挂载时初始化 SSE
  onMounted(async () => {
    try {
      // 连接 SSE
      if (!sseService.isConnected()) {
        await sseService.connect()
        console.log('SSE 连接已建立')
      }

      // 订阅文件系统变化事件
      sseService.onFileSystemChange(handleFileSystemChange)
    } catch (err) {
      console.error('SSE 连接失败:', err)
    }
  })

  // 组件卸载时清理
  onUnmounted(() => {
    // 取消监听当前路径
    unwatchCurrentPath()
  })

  return {
    // 状态
    currentPath,
    parentPath,
    items,
    fileItems,
    loading,
    error,

    // 方法
    browse,
    createDirectory,
    deleteItem,
    moveItem,
    copyItem,
    searchFiles,
    getDiskUsage,
    convertToFileItem,
    watchCurrentPath,
    unwatchCurrentPath
  }
}
