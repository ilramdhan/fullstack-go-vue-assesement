<script setup lang="ts">
import { CheckCircle2, Clock, Receipt, XCircle } from 'lucide-vue-next'
import { Card, CardContent } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { usePaymentsSummary } from './queries'

const { data, isLoading, isError } = usePaymentsSummary()

interface Tile {
  label: string
  key: 'total' | 'completed' | 'processing' | 'failed'
  icon: typeof Receipt
  iconClass: string
}

const tiles: Tile[] = [
  { label: 'Total payments', key: 'total', icon: Receipt, iconClass: 'text-foreground' },
  { label: 'Completed', key: 'completed', icon: CheckCircle2, iconClass: 'text-emerald-600' },
  { label: 'Processing', key: 'processing', icon: Clock, iconClass: 'text-amber-600' },
  { label: 'Failed', key: 'failed', icon: XCircle, iconClass: 'text-rose-600' },
]
</script>

<template>
  <section class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4" aria-label="Payment summary">
    <Card v-for="t in tiles" :key="t.key">
      <CardContent class="p-5">
        <div class="flex items-center justify-between">
          <p class="text-xs font-medium uppercase tracking-wide text-muted-foreground">
            {{ t.label }}
          </p>
          <component :is="t.icon" class="h-4 w-4" :class="t.iconClass" />
        </div>
        <Skeleton v-if="isLoading" class="mt-3 h-8 w-20" />
        <p v-else-if="isError" class="mt-3 text-sm text-destructive">Failed to load</p>
        <p v-else class="mt-2 text-3xl font-semibold tabular-nums">
          {{ data?.[t.key] ?? 0 }}
        </p>
      </CardContent>
    </Card>
  </section>
</template>
