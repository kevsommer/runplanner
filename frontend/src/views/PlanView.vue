<template>
  <div class="flex justify-content-center">
    <div class="p-4 w-full md:w-8 lg:w-6" v-if="plan">
      <h1 class="text-3xl font-bold mb-4">{{ plan.name }}</h1>
      <p class="text-color-secondary mb-2">
        {{ formatDate(plan.startDate) }} - {{ formatDate(plan.endDate) }}
      </p>
      <p class="mb-4">Duration: {{ plan.weeks }} weeks</p>

      <Accordion :value="currentWeekValue">
        <AccordionPanel
          v-for="week in weeks"
          :key="week.number"
          :value="String(week.number - 1)"
        >
          <AccordionHeader>
            <div class="flex align-items-center gap-2 w-full">
              <span>Week {{ week.number }}</span>
              <i v-if="week.allDone" class="pi pi-check-circle text-green-500" />
              <Badge v-if="week.number - 1 === currentWeekIndex" value="Current" severity="info" />
              <span class="ml-auto text-sm text-color-secondary px-2">
                {{ week.doneKm.toFixed(0) }} / {{ week.plannedKm.toFixed(0) }} km
              </span>
            </div>
          </AccordionHeader>
          <AccordionContent>
            <div class="flex flex-column gap-3">
              <DayCard
                v-for="day in week.days"
                :key="day.date"
                :day-name="day.dayName"
                :date="day.date"
                :workouts="day.workouts"
                :plan-id="String(planId)"
                @workout-created="fetchWorkouts"
                @workout-updated="fetchWorkouts"
              />
            </div>
          </AccordionContent>
        </AccordionPanel>
      </Accordion>
    </div>
  </div>
</template>

<script setup lang="ts">
import { api } from "@/api";
import DayCard from "@/components/DayCard.vue";
import { useApi } from "@/composables/useApi";
import { type Workout } from "@/components/WorkoutCard.vue";
import { formatDate, formatDateISO } from "@/utils";
import Accordion from "primevue/accordion";
import AccordionPanel from "primevue/accordionpanel";
import AccordionHeader from "primevue/accordionheader";
import AccordionContent from "primevue/accordioncontent";
import Badge from "primevue/badge";
import { ref, computed } from "vue";
import { useRouter } from "vue-router";

const router = useRouter();

type Plan = {
  id: number;
  name: string;
  startDate: string;
  endDate: string;
  weeks: number;
};

type WeekDay = {
  date: string;
  dayName: string;
  workouts: Workout[];
};

type Week = {
  number: number;
  days: WeekDay[];
  plannedKm: number;
  doneKm: number;
  allDone: boolean;
};

const plan = ref<Plan>();
const workouts = ref<Workout[]>([]);

const planId = router.currentRoute.value.params.id;

const dayNames = [
  "Monday",
  "Tuesday",
  "Wednesday",
  "Thursday",
  "Friday",
  "Saturday",
  "Sunday",
];

const currentWeekIndex = computed<number | null>(() => {
  if (!plan.value) return null;

  const today = new Date();
  const startDate = new Date(plan.value.startDate);
  const diffTime = today.getTime() - startDate.getTime();
  const diffDays = Math.floor(diffTime / (1000 * 60 * 60 * 24));

  if (diffDays < 0) return 0;

  const weekIndex = Math.floor(diffDays / 7);
  if (weekIndex >= plan.value.weeks) return plan.value.weeks - 1;

  return weekIndex;
});

const currentWeekValue = computed<string | null>(() => {
  return currentWeekIndex.value !== null ? String(currentWeekIndex.value) : null;
});

const weeks = computed<Week[]>(() => {
  if (!plan.value) return [];

  const result: Week[] = [];
  const startDate = new Date(plan.value.startDate);

  for (let weekNum = 1; weekNum <= plan.value.weeks; weekNum++) {
    const days: WeekDay[] = [];

    for (let dayIndex = 0; dayIndex < 7; dayIndex++) {
      const currentDate = new Date(startDate);
      currentDate.setDate(startDate.getDate() + (weekNum - 1) * 7 + dayIndex);
      const dateStr = formatDateISO(currentDate);

      days.push({
        date: dateStr,
        dayName: dayNames[dayIndex],
        workouts: workouts.value.filter(
          (w) => w.day.substring(0, 10) === dateStr,
        ),
      });
    }

    const weekWorkouts = days.flatMap((d) => d.workouts);
    const plannedKm = weekWorkouts.reduce((sum, w) => sum + (w.distance || 0), 0);
    const doneKm = weekWorkouts.filter((w) => w.status === 'completed').reduce((sum, w) => sum + (w.distance || 0), 0);

    const allDone = weekWorkouts.length > 0 && weekWorkouts.every((w) => w.status === 'completed');

    result.push({ number: weekNum, days, plannedKm, doneKm, allDone });
  }

  return result;
});

const { exec: fetchTrainingPlan } = useApi({
  exec: () => api.get(`/plans/${planId}`),
  onSuccess: ({ data }) => {
    plan.value = data.plan;
  },
});

const { exec: fetchWorkouts } = useApi({
  exec: () => api.get(`/plans/${planId}/workouts`),
  onSuccess: ({ data }) => {
    workouts.value = data.workouts;
  },
});

fetchTrainingPlan();
fetchWorkouts();
</script>
