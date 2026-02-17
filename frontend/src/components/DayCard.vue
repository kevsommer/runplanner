<template>
  <div
    class="border-1 border-round surface-border p-3"
    :style="dragOver ? { backgroundColor: 'rgba(255, 255, 255, 0.05)' } : {}"
    @dragover.prevent="onDragOver"
    @dragenter.prevent
    @dragleave="onDragLeave"
    @drop="onDrop"
  >
    <div class="flex justify-content-between align-items-center mb-2">
      <div>
        <span class="font-semibold">{{ dayName }}</span>
        <span class="text-color-secondary ml-2">{{ formatDate(date) }}</span>
      </div>
      <Button
        v-if="!showForm"
        icon="pi pi-plus"
        text
        size="small"
        aria-label="Add workout"
        @click="showForm = true"
      />
    </div>

    <div
      v-if="workouts.length > 0"
      class="flex flex-column gap-2">
      <WorkoutCard
        v-for="workout in workouts"
        :key="workout.id"
        :workout="workout"
        @updated="emit('workoutUpdated')"
        @edit="editingWorkoutId = workout.id"
      />
    </div>
    <p
      v-else
      class="text-color-secondary text-sm mb-0">Rest day</p>

    <Dialog
      v-model:visible="showForm"
      header="Add workout"
      modal
      class="w-full md:w-6 lg:w-4">
      <CreateWorkoutForm
        :plan-id="planId"
        :initial-date="date"
        @created="handleCreated"
        @cancel="showForm = false"
      />
    </Dialog>

    <Dialog
      v-model:visible="showEditForm"
      header="Edit workout"
      modal
      class="w-full md:w-6 lg:w-4">
      <EditWorkoutForm
        v-if="editingWorkout"
        :workout="editingWorkout"
        @updated="handleWorkoutUpdated"
        @cancel="editingWorkoutId = null"
      />
    </Dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import Button from "primevue/button";
import Dialog from "primevue/dialog";
import CreateWorkoutForm from "@/components/CreateWorkoutForm.vue";
import EditWorkoutForm from "@/components/EditWorkoutForm.vue";
import WorkoutCard, { type Workout } from "@/components/WorkoutCard.vue";
import { formatDate } from "@/utils";
import { api } from "@/api";
import { useApi } from "@/composables/useApi";

const props = defineProps<{
  dayName: string;
  date: string;
  workouts: Workout[];
  planId: string;
}>();

const emit = defineEmits<{
  (e: "workoutCreated"): void;
  (e: "workoutUpdated"): void;
}>();

const showForm = ref(false);
const editingWorkoutId = ref<string | null>(null);
const dragOver = ref(false);

const editingWorkout = computed(() =>
  props.workouts.find((w) => w.id === editingWorkoutId.value) ?? null
);

const showEditForm = computed({
  get: () => editingWorkoutId.value !== null,
  set: (val: boolean) => { if (!val) editingWorkoutId.value = null; },
});

const { exec: moveWorkout } = useApi({
  exec: (workoutId: string) => api.put(`/workouts/${workoutId}`, { day: props.date }),
  onSuccess: () => {
    emit("workoutUpdated");
  },
});

function handleCreated() {
  showForm.value = false;
  emit("workoutCreated");
}

function handleWorkoutUpdated() {
  editingWorkoutId.value = null;
  emit("workoutUpdated");
}

function onDragOver(event: DragEvent) {
  event.dataTransfer!.dropEffect = "move";
  dragOver.value = true;
}

function onDragLeave(event: DragEvent) {
  const target = event.currentTarget as HTMLElement;
  if (!target.contains(event.relatedTarget as Node)) {
    dragOver.value = false;
  }
}

function onDrop(event: DragEvent) {
  dragOver.value = false;
  const workoutId = event.dataTransfer!.getData("workout-id");
  if (!workoutId) return;
  moveWorkout(workoutId);
}
</script>
