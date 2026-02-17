<template>
  <form
    class="flex flex-column gap-3 w-full"
    @submit.prevent="onSubmit">
    <div
      v-if="!initialDate"
      class="flex flex-column gap-2"
      data-test="day-field">
      <label for="day">Day</label>
      <DatePicker
        id="day"
        v-model="form.day"
        dateFormat="yy-mm-dd"
        showIcon />
    </div>

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
        :loading
        label="Create workout"
        data-test="submit-button" />
    </div>
  </form>
</template>

<script setup lang="ts">
import { computed, reactive } from "vue";
import DatePicker from "primevue/datepicker";
import Button from "primevue/button";
import WorkoutFormFields from "@/components/WorkoutFormFields.vue";
import { api } from "@/api";
import { useApi } from "@/composables/useApi";
import { formatDateToYYYYMMDD } from "@/utils";

const props = defineProps<{
  planId: string;
  initialDate?: string;
}>();

const emit = defineEmits<{
  (e: "created"): void;
  (e: "cancel"): void;
}>();

const form = reactive({
  day: props.initialDate ? new Date(props.initialDate + "T00:00:00") : new Date(),
  runType: "easy_run",
  description: "",
  distance: 5,
});

const distance = computed(() => form.runType === "strength_training" ? 0 : form.distance);

const { exec: submitWorkout, loading } = useApi({
  exec: (payload: Record<string, any>) => api.post("/workouts/", payload),
  successToast: "Workout created",
  onSuccess: () => {
    emit("created");
  },
});

function onSubmit() {
  submitWorkout({
    planId: props.planId,
    runType: form.runType,
    day: formatDateToYYYYMMDD(form.day),
    description: form.description,
    distance: distance.value,
  });
}
</script>
