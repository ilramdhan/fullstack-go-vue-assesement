import { describe, expect, it } from 'vitest'
import { fireEvent, screen, waitFor } from '@testing-library/vue'
import { renderWithStack } from '@/test/utils'
import LoginPage from './LoginPage.vue'
import { useAuthStore } from '@/stores/auth'
import '@/api/http'

describe('LoginPage', () => {
  it('signs in successfully and updates auth store', async () => {
    renderWithStack(LoginPage)

    const auth = useAuthStore()
    expect(auth.isAuthenticated).toBe(false)

    await fireEvent.update(screen.getByLabelText(/email/i), 'operation@test.com')
    await fireEvent.update(screen.getByLabelText(/password/i), 'password')
    await fireEvent.click(screen.getByRole('button', { name: /sign in/i }))

    await waitFor(() => expect(auth.isAuthenticated).toBe(true))
    expect(auth.email).toBe('operation@test.com')
    expect(auth.role).toBe('operation')
  })

  it('shows validation error for invalid email', async () => {
    renderWithStack(LoginPage)
    await fireEvent.update(screen.getByLabelText(/email/i), 'not-an-email')
    await fireEvent.update(screen.getByLabelText(/password/i), 'x')
    await fireEvent.click(screen.getByRole('button', { name: /sign in/i }))

    expect(await screen.findByText(/valid email/i)).toBeInTheDocument()
    const auth = useAuthStore()
    expect(auth.isAuthenticated).toBe(false)
  })

  it('does not authenticate on wrong password', async () => {
    renderWithStack(LoginPage)
    await fireEvent.update(screen.getByLabelText(/email/i), 'cs@test.com')
    await fireEvent.update(screen.getByLabelText(/password/i), 'WRONG')
    await fireEvent.click(screen.getByRole('button', { name: /sign in/i }))

    const auth = useAuthStore()
    await waitFor(() => {
      expect(auth.isAuthenticated).toBe(false)
    })
  })

  it('demo credentials button fills inputs', async () => {
    renderWithStack(LoginPage)
    await fireEvent.click(screen.getByRole('button', { name: /cs@test.com/i }))
    expect(screen.getByLabelText<HTMLInputElement>(/email/i).value).toBe('cs@test.com')
    expect(screen.getByLabelText<HTMLInputElement>(/password/i).value).toBe('password')
  })
})
