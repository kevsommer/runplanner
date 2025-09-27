<template>
  <div class="flex justify-content-center">
    <CreateTrainingPlanForm v-if="formVisible" />
    <div v-else>
      <Button label="Create Training Plan" class="mb-4" @click="formVisible = true" />
      <div v-for="plan in plans" :key="plan.id" class="p-4 mb-4 border-1 border-400 border-round">
        <h2 class="text-2xl font-bold mb-2">{{ plan.name }}</h2>
        <p class="mb-1"><strong>Goal:</strong> {{ plan.goal }}</p>
        <p class="mb-1"><strong>Start Date:</strong> {{ plan.startDate }}</p>
        <p class="mb-1"><strong>Activities Per Week:</strong> {{ plan.activitiesPerWeek }}</p>
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
  goal: string;
  startDate: string;
  activitiesPerWeek: number;
};

const plans = ref<Plan[]>([]);

function fetchTrainingPlans() {
  api.get("/plans/").then((response) => {
    plans.value = response.data.plans;
  });
}

fetchTrainingPlans();
</script>
