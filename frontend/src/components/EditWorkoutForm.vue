<template>
  <Card class="w-full">
    <template #title>Edit workout</template>

    <template #content>
      <form @submit.prevent="onSubmit" class="flex flex-column gap-3">
        <WorkoutFormFields :form="form" />

        <div class="flex gap-2">
          <Button type="submit" :loading="loading" label="Save" data-test="submit-button" />
          <Button type="button" label="Cancel" severity="secondary" text @click="emit('cancel')" data-test="cancel-button" />
        </div>
      </form>
    </template>
  </Card>
</template>

<script setup lang="ts">
import { computed, reactive } from "vue";
import Button from "primevue/button";
import Card from "primevue/card";
import WorkoutFormFields from "@/components/WorkoutFormFields.vue";
import { api } from "@/api";
import { useApi } from "@/composables/useApi";
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

const { exec: submitEdit, loading } = useApi({
  exec: (payload: Record<string, any>) => api.put(`/workouts/${props.workout.id}`, payload),
  successToast: "Workout updated",
  onSuccess: () => {
    emit("updated");
  },
});

function onSubmit() {
  submitEdit({
    runType: form.runType,
    description: form.description,
    distance: distance.value,
  });
}
</script>
