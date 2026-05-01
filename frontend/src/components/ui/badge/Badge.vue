<script setup lang="ts">
import { type HTMLAttributes, computed } from 'vue'
import { cn } from '@/lib/utils'
import { type BadgeVariants, badgeVariants } from './variants'

const props = withDefaults(
  defineProps<{
    variant?: BadgeVariants['variant']
    class?: HTMLAttributes['class']
  }>(),
  { variant: 'default' },
)

const classes = computed(() => cn(badgeVariants({ variant: props.variant }), props.class))
</script>

<template>
  <span :class="classes"><slot /></span>
</template>
