<script setup lang="ts">
import { useRouter } from 'vue-router'
import { Button } from '@/components/ui/button'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const router = useRouter()

function logout() {
  auth.clear()
  void router.push({ name: 'login' })
}
</script>

<template>
  <main class="min-h-screen bg-muted/30 p-8">
    <header class="mb-8 flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-semibold">Payment Dashboard</h1>
        <p class="text-sm text-muted-foreground">
          Signed in as <span class="font-medium text-foreground">{{ auth.email }}</span>
          (role: {{ auth.role }})
        </p>
      </div>
      <Button variant="outline" @click="logout">Sign out</Button>
    </header>
    <p class="text-sm text-muted-foreground">Dashboard content arrives in the next phase.</p>
  </main>
</template>
