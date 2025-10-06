import { httpClient } from '@/utils/http'

export interface PreviewConfig {
  filePath: string
  fileSize?: number // 文件大小（字节）
}

export interface PreviewUrls {
  thumbnail: string   // 初始缩略图 (720p)
  highRes: string     // 高清版本 (2K)
  original: string    // 原图
}

class PhotoPreviewService {
  private baseUrl = '/v1/photo/preview'

  /**
   * 根据文件大小智能选择预览策略
   * @param filePath 文件路径
   * @param fileSize 文件大小（字节）
   * @returns 预览 URL
   */
  getPreviewUrl(filePath: string, fileSize?: number): string {
    const encodedPath = encodeURIComponent(filePath)

    if (!fileSize) {
      // 未知文件大小，默认加载 720p 缩略图
      return `${httpClient.getAxiosInstance().defaults.baseURL}${this.baseUrl}?path=${encodedPath}&size=720`
    }

    const MB = 1024 * 1024

    if (fileSize < 5 * MB) {
      // 小于 5MB：直接加载原图
      return `${httpClient.getAxiosInstance().defaults.baseURL}${this.baseUrl}?path=${encodedPath}&size=0`
    } else if (fileSize < 20 * MB) {
      // 5-20MB：加载 1080p 缩略图
      return `${httpClient.getAxiosInstance().defaults.baseURL}${this.baseUrl}?path=${encodedPath}&size=1080`
    } else {
      // 超过 20MB：加载 720p 缩略图
      return `${httpClient.getAxiosInstance().defaults.baseURL}${this.baseUrl}?path=${encodedPath}&size=720`
    }
  }

  /**
   * 获取所有级别的预览 URL
   * @param filePath 文件路径
   * @returns 包含不同分辨率的 URL 对象
   */
  getAllPreviewUrls(filePath: string): PreviewUrls {
    const encodedPath = encodeURIComponent(filePath)
    const baseURL = httpClient.getAxiosInstance().defaults.baseURL

    return {
      thumbnail: `${baseURL}${this.baseUrl}?path=${encodedPath}&size=720`,
      highRes: `${baseURL}${this.baseUrl}?path=${encodedPath}&size=2048`,
      original: `${baseURL}${this.baseUrl}?path=${encodedPath}&size=0`
    }
  }

  /**
   * 获取高清版本 URL（用户放大时加载）
   * @param filePath 文件路径
   * @returns 高清 URL
   */
  getHighResUrl(filePath: string): string {
    const encodedPath = encodeURIComponent(filePath)
    return `${httpClient.getAxiosInstance().defaults.baseURL}${this.baseUrl}?path=${encodedPath}&size=2048`
  }

  /**
   * 获取原图 URL（用户明确需要时加载）
   * @param filePath 文件路径
   * @returns 原图 URL
   */
  getOriginalUrl(filePath: string): string {
    const encodedPath = encodeURIComponent(filePath)
    return `${httpClient.getAxiosInstance().defaults.baseURL}${this.baseUrl}?path=${encodedPath}&size=0`
  }

  /**
   * 获取自定义尺寸的预览 URL
   * @param filePath 文件路径
   * @param size 最长边尺寸（像素）
   * @returns 预览 URL
   */
  getCustomSizeUrl(filePath: string, size: number): string {
    const encodedPath = encodeURIComponent(filePath)
    return `${httpClient.getAxiosInstance().defaults.baseURL}${this.baseUrl}?path=${encodedPath}&size=${size}`
  }

  /**
   * 预加载图片（用于提前缓存）
   * @param url 图片 URL
   * @returns Promise
   */
  preloadImage(url: string): Promise<void> {
    return new Promise((resolve, reject) => {
      const img = new Image()
      img.onload = () => resolve()
      img.onerror = () => reject(new Error(`Failed to preload image: ${url}`))
      img.src = url
    })
  }

  /**
   * 批量预加载图片
   * @param urls 图片 URL 数组
   * @returns Promise
   */
  async preloadImages(urls: string[]): Promise<void> {
    await Promise.all(urls.map(url => this.preloadImage(url)))
  }
}

export const photoPreviewService = new PhotoPreviewService()
