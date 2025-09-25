<template>
  <div class="flex justify-content-center align-items-center min-h-screen surface-ground">
    <Card class="w-full sm:w-8 md:w-6 lg:w-4 p-4">
      <template #title>
        <h2 class="text-center">Logout</h2>
      </template>

      <template #content>
        <p>You have been logged out.</p>
        <p class="mt-2">
          <RouterLink to="/login">Login again</RouterLink>
        </p>
      </template>
    </Card>
  </div>
</template>

<script setup lang="ts">
import { useAuth } from "@/composables/useAuth";

const { logout } = useAuth();

logout();
</script>
