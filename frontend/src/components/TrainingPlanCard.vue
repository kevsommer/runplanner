<template>
  <div
    class="surface-card border-round p-4 cursor-pointer training-plan-card"
    @click="router.push(`/plans/${plan.id}`)"
  >
    <div class="flex justify-content-between align-items-start">
      <div class="flex align-items-center gap-2">
        <h3 class="text-xl font-bold m-0">{{ plan.name }}</h3>
      </div>
      <div class="flex align-items-center gap-2">
        <Button
          icon="pi pi-verified"
          :severity="isSelectedActive ? 'warning' : 'secondary'"
          text
          rounded
          size="small"
          :loading="activateLoading"
          :aria-label="isSelectedActive ? 'Remove active plan' : 'Set as active plan'"
          class="activate-btn"
          @click.stop="toggleActivate"
        />
        <Button
          icon="pi pi-trash"
          severity="danger"
          text
          rounded
          size="small"
          class="delete-btn"
          aria-label="Delete plan"
          data-test="delete-button"
          @click.stop="deletePlan"
        />
      </div>
    </div>

    <div class="flex align-items-center gap-2 mb-3">
      <Badge
        v-if="isSelectedActive"
        value="Active"
        severity="success" />
      <Badge
        v-if="isCurrentlyRunning"
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
      <div class="flex align-items-center gap-2">
        <i class="pi pi-chart-line" />
        <span>{{ plan.totalDoneKm.toFixed(0) }} / {{ plan.totalPlannedKm.toFixed(0) }} km</span>
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
import { api } from "@/api";
import { useApi } from "@/composables/useApi";
import Badge from "primevue/badge";
import Button from "primevue/button";
import { useConfirm } from "primevue/useconfirm";
import ProgressBar from "primevue/progressbar";
import { computed } from "vue";
import { useRouter } from "vue-router";
import { formatDate } from "@/utils";

export type Plan = {
  id: string;
  name: string;
  startDate: string;
  endDate: string;
  weeks: number;
  totalPlannedKm: number;
  totalDoneKm: number;
};

const props = defineProps<{
  plan: Plan;
  activePlanId?: string | null;
}>();

const emit = defineEmits<{
  (e: "deleted"): void;
  (e: "activated", activePlanId: string | null): void;
}>();

const router = useRouter();
const confirm = useConfirm();

const { exec: deleteExec, loading: deleteLoading } = useApi({
  exec: () => api.delete(`/plans/${props.plan.id}`),
  onSuccess: () => emit("deleted"),
});

function deletePlan() {
  if (deleteLoading.value) return;
  confirm.require({
    message: "Are you sure you want to delete this training plan?",
    header: "Delete Training Plan",
    icon: "pi pi-exclamation-triangle",
    rejectProps: { label: "Cancel", severity: "secondary", outlined: true },
    acceptProps: { label: "Delete", severity: "danger" },
    accept: () => deleteExec(),
  });
}

const { exec: activateExec, loading: activateLoading } = useApi({
  exec: () => api.post(`/plans/${props.plan.id}/activate`),
  onSuccess: ({ data }) => emit("activated", data.activePlanId ?? null),
});

function toggleActivate() {
  if (activateLoading.value) return;
  activateExec();
}

const today = new Date();

const startDate = computed(() => new Date(props.plan.startDate));
const endDate = computed(() => new Date(props.plan.endDate));

const isSelectedActive = computed(() => props.activePlanId === props.plan.id);

const isCurrentlyRunning = computed(() => {
  return today >= startDate.value && today <= endDate.value;
});

const currentWeek = computed(() => {
  if (!isCurrentlyRunning.value) return 0;
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
