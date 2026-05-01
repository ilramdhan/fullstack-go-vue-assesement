import '@testing-library/jest-dom/vitest'
import { afterAll, afterEach, beforeAll } from 'vitest'
import { server } from './msw/server'
import { resetMockState } from './msw/handlers'

beforeAll(() => server.listen({ onUnhandledRequest: 'error' }))
afterEach(() => {
  server.resetHandlers()
  resetMockState()
  localStorage.clear()
})
afterAll(() => server.close())
