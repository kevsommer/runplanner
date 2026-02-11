<template>
  <div class="border-1 border-round surface-border p-3">
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
        @click="showForm = true"
        aria-label="Add workout"
      />
    </div>

    <div v-if="workouts.length > 0" class="flex flex-column gap-2">
      <WorkoutCard v-for="workout in workouts" :key="workout.id" :workout="workout" />
    </div>
    <p v-else class="text-color-secondary text-sm mb-0">Rest day</p>

    <div v-if="showForm" class="mt-3">
      <CreateWorkoutForm
        :plan-id="planId"
        :initial-date="date"
        @created="handleCreated"
        @cancel="showForm = false"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue";
import Button from "primevue/button";
import CreateWorkoutForm from "@/components/CreateWorkoutForm.vue";
import WorkoutCard, { type Workout } from "@/components/WorkoutCard.vue";
import { formatDate } from "@/utils";

defineProps<{
  dayName: string;
  date: string;
  workouts: Workout[];
  planId: string;
}>();

const emit = defineEmits<{
  (e: "workoutCreated"): void;
}>();

const showForm = ref(false);

function handleCreated() {
  showForm.value = false;
  emit("workoutCreated");
}
</script>
