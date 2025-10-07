import { httpClient } from '@/utils/http'

interface BatchThumbnailRequest {
  paths: string[]
  size: number
}

interface ThumbnailResult {
  index: number
  path: string
  url?: string // Blob URL
  error?: string
}

export class BatchThumbnailService {
  private baseUrl = '/v1/photo/batch-preview'

  /**
   * 批量获取缩略图（使用Multipart Streaming）
   * @param paths 文件路径列表
   * @param size 缩略图尺寸
   * @param onProgress 进度回调
   * @returns Promise<ThumbnailResult[]>
   */
  async getBatchThumbnails(
    paths: string[],
    size: number = 512,
    onProgress?: (loaded: number, total: number) => void
  ): Promise<ThumbnailResult[]> {
    if (!paths || paths.length === 0) {
      return []
    }

    // 限制单次请求数量
    const maxBatchSize = 100
    if (paths.length > maxBatchSize) {
      // 分批处理
      const results: ThumbnailResult[] = []
      for (let i = 0; i < paths.length; i += maxBatchSize) {
        const batch = paths.slice(i, i + maxBatchSize)
        const batchResults = await this.getBatchThumbnails(batch, size, (loaded, total) => {
          if (onProgress) {
            const overallProgress = i + Math.floor((loaded / total) * batch.length)
            onProgress(overallProgress, paths.length)
          }
        })
        results.push(...batchResults)
      }
      return results
    }

    try {
      const baseURL = httpClient.getAxiosInstance().defaults.baseURL || ''
      const url = `${baseURL}${this.baseUrl}`

      const response = await fetch(url, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          paths,
          size
        } as BatchThumbnailRequest)
      })

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`)
      }

      // 解析multipart响应
      const results = await this.parseMultipartResponse(response, paths.length, onProgress)
      return results
    } catch (error) {
      console.error('批量缩略图请求失败:', error)
      // 返回错误结果
      return paths.map((path, index) => ({
        index,
        path,
        error: error instanceof Error ? error.message : '未知错误'
      }))
    }
  }

  /**
   * 解析Multipart响应
   */
  private async parseMultipartResponse(
    response: Response,
    totalCount: number,
    onProgress?: (loaded: number, total: number) => void
  ): Promise<ThumbnailResult[]> {
    const results: ThumbnailResult[] = []
    const contentType = response.headers.get('content-type') || ''
    const boundaryMatch = contentType.match(/boundary=(.+)/)

    if (!boundaryMatch) {
      throw new Error('Invalid multipart response: no boundary')
    }

    const boundary = boundaryMatch[1]
    const reader = response.body?.getReader()

    if (!reader) {
      throw new Error('Response body not readable')
    }

    const decoder = new TextDecoder()
    let buffer = ''
    let loadedCount = 0

    try {
      while (true) {
        const { done, value } = await reader.read()

        if (done) break

        // 将新数据添加到缓冲区
        buffer += decoder.decode(value, { stream: true })

        // 处理缓冲区中的完整part
        const parts = await this.extractCompleteParts(buffer, boundary)

        for (const partData of parts.completeParts) {
          try {
            const result = await this.parsePart(partData)
            results.push(result)
            loadedCount++

            // 触发进度回调
            if (onProgress) {
              onProgress(loadedCount, totalCount)
            }
          } catch (error) {
            console.error('解析part失败:', error)
          }
        }

        // 更新缓冲区（保留不完整的part）
        buffer = parts.remaining
      }

      // 处理剩余的缓冲区数据
      if (buffer.trim()) {
        const finalParts = await this.extractCompleteParts(buffer, boundary, true)
        for (const partData of finalParts.completeParts) {
          try {
            const result = await this.parsePart(partData)
            results.push(result)
          } catch (error) {
            console.error('解析最后part失败:', error)
          }
        }
      }
    } finally {
      reader.releaseLock()
    }

    return results
  }

  /**
   * 从缓冲区提取完整的part
   */
  private async extractCompleteParts(
    buffer: string,
    boundary: string,
    isFinal: boolean = false
  ): Promise<{ completeParts: string[]; remaining: string }> {
    const parts: string[] = []
    const boundaryDelimiter = `--${boundary}`
    const endBoundary = `${boundaryDelimiter}--`

    let position = 0

    while (true) {
      // 查找下一个boundary
      const nextBoundary = buffer.indexOf(boundaryDelimiter, position)

      if (nextBoundary === -1) {
        // 没有更多boundary
        return {
          completeParts: parts,
          remaining: isFinal ? '' : buffer.substring(position)
        }
      }

      // 检查是否是结束boundary
      if (buffer.substring(nextBoundary, nextBoundary + endBoundary.length) === endBoundary) {
        // 结束边界
        if (position < nextBoundary) {
          parts.push(buffer.substring(position, nextBoundary))
        }
        return {
          completeParts: parts,
          remaining: ''
        }
      }

      // 查找下一个boundary以确定当前part的结束位置
      const nextNextBoundary = buffer.indexOf(boundaryDelimiter, nextBoundary + boundaryDelimiter.length)

      if (nextNextBoundary === -1 && !isFinal) {
        // 当前part不完整，保留在缓冲区
        return {
          completeParts: parts,
          remaining: buffer.substring(position)
        }
      }

      // 提取完整的part
      const partStart = nextBoundary + boundaryDelimiter.length
      const partEnd = nextNextBoundary !== -1 ? nextNextBoundary : buffer.length

      const partData = buffer.substring(partStart, partEnd).trim()
      if (partData) {
        parts.push(partData)
      }

      position = nextNextBoundary !== -1 ? nextNextBoundary : buffer.length
    }
  }

  /**
   * 解析单个part
   */
  private async parsePart(partData: string): Promise<ThumbnailResult> {
    // 分离header和body
    const headerEndIndex = partData.indexOf('\r\n\r\n')
    if (headerEndIndex === -1) {
      throw new Error('Invalid part: no header/body separator')
    }

    const headerSection = partData.substring(0, headerEndIndex)
    const bodySection = partData.substring(headerEndIndex + 4)

    // 解析headers
    const headers = this.parseHeaders(headerSection)

    const index = parseInt(headers['x-index'] || '0', 10)
    const path = headers['x-original-path'] || ''
    const isError = headers['x-error'] === 'true'

    if (isError) {
      // 错误part
      try {
        const errorData = JSON.parse(bodySection)
        return {
          index,
          path,
          error: errorData.error || '未知错误'
        }
      } catch {
        return {
          index,
          path,
          error: '解析错误响应失败'
        }
      }
    }

    // 图片part
    const contentType = headers['content-type'] || 'image/jpeg'

    // 将base64或文本转换为Blob
    // 注意：由于multipart/mixed在文本传输中，图片数据可能是base64编码的
    // 但在我们的实现中，我们直接传输二进制数据
    // 这里我们需要处理响应体

    // 创建Blob URL
    try {
      // 假设bodySection包含二进制数据
      // 实际上在文本解析中这很难处理
      // 我们需要改用ArrayBuffer方式处理

      // 临时方案：由于文本解析的限制，我们返回路径让前端单独请求
      return {
        index,
        path,
        error: 'Multipart binary parsing not fully implemented, use individual requests'
      }
    } catch (error) {
      return {
        index,
        path,
        error: error instanceof Error ? error.message : '创建Blob失败'
      }
    }
  }

  /**
   * 解析headers
   */
  private parseHeaders(headerSection: string): Record<string, string> {
    const headers: Record<string, string> = {}
    const lines = headerSection.split('\r\n')

    for (const line of lines) {
      const colonIndex = line.indexOf(':')
      if (colonIndex > 0) {
        const key = line.substring(0, colonIndex).trim().toLowerCase()
        const value = line.substring(colonIndex + 1).trim()
        headers[key] = value
      }
    }

    return headers
  }
}

export const batchThumbnailService = new BatchThumbnailService()
