<template>
  <div
    class="surface-ground border-round p-2 workout-card"
    :class="{
      'workout-done': workout.status === 'completed',
      'workout-skipped': workout.status === 'skipped',
      'cursor-move': allowDrag && workout.status === 'pending',
    }"
    :draggable="allowDrag && workout.status === 'pending'"
    @dragstart="onDragStart"
  >
    <div class="flex justify-content-between align-items-center">
      <div>
        <Tag
          :value="formatRunType(workout.runType)"
          :severity="runTypeSeverity(workout.runType)"
          class="mr-2" />
        <span
          v-if="workout.distance"
          :class="{ 'line-through': workout.status === 'completed' || workout.status === 'skipped' }"
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
          v-if="workout.status !== 'skipped'"
          class="pi cursor-pointer ml-2"
          :class="
            workout.status === 'completed'
              ? 'pi-check-circle text-green-500'
              : 'pi-circle text-color-secondary'
          "
          :style="{ opacity: loading ? 0.5 : 1 }"
          data-test="complete-button"
          @click="toggleDone"
        />
        <i
          v-if=" workout.status !== 'completed'"
          class="pi pi-forward cursor-pointer ml-2"
          :class="workout.status === 'skipped' ? 'text-orange-500' : 'text-color-secondary'"
          :style="{ opacity: loading ? 0.5 : 1 }"
          data-test="skip-button"
          @click="skipWorkout"
        />
        <i
          class="pi cursor-pointer pi-trash text-red-500 ml-2"
          data-test="delete-button"
          @click="deleteWorkout" />
      </div>
    </div>
    <p
      v-if="workout.description"
      class="p-2 mt-1 mb-0 text-sm text-color-secondary"
      style="white-space: pre-line"
    >
      {{ workout.description }}
    </p>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import Tag from "primevue/tag";
import { useConfirm } from "primevue/useconfirm";
import { api } from "@/api";
import { useApi } from "@/composables/useApi";

export type Workout = {
  id: string;
  planId: string;
  runType: string;
  day: string;
  description: string;
  notes: string;
  status: 'completed' | 'pending' | 'skipped';
  distance: number;
};

const props = withDefaults(defineProps<{
  workout: Workout;
  allowDrag?: boolean;
}>(), { allowDrag: true });

const emit = defineEmits<{
  (e: "updated"): void;
  (e: "edit"): void;
}>();

const confirm = useConfirm();

const { exec: toggleDoneExec, loading: toggleDoneLoading } = useApi({
  exec: (newStatus: string) => api.put(`/workouts/${props.workout.id}`, { status: newStatus }),
  onSuccess: () => emit("updated"),
});

const { exec: skipExec, loading: skipLoading } = useApi({
  exec: (newStatus: string) => api.put(`/workouts/${props.workout.id}`, { status: newStatus }),
  onSuccess: () => emit("updated"),
});

const { exec: deleteExec, loading: deleteLoading } = useApi({
  exec: () => api.delete(`/workouts/${props.workout.id}`),
  onSuccess: () => emit("updated"),
});

const loading = computed(() => toggleDoneLoading.value || skipLoading.value || deleteLoading.value);

function toggleDone() {
  if (loading.value) return;
  const newStatus = props.workout.status === 'pending' ? 'completed' : 'pending';
  toggleDoneExec(newStatus);
}

function skipWorkout() {
  if (loading.value) return;
  const newStatus = props.workout.status === 'skipped' ? 'pending' : 'skipped';
  skipExec(newStatus);
}

function deleteWorkout() {
  if (loading.value) return;
  confirm.require({
    message: "Are you sure you want to delete this workout?",
    header: "Delete Workout",
    icon: "pi pi-exclamation-triangle",
    rejectProps: { label: "Cancel", severity: "secondary", outlined: true },
    acceptProps: { label: "Delete", severity: "danger" },
    accept: () => deleteExec(),
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
    strength_training: "secondary",
    race: "contrast",
  };
  return severities[runType];
}
</script>

<style scoped>
.workout-skipped {
  opacity: 0.6;
  border-left: 3px dashed var(--p-orange-500);
}
</style>
