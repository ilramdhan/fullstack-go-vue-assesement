<script setup lang="ts">
import { type HTMLAttributes, computed, useAttrs } from 'vue'
import { cn } from '@/lib/utils'

interface Props {
  modelValue?: string | number
  class?: HTMLAttributes['class']
}

const props = defineProps<Props>()
const emit = defineEmits<(e: 'update:modelValue', value: string | number) => void>()

defineOptions({ inheritAttrs: false })
const attrs = useAttrs()

const classes = computed(() =>
  cn(
    'flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50',
    props.class,
  ),
)
</script>

<template>
  <input
    :value="modelValue"
    :class="classes"
    v-bind="attrs"
    @input="emit('update:modelValue', ($event.target as HTMLInputElement).value)"
  >
</template>
