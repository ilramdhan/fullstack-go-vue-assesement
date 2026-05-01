import { beforeEach, describe, expect, it } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useAuthStore } from './auth'

describe('auth store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    localStorage.clear()
  })

  it('starts unauthenticated', () => {
    const auth = useAuthStore()
    expect(auth.isAuthenticated).toBe(false)
    expect(auth.token).toBeNull()
  })

  it('setSession persists to localStorage', () => {
    const auth = useAuthStore()
    auth.setSession({ token: 't', email: 'cs@test.com', role: 'cs' })

    expect(auth.isAuthenticated).toBe(true)
    expect(auth.role).toBe('cs')
    expect(auth.isOperation).toBe(false)

    const persisted = JSON.parse(localStorage.getItem('auth.session.v1') ?? '{}')
    expect(persisted).toEqual({ token: 't', email: 'cs@test.com', role: 'cs' })
  })

  it('clear removes the persisted session', () => {
    const auth = useAuthStore()
    auth.setSession({ token: 't', email: 'op@test.com', role: 'operation' })
    expect(localStorage.getItem('auth.session.v1')).not.toBeNull()

    auth.clear()
    expect(auth.isAuthenticated).toBe(false)
    expect(localStorage.getItem('auth.session.v1')).toBeNull()
  })

  it('rehydrates from localStorage on init', () => {
    localStorage.setItem(
      'auth.session.v1',
      JSON.stringify({ token: 't', email: 'op@test.com', role: 'operation' }),
    )
    const auth = useAuthStore()
    expect(auth.isAuthenticated).toBe(true)
    expect(auth.isOperation).toBe(true)
    expect(auth.email).toBe('op@test.com')
  })

  it('ignores corrupt storage payloads', () => {
    localStorage.setItem('auth.session.v1', '{broken json')
    const auth = useAuthStore()
    expect(auth.isAuthenticated).toBe(false)
  })
})
