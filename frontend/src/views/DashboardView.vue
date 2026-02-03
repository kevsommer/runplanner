<template>
  <div class="flex justify-content-center">
    <CreateTrainingPlanForm v-if="formVisible" />
    <div v-else>
      <Button label="Create Training Plan" class="mb-4" @click="formVisible = true" />
      <div v-for="plan in plans" :key="plan.id" class="p-4 mb-4 border-1 border-400 border-round">
        <h2 class="text-2xl font-bold mb-2">{{ plan.name }}</h2>
        <p class="mb-1"><strong>Race Date:</strong> {{ formattedDate(plan.endDate) }}</p>
        <p class="mb-1"><strong>Weeks</strong> {{ plan.weeks }}</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { api } from "@/api";
import CreateTrainingPlanForm from "@/components/CreateTrainingPlanForm.vue";
import Button from "primevue/button";
import { ref } from "vue";

const formVisible = ref(false);

type Plan = {
  id: number;
  name: string;
  startDate: string;
  endDate: string;
  weeks: number;
};

const plans = ref<Plan[]>([]);

function formattedDate(dateStr: string): string {
  const date = new Date(dateStr);
  return date.toLocaleDateString();
}

function fetchTrainingPlans() {
  api.get("/plans").then((response) => {
    plans.value = response.data.plans;
  });
}

fetchTrainingPlans();
</script>
