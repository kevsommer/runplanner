<template>
  <Card class="w-full md:w-6 lg:w-4">
    <template #title>Add workout</template>

    <template #content>
      <form @submit.prevent="onSubmit" class="flex flex-column gap-3">
        <Message v-if="error" severity="error" :closable="false">{{ error }}</Message>

        <div v-if="!initialDate" class="flex flex-column gap-2">
          <label for="day">Day</label>
          <DatePicker id="day" v-model="form.day" dateFormat="yy-mm-dd" showIcon />
        </div>

        <div class="flex flex-column gap-2">
          <label for="runType">Run type</label>
          <SelectButton
            id="runType"
            v-model="form.runType"
            :options="runTypeOptions"
            optionLabel="label"
            optionValue="value"
          />
        </div>

        <div class="flex flex-column gap-2">
          <label for="description">Description</label>
          <InputText
            id="description"
            v-model="form.description"
            type="text"
            placeholder="Workout description"
          />
        </div>

        <div class="flex flex-column gap-2">
          <label for="distance">Distance (km)</label>
          <InputNumber
            id="distance"
            v-model="form.distance"
            :min="0"
            :step="0.5"
            mode="decimal"
            :maxFractionDigits="2"
          />
        </div>

        <div class="flex gap-2">
          <Button type="submit" :loading label="Create workout" />
          <Button type="button" label="Cancel" severity="secondary" text @click="emit('cancel')" />
        </div>
      </form>
    </template>
  </Card>
</template>

<script setup lang="ts">
import { reactive, ref } from "vue";
import InputText from "primevue/inputtext";
import DatePicker from "primevue/datepicker";
import Button from "primevue/button";
import SelectButton from "primevue/selectbutton";
import Card from "primevue/card";
import InputNumber from "primevue/inputnumber";
import Message from "primevue/message";
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

const runTypeOptions = [
  { label: "Easy run", value: "easy_run" },
  { label: "Intervals", value: "intervals" },
  { label: "Long run", value: "long_run" },
  { label: "Tempo run", value: "tempo_run" },
];

const form = reactive({
  day: props.initialDate ? new Date(props.initialDate + "T00:00:00") : new Date(),
  runType: runTypeOptions[0].value as string,
  description: "",
  distance: 5,
});

const loading = ref(false);
const error = ref<string | null>(null);

function onSubmit() {
  const payload = {
    planId: props.planId,
    runType: form.runType,
    day: formatDateToYYYYMMDD(form.day),
    description: form.description,
    distance: form.distance,
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
