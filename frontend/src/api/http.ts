import axios, { type AxiosError } from 'axios'
import { toast } from 'vue-sonner'
import { useAuthStore } from '@/stores/auth'
import router from '@/router'
import { client } from './generated/client.gen'

const baseURL = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080'

export const http = axios.create({
  baseURL,
  headers: { 'Content-Type': 'application/json' },
})

http.interceptors.request.use((config) => {
  const auth = useAuthStore()
  if (auth.token) {
    config.headers = config.headers ?? {}
    config.headers.Authorization = `Bearer ${auth.token}`
  }
  return config
})

http.interceptors.response.use(
  (res) => res,
  (err: AxiosError<{ message?: string }>) => {
    const status = err.response?.status
    const auth = useAuthStore()
    const onLogin = router.currentRoute.value.name === 'login'

    if (status === 401 && auth.isAuthenticated && !onLogin) {
      auth.clear()
      toast.error('Your session has expired. Please sign in again.')
      void router.push({ name: 'login' })
    }
    return Promise.reject(err)
  },
)

client.setConfig({ axios: http, baseURL })

export function apiErrorMessage(err: unknown, fallback = 'Something went wrong'): string {
  if (axios.isAxiosError(err)) {
    const data = err.response?.data as { message?: string } | undefined
    if (data?.message) return data.message
    if (err.message) return err.message
  }
  return fallback
}
