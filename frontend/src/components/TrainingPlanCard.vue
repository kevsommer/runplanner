<template>
  <div
    class="surface-card border-round p-4 cursor-pointer training-plan-card"
    @click="router.push(`/plans/${plan.id}`)"
  >
    <div class="flex justify-content-between align-items-start mb-3">
      <h3 class="text-xl font-bold m-0">{{ plan.name }}</h3>
      <Badge
        v-if="isActive"
        :value="`Week ${currentWeek}`"
        severity="info" />
    </div>

    <div class="flex flex-column gap-2 text-color-secondary text-sm mb-3">
      <div class="flex align-items-center gap-2">
        <i class="pi pi-calendar" />
        <span>{{ formatDate(plan.startDate) }} - {{ formatDate(plan.endDate) }}</span>
      </div>
      <div class="flex align-items-center gap-2">
        <i class="pi pi-clock" />
        <span>{{ plan.weeks }} weeks</span>
      </div>
      <div
        v-if="daysRemaining !== null"
        class="flex align-items-center gap-2">
        <i class="pi pi-flag" />
        <span v-if="daysRemaining > 0">{{ daysRemaining }} days until race day</span>
        <span v-else-if="daysRemaining === 0">Race day is today!</span>
        <span v-else>Plan completed</span>
      </div>
    </div>

    <ProgressBar
      v-if="progressPercent !== null"
      :value="progressPercent"
      :showValue="false"
      style="height: 6px"
    />
  </div>
</template>

<script setup lang="ts">
import Badge from "primevue/badge";
import ProgressBar from "primevue/progressbar";
import { computed } from "vue";
import { useRouter } from "vue-router";
import { formatDate } from "@/utils";

export type Plan = {
  id: number;
  name: string;
  startDate: string;
  endDate: string;
  weeks: number;
};

const props = defineProps<{
  plan: Plan;
}>();

const router = useRouter();

const today = new Date();

const startDate = computed(() => new Date(props.plan.startDate));
const endDate = computed(() => new Date(props.plan.endDate));

const isActive = computed(() => {
  return today >= startDate.value && today <= endDate.value;
});

const currentWeek = computed(() => {
  if (!isActive.value) return 0;
  const msPerDay = 1000 * 60 * 60 * 24;
  const daysSinceStart = Math.floor((today.getTime() - startDate.value.getTime()) / msPerDay);
  return Math.floor(daysSinceStart / 7) + 1;
});

const daysRemaining = computed(() => {
  const msPerDay = 1000 * 60 * 60 * 24;
  return Math.ceil((endDate.value.getTime() - today.getTime()) / msPerDay);
});

const progressPercent = computed(() => {
  const totalMs = endDate.value.getTime() - startDate.value.getTime();
  if (totalMs <= 0) return null;
  const elapsedMs = today.getTime() - startDate.value.getTime();
  if (elapsedMs < 0) return 0;
  if (elapsedMs > totalMs) return 100;
  return Math.round((elapsedMs / totalMs) * 100);
});
</script>

<style scoped>
.training-plan-card {
  transition: transform 0.15s ease, box-shadow 0.15s ease;
  border: 1px solid var(--surface-border);
}

.training-plan-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}
</style>
