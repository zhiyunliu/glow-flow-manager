import { createRouter, createWebHashHistory } from 'vue-router'

// 使用 hash 路由：页面嵌入在 Go 服务的 /api/manager-glow/web/ 下，不依赖服务端路由回退
const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    { path: '/', redirect: '/chains' },
    { path: '/chains', name: 'chains', component: () => import('@/views/ChainList.vue'), meta: { title: '流程定义' } },
    {
      path: '/designer/:versionId',
      name: 'designer',
      component: () => import('@/views/Designer.vue'),
      meta: { title: '流程设计器', fullscreen: true },
    },
    { path: '/history', name: 'history', component: () => import('@/views/History.vue'), meta: { title: '执行历史' } },
    { path: '/node-types', name: 'node-types', component: () => import('@/views/NodeTypes.vue'), meta: { title: '节点注册' } },
    { path: '/audit', name: 'audit', component: () => import('@/views/AuditLog.vue'), meta: { title: '操作审计' } },
    { path: '/:pathMatch(.*)*', redirect: '/chains' },
  ],
})

router.afterEach((to) => {
  const title = to.meta.title as string | undefined
  document.title = title ? `${title} - 流程编排管理后台` : '流程编排管理后台'
})

export default router
