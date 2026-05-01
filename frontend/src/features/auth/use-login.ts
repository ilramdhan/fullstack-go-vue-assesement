import { useMutation } from '@tanstack/vue-query'
import { useAuthStore, type Role } from '@/stores/auth'
import { loginUser } from '@/api/generated'

interface LoginInput {
  email: string
  password: string
}

export function useLogin() {
  const auth = useAuthStore()

  return useMutation({
    mutationFn: async ({ email, password }: LoginInput) => {
      const { data, error } = await loginUser({ body: { email, password } })
      if (error || !data) throw error ?? new Error('login failed')
      return data
    },
    onSuccess: (data) => {
      auth.setSession({
        token: data.token,
        email: data.user.email,
        role: data.user.role as Role,
      })
    },
  })
}
