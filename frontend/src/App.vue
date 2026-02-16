<template>
  <div class="min-h-screen surface-ground">
    <Toast />
    <Menubar
      :model="items"
      class="mb-4">
      <template #item="{ label, item }">
        <i
          :class="item.icon"
          class="mr-2" />
        <RouterLink
          :to="item.route"
          class="c-link p-2">
          {{ label }}
        </RouterLink>
      </template>
    </Menubar>
    <RouterView />
  </div>
</template>

<script setup lang="ts">
import Menubar from "primevue/menubar";
import Toast from "primevue/toast";
import { computed } from "vue";
import { RouterLink } from "vue-router";
import { useAuth } from "./composables/useAuth";

const { isAuthed } = useAuth();

const items = computed(() =>
  isAuthed.value
    ? [
      { label: "Dashboard", icon: "pi pi-chart-bar", route: "/dashboard" },
      { label: "Logout", icon: "pi pi-sign-out", route: "/logout" },
    ]
    : [
      {
        label: "Login",
        icon: "pi pi-sign-in",
        route: "/login",
      },
      {
        label: "Register",
        icon: "pi pi-user-plus",
        route: "/register",
      },
    ],
);
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
  margin: 0;
  font-family:
    -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
}
</style>
