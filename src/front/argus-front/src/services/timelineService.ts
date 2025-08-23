import { httpClient } from '@/utils/http'
import type {
  TimelineParams,
  TimelineResponse,
  PhotosParams,
  PhotosResponse,
  Photo,
  DayGroup,
  MonthGroup, PhotosData
} from '@/types/api'
import type { ApiResponse } from '@/types/http.ts'

/**
 * 获取照片时间线统计数据
 */
export async function getTimeline(params?: TimelineParams): Promise<ApiResponse<TimelineResponse>> {
  const response = await httpClient.get<TimelineResponse>('/v1/photos/timeline', params)
  return response
}

/**
 * 获取照片列表数据
 */
export async function getPhotos(params?: PhotosParams): Promise<ApiResponse<PhotosData>> {
  const response = await httpClient.get<PhotosData>('/v1/photos', params)
  return response
}

/**
 * 根据时间线数据获取具体某天的照片
 */
export async function getPhotosByDate(date: string, limit = 50, offset = 0): Promise<Photo[]> {
  const response:ApiResponse<PhotosData> = await getPhotos({
    start_date: date,
    end_date: date,
    limit,
    offset
  })

  if (response.code !== 200 || !response.data) {
    return []
  }
  const { hash, isImage, takenAt, ratio } = response.data

  // 转换列式存储数据为行式存储
  const photos: Photo[] = []
  const count = Math.min(hash.length, isImage.length, takenAt.length, ratio.length)

  for (let i = 0; i < count; i++) {
    // 为justified-layout计算实际尺寸
    const photoRatio = ratio[i] || 1
    const baseHeight = 200
    const width = Math.round(baseHeight * photoRatio)
    const height = baseHeight

    photos.push({
      hash: hash[i],
      isImage: isImage[i],
      takenAt: takenAt[i],
      ratio: photoRatio,
      width,
      height
    })
  }

  return photos
}

/**
 * 获取照片缩略图URL
 */
export function getThumbnailUrl(hash: string, size = '400x400'): string {
  const baseUrl = import.meta.env.VITE_APP_API_URL
  return `${baseUrl}/v1/photo/${hash}?format=thumbnail&size=${size}`
}

/**
 * 获取照片原图URL
 */
export function getOriginalUrl(hash: string): string {
  const baseUrl = import.meta.env.VITE_APP_API_URL
  return `${baseUrl}/v1/photo/${hash}?format=original`
}

/**
 * 将时间线数据转换为月份分组数据
 */
export function groupTimelineByMonth(timelineData: { date: string; count: number }[]): MonthGroup[] {
  const monthMap = new Map<string, MonthGroup>()

  for (const item of timelineData) {
    const date = new Date(item.date)
    const year = date.getFullYear()
    const month = date.getMonth() + 1
    const monthKey = `${year}-${month.toString().padStart(2, '0')}`

    if (!monthMap.has(monthKey)) {
      monthMap.set(monthKey, {
        year,
        month,
        title: `${year}年${month}月`,
        subtitle: '0张照片',
        days: [],
        totalCount: 0
      })
    }

    const monthGroup = monthMap.get(monthKey)!
    monthGroup.totalCount += item.count

    // 添加日期组
    monthGroup.days.push({
      date: item.date,
      photos: [], // 这里暂时为空，需要时再加载
      count: item.count
    })
  }

  // 更新subtitle并排序
  const result = Array.from(monthMap.values())
    .map(month => ({
      ...month,
      subtitle: `${month.totalCount}张照片`,
      days: month.days.sort((a, b) => b.date.localeCompare(a.date)) // 按日期倒序
    }))
    .sort((a, b) => {
      // 按年月倒序
      if (a.year !== b.year) return b.year - a.year
      return b.month - a.month
    })

  return result
}

/**
 * 获取完整的时间线数据（包含照片）
 */
export async function getFullTimeline(params?: TimelineParams): Promise<MonthGroup[]> {
  console.log('正在获取时间线数据，参数:', params)

  try {
    const timelineResponse = await getTimeline(params)
    console.log('时间线API响应:', timelineResponse)

    // 检查响应结构
    if (!timelineResponse) {
      throw new Error('API响应为空')
    }

    if (timelineResponse.code !== 200) {
      throw new Error(timelineResponse.message || '获取时间线数据失败')
    }

    if (!Array.isArray(timelineResponse.data)) {
      throw new Error('时间线数据格式错误')
    }

    const monthGroups = groupTimelineByMonth(timelineResponse.data)
    console.log('处理后的月份分组:', monthGroups)

    // 为每个日期加载照片数据（限制加载数量，避免一次性加载太多）
    for (const monthGroup of monthGroups) {
      for (const dayGroup of monthGroup.days) {
        try {
          // 每天最多加载30张照片用于展示
          const photos = await getPhotosByDate(dayGroup.date, 30, 0)
          dayGroup.photos = photos
        } catch (error) {
          console.error(`Failed to load photos for ${dayGroup.date}:`, error)
          dayGroup.photos = []
        }
      }
    }

    return monthGroups
  } catch (error) {
    console.error('获取时间线数据失败:', error)
    throw error
  }
}
