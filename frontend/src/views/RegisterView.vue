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
  name: '',
  email: '',
  password: '',
  confirm: '',
})

const loading = ref(false)
const error = ref<string | null>(null)

function validate() {
  error.value = null
  if (!form.name) {
    error.value = 'Name is required.'
    return false
  }
  if (!form.email || !/.+@.+\..+/.test(form.email)) {
    error.value = 'Please enter a valid email.'
    return false
  }
  if (!form.password || form.password.length < 8) {
    error.value = 'Password must be at least 8 characters.'
    return false
  }
  if (form.password !== form.confirm) {
    error.value = 'Passwords do not match.'
    return false
  }
  return true
}

async function onSubmit() {
  if (!validate()) return
  loading.value = true
  try {
    await fetch("http://localhost:8080/api/auth/register", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        email: form.email,
        password: form.password,
      }),
    }).then(res => res.json())

    router.push({ name: 'login' })
  } catch (e) {
    error.value = 'Registration failed. Please try again.'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="flex justify-content-center">
    <Card class="w-full md:w-6 lg:w-4">
      <template #title>Create your account</template>
      <template #subtitle>It only takes a minute</template>

      <template #content>
        <form @submit.prevent="onSubmit" class="flex flex-column gap-3">
          <Message v-if="error" severity="error" :closable="false">{{ error }}</Message>

          <div class="flex flex-column gap-2">
            <label for="name">Name</label>
            <InputText id="name" v-model="form.name" placeholder="Jane Doe" />
          </div>

          <div class="flex flex-column gap-2">
            <label for="email">Email</label>
            <InputText id="email" v-model="form.email" type="email" placeholder="you@example.com" />
          </div>

          <div class="flex flex-column gap-2">
            <label for="password">Password</label>
            <Password id="password" v-model="form.password" :feedback="true" toggleMask />
          </div>

          <div class="flex flex-column gap-2">
            <label for="confirm">Confirm Password</label>
            <Password id="confirm" v-model="form.confirm" :feedback="false" toggleMask />
          </div>

          <Button type="submit" :loading="loading" label="Create account" />
        </form>

        <Divider />
        <p class="mt-2">
          Already have an account?
          <RouterLink to="/login">Log in</RouterLink>
        </p>
      </template>
    </Card>
  </div>
</template>
