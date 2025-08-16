import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      redirect: '/photos'
    },
    {
      path: '/photos',
      name: 'photos',
      component: () => import('@/views/PhotosView.vue'),
      meta: { title: '照片' }
    },
    {
      path: '/search',
      name: 'search',
      component: () => import('@/views/SearchView.vue'),
      meta: { title: '搜索' }
    },
    {
      path: '/people',
      name: 'people',
      component: () => import('@/views/PeopleView.vue'),
      meta: { title: '人物' }
    },
    {
      path: '/field',
      name: 'field',
      component: () => import('@/views/FieldView.vue'),
      meta: { title: '文件夹' }
    },
    {
      path: '/local-files',
      name: 'local-files',
      component: () => import('@/views/LocalFilesView.vue'),
      meta: { title: '本地文件' }
    },

    {
      path: '/library',
      name: 'library',
      component: () => import('@/views/LibraryView.vue'),
      meta: { title: '资料库' }
    },
    {
      path: '/settings',
      name: 'settings',
      component: () => import('@/views/SettingsView.vue'),
      meta: { title: '设置' }
    }
  ]
})


export default router
