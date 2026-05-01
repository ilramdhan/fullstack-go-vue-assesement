import { afterEach, beforeEach, describe, expect, it } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { createMemoryHistory, createRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

function build() {
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/', redirect: '/dashboard' },
      {
        path: '/login',
        name: 'login',
        component: { template: '<div />' },
        meta: { public: true },
      },
      {
        path: '/dashboard',
        name: 'dashboard',
        component: { template: '<div />' },
        meta: { requiresAuth: true },
      },
    ],
  })

  router.beforeEach((to) => {
    const auth = useAuthStore()
    if (to.meta.requiresAuth && !auth.isAuthenticated) {
      return { name: 'login', query: { redirect: to.fullPath } }
    }
    if (to.name === 'login' && auth.isAuthenticated) {
      return { name: 'dashboard' }
    }
  })

  return router
}

describe('router guards', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    localStorage.clear()
  })

  afterEach(() => {
    localStorage.clear()
  })

  it('redirects unauthenticated users from /dashboard to /login', async () => {
    const router = build()
    await router.push('/dashboard')
    await router.isReady()
    expect(router.currentRoute.value.name).toBe('login')
    expect(router.currentRoute.value.query.redirect).toBe('/dashboard')
  })

  it('keeps authenticated users on /dashboard', async () => {
    const router = build()
    useAuthStore().setSession({ token: 't', email: 'cs@test.com', role: 'cs' })

    await router.push('/dashboard')
    await router.isReady()
    expect(router.currentRoute.value.name).toBe('dashboard')
  })

  it('bounces authenticated users away from /login', async () => {
    const router = build()
    useAuthStore().setSession({ token: 't', email: 'cs@test.com', role: 'cs' })

    await router.push('/login')
    await router.isReady()
    expect(router.currentRoute.value.name).toBe('dashboard')
  })
})
