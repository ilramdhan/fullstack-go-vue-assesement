import { computed, ref, watch } from 'vue'
import { defineStore } from 'pinia'

const STORAGE_KEY = 'auth.session.v1'

type Role = 'cs' | 'operation'

interface PersistedSession {
  token: string
  email: string
  role: Role
}

function loadFromStorage(): PersistedSession | null {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) return null
    const parsed = JSON.parse(raw) as PersistedSession
    if (!parsed?.token || !parsed?.email || !parsed?.role) return null
    return parsed
  } catch {
    return null
  }
}

export const useAuthStore = defineStore('auth', () => {
  const initial = loadFromStorage()
  const token = ref<string | null>(initial?.token ?? null)
  const email = ref<string | null>(initial?.email ?? null)
  const role = ref<Role | null>(initial?.role ?? null)

  const isAuthenticated = computed(() => !!token.value)
  const isOperation = computed(() => role.value === 'operation')

  function setSession(next: PersistedSession) {
    token.value = next.token
    email.value = next.email
    role.value = next.role
  }

  function clear() {
    token.value = null
    email.value = null
    role.value = null
  }

  watch(
    [token, email, role],
    ([t, e, r]) => {
      if (t && e && r) {
        localStorage.setItem(STORAGE_KEY, JSON.stringify({ token: t, email: e, role: r }))
      } else {
        localStorage.removeItem(STORAGE_KEY)
      }
    },
    { flush: 'sync' },
  )

  return { token, email, role, isAuthenticated, isOperation, setSession, clear }
})

export type { Role }
