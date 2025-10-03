/**
 * 文件系统服务
 * 提供文件和文件夹的浏览、操作功能
 */
import { httpClient } from '@/utils/http'

// ==================== 类型定义 ====================

/** 文件系统项目类型 */
export type FileSystemItemType = 'drive' | 'directory' | 'file'

/** 驱动器信息 */
export interface DriveInfo {
  label: string           // 卷标
  file_system: string     // 文件系统类型
  total_space: number     // 总空间 (字节)
  free_space: number      // 可用空间
  used_space: number      // 已用空间
  usage_percent: number   // 使用百分比
  is_removable: boolean   // 是否可移动
  drive_type: string      // 驱动器类型
}

/** 文件信息 */
export interface FileInfo {
  extension: string       // 文件扩展名
  mime_type: string       // MIME类型
}

/** 目录信息 */
export interface DirectoryInfo {
  item_count: number      // 子项目数量
}

/** 文件系统项目 */
export interface FileSystemItem {
  id: string                        // 项目唯一标识
  name: string                      // 显示名称
  path: string                      // 完整路径
  type: FileSystemItemType          // 项目类型
  size: number                      // 大小 (字节)
  mod_time: string                  // 修改时间
  is_accessible: boolean            // 是否可访问
  drive_info?: DriveInfo            // 驱动器信息 (仅drive类型)
  file_info?: FileInfo              // 文件信息 (仅file类型)
  directory_info?: DirectoryInfo    // 目录信息 (仅directory类型)
}

/** 浏览结果摘要 */
export interface BrowseSummary {
  total_items: number       // 总项目数
  directory_count: number   // 目录数量
  file_count: number        // 文件数量
  drive_count: number       // 驱动器数量
}

/** 浏览结果 */
export interface BrowseResult {
  current_path: string      // 当前路径
  parent_path: string       // 父路径
  items: FileSystemItem[]   // 项目列表
  summary: BrowseSummary    // 摘要信息
}

/** 磁盘使用情况 */
export interface DiskUsage {
  path: string              // 路径
  total: number             // 总空间 (字节)
  free: number              // 可用空间
  used: number              // 已用空间
  usage_percent: number     // 使用百分比
}

/** 搜索结果 */
export interface SearchResult {
  items: FileSystemItem[]   // 匹配的项目列表
  total_count: number       // 总匹配数
}

/** 操作结果 */
export interface OperationResult {
  success: boolean          // 操作是否成功
  message: string           // 结果消息
  path?: string             // 相关路径
  operation_id?: string     // 操作ID
}

// ==================== API 服务 ====================

class FileSystemService {
  /**
   * 浏览文件系统
   * @param path 路径 (为空则返回根级别/所有硬盘)
   */
  async browse(path?: string): Promise<BrowseResult> {
    const response = await httpClient.get<BrowseResult>('/v1/filesystem/browse', {
      path: path || ''
    })
    return response.data
  }

  /**
   * 获取磁盘使用情况
   * @param path 路径
   */
  async getDiskUsage(path: string): Promise<DiskUsage> {
    const response = await httpClient.get<DiskUsage>('/v1/filesystem/disk-usage', {
      path
    })
    return response.data
  }

  /**
   * 获取文件系统项目信息
   * @param path 项目路径
   */
  async getItem(path: string): Promise<BrowseResult> {
    const response = await httpClient.get<BrowseResult>('/v1/filesystem/item', {
      path
    })
    return response.data
  }

  /**
   * 搜索文件
   * @param path 搜索路径
   * @param pattern 匹配模式 (例如: *.txt)
   * @param type 文件类型筛选
   * @param recursive 是否递归搜索
   */
  async searchFiles(
    path: string,
    pattern: string = '*',
    type?: string,
    recursive: boolean = false
  ): Promise<SearchResult> {
    const response = await httpClient.get<SearchResult>('/v1/filesystem/search', {
      path,
      pattern,
      type: type || '',
      recursive: recursive.toString()
    })
    return response.data
  }

  /**
   * 创建目录
   * @param path 目录路径
   */
  async createDirectory(path: string): Promise<OperationResult> {
    const response = await httpClient.post<OperationResult>('/v1/filesystem/directory', {
      path
    })
    return response.data
  }

  /**
   * 删除文件或目录
   * @param path 路径
   * @param operationId 操作ID (用于进度追踪)
   * @param recursive 是否递归删除 (默认true)
   */
  async deleteItem(
    path: string,
    operationId?: string,
    recursive: boolean = true
  ): Promise<OperationResult> {
    const response = await httpClient.delete<OperationResult>('/v1/filesystem/item', {
      path,
      operation_id: operationId || '',
      recursive: recursive.toString()
    })
    return response.data
  }

  /**
   * 移动/重命名文件或目录
   * @param source 源路径
   * @param destination 目标路径
   * @param operationId 操作ID
   * @param overwrite 是否覆盖已存在文件
   */
  async moveItem(
    source: string,
    destination: string,
    operationId?: string,
    overwrite: boolean = false
  ): Promise<OperationResult> {
    const response = await httpClient.put<OperationResult>('/v1/filesystem/item/move', {
      source,
      destination,
      operation_id: operationId || '',
      overwrite
    })
    return response.data
  }

  /**
   * 复制文件或目录
   * @param source 源路径
   * @param destination 目标路径
   * @param operationId 操作ID
   * @param overwrite 是否覆盖已存在文件
   */
  async copyItem(
    source: string,
    destination: string,
    operationId?: string,
    overwrite: boolean = false
  ): Promise<OperationResult> {
    const response = await httpClient.post<OperationResult>('/v1/filesystem/item/copy', {
      source,
      destination,
      operation_id: operationId || '',
      overwrite
    })
    return response.data
  }
}

// 导出单例实例
export const fileSystemService = new FileSystemService()

// 导出服务类 (用于测试)
export default FileSystemService
