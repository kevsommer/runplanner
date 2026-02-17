<template>
  <div class="flex justify-content-center p-3 md:p-4">
    <div
      v-if="plan"
      class="w-full md:w-8 lg:w-6">
      <div class="flex align-items-center justify-content-between mb-3 gap-3">
        <div>
          <h1 class="text-xl font-bold mb-1">{{ plan.name }}</h1>
          <span class="text-color-secondary text-sm">
            {{ formatDate(plan.startDate) }} - {{ formatDate(plan.endDate) }} · {{ plan.weeks }} weeks
          </span>
        </div>
        <Select
          v-model="selectedWeekIndex"
          :options="weekOptions"
          option-label="label"
          option-value="value"
          @change="onWeekSelected"
        />
      </div>

      <WeeklyKmChart
        :weeks-summary="plan.weeksSummary"
        :current-week-index="currentWeekIndex"
        class="mb-3"
        @week-selected="onChartWeekSelected"
      />

      <div class="flex align-items-center gap-2 mb-3">
        <Button
          icon="pi pi-chevron-left"
          class="px-4"
          :severity="isFirstWeek ? 'secondary' : 'primary'"
          outlined
          @click="goToPreviousWeek"
        />
        <span class="font-semibold text-lg">Week {{ selectedWeek.number }}</span>
        <i
          v-if="selectedWeek.allDone"
          class="pi pi-check-circle text-green-500" />
        <Badge
          v-if="selectedWeekIndex === currentWeekIndex"
          value="Current"
          severity="info" />
        <span class="ml-auto text-sm text-color-secondary">
          {{ selectedWeek.doneKm.toFixed(0) }} / {{ selectedWeek.plannedKm.toFixed(0) }} km
        </span>
        <Button
          icon="pi pi-chevron-right"
          class="px-4"
          :severity="isLastWeek ? 'secondary' : 'primary'"
          outlined
          @click="goToNextWeek"
        />
      </div>

      <Transition
        :name="slideDirection === 'up' ? 'slide-up' : 'slide-down'"
        mode="out-in">
        <div
          :key="selectedWeekIndex"
          class="flex flex-column gap-3">
          <DayCard
            v-for="day in selectedWeek.days"
            :key="day.date"
            :day-name="day.dayName"
            :date="day.date"
            :workouts="day.workouts"
            :plan-id="String(planId)"
            @workout-created="fetchTrainingPlan"
            @workout-updated="fetchTrainingPlan"
          />
        </div>
      </Transition>

    </div>
  </div>
</template>

<script setup lang="ts">
import { api } from "@/api";
import DayCard from "@/components/DayCard.vue";
import WeeklyKmChart from "@/components/WeeklyKmChart.vue";
import { useApi } from "@/composables/useApi";
import { type Workout } from "@/components/WorkoutCard.vue";
import { formatDate } from "@/utils";
import Select from "primevue/select";
import Button from "primevue/button";
import Badge from "primevue/badge";
import { ref, computed } from "vue";
import { useRouter } from "vue-router";

const router = useRouter();

type WeekDay = {
  date: string;
  dayName: string;
  workouts: Workout[];
};

type WeekSummary = {
  number: number;
  days: WeekDay[];
  plannedKm: number;
  doneKm: number;
  allDone: boolean;
};

type Plan = {
  id: string;
  name: string;
  startDate: string;
  endDate: string;
  weeks: number;
  weeksSummary: WeekSummary[];
};

const plan = ref<Plan>();
const selectedWeekIndex = ref(0);
const slideDirection = ref<"up" | "down">("up");
const initialLoadDone = ref(false);

const planId = router.currentRoute.value.params.id;

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

const selectedWeek = computed<WeekSummary>(() => {
  return plan.value!.weeksSummary[selectedWeekIndex.value];
});

const weekOptions = computed(() => {
  if (!plan.value) return [];
  return plan.value.weeksSummary.map((week, index) => ({
    label: `Week ${week.number}`,
    value: index,
  }));
});

const isFirstWeek = computed(() => selectedWeekIndex.value === 0);

const isLastWeek = computed(() => {
  if (!plan.value) return true;
  return selectedWeekIndex.value === plan.value.weeksSummary.length - 1;
});

function goToPreviousWeek() {
  if (isFirstWeek.value) return;
  slideDirection.value = "down";
  selectedWeekIndex.value--;
}

function goToNextWeek() {
  if (isLastWeek.value) return;
  slideDirection.value = "up";
  selectedWeekIndex.value++;
}

function onWeekSelected(event: { value: number }) {
  const newIndex = event.value;
  slideDirection.value = newIndex > selectedWeekIndex.value ? "up" : "down";
  selectedWeekIndex.value = newIndex;
}

function onChartWeekSelected(index: number) {
  slideDirection.value = index > selectedWeekIndex.value ? "up" : "down";
  selectedWeekIndex.value = index;
}

const { exec: fetchTrainingPlan } = useApi({
  exec: () => api.get(`/plans/${planId}`),
  onSuccess: ({ data }) => {
    plan.value = data.plan;
    if (!initialLoadDone.value) {
      initialLoadDone.value = true;
      selectedWeekIndex.value = currentWeekIndex.value ?? 0;
    }
  },
});

fetchTrainingPlan();
</script>

<style scoped>
.slide-up-enter-active,
.slide-up-leave-active,
.slide-down-enter-active,
.slide-down-leave-active {
  transition: all 0.3s ease;
}

.slide-up-enter-from {
  opacity: 0;
  transform: translateY(30px);
}

.slide-up-leave-to {
  opacity: 0;
  transform: translateY(-30px);
}

.slide-down-enter-from {
  opacity: 0;
  transform: translateY(-30px);
}

.slide-down-leave-to {
  opacity: 0;
  transform: translateY(30px);
}
</style>
