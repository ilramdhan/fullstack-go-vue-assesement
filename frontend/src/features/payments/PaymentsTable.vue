<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import {
  ArrowDown,
  ArrowUp,
  ArrowUpDown,
  Copy,
  Download,
  RefreshCw,
  Search,
  X,
} from 'lucide-vue-next'
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
import { usePayments, useReviewPayment } from './queries'
import { exportPaymentsToExcel } from './export'
import type { Payment, PaymentStatus, ReviewDecision } from '@/api/generated'

type SortField = 'created_at' | 'amount' | 'merchant' | 'status'
type SortKey = SortField | `-${SortField}`

const auth = useAuthStore()

const filter = reactive({
  status: '' as PaymentStatus | '',
  sort: '-created_at' as SortKey,
  from: '',
  to: '',
})

const idInput = ref('')
const merchantInput = ref('')
const idQuery = ref('')
const merchantQuery = ref('')

const setIdQuery = useDebounceFn((v: string) => {
  idQuery.value = v.trim()
  page.value = 1
}, 250)
const setMerchantQuery = useDebounceFn((v: string) => {
  merchantQuery.value = v.trim()
  page.value = 1
}, 250)
watch(idInput, (v) => setIdQuery(v))
watch(merchantInput, (v) => setMerchantQuery(v))

const { data, isLoading, isError, error, refetch, isFetching } = usePayments(() => ({
  status: filter.status,
}))

const review = useReviewPayment()

const allRows = computed(() => data.value?.data ?? [])

const filteredRows = computed(() => {
  const id = idQuery.value.toLowerCase()
  const merchant = merchantQuery.value.toLowerCase()
  const fromTs = filter.from ? new Date(filter.from + 'T00:00:00').getTime() : null
  const toTs = filter.to ? new Date(filter.to + 'T23:59:59.999').getTime() : null

  return allRows.value.filter((p) => {
    if (id && !p.id.toLowerCase().includes(id)) return false
    if (merchant && !p.merchant.toLowerCase().includes(merchant)) return false
    if (fromTs !== null || toTs !== null) {
      const t = new Date(p.created_at).getTime()
      if (fromTs !== null && t < fromTs) return false
      if (toTs !== null && t > toTs) return false
    }
    return true
  })
})

const sortedRows = computed(() => {
  const sort = filter.sort
  if (!sort) return filteredRows.value
  const desc = sort.startsWith('-')
  const field = (desc ? sort.slice(1) : sort) as SortField
  const arr = [...filteredRows.value]
  arr.sort((a, b) => {
    let cmp = 0
    if (field === 'amount') {
      cmp = a.amount - b.amount
    } else if (field === 'merchant') {
      cmp = a.merchant.localeCompare(b.merchant)
    } else if (field === 'status') {
      cmp = a.status.localeCompare(b.status)
    } else {
      cmp = new Date(a.created_at).getTime() - new Date(b.created_at).getTime()
    }
    return desc ? -cmp : cmp
  })
  return arr
})

const PAGE_SIZE = 10
const page = ref(1)
const totalPages = computed(() => Math.max(1, Math.ceil(sortedRows.value.length / PAGE_SIZE)))
const visibleRows = computed(() =>
  sortedRows.value.slice((page.value - 1) * PAGE_SIZE, page.value * PAGE_SIZE),
)

watch([() => filter.status, () => filter.from, () => filter.to, () => filter.sort], () => {
  if (page.value > totalPages.value) page.value = 1
})

function setStatus(next: PaymentStatus | '') {
  filter.status = next
  page.value = 1
}

function toggleSort(field: SortField) {
  if (filter.sort === field) filter.sort = `-${field}`
  else if (filter.sort === `-${field}`) filter.sort = field
  else filter.sort = `-${field}`
}

