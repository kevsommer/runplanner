<template>
  <div class="min-h-screen surface-ground">
    <Toast />
    <div class="menubar-wrapper">
      <Menubar :model="items">
        <template #start>
          <span class="app-brand">
            <i class="pi pi-directions-run" />
            RunPlanner
          </span>
        </template>
        <template #item="{ label, item }">
          <RouterLink
            :to="item.route"
            class="menu-item">
            <i :class="item.icon" />
            <span>{{ label }}</span>
          </RouterLink>
        </template>
      </Menubar>
    </div>
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
.menubar-wrapper {
  margin-bottom: 0.5rem;
}

.menubar-wrapper :deep(.p-menubar) {
  border: none;
  border-radius: 0;
  background: var(--p-primary-color);
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.12);
  padding: 0.5rem 1rem;
}

.app-brand {
  font-weight: 700;
  font-size: 1.15rem;
  color: #fff;
  display: flex;
  align-items: center;
  gap: 0.4rem;
  margin-right: 1.5rem;
  user-select: none;
}

.menu-item {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  text-decoration: none;
  color: rgba(255, 255, 255, 0.85);
  padding: 0.5rem 0.75rem;
  border-radius: 0.375rem;
  font-weight: 500;
  font-size: 0.9rem;
  transition: background 0.15s, color 0.15s;
}

.menu-item:hover {
  background: rgba(255, 255, 255, 0.15);
  color: #fff;
}

@media (max-width: 768px) {
  .menubar-wrapper :deep(.p-menubar) {
    padding: 0.4rem 0.75rem;
  }

  .app-brand {
    font-size: 1.05rem;
    margin-right: 0.75rem;
  }

  .menu-item {
    padding: 0.5rem 0.5rem;
    font-size: 0.85rem;
  }
}
</style>

<style>
.surface-ground {
  background: var(--p-surface-ground);
}

body {
  margin: 0;
  font-family:
    -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
}
</style>
