<template>
  <div class="flex justify-content-center">
    <div class="p-4">
      <h1 class="text-3xl font-bold mb-4">{{ plan.name }}</h1>
      <p>{{ plan.endDate }}</p>
      <p>Duration: {{ plan.weeks }} weeks</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { api } from "@/api";
import { ref } from "vue";
import { useRouter } from "vue-router";

const router = useRouter();

const formVisible = ref(false);

type Plan = {
  id: number;
  name: string;
  startDate: string;
  endDate: string;
  weeks: number;
};

const plan = ref<Plan>();

const planId = router.currentRoute.value.params.id;

function fetchTrainingPlan() {
  api.get(`/plans/${planId}`).then((response) => {
    plan.value = response.data.plan;
  });
}

fetchTrainingPlan();
</script>
