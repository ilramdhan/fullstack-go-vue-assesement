<script setup lang="ts">
import { type HTMLAttributes, computed } from 'vue'
import { cn } from '@/lib/utils'

const props = defineProps<{ class?: HTMLAttributes['class'] }>()
const classes = computed(() => cn('flex flex-col space-y-1.5 p-6', props.class))
</script>

<template>
  <div :class="classes"><slot /></div>
</template>
