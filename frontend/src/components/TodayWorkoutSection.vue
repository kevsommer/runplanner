<template>
  <div
    v-if="plan"
    class="surface-card border-round p-4 mb-4"
    style="border: 1px solid var(--surface-border)"
  >
    <div class="flex justify-content-between align-items-center mb-3">
      <div>
        <h2 class="text-xl font-bold m-0">Today's Workout</h2>
        <div class="text-color-secondary text-sm mt-1">
          {{ todayLabel }} &middot;
          <span class="font-medium text-color">{{ plan.name }}</span>
        </div>
      </div>
      <Button
        label="View plan"
        icon="pi pi-arrow-right"
        iconPos="right"
        text
        size="small"
        @click="router.push(`/plans/${plan.id}`)"
      />
    </div>

    <div
      v-if="loading"
      class="flex flex-column gap-2"
    >
      <Skeleton height="2.5rem" border-radius="6px" />
      <Skeleton height="2.5rem" border-radius="6px" />
    </div>

    <div
      v-else-if="todayWorkouts.length > 0"
      class="flex flex-column gap-2"
    >
      <WorkoutCard
        v-for="workout in todayWorkouts"
        :key="workout.id"
        :workout="workout"
        @updated="fetchWorkouts"
        @edit="router.push(`/plans/${plan.id}`)"
      />
    </div>

    <div
      v-else
      class="flex align-items-center gap-3 text-color-secondary p-2"
    >
      <i class="pi pi-sun text-2xl" />
      <span>Rest day — no workouts scheduled today.</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { useRouter } from "vue-router";
import Button from "primevue/button";
import Skeleton from "primevue/skeleton";
import WorkoutCard from "@/components/WorkoutCard.vue";
import type { Workout } from "@/components/WorkoutCard.vue";
import { api } from "@/api";
import { useApi } from "@/composables/useApi";
import type { Plan } from "@/components/TrainingPlanCard.vue";

const props = defineProps<{
  plan: Plan | null;
}>();

const router = useRouter();

const today = new Date();

const todayStr = `${today.getFullYear()}-${String(today.getMonth() + 1).padStart(2, "0")}-${String(today.getDate()).padStart(2, "0")}`;

const todayLabel = today.toLocaleDateString("en-US", {
  weekday: "long",
  month: "long",
  day: "numeric",
});

const allWorkouts = ref<Workout[]>([]);

const todayWorkouts = computed(() =>
  allWorkouts.value.filter((w) => w.day.slice(0, 10) === todayStr)
);

const { exec: fetchWorkouts, loading } = useApi({
  exec: () => api.get(`/plans/${props.plan!.id}/workouts`),
  onSuccess: ({ data }) => {
    allWorkouts.value = data.workouts ?? [];
  },
});

watch(
  () => props.plan,
  (newPlan) => {
    if (newPlan) {
      fetchWorkouts();
    } else {
      allWorkouts.value = [];
    }
  },
  { immediate: true }
);
</script>
