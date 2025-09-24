<template>
  <div class="flex justify-content-center">
    <Card class="w-full md:w-6 lg:w-4">
      <template #title>Welcome back</template>
      <template #subtitle>Log in to your account</template>

      <template #content>
        <form @submit.prevent="onSubmit" class="flex flex-column gap-3">
          <Message v-if="error" severity="error" :closable="false">{{ error }}</Message>

          <div class="flex flex-column gap-2">
            <label for="email">Email</label>
            <InputText id="email" v-model="form.email" type="email" placeholder="you@example.com" />
          </div>

          <div class="flex flex-column gap-2">
            <label for="password">Password</label>
            <Password id="password" v-model="form.password" :feedback="false" toggleMask />
          </div>

          <Button type="submit" :loading="loading" label="Login" />
        </form>

        <Divider />
        <p class="mt-2">
          Don’t have an account?
          <RouterLink to="/register">Create one</RouterLink>
        </p>
      </template>
    </Card>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'

import Card from 'primevue/card'
import InputText from 'primevue/inputtext'
import Password from 'primevue/password'
import Button from 'primevue/button'
import Message from 'primevue/message'
import Divider from 'primevue/divider'

const router = useRouter()

const form = reactive({
  email: '',
  password: '',
})
const loading = ref(false)
const error = ref<string | null>(null)

function validate() {
  error.value = null
  if (!form.email || !/.+@.+\..+/.test(form.email)) {
    error.value = 'Please enter a valid email.'
    return false
  }
  if (!form.password) {
    error.value = 'Password is required.'
    return false
  }
  return true
}

async function onSubmit() {
  if (!validate()) return
  loading.value = true
  try {
    // TODO: call your real API here
    // Example mock:
    await new Promise((r) => setTimeout(r, 600))
    localStorage.setItem('token', 'mock-token')
    router.push({ name: 'dashboard' }) // or wherever
  } catch (e) {
    error.value = 'Login failed. Please try again.'
  } finally {
    loading.value = false
  }
}
</script>
