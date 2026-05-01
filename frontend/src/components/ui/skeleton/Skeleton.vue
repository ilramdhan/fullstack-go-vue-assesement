<script setup lang="ts">
import { type HTMLAttributes, computed } from 'vue'
import { cn } from '@/lib/utils'

const props = defineProps<{ class?: HTMLAttributes['class'] }>()
const classes = computed(() => cn('animate-pulse rounded-md bg-muted', props.class))
</script>

<template>
  <div :class="classes" />
</template>
