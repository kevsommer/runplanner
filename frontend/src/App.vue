<template>
  <div class="min-h-screen surface-ground p-4">
    <nav class="mb-4 flex gap-3">
      <RouterLink to="/login">Login</RouterLink>
      <RouterLink to="/register">Register</RouterLink>
    </nav>
    <RouterView />
  </div>
</template>

<script setup lang="ts"></script>

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
