<template>
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
    <Textarea
      id="description"
      v-model="form.description"
      placeholder="Workout description"
      rows="5"
      autoResize
    />
  </div>

  <div
    v-if="form.runType !== 'strength_training'"
    class="flex flex-column gap-2">
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
</template>

<script setup lang="ts">
import Textarea from "primevue/textarea";
import SelectButton from "primevue/selectbutton";
import InputNumber from "primevue/inputnumber";

const runTypeOptions = [
  { label: "Easy run", value: "easy_run" },
  { label: "Intervals", value: "intervals" },
  { label: "Long run", value: "long_run" },
  { label: "Tempo run", value: "tempo_run" },
  { label: "Strength", value: "strength_training" },
  { label: "Race", value: "race" },
];

defineProps<{
  form: {
    runType: string;
    description: string;
    distance: number;
  };
}>();
</script>
