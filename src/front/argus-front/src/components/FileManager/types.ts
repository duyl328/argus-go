// 文件管理器类型定义

export interface FileItem {
  name: string
  type: 'folder' | 'photo' | 'video' | 'file'
  size?: string
  date?: string
  children?: Record<string, FileItem>
}

export interface FolderStructure {
  [key: string]: FileItem
}

export type ViewMode = 'grid' | 'list'
export type ThumbnailSize = 'small' | 'medium' | 'large'
export type LayoutMode = 'single' | 'horizontal' | 'vertical'
export type PaneId = 'left' | 'right'

// 排序和筛选相关类型
export type SortField = 'name' | 'extension' | 'date' | 'size' | 'type'
export type SortOrder = 'asc' | 'desc'

export interface SortOptions {
  field: SortField
  order: SortOrder
}

export interface FilterOptions {
  nameQuery: string
  fileType: 'all' | 'photo' | 'video' | 'folder' | 'file'
}

export interface SelectionState {
  selectedItems: Set<string>
  focusedItem: { name: string; index: number } | null
  anchorItem: { name: string; index: number } | null
}

export interface ContextMenuPosition {
  x: number
  y: number
}

export interface ContextMenuItem {
  label: string
  icon: string
  action: () => void
  disabled?: boolean
  separator?: boolean
}