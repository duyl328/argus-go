import type { FolderStructure } from './types'

// 生成模拟数据
export const mockFolderStructure: FolderStructure = {
  Home: {
    name: 'Home',
    type: 'folder',
    children: {
      'My Photos': {
        name: 'My Photos',
        type: 'folder',
        children: {
          'Vacation 2023': {
            name: 'Vacation 2023',
            type: 'folder',
            children: {
              Beach: {
                name: 'Beach',
                type: 'folder',
                children: {
                  'photo1.jpg': { name: 'photo1.jpg', type: 'photo', size: '2.4 MB', date: '2023-07-15' },
                  'photo2.jpg': { name: 'photo2.jpg', type: 'photo', size: '3.1 MB', date: '2023-07-15' },
                  'photo3.jpg': { name: 'photo3.jpg', type: 'photo', size: '2.8 MB', date: '2023-07-16' },
                  'beach_sunset.jpg': { name: 'beach_sunset.jpg', type: 'photo', size: '5.2 MB', date: '2023-07-17' },
                  'waves.jpg': { name: 'waves.jpg', type: 'photo', size: '3.8 MB', date: '2023-07-18' }
                }
              },
              Mountains: {
                name: 'Mountains',
                type: 'folder',
                children: {
                  'sunrise.jpg': { name: 'sunrise.jpg', type: 'photo', size: '4.2 MB', date: '2023-07-20' },
                  'peak_view.jpg': { name: 'peak_view.jpg', type: 'photo', size: '6.1 MB', date: '2023-07-21' }
                }
              }
            }
          },
          Family: {
            name: 'Family',
            type: 'folder',
            children: {
              'dinner.jpg': { name: 'dinner.jpg', type: 'photo', size: '2.2 MB', date: '2023-08-15' },
              'family_portrait.jpg': { name: 'family_portrait.jpg', type: 'photo', size: '4.1 MB', date: '2023-08-20' }
            }
          }
        }
      },
      'Camera Roll': {
        name: 'Camera Roll',
        type: 'folder',
        children: (() => {
          const items: Record<string, any> = {
            Recent: {
              name: 'Recent',
              type: 'folder',
              children: {}
            },
            Favorites: {
              name: 'Favorites',
              type: 'folder',
              children: {}
            }
          }

          // 生成100张照片
          for (let i = 1; i <= 100; i++) {
            const size = (Math.random() * 3 + 1).toFixed(1)
            const day = Math.floor(Math.random() * 30) + 1
            items[`IMG_${String(i).padStart(4, '0')}.jpg`] = {
              name: `IMG_${String(i).padStart(4, '0')}.jpg`,
              type: 'photo',
              size: `${size} MB`,
              date: `2023-09-${String(day).padStart(2, '0')}`
            }
          }

          // Recent文件夹添加照片
          for (let i = 1; i <= 30; i++) {
            const size = (Math.random() * 3 + 1).toFixed(1)
            items.Recent.children[`Recent_${String(i).padStart(3, '0')}.jpg`] = {
              name: `Recent_${String(i).padStart(3, '0')}.jpg`,
              type: 'photo',
              size: `${size} MB`,
              date: `2023-12-${String(Math.floor(Math.random() * 28) + 1).padStart(2, '0')}`
            }
          }

          return items
        })()
      },
      'Work Projects': {
        name: 'Work Projects',
        type: 'folder',
        children: {
          Screenshots: {
            name: 'Screenshots',
            type: 'folder',
            children: {
              'design1.png': { name: 'design1.png', type: 'photo', size: '1.2 MB', date: '2023-09-01' },
              'design2.png': { name: 'design2.png', type: 'photo', size: '1.5 MB', date: '2023-09-02' }
            }
          }
        }
      }
    }
  }
}