import type { FileItem, FolderStructure } from '@/components/FileManager/types'

/**
 * 文件管理器操作工具
 * 提供文件/文件夹的基本操作功能
 */

/**
 * 根据路径获取文件夹对象
 * @param structure 文件夹结构对象
 * @param path 路径数组
 * @returns 文件夹对象，如果路径无效则返回 null
 */
export function getFolderByPath(
  structure: FolderStructure,
  path: string[]
): Record<string, FileItem> | null {
  let folder: any = structure
  for (const segment of path) {
    if (folder[segment] && folder[segment].children) {
      folder = folder[segment].children
    } else {
      return null
    }
  }
  return folder
}

/**
 * 移动文件/文件夹
 * @param structure 文件夹结构对象（必须是响应式对象）
 * @param itemNames 要移动的项目名称数组
 * @param sourcePath 源路径
 * @param targetPath 目标路径
 * @returns 是否成功移动
 */
export function moveItems(
  structure: FolderStructure,
  itemNames: string[],
  sourcePath: string[],
  targetPath: string[]
): boolean {
  const sourceFolder = getFolderByPath(structure, sourcePath)
  const targetFolder = getFolderByPath(structure, targetPath)

  if (!sourceFolder || !targetFolder) {
    console.error('[文件操作] 移动失败：路径无效', {
      sourcePath,
      targetPath
    })
    return false
  }

  // 检查是否尝试将文件夹移动到自己内部
  if (isPathDescendant(sourcePath, targetPath)) {
    console.error('[文件操作] 移动失败：不能将文件夹移动到自己内部')
    return false
  }

  // 收集要移动的项目
  const itemsToMove: Record<string, FileItem> = {}
  for (const name of itemNames) {
    if (sourceFolder[name]) {
      itemsToMove[name] = sourceFolder[name]
    }
  }

  // 从源文件夹删除
  for (const name of itemNames) {
    if (sourceFolder[name]) {
      delete sourceFolder[name]
    }
  }

  // 添加到目标文件夹
  for (const [name, item] of Object.entries(itemsToMove)) {
    // 处理重名冲突
    let finalName = name
    let counter = 1
    while (targetFolder[finalName]) {
      const ext = name.lastIndexOf('.') > 0 ? name.substring(name.lastIndexOf('.')) : ''
      const base = ext ? name.substring(0, name.lastIndexOf('.')) : name
      finalName = `${base} (${counter})${ext}`
      counter++
    }
    targetFolder[finalName] = item
  }

  console.log(`[文件操作] 成功移动 ${itemNames.length} 个项目`, {
    from: sourcePath.join('/'),
    to: targetPath.join('/')
  })

  return true
}

/**
 * 复制文件/文件夹
 * @param structure 文件夹结构对象（必须是响应式对象）
 * @param itemNames 要复制的项目名称数组
 * @param sourcePath 源路径
 * @param targetPath 目标路径
 * @returns 是否成功复制
 */
export function copyItems(
  structure: FolderStructure,
  itemNames: string[],
  sourcePath: string[],
  targetPath: string[]
): boolean {
  const sourceFolder = getFolderByPath(structure, sourcePath)
  const targetFolder = getFolderByPath(structure, targetPath)

  if (!sourceFolder || !targetFolder) {
    console.error('[文件操作] 复制失败：路径无效')
    return false
  }

  // 深拷贝并添加到目标文件夹
  for (const name of itemNames) {
    if (sourceFolder[name]) {
      const item = JSON.parse(JSON.stringify(sourceFolder[name]))

      // 处理重名冲突
      let finalName = name
      let counter = 1
      while (targetFolder[finalName]) {
        const ext = name.lastIndexOf('.') > 0 ? name.substring(name.lastIndexOf('.')) : ''
        const base = ext ? name.substring(0, name.lastIndexOf('.')) : name
        finalName = `${base} (${counter})${ext}`
        counter++
      }

      targetFolder[finalName] = item
    }
  }

  console.log(`[文件操作] 成功复制 ${itemNames.length} 个项目`)
  return true
}

/**
 * 删除文件/文件夹
 * @param structure 文件夹结构对象（必须是响应式对象）
 * @param itemNames 要删除的项目名称数组
 * @param path 路径
 * @returns 是否成功删除
 */
export function deleteItems(
  structure: FolderStructure,
  itemNames: string[],
  path: string[]
): boolean {
  const folder = getFolderByPath(structure, path)

  if (!folder) {
    console.error('[文件操作] 删除失败：路径无效')
    return false
  }

  for (const name of itemNames) {
    if (folder[name]) {
      delete folder[name]
    }
  }

  console.log(`[文件操作] 成功删除 ${itemNames.length} 个项目`)
  return true
}

/**
 * 重命名文件/文件夹
 * @param structure 文件夹结构对象（必须是响应式对象）
 * @param oldName 旧名称
 * @param newName 新名称
 * @param path 路径
 * @returns 是否成功重命名
 */
export function renameItem(
  structure: FolderStructure,
  oldName: string,
  newName: string,
  path: string[]
): boolean {
  const folder = getFolderByPath(structure, path)

  if (!folder || !folder[oldName]) {
    console.error('[文件操作] 重命名失败：项目不存在')
    return false
  }

  if (folder[newName]) {
    console.error('[文件操作] 重命名失败：目标名称已存在')
    return false
  }

  folder[newName] = folder[oldName]
  folder[newName].name = newName
  delete folder[oldName]

  console.log(`[文件操作] 成功重命名: ${oldName} -> ${newName}`)
  return true
}

/**
 * 创建新文件夹
 * @param structure 文件夹结构对象（必须是响应式对象）
 * @param folderName 文件夹名称
 * @param path 路径
 * @returns 是否成功创建
 */
export function createFolder(
  structure: FolderStructure,
  folderName: string,
  path: string[]
): boolean {
  const folder = getFolderByPath(structure, path)

  if (!folder) {
    console.error('[文件操作] 创建文件夹失败：路径无效')
    return false
  }

  // 处理重名
  let finalName = folderName
  let counter = 1
  while (folder[finalName]) {
    finalName = `${folderName} (${counter})`
    counter++
  }

  folder[finalName] = {
    name: finalName,
    type: 'folder',
    children: {}
  }

  console.log(`[文件操作] 成功创建文件夹: ${finalName}`)
  return true
}

/**
 * 检查 targetPath 是否是 sourcePath 的后代路径
 */
function isPathDescendant(sourcePath: string[], targetPath: string[]): boolean {
  if (targetPath.length <= sourcePath.length) {
    return false
  }

  for (let i = 0; i < sourcePath.length; i++) {
    if (sourcePath[i] !== targetPath[i]) {
      return false
    }
  }

  return true
}

/**
 * 获取文件夹内的所有项目
 * @param structure 文件夹结构对象
 * @param path 路径
 * @returns 项目数组
 */
export function getItems(
  structure: FolderStructure,
  path: string[]
): FileItem[] {
  const folder = getFolderByPath(structure, path)

  if (!folder) {
    return []
  }

  return Object.values(folder)
}

/**
 * 搜索文件/文件夹
 * @param structure 文件夹结构对象
 * @param path 搜索路径
 * @param query 搜索关键词
 * @returns 匹配的项目数组
 */
export function searchItems(
  structure: FolderStructure,
  path: string[],
  query: string
): FileItem[] {
  const items = getItems(structure, path)

  if (!query.trim()) {
    return items
  }

  const lowerQuery = query.toLowerCase()
  return items.filter(item =>
    item.name.toLowerCase().includes(lowerQuery)
  )
}