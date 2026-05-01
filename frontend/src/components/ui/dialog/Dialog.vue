<script setup lang="ts">
import { onKeyStroke } from '@vueuse/core'

const props = defineProps<{
  open: boolean
  title?: string
  description?: string
}>()
const emit = defineEmits<(e: 'update:open', value: boolean) => void>()

function close() {
  emit('update:open', false)
}

onKeyStroke('Escape', () => {
  if (props.open) close()
})
</script>

<template>
  <Teleport to="body">
    <Transition
      enter-active-class="transition duration-150"
      enter-from-class="opacity-0"
      enter-to-class="opacity-100"
      leave-active-class="transition duration-100"
      leave-from-class="opacity-100"
      leave-to-class="opacity-0"
    >
      <div
        v-if="open"
        class="fixed inset-0 z-50 flex items-end justify-center bg-black/50 p-4 sm:items-center"
        role="dialog"
        aria-modal="true"
        :aria-labelledby="title ? 'dialog-title' : undefined"
        :aria-describedby="description ? 'dialog-desc' : undefined"
        @click.self="close"
      >
        <div
          class="w-full max-w-md rounded-lg border bg-background p-6 shadow-lg"
          @click.stop
        >
          <header v-if="title || description" class="mb-4 space-y-1">
            <h2 v-if="title" id="dialog-title" class="text-lg font-semibold">{{ title }}</h2>
            <p v-if="description" id="dialog-desc" class="text-sm text-muted-foreground">
              {{ description }}
            </p>
          </header>
          <slot />
        </div>
      </div>
    </Transition>
  </Teleport>
</template>
