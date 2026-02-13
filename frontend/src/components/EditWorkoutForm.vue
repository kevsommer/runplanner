<template>
  <Card class="w-full">
    <template #title>Edit workout</template>

    <template #content>
      <form @submit.prevent="onSubmit" class="flex flex-column gap-3">
        <Message v-if="error" severity="error" :closable="false">{{ error }}</Message>

        <WorkoutFormFields :form="form" />

        <div class="flex gap-2">
          <Button type="submit" :loading="loading" label="Save" />
          <Button type="button" label="Cancel" severity="secondary" text @click="emit('cancel')" />
        </div>
      </form>
    </template>
  </Card>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from "vue";
import Button from "primevue/button";
import Card from "primevue/card";
import Message from "primevue/message";
import WorkoutFormFields from "@/components/WorkoutFormFields.vue";
import { api } from "@/api";
import type { Workout } from "@/components/WorkoutCard.vue";

const props = defineProps<{
  workout: Workout;
}>();

const emit = defineEmits<{
  (e: "updated"): void;
  (e: "cancel"): void;
}>();

const form = reactive({
  runType: props.workout.runType,
  description: props.workout.description,
  distance: props.workout.distance,
});

const distance = computed(() => form.runType === "strength_training" ? 0 : form.distance);

const loading = ref(false);
const error = ref<string | null>(null);

function onSubmit() {
  loading.value = true;
  error.value = null;

  api
    .put(`/workouts/${props.workout.id}`, {
      runType: form.runType,
      description: form.description,
      distance: distance.value,
    })
    .then(() => {
      emit("updated");
    })
    .catch(() => {
      error.value = "Failed to update workout. Please try again.";
    })
    .finally(() => {
      loading.value = false;
    });
}
</script>
