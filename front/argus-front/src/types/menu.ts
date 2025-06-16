// # 菜单类型定义
export interface MenuItem {
  key: string
  label: string
  icon?: string
  path: string
}

export interface MenuGroup {
  key: string
  label: string
  children: MenuItem[]
}

export type MenuOption = MenuItem | MenuGroup

export const menuOptions: MenuOption[] = [
  // 默认分组菜单
  {
    key: 'photos',
    label: '照片',
    icon: 'image-outline',
    path: '/photos',
  },
  {
    key: 'search',
    label: '搜索',
    icon: 'search-outline',
    path: '/search',
  },
  {
    key: 'people',
    label: '人物',
    icon: 'people-outline',
    path: '/people',
  },
  // 媒体分组
  {
    key: 'media',
    label: '媒体',
    children: [],
  },

  {
    key: 'favorites',
    label: '收藏',
    icon: 'heart-outline',
    path: '/favorites',
  },
  {
    key: 'places',
    label: '地点',
    icon: 'location-outline',
    path: '/places',
  },
  {
    key: 'calendar',
    label: '日历',
    icon: 'calendar-outline',
    path: '/calendar',
  },
  {
    key: 'tags',
    label: '标签',
    icon: 'pricetag-outline',
    path: '/tags',
  },
  {
    key: 'similar',
    label: '相似照片',
    icon: 'eye-off-outline',
    path: '/similar',
  },

  // 管理分组
  {
    key: 'management',
    label: '管理',
    children: [],
  },

  {
    key: 'library',
    label: '资料库',
    icon: 'library-outline',
    path: '/library',
  },
  {
    key: 'settings',
    label: '设置',
    icon: 'settings-outline',
    path: '/settings',
  },
]
