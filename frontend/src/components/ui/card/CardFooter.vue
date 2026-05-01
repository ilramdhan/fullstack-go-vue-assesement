<script setup lang="ts">
import { type HTMLAttributes, computed } from 'vue'
import { cn } from '@/lib/utils'

const props = defineProps<{ class?: HTMLAttributes['class'] }>()
const classes = computed(() => cn('flex items-center p-6 pt-0', props.class))
</script>

<template>
  <div :class="classes"><slot /></div>
</template>
