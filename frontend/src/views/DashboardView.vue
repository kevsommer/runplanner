<template>
  <div class="flex justify-content-center">
    <div class="w-full" style="max-width: 900px">
      <div class="flex justify-content-between align-items-center mb-4">
        <h1 class="text-2xl font-bold m-0">Your Training Plans</h1>
        <Button
          v-if="!formVisible"
          label="Create Training Plan"
          icon="pi pi-plus"
          @click="formVisible = true"
        />
      </div>

      <CreateTrainingPlanForm v-if="formVisible" />

      <template v-else>
        <div v-if="plans.length > 0" class="grid">
          <div v-for="plan in plans" :key="plan.id" class="col-12 md:col-6">
            <TrainingPlanCard :plan="plan" />
          </div>
        </div>
        <p v-else class="text-color-secondary text-center mt-5">
          No training plans yet. Create one to get started!
        </p>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { api } from "@/api";
import CreateTrainingPlanForm from "@/components/CreateTrainingPlanForm.vue";
import TrainingPlanCard from "@/components/TrainingPlanCard.vue";
import type { Plan } from "@/components/TrainingPlanCard.vue";
import { useApi } from "@/composables/useApi";
import Button from "primevue/button";
import { ref } from "vue";

const formVisible = ref(false);

const plans = ref<Plan[]>([]);

const { exec: fetchTrainingPlans } = useApi({
  exec: () => api.get("/plans"),
  onSuccess: ({ data }) => {
    plans.value = data.plans;
  },
});

fetchTrainingPlans();
</script>
