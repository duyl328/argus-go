/**
 * 文件类型配置
 * 定义照片管理系统支持的文件类型
 */

/** 支持的图片格式 */
export const IMAGE_EXTENSIONS = [
  // 常见格式
  'jpg',
  'jpeg',
  'png',
  'gif',
  'bmp',
  'webp',
  'svg',

  // RAW 格式
  'raw',
  'cr2',
  'cr3',
  'nef',
  'arw',
  'dng',
  'orf',
  'rw2',
  'pef',
  'srw',
  'raf',

  // 其他
  'tiff',
  'tif',
  'heic',
  'heif',
  'avif',
  'jxl'
] as const

/** 支持的视频格式 */
export const VIDEO_EXTENSIONS = [
  // 常见格式
  'mp4',
  'avi',
  'mov',
  'mkv',
  'wmv',
  'flv',
  'webm',
  'm4v',

  // 高清格式
  'mts',
  'm2ts',
  'ts',

  // 其他
  '3gp',
  'mpg',
  'mpeg',
  'vob',
  'ogv',
  'rm',
  'rmvb',
  'asf'
] as const

/** 文件类型 MIME 映射 */
export const MIME_TYPE_MAP = {
  // 图片
  'image/jpeg': 'photo',
  'image/png': 'photo',
  'image/gif': 'photo',
  'image/bmp': 'photo',
  'image/webp': 'photo',
  'image/svg+xml': 'photo',
  'image/tiff': 'photo',
  'image/heic': 'photo',
  'image/heif': 'photo',
  'image/avif': 'photo',

  // 视频
  'video/mp4': 'video',
  'video/avi': 'video',
  'video/quicktime': 'video',
  'video/x-matroska': 'video',
  'video/x-msvideo': 'video',
  'video/x-ms-wmv': 'video',
  'video/x-flv': 'video',
  'video/webm': 'video',
  'video/3gpp': 'video',
  'video/mpeg': 'video'
} as const

/** 文件类型 */
export type FileType = 'folder' | 'photo' | 'video' | 'file'

/**
 * 根据扩展名判断文件类型
 * @param extension 文件扩展名 (小写，不含点)
 * @returns 文件类型
 */
export function getFileTypeByExtension(extension: string): FileType {
  const ext = extension.toLowerCase()

  if (IMAGE_EXTENSIONS.includes(ext as any)) {
    return 'photo'
  }

  if (VIDEO_EXTENSIONS.includes(ext as any)) {
    return 'video'
  }

  return 'file'
}

/**
 * 根据 MIME 类型判断文件类型
 * @param mimeType MIME 类型
 * @returns 文件类型
 */
export function getFileTypeByMime(mimeType: string): FileType {
  const mime = mimeType.toLowerCase()

  // 精确匹配
  if (MIME_TYPE_MAP[mime as keyof typeof MIME_TYPE_MAP]) {
    return MIME_TYPE_MAP[mime as keyof typeof MIME_TYPE_MAP] as FileType
  }

  // 模糊匹配
  if (mime.startsWith('image/')) {
    return 'photo'
  }

  if (mime.startsWith('video/')) {
    return 'video'
  }

  return 'file'
}

/**
 * 判断是否为支持的媒体文件
 * @param extension 文件扩展名
 * @returns 是否支持
 */
export function isSupportedMediaFile(extension: string): boolean {
  const ext = extension.toLowerCase()
  return (
    IMAGE_EXTENSIONS.includes(ext as any) ||
    VIDEO_EXTENSIONS.includes(ext as any)
  )
}

/**
 * 获取文件扩展名
 * @param filename 文件名
 * @returns 扩展名 (小写，不含点)
 */
export function getExtension(filename: string): string {
  const lastDot = filename.lastIndexOf('.')
  if (lastDot === -1 || lastDot === 0) return ''
  return filename.substring(lastDot + 1).toLowerCase()
}

/**
 * 获取文件图标/显示
 * @param type 文件类型
 * @param extension 文件扩展名 (用于未支持文件)
 * @returns 图标字符串或扩展名
 */
export function getFileIcon(type: FileType, extension?: string): string {
  switch (type) {
    case 'folder':
      return '📁'
    case 'photo':
      return '🖼️'
    case 'video':
      return '🎬'
    case 'file':
      // 不支持的文件显示扩展名
      return extension ? extension.toUpperCase() : '📄'
    default:
      return '📄'
  }
}

/**
 * 获取文件类型的描述
 * @param type 文件类型
 * @returns 描述文本
 */
export function getFileTypeDescription(type: FileType): string {
  switch (type) {
    case 'folder':
      return '文件夹'
    case 'photo':
      return '图片'
    case 'video':
      return '视频'
    case 'file':
      return '文件'
    default:
      return '未知'
  }
}
