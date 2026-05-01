import { http, HttpResponse } from 'msw'
import { fixturePayments } from '../fixtures'
import type { Payment, PaymentStatus } from '@/api/generated'

const BASE = 'http://localhost:8080'

interface State {
  payments: Payment[]
  badPassword: boolean
}

const state: State = {
  payments: [...fixturePayments],
  badPassword: false,
}

export function resetMockState() {
  state.payments = fixturePayments.map((p) => ({ ...p }))
  state.badPassword = false
}

export function setBadPassword(v: boolean) {
  state.badPassword = v
}

export const handlers = [
  http.post(`${BASE}/dashboard/v1/auth/login`, async ({ request }) => {
    const body = (await request.json()) as { email: string; password: string }
    if (state.badPassword || body.password !== 'password') {
      return HttpResponse.json(
        { code: 401, message: 'invalid credentials' },
        { status: 401 },
      )
    }
    const role = body.email.startsWith('operation') ? 'operation' : 'cs'
    return HttpResponse.json({
      token: 'test-token',
      user: { email: body.email, role },
    })
  }),

  http.get(`${BASE}/dashboard/v1/payments`, ({ request }) => {
    const url = new URL(request.url)
    const status = url.searchParams.get('status') as PaymentStatus | null
    const id = url.searchParams.get('id')

    let data = state.payments
    if (status) data = data.filter((p) => p.status === status)
    if (id) data = data.filter((p) => p.id === id)
    return HttpResponse.json({ data, total: data.length })
  }),

  http.get(`${BASE}/dashboard/v1/payments/summary`, () => {
    const sum = state.payments.reduce(
      (acc, p) => {
        acc.total++
        acc[p.status]++
        return acc
      },
      { total: 0, completed: 0, processing: 0, failed: 0 },
    )
    return HttpResponse.json(sum)
  }),

  http.put(`${BASE}/dashboard/v1/payments/:id/review`, async ({ params, request }) => {
    const id = params.id as string
    const body = (await request.json()) as { decision: 'approve' | 'reject' }
    const idx = state.payments.findIndex((p) => p.id === id)
    if (idx < 0) {
      return HttpResponse.json({ code: 404, message: 'payment not found' }, { status: 404 })
    }
    const cur = state.payments[idx]
    if (cur.status !== 'processing') {
      return HttpResponse.json(
        { code: 409, message: 'cannot review' },
        { status: 409 },
      )
    }
    state.payments[idx] = {
      ...cur,
      status: body.decision === 'approve' ? 'completed' : 'failed',
      reviewed_by: 'operation@test.com',
      reviewed_at: new Date().toISOString(),
    }
    return HttpResponse.json(state.payments[idx])
  }),
]

export { fixturePayments } from '../fixtures'
