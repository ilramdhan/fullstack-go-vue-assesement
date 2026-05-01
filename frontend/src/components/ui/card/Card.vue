<script setup lang="ts">
import { type HTMLAttributes, computed } from 'vue'
import { cn } from '@/lib/utils'

const props = defineProps<{ class?: HTMLAttributes['class'] }>()
const classes = computed(() =>
  cn('rounded-lg border bg-card text-card-foreground shadow-sm', props.class),
)
</script>

<template>
  <div :class="classes"><slot /></div>
</template>
