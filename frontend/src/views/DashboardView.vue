<template>
  <div class="flex justify-content-center">
    <CreateTrainingPlanForm v-if="formVisible" />
    <Button v-else label="Create Training Plan" class="mb-4" @click="formVisible = true" />
  </div>
</template>

<script setup lang="ts">
import CreateTrainingPlanForm from "@/components/CreateTrainingPlanForm.vue";
import Button from "primevue/button";
import { ref } from "vue";

const formVisible = ref(false);
</script>
