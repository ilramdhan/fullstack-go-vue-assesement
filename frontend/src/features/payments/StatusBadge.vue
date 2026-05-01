<script setup lang="ts">
import { computed } from 'vue'
import { Badge } from '@/components/ui/badge'

const props = defineProps<{ status: string }>()

const variant = computed(() => {
  switch (props.status) {
    case 'completed':
      return 'success' as const
    case 'processing':
      return 'warning' as const
    case 'failed':
      return 'destructive' as const
    default:
      return 'secondary' as const
  }
})
</script>

<template>
  <Badge :variant="variant" class="capitalize">{{ status }}</Badge>
</template>
