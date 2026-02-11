<template>
  <div class="surface-ground border-round p-2">
    <div class="flex justify-content-between align-items-center">
      <div>
        <Tag :value="formatRunType(workout.runType)" class="mr-2" />
        <span v-if="workout.distance">{{ workout.distance }} km</span>
      </div>
      <i v-if="workout.done" class="pi pi-check-circle text-green-500" />
    </div>
    <p v-if="workout.description" class="mt-1 mb-0 text-sm">
      {{ workout.description }}
    </p>
  </div>
</template>

<script setup lang="ts">
import Tag from "primevue/tag";

export type Workout = {
  id: string;
  planId: string;
  runType: string;
  day: string;
  description: string;
  notes: string;
  done: boolean;
  distance: number;
};

defineProps<{
  workout: Workout;
}>();

function formatRunType(runType: string): string {
  return runType.replace(/_/g, " ").replace(/\b\w/g, (c) => c.toUpperCase());
}
</script>
