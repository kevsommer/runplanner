<template>
  <div
    class="surface-ground border-round p-2"
    :class="{ 'opacity-60': workout.done }"
  >
    <div class="flex justify-content-between align-items-center">
      <div>
        <Tag :value="formatRunType(workout.runType)" class="mr-2" />
        <span
          v-if="workout.distance"
          :class="{ 'line-through': workout.done }"
        >
          {{ workout.distance }} km
        </span>
      </div>
      <i
        class="pi cursor-pointer"
        :class="
          workout.done
            ? 'pi-check-circle text-green-500'
            : 'pi-circle text-color-secondary'
        "
        :style="{ opacity: loading ? 0.5 : 1 }"
        @click="toggleDone"
      />
    </div>
    <p
      v-if="workout.description"
      class="mt-1 mb-0 text-sm"
      :class="{ 'line-through': workout.done }"
    >
      {{ workout.description }}
    </p>
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue";
import Tag from "primevue/tag";
import { api } from "@/api";

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

const props = defineProps<{
  workout: Workout;
}>();

const emit = defineEmits<{
  (e: "updated"): void;
}>();

const loading = ref(false);

function toggleDone() {
  if (loading.value) return;
  loading.value = true;
  api
    .put(`/workouts/${props.workout.id}`, { done: !props.workout.done })
    .then(() => {
      emit("updated");
    })
    .finally(() => {
      loading.value = false;
    });
}

function formatRunType(runType: string): string {
  return runType.replace(/_/g, " ").replace(/\b\w/g, (c) => c.toUpperCase());
}
</script>