function sortIcon(field: SortField) {
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

const hasActiveFilters = computed(
  () =>
    !!filter.status ||
    !!filter.from ||
    !!filter.to ||
    !!idQuery.value ||
    !!merchantQuery.value,
)

function clearFilters() {
  filter.status = ''
  filter.from = ''
  filter.to = ''
  idInput.value = ''
  merchantInput.value = ''
  idQuery.value = ''
  merchantQuery.value = ''
  page.value = 1
}

const minSpin = ref(false)
async function handleRefresh() {
  minSpin.value = true
  setTimeout(() => {
    minSpin.value = false
  }, 700)
  await refetch()
}
const isRefreshing = computed(() => minSpin.value || isFetching.value)

function handleExport() {
  if (!sortedRows.value.length) {
    toast.info('Nothing to export with the current filters')
    return
  }
  try {
    exportPaymentsToExcel(sortedRows.value)
    toast.success(`Exported ${sortedRows.value.length} payment(s)`)
  } catch (e) {
    toast.error(apiErrorMessage(e, 'Could not generate the Excel file'))
  }
}

async function copyId(id: string) {
  try {
    await navigator.clipboard.writeText(id)
    toast.success('Payment ID copied')
  } catch {
    toast.error('Clipboard not available')
  }
}

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

const totalCount = computed(() => sortedRows.value.length)
const startIdx = computed(() => (totalCount.value === 0 ? 0 : (page.value - 1) * PAGE_SIZE + 1))
const endIdx = computed(() => Math.min(page.value * PAGE_SIZE, totalCount.value))
</script>

<template>
  <Card>
    <CardContent class="p-0">
      <div class="space-y-4 border-b p-5">
        <div class="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
          <div class="space-y-1.5">
            <Label for="search-id">Payment ID</Label>
            <div class="relative">
              <Search class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                id="search-id"
                v-model="idInput"
                placeholder="paste full or partial id"
                class="pl-9"
              />
            </div>
          </div>

          <div class="space-y-1.5">
            <Label for="search-merchant">Merchant</Label>
            <div class="relative">
              <Search class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                id="search-merchant"
                v-model="merchantInput"
                placeholder="e.g. Tokopedia"
                class="pl-9"
              />
            </div>
          </div>

          <div class="space-y-1.5">
            <Label for="date-from">From</Label>
            <Input id="date-from" v-model="filter.from" type="date" />
          </div>

          <div class="space-y-1.5">
            <Label for="date-to">To</Label>
            <Input id="date-to" v-model="filter.to" type="date" />
          </div>
        </div>

        <div class="flex flex-wrap items-end justify-between gap-3">
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

          <div class="flex flex-wrap items-center gap-2">
            <Button
              v-if="hasActiveFilters"
              variant="ghost"
              size="sm"
              type="button"
              @click="clearFilters"
            >
              <X class="h-4 w-4" />
              Clear filters
            </Button>
            <Button
              variant="outline"
              size="sm"
              type="button"
              :disabled="isRefreshing"
              @click="handleRefresh"
            >
              <RefreshCw class="h-4 w-4" :class="{ 'animate-spin': isRefreshing }" />
              Refresh
            </Button>
            <Button
              variant="outline"
              size="sm"
              type="button"
              :disabled="!sortedRows.length"
              @click="handleExport"
            >
              <Download class="h-4 w-4" />
              Export Excel
            </Button>
          </div>
        </div>
      </div>

      <div class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead class="bg-muted/40 text-left text-xs uppercase tracking-wide text-muted-foreground">
            <tr>
              <th scope="col" class="px-5 py-3 font-medium">Payment ID</th>
              <th scope="col" class="px-5 py-3 font-medium">
                <button
                  type="button"
                  class="inline-flex items-center gap-1 hover:text-foreground"
                  @click="toggleSort('merchant')"
                >
                  Merchant
                  <component :is="sortIcon('merchant')" class="h-3 w-3" />
                </button>
              </th>
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
              <th scope="col" class="px-5 py-3 font-medium">
                <button
                  type="button"
                  class="inline-flex items-center gap-1 hover:text-foreground"
                  @click="toggleSort('status')"
                >
                  Status
                  <component :is="sortIcon('status')" class="h-3 w-3" />
                </button>
              </th>
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
                  <Button size="sm" variant="outline" @click="handleRefresh">Try again</Button>
                </div>
              </td>
            </tr>

            <tr v-else-if="!sortedRows.length">
              <td :colspan="auth.isOperation ? 6 : 5" class="px-5 py-12 text-center text-sm text-muted-foreground">
                <p>No payments match the current filters.</p>
                <Button v-if="hasActiveFilters" variant="link" size="sm" class="mt-1" @click="clearFilters">
                  Clear filters
                </Button>
              </td>
            </tr>

            <tr v-for="p in visibleRows" v-else :key="p.id" class="hover:bg-muted/30">
              <td class="px-5 py-3 font-mono text-xs">
                <button
                  type="button"
                  class="group inline-flex items-center gap-1.5 hover:text-foreground"
                  :title="`${p.id} (click to copy)`"
                  @click="copyId(p.id)"
                >
                  <span>{{ shortId(p.id) }}</span>
                  <Copy class="h-3 w-3 opacity-0 transition-opacity group-hover:opacity-60" />
                </button>
              </td>
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
        v-if="sortedRows.length"
        class="flex flex-col gap-2 border-t px-5 py-3 text-xs text-muted-foreground sm:flex-row sm:items-center sm:justify-between"
      >
        <p>
          Showing
          <span class="font-medium text-foreground">{{ startIdx }}</span>
          –
          <span class="font-medium text-foreground">{{ endIdx }}</span>
          of
          <span class="font-medium text-foreground">{{ totalCount }}</span>
          <span v-if="totalCount !== allRows.length" class="text-muted-foreground">
            (filtered from {{ allRows.length }})
          </span>
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
