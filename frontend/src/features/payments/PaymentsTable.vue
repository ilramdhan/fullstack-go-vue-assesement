<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { ArrowDown, ArrowUp, ArrowUpDown, RefreshCw, Search } from 'lucide-vue-next'
import { useDebounceFn } from '@vueuse/core'
import { toast } from 'vue-sonner'

import { Card, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { Dialog } from '@/components/ui/dialog'

import { formatCurrency, formatDate, shortId } from '@/lib/format'
import { apiErrorMessage } from '@/api/http'
import { useAuthStore } from '@/stores/auth'

import StatusBadge from './StatusBadge.vue'
import { usePayments, useReviewPayment, type PaymentsFilter } from './queries'
import type { Payment, PaymentStatus, ReviewDecision } from '@/api/generated'

const auth = useAuthStore()

const filter = reactive<PaymentsFilter>({ status: '', id: '', sort: '-created_at' })

const searchInput = ref('')
const setSearch = useDebounceFn((v: string) => {
  filter.id = v.trim()
  page.value = 1
}, 300)
watch(searchInput, (v) => setSearch(v))

const { data, isLoading, isError, error, refetch, isFetching } = usePayments(() => ({
  status: filter.status,
  id: filter.id,
  sort: filter.sort,
}))

const review = useReviewPayment()

const PAGE_SIZE = 10
const page = ref(1)
const rows = computed(() => data.value?.data ?? [])
const totalPages = computed(() => Math.max(1, Math.ceil(rows.value.length / PAGE_SIZE)))
const visibleRows = computed(() =>
  rows.value.slice((page.value - 1) * PAGE_SIZE, page.value * PAGE_SIZE),
)

watch([() => filter.status, () => filter.sort], () => {
  page.value = 1
})

function setStatus(next: PaymentStatus | '') {
  filter.status = next
  page.value = 1
}

function toggleSort(field: 'created_at' | 'amount') {
  const current = filter.sort
  filter.sort = current === field ? `-${field}` : current === `-${field}` ? field : `-${field}`
}

function sortIcon(field: 'created_at' | 'amount') {
  if (filter.sort === field) return ArrowUp
  if (filter.sort === `-${field}`) return ArrowDown
  return ArrowUpDown
}

const STATUS_OPTIONS: Array<{ label: string; value: PaymentStatus | '' }> = [
  { label: 'All', value: '' },
  { label: 'Completed', value: 'completed' },
  { label: 'Processing', value: 'processing' },
  { label: 'Failed', value: 'failed' },
]

const pendingReview = ref<{ payment: Payment; decision: ReviewDecision } | null>(null)
const dialogOpen = ref(false)

function askReview(payment: Payment, decision: ReviewDecision) {
  pendingReview.value = { payment, decision }
  dialogOpen.value = true
}

async function confirmReview() {
  if (!pendingReview.value) return
  const { payment, decision } = pendingReview.value
  dialogOpen.value = false
  try {
    await review.mutateAsync({ id: payment.id, decision })
    toast.success(decision === 'approve' ? 'Payment approved' : 'Payment rejected')
  } catch (err) {
    toast.error(apiErrorMessage(err, 'Could not update payment'))
  } finally {
    pendingReview.value = null
  }
}
</script>

<template>
  <Card>
    <CardContent class="p-0">
      <div class="flex flex-col gap-4 border-b p-5 sm:flex-row sm:items-end sm:justify-between">
        <div class="flex flex-1 flex-col gap-3 sm:flex-row sm:items-end">
          <div class="space-y-1.5 sm:max-w-sm sm:flex-1">
            <Label for="search-id">Search by payment ID</Label>
            <div class="relative">
              <Search class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                id="search-id"
                v-model="searchInput"
                placeholder="paste a uuid…"
                class="pl-9"
              />
            </div>
          </div>

          <div class="space-y-1.5">
            <Label>Status</Label>
            <div class="flex flex-wrap gap-1.5" role="group" aria-label="Filter by status">
              <Button
                v-for="opt in STATUS_OPTIONS"
                :key="opt.value || 'all'"
                type="button"
                size="sm"
                :variant="filter.status === opt.value ? 'default' : 'outline'"
                @click="setStatus(opt.value)"
              >
                {{ opt.label }}
              </Button>
            </div>
          </div>
        </div>

        <Button variant="outline" size="sm" :disabled="isFetching" @click="() => refetch()">
          <RefreshCw class="h-4 w-4" :class="{ 'animate-spin': isFetching }" />
          Refresh
        </Button>
      </div>

      <div class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead class="bg-muted/40 text-left text-xs uppercase tracking-wide text-muted-foreground">
            <tr>
              <th scope="col" class="px-5 py-3 font-medium">Payment ID</th>
              <th scope="col" class="px-5 py-3 font-medium">Merchant</th>
              <th scope="col" class="px-5 py-3 font-medium">
                <button
                  type="button"
                  class="inline-flex items-center gap-1 hover:text-foreground"
                  @click="toggleSort('created_at')"
                >
                  Date
                  <component :is="sortIcon('created_at')" class="h-3 w-3" />
                </button>
              </th>
              <th scope="col" class="px-5 py-3 text-right font-medium">
                <button
                  type="button"
                  class="ml-auto inline-flex items-center gap-1 hover:text-foreground"
                  @click="toggleSort('amount')"
                >
                  Amount
                  <component :is="sortIcon('amount')" class="h-3 w-3" />
                </button>
              </th>
              <th scope="col" class="px-5 py-3 font-medium">Status</th>
              <th v-if="auth.isOperation" scope="col" class="px-5 py-3 font-medium text-right">
                Actions
              </th>
            </tr>
          </thead>
          <tbody class="divide-y">
            <template v-if="isLoading">
              <tr v-for="i in 6" :key="i">
                <td v-for="c in auth.isOperation ? 6 : 5" :key="c" class="px-5 py-3">
                  <Skeleton class="h-4 w-full" />
                </td>
              </tr>
            </template>

            <tr v-else-if="isError">
              <td :colspan="auth.isOperation ? 6 : 5" class="px-5 py-12 text-center">
                <div class="space-y-2">
                  <p class="text-sm text-destructive">{{ apiErrorMessage(error) }}</p>
                  <Button size="sm" variant="outline" @click="() => refetch()">Try again</Button>
                </div>
              </td>
            </tr>

            <tr v-else-if="!rows.length">
              <td :colspan="auth.isOperation ? 6 : 5" class="px-5 py-12 text-center text-sm text-muted-foreground">
                No payments match the current filter.
              </td>
            </tr>

            <tr v-for="p in visibleRows" v-else :key="p.id" class="hover:bg-muted/30">
              <td class="px-5 py-3 font-mono text-xs" :title="p.id">{{ shortId(p.id) }}</td>
              <td class="px-5 py-3">{{ p.merchant }}</td>
              <td class="px-5 py-3 text-muted-foreground">{{ formatDate(p.created_at) }}</td>
              <td class="px-5 py-3 text-right font-medium tabular-nums">
                {{ formatCurrency(p.amount, p.currency) }}
              </td>
              <td class="px-5 py-3">
                <StatusBadge :status="p.status" />
              </td>
              <td v-if="auth.isOperation" class="px-5 py-3 text-right">
                <div v-if="p.status === 'processing'" class="flex justify-end gap-2">
                  <Button
                    size="sm"
                    variant="outline"
                    :disabled="review.isPending.value"
                    @click="askReview(p, 'reject')"
                  >
                    Reject
                  </Button>
                  <Button
                    size="sm"
                    :disabled="review.isPending.value"
                    @click="askReview(p, 'approve')"
                  >
                    Approve
                  </Button>
                </div>
                <span v-else-if="p.reviewed_by" class="text-xs text-muted-foreground">
                  Reviewed by {{ p.reviewed_by }}
                </span>
                <span v-else class="text-xs text-muted-foreground">—</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <footer
        v-if="rows.length"
        class="flex flex-col gap-2 border-t px-5 py-3 text-xs text-muted-foreground sm:flex-row sm:items-center sm:justify-between"
      >
        <p>
          Showing
          <span class="font-medium text-foreground">{{ (page - 1) * PAGE_SIZE + 1 }}</span>
          –
          <span class="font-medium text-foreground">
            {{ Math.min(page * PAGE_SIZE, rows.length) }}
          </span>
          of
          <span class="font-medium text-foreground">{{ data?.total ?? rows.length }}</span>
        </p>
        <div class="flex items-center gap-2">
          <Button size="sm" variant="outline" :disabled="page === 1" @click="page--">
            Previous
          </Button>
          <span class="tabular-nums">Page {{ page }} of {{ totalPages }}</span>
          <Button size="sm" variant="outline" :disabled="page >= totalPages" @click="page++">
            Next
          </Button>
        </div>
      </footer>
    </CardContent>
  </Card>

  <Dialog
    v-model:open="dialogOpen"
    :title="pendingReview?.decision === 'approve' ? 'Approve payment?' : 'Reject payment?'"
    :description="
      pendingReview
        ? `${pendingReview.payment.merchant} · ${formatCurrency(
            pendingReview.payment.amount,
            pendingReview.payment.currency,
          )}`
        : ''
    "
  >
    <p class="mb-6 text-sm text-muted-foreground">
      This will mark the payment as
      <span class="font-medium text-foreground">
        {{ pendingReview?.decision === 'approve' ? 'completed' : 'failed' }}
      </span>
      and record you as the reviewer. This action cannot be undone.
    </p>
    <div class="flex justify-end gap-2">
      <Button variant="outline" @click="dialogOpen = false">Cancel</Button>
      <Button
        :variant="pendingReview?.decision === 'reject' ? 'destructive' : 'default'"
        :disabled="review.isPending.value"
        @click="confirmReview"
      >
        {{ pendingReview?.decision === 'approve' ? 'Approve' : 'Reject' }}
      </Button>
    </div>
  </Dialog>
</template>
