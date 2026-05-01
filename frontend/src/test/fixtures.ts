import type { Payment, PaymentSummary } from '@/api/generated'

export function makePayment(overrides: Partial<Payment> = {}): Payment {
  return {
    id: 'p-' + Math.random().toString(36).slice(2, 10),
    merchant: 'Tokopedia',
    amount: 1_500_000_00,
    currency: 'IDR',
    status: 'completed',
    created_at: '2026-04-15T10:00:00Z',
    reviewed_by: null,
    reviewed_at: null,
    ...overrides,
  }
}

export const fixturePayments: Payment[] = [
  makePayment({ id: 'p-completed', merchant: 'Tokopedia', status: 'completed', amount: 1_000_000_00 }),
  makePayment({ id: 'p-processing', merchant: 'Shopee', status: 'processing', amount: 2_000_000_00 }),
  makePayment({ id: 'p-failed', merchant: 'Lazada', status: 'failed', amount: 3_000_000_00 }),
  makePayment({ id: 'p-completed-2', merchant: 'Blibli', status: 'completed', amount: 500_000_00 }),
]

export const fixtureSummary: PaymentSummary = {
  total: 4,
  completed: 2,
  processing: 1,
  failed: 1,
}
