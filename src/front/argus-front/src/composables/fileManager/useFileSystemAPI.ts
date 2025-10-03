/**
 * 文件系统 API 集成 Composable
 * 用于在 FileManager 中集成后端文件系统 API
 */
import { ref, computed } from 'vue'
import { fileSystemService } from '@/services/fileSystemService'
import type {
  FileSystemItem,
  BrowseResult,
  OperationResult
} from '@/services/fileSystemService'
import type { FileItem } from '@/components/FileManager/types'
import { getFileTypeByMime, getFileTypeByExtension, getExtension } from '@/config/fileTypes'

export function useFileSystemAPI() {
  // 状态
  const currentPath = ref<string>('')  // 当前路径
  const items = ref<FileSystemItem[]>([])  // 当前目录下的项目
  const loading = ref(false)
  const error = ref<string | null>(null)
  const parentPath = ref<string>('')  // 父路径

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

    try {
      const result: BrowseResult = await fileSystemService.browse(path)

      currentPath.value = result.current_path
      parentPath.value = result.parent_path

      // 转换项目列表
      items.value = result.items

      return result
    } catch (err: any) {
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

      // 刷新当前目录
      await browse(currentPath.value)

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

      // 刷新当前目录
      await browse(currentPath.value)

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

      // 刷新当前目录
      await browse(currentPath.value)

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

      // 刷新当前目录
      await browse(currentPath.value)

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
    return items.value.map(convertToFileItem)
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
    convertToFileItem
  }
}
