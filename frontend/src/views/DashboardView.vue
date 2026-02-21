<template>
  <div class="flex justify-content-center p-4">
    <div
      class="w-full"
      style="max-width: 900px">
      <div class="flex justify-content-between align-items-center mb-4">
        <h1 class="text-2xl font-bold m-0">Your Training Plans</h1>
        <Button
          label="Create Training Plan"
          icon="pi pi-plus"
          @click="formVisible = true"
        />
      </div>

      <TodayWorkoutSection :plan="activePlan" />

      <div
        v-if="plans.length > 0"
        class="grid">
        <div
          v-for="plan in plans"
          :key="plan.id"
          class="col-12 md:col-6">
          <TrainingPlanCard
            :plan="plan"
            :activePlanId="activePlanId"
            @deleted="fetchTrainingPlans"
            @activated="(id) => setActivePlanId(id)" />
        </div>
      </div>
      <p
        v-else
        class="text-color-secondary text-center mt-5">
        No training plans yet. Create one to get started!
      </p>

      <Dialog
        v-model:visible="formVisible"
        header="Create Training Plan"
        modal
        class="w-full md:w-8 lg:w-5">
        <CreateTrainingPlanForm />
      </Dialog>
    </div>
  </div>
</template>

<script setup lang="ts">
import { api } from "@/api";
import CreateTrainingPlanForm from "@/components/CreateTrainingPlanForm.vue";
import TodayWorkoutSection from "@/components/TodayWorkoutSection.vue";
import TrainingPlanCard from "@/components/TrainingPlanCard.vue";
import type { Plan } from "@/components/TrainingPlanCard.vue";
import { useApi } from "@/composables/useApi";
import { useAuth } from "@/composables/useAuth";
import Button from "primevue/button";
import Dialog from "primevue/dialog";
import { computed, ref } from "vue";

const { user, setActivePlanId } = useAuth();
const activePlanId = computed(() => user.value?.user?.activePlanId ?? null);

const formVisible = ref(false);

const plans = ref<Plan[]>([]);
const activePlan = computed(() => plans.value.find((p) => p.id === activePlanId.value) ?? null);

const { exec: fetchTrainingPlans } = useApi({
  exec: () => api.get("/plans"),
  onSuccess: ({ data }) => {
    plans.value = data.plans || [];
  },
});

fetchTrainingPlans();
</script>
