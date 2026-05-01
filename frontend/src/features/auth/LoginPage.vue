<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useForm } from 'vee-validate'
import { toTypedSchema } from '@vee-validate/zod'
import { z } from 'zod'
import { toast } from 'vue-sonner'
import { LogIn, Loader2 } from 'lucide-vue-next'

import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { useLogin } from './use-login'
import { apiErrorMessage } from '@/api/http'

const router = useRouter()
const login = useLogin()

const showPassword = ref(false)

const schema = toTypedSchema(
  z.object({
    email: z.string().email('Enter a valid email'),
    password: z.string().min(1, 'Password is required'),
  }),
)

const { handleSubmit, errors, defineField } = useForm({
  validationSchema: schema,
  initialValues: { email: '', password: '' },
})

const [email, emailAttrs] = defineField('email')
const [password, passwordAttrs] = defineField('password')

const onSubmit = handleSubmit(async (values) => {
  try {
    await login.mutateAsync(values)
    toast.success('Signed in')
    await router.push({ name: 'dashboard' })
  } catch (err) {
    toast.error(apiErrorMessage(err, 'Sign in failed'))
  }
})

function fill(role: 'cs' | 'operation') {
  email.value = role === 'cs' ? 'cs@test.com' : 'operation@test.com'
  password.value = 'password'
}
</script>

<template>
  <main class="grid min-h-screen place-items-center bg-muted/30 px-4 py-12">
    <Card class="w-full max-w-md">
      <CardHeader class="space-y-1">
        <div class="mb-2 inline-flex h-10 w-10 items-center justify-center rounded-md bg-primary text-primary-foreground">
          <LogIn class="h-5 w-5" />
        </div>
        <CardTitle>Payment Dashboard</CardTitle>
        <CardDescription>Sign in to monitor and review payments.</CardDescription>
      </CardHeader>

      <form @submit="onSubmit">
        <CardContent class="space-y-4">
          <div class="space-y-2">
            <Label for="email">Email</Label>
            <Input
              id="email"
              v-model="email"
              v-bind="emailAttrs"
              type="email"
              autocomplete="username"
              placeholder="you@durianpay.id"
              :disabled="login.isPending.value"
            />
            <p v-if="errors.email" class="text-sm text-destructive">{{ errors.email }}</p>
          </div>

          <div class="space-y-2">
            <div class="flex items-center justify-between">
              <Label for="password">Password</Label>
              <button
                type="button"
                class="text-xs text-muted-foreground hover:text-foreground"
                @click="showPassword = !showPassword"
              >
                {{ showPassword ? 'Hide' : 'Show' }}
              </button>
            </div>
            <Input
              id="password"
              v-model="password"
              v-bind="passwordAttrs"
              :type="showPassword ? 'text' : 'password'"
              autocomplete="current-password"
              :disabled="login.isPending.value"
            />
            <p v-if="errors.password" class="text-sm text-destructive">{{ errors.password }}</p>
          </div>

          <div class="rounded-md border bg-muted/50 p-3 text-xs text-muted-foreground">
            <p class="mb-2 font-medium text-foreground">Demo accounts</p>
            <div class="flex flex-wrap gap-2">
              <button
                type="button"
                class="rounded border bg-background px-2 py-1 hover:bg-accent"
                @click="fill('cs')"
              >
                cs@test.com (read-only)
              </button>
              <button
                type="button"
                class="rounded border bg-background px-2 py-1 hover:bg-accent"
                @click="fill('operation')"
              >
                operation@test.com (can review)
              </button>
            </div>
          </div>
        </CardContent>

        <CardFooter>
          <Button type="submit" class="w-full" :disabled="login.isPending.value">
            <Loader2 v-if="login.isPending.value" class="h-4 w-4 animate-spin" />
            <span>{{ login.isPending.value ? 'Signing in…' : 'Sign in' }}</span>
          </Button>
        </CardFooter>
      </form>
    </Card>
  </main>
</template>
