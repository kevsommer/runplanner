<template>
  <div class="min-h-screen surface-ground">
    <Menubar :model="items" class="mb-4">
      <template #item="{ label, item }">
        <RouterLink :to="item.route" class="c-link p-2">
          {{ label }}
        </RouterLink>
      </template>
    </Menubar>
    <RouterView />
  </div>
</template>

<script setup lang="ts">
import Menubar from 'primevue/menubar'
import { RouterLink } from 'vue-router';

const items = [
  {
    label: 'Login',
    icon: 'pi pi-sign-in',
    route: '/login'
  },
  {
    label: 'Register',
    icon: 'pi pi-user-plus',
    route: '/register'
  }
]
</script>

<style scoped>
.c-link {
  text-decoration: none;
  color: var(--p-primary-color);
}
</style>

<style>
/* optional small helpers */
.surface-ground {
  background: var(--p-surface-ground);
}

body {
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto,
    "Helvetica Neue", Arial, sans-serif;
}
</style>
