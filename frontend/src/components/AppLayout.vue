<script setup lang="ts">
import { useRouter } from 'vue-router'
import { LogOut, Wallet } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const router = useRouter()

function logout() {
  auth.clear()
  void router.push({ name: 'login' })
}
</script>

<template>
  <div class="min-h-screen bg-muted/30">
    <header class="sticky top-0 z-10 border-b bg-background/80 backdrop-blur">
      <div class="container mx-auto flex h-16 items-center justify-between px-4">
        <div class="flex items-center gap-2">
          <div class="grid h-8 w-8 place-items-center rounded-md bg-primary text-primary-foreground">
            <Wallet class="h-4 w-4" />
          </div>
          <div>
            <h1 class="text-sm font-semibold leading-tight">Payment Dashboard</h1>
            <p class="text-xs text-muted-foreground">Durianpay internal monitoring</p>
          </div>
        </div>

        <div class="flex items-center gap-3">
          <div class="hidden text-right sm:block">
            <p class="text-sm font-medium leading-tight">{{ auth.email }}</p>
            <Badge variant="outline" class="mt-1 capitalize">{{ auth.role }}</Badge>
          </div>
          <Button variant="outline" size="sm" @click="logout">
            <LogOut class="h-4 w-4" />
            <span class="hidden sm:inline">Sign out</span>
          </Button>
        </div>
      </div>
    </header>

    <main class="container mx-auto px-4 py-8">
      <slot />
    </main>
  </div>
</template>
