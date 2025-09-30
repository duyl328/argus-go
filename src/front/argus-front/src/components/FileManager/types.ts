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