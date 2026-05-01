import { computed, type MaybeRefOrGetter, toValue } from 'vue'
import { useMutation, useQuery, useQueryClient } from '@tanstack/vue-query'
import {
  getPaymentSummary,
  listPayments,
  reviewPayment,
  type Payment,
  type PaymentList,
  type PaymentStatus,
  type ReviewDecision,
} from '@/api/generated'

export interface PaymentsFilter {
  status?: PaymentStatus | ''
  id?: string
  sort?: string
}

export const paymentKeys = {
  all: ['payments'] as const,
  list: (filter: PaymentsFilter) => ['payments', 'list', filter] as const,
  summary: () => ['payments', 'summary'] as const,
}

export function usePayments(filter: MaybeRefOrGetter<PaymentsFilter>) {
  return useQuery({
    queryKey: computed(() => paymentKeys.list(toValue(filter))),
    queryFn: async () => {
      const f = toValue(filter)
      const { data, error } = await listPayments({
        query: {
          status: f.status || undefined,
          id: f.id || undefined,
          sort: f.sort || undefined,
        },
      })
      if (error || !data) throw error ?? new Error('failed to load payments')
      return data
    },
  })
}

export function usePaymentsSummary() {
  return useQuery({
    queryKey: paymentKeys.summary(),
    queryFn: async () => {
      const { data, error } = await getPaymentSummary()
      if (error || !data) throw error ?? new Error('failed to load summary')
      return data
    },
  })
}

interface ReviewInput {
  id: string
  decision: ReviewDecision
}

export function useReviewPayment() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({ id, decision }: ReviewInput) => {
      const { data, error } = await reviewPayment({
        path: { id },
        body: { decision },
      })
      if (error || !data) throw error ?? new Error('review failed')
      return data
    },
    onMutate: async ({ id, decision }) => {
      await queryClient.cancelQueries({ queryKey: paymentKeys.all })

      const optimistic: Partial<Payment> = {
        status: decision === 'approve' ? 'completed' : 'failed',
        reviewed_at: new Date().toISOString(),
      }

      const lists = queryClient.getQueriesData<PaymentList>({ queryKey: ['payments', 'list'] })
      for (const [key, value] of lists) {
        if (!value) continue
        queryClient.setQueryData<PaymentList>(key, {
          ...value,
          data: value.data.map((p) => (p.id === id ? { ...p, ...optimistic } : p)),
        })
      }
      return { lists }
    },
    onError: (_err, _vars, ctx) => {
      if (!ctx?.lists) return
      for (const [key, value] of ctx.lists) {
        queryClient.setQueryData(key, value)
      }
    },
    onSettled: () => {
      void queryClient.invalidateQueries({ queryKey: paymentKeys.all })
    },
  })
}
