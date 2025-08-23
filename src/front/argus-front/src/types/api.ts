// API相关类型定义

// 时间线API响应
export interface TimelineItem {
  date: string      // YYYY-MM-DD格式
  count: number     // 该日期的照片数量
}

export interface TimelineResponse {
  code: number
  message: string
  data: TimelineItem[]
}

// 照片列表API响应（列式存储格式）
export interface PhotosData {
  hash: string[]        // 照片hash数组
  isImage: boolean[]    // 是否为图片数组
  takenAt: string[]     // 拍摄时间数组 ISO格式
  ratio: number[]       // 宽高比数组
}

export interface PhotosResponse {
  code: number
  message: string
  data: PhotosData
}

// 照片详情响应
export interface PhotoDetails {
  // 这里需要根据实际API响应结构补充
  // 目前从API文档看只有获取照片文件的接口，没有详情结构
}

// API请求参数
export interface TimelineParams {
  start_date?: string   // YYYY-MM-DD格式
  end_date?: string     // YYYY-MM-DD格式
}

export interface PhotosParams {
  limit?: number        // 分页大小，最大1000
  offset?: number       // 偏移量
  start_date?: string   // YYYY-MM-DD格式
  end_date?: string     // YYYY-MM-DD格式
}

export interface PhotoRequestParams {
  format?: 'original' | 'thumbnail'
  size?: string         // 格式：WIDTHxHEIGHT，如 400x400
}

// 内部使用的照片数据结构
export interface Photo {
  hash: string
  isImage: boolean
  takenAt: string
  ratio: number
  // 为justified-layout计算用的尺寸信息
  width: number
  height: number
}

// 时间线组织数据结构
export interface DayGroup {
  date: string          // YYYY-MM-DD
  photos: Photo[]
  count: number
}

export interface MonthGroup {
  year: number
  month: number
  title: string         // 如："2024年1月"
  subtitle: string      // 如："128张照片"
  days: DayGroup[]
  totalCount: number
}