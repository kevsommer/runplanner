<template>
  <div
    class="surface-ground border-round p-2"
    :class="{ 'opacity-60': workout.done }"
    draggable="true"
    @dragstart="onDragStart"
  >
    <div class="flex justify-content-between align-items-center">
      <div>
        <Tag :value="formatRunType(workout.runType)" :severity="runTypeSeverity(workout.runType)" class="mr-2" />
        <span
          v-if="workout.distance"
          :class="{ 'line-through': workout.done }"
        >
          {{ workout.distance }} km
        </span>
      </div>
      <div class="flex align-items-center gap-2">
        <i
          class="pi pi-pencil cursor-pointer text-color-secondary"
          @click="emit('edit')"
        />
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

        <i
          class="pi cursor-pointer pi-trash text-red-500"
          @click="deleteWorkout" />
      </div>
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
  (e: "edit"): void;
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

function deleteWorkout() {
  if (loading.value) return;
  loading.value = true;
  api
    .delete(`/workouts/${props.workout.id}`)
    .then(() => {
      emit("updated");
    })
    .finally(() => {
      loading.value = false;
    });
}

function onDragStart(event: DragEvent) {
  event.dataTransfer!.effectAllowed = "move";
  event.dataTransfer!.setData("workout-id", props.workout.id);
}

function formatRunType(runType: string): string {
  return runType.replace(/_/g, " ").replace(/\b\w/g, (c) => c.toUpperCase());
}

function runTypeSeverity(runType: string): string | undefined {
  const severities: Record<string, string> = {
    easy_run: "success",
    long_run: "warn",
    intervals: "danger",
    tempo_run: "info",
  };
  return severities[runType];
}
</script>
