<template>
  <form
    class="flex flex-column gap-3 w-full"
    @submit.prevent="onSubmit">
    <WorkoutFormFields :form="form" />

    <div class="flex gap-2 justify-content-end">
      <Button
        type="button"
        label="Cancel"
        severity="secondary"
        text
        data-test="cancel-button"
        @click="emit('cancel')" />
      <Button
        type="submit"
        :loading="loading"
        label="Save"
        data-test="submit-button" />
    </div>
  </form>
</template>

<script setup lang="ts">
import { computed, reactive } from "vue";
import Button from "primevue/button";
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
