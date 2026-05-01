import { render, type RenderOptions } from '@testing-library/vue'
import { QueryClient, VueQueryPlugin } from '@tanstack/vue-query'
import { createPinia } from 'pinia'
import { createMemoryHistory, createRouter, type RouteRecordRaw } from 'vue-router'
import type { Component } from 'vue'

const defaultRoutes: RouteRecordRaw[] = [
  { path: '/', component: { template: '<div />' } },
  { path: '/login', name: 'login', component: { template: '<div />' } },
  { path: '/dashboard', name: 'dashboard', component: { template: '<div />' } },
]

export function renderWithStack(component: Component, options: RenderOptions<Component> = {}) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  })
  const router = createRouter({ history: createMemoryHistory(), routes: defaultRoutes })

  return {
    queryClient,
    router,
    ...render(component, {
      global: {
        plugins: [createPinia(), [VueQueryPlugin, { queryClient }], router],
        ...options.global,
      },
      ...options,
    }),
  }
}
