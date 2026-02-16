<template>
  <Card class="w-full md:w-6 lg:w-4">
    <template #title>Add workout</template>

    <template #content>
      <form @submit.prevent="onSubmit" class="flex flex-column gap-3">
        <Message v-if="error" severity="error" :closable="false" data-test="error-message">{{ error }}</Message>

        <div v-if="!initialDate" class="flex flex-column gap-2" data-test="day-field">
          <label for="day">Day</label>
          <DatePicker id="day" v-model="form.day" dateFormat="yy-mm-dd" showIcon />
        </div>

        <WorkoutFormFields :form="form" />

        <div class="flex gap-2">
          <Button type="submit" :loading label="Create workout" data-test="submit-button" />
          <Button type="button" label="Cancel" severity="secondary" text @click="emit('cancel')" data-test="cancel-button" />
        </div>
      </form>
    </template>
  </Card>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from "vue";
import DatePicker from "primevue/datepicker";
import Button from "primevue/button";
import Card from "primevue/card";
import Message from "primevue/message";
import WorkoutFormFields from "@/components/WorkoutFormFields.vue";
import { api } from "@/api";
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
const loading = ref(false);
const error = ref<string | null>(null);

function onSubmit() {
  const payload = {
    planId: props.planId,
    runType: form.runType,
    day: formatDateToYYYYMMDD(form.day),
    description: form.description,
    distance: distance.value,
  };

  loading.value = true;
  error.value = null;

  api
    .post("/workouts/", payload)
    .then(() => {
      emit("created");
    })
    .catch(() => {
      error.value = "Failed to create workout. Please try again.";
    })
    .finally(() => {
      loading.value = false;
    });
}
</script>
