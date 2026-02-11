<template>
  <div class="flex justify-content-center">
    <div class="p-4 w-full md:w-8 lg:w-6" v-if="plan">
      <h1 class="text-3xl font-bold mb-4">{{ plan.name }}</h1>
      <p class="text-color-secondary mb-2">
        {{ formatDate(plan.startDate) }} - {{ formatDate(plan.endDate) }}
      </p>
      <p class="mb-4">Duration: {{ plan.weeks }} weeks</p>

      <Accordion :multiple="false" :activeIndex="currentWeekIndex">
        <AccordionTab
          v-for="week in weeks"
          :key="week.number"
          :header="`Week ${week.number}`"
        >
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
        </AccordionTab>
      </Accordion>
    </div>
  </div>
</template>

<script setup lang="ts">
import { api } from "@/api";
import DayCard from "@/components/DayCard.vue";
import { type Workout } from "@/components/WorkoutCard.vue";
import { formatDate, formatDateISO } from "@/utils";
import Accordion from "primevue/accordion";
import AccordionTab from "primevue/accordiontab";
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

    result.push({ number: weekNum, days });
  }

  return result;
});

function fetchTrainingPlan() {
  api.get(`/plans/${planId}`).then((response) => {
    plan.value = response.data.plan;
  });
}

function fetchWorkouts() {
  api.get(`/plans/${planId}/workouts`).then((response) => {
    workouts.value = response.data.workouts;
  });
}

fetchTrainingPlan();
fetchWorkouts();
</script>
