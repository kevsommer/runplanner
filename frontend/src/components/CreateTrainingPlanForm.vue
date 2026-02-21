<template>
  <form
    class="flex flex-column gap-3"
    @submit.prevent="onSubmit">
    <div class="flex flex-column gap-2">
      <label>Mode</label>
      <SelectButton
        v-model="mode"
        :options="modeOptions"
        optionLabel="label"
        optionValue="value" />
    </div>

    <div class="flex flex-column gap-2">
      <label for="name">Name</label>
      <InputText
        id="name"
        v-model="form.name"
        type="text"
        placeholder="Training Plan Name" />
    </div>

    <div class="flex flex-column gap-2">
      <label for="endDate">End Date</label>
      <DatePicker
        id="endDate"
        v-model="form.endDate"
        dateFormat="yy-mm-dd"
        showIcon />
    </div>

    <div class="flex flex-column gap-2">
      <label class="text-color-secondary text-sm">Start Date</label>
      <span class="text-color-secondary">{{ startDate }}</span>
    </div>

    <div class="flex flex-column gap-2">
      <label for="weeks">Number of Weeks</label>
      <InputNumber
        id="weeks"
        v-model="form.weeks"
        show-buttons
        :min="mode === 'ai' ? 6 : 1"
        :max="30" />
    </div>

    <div class="flex flex-column gap-2">
      <label>Race Goal</label>
      <SelectButton
        v-model="form.raceGoal"
        :options="raceGoalOptions"
        optionLabel="label"
        optionValue="value" />
    </div>

    <template v-if="mode === 'ai'">
      <div class="flex flex-column gap-2">
        <label for="baseKm">Base km/week</label>
        <InputNumber
          id="baseKm"
          v-model="form.baseKmPerWeek"
          show-buttons
          :min="5"
          :max="200"
          suffix=" km" />
      </div>

      <div class="flex flex-column gap-2">
        <label for="runsPerWeek">Runs per week</label>
        <InputNumber
          id="runsPerWeek"
          v-model="form.runsPerWeek"
          show-buttons
          :min="2"
          :max="7" />
      </div>
    </template>

    <template v-else>
      <div class="flex flex-column gap-2">
        <label for="importJson">Import Workouts (JSON)</label>
        <Textarea
          id="importJson"
          v-model="form.importJson"
          rows="6"
          placeholder='{"workouts": [{"runType": "easy_run", "week": 1, "dayOfWeek": 1, "description": "", "distance": 8.0}]}' />
        <small class="text-color-secondary">
          Optional. Paste JSON to bulk-create workouts with the plan.
        </small>
      </div>
    </template>

    <Button
      type="submit"
      :loading="loading"
      :label="mode === 'ai' ? 'Generate Plan' : 'Create Plan'" />

    <div
      v-if="generateLoading"
      class="flex align-items-center gap-2 text-color-secondary">
      <i class="pi pi-spin pi-spinner" />
      <span>Generating your training plan with AI — this may take up to a minute...</span>
    </div>
  </form>
</template>

<script setup lang="ts">
import { reactive, ref, computed } from "vue";
import InputText from "primevue/inputtext";
import DatePicker from "primevue/datepicker";
import Button from "primevue/button";
import InputNumber from "primevue/inputnumber";
import Textarea from "primevue/textarea";
import SelectButton from "primevue/selectbutton";
import { useRouter } from "vue-router";
import { api } from "@/api";
import { useApi } from "@/composables/useApi";
import { formatDateToYYYYMMDD, calcStartDate } from "@/utils";

const router = useRouter();

const mode = ref<"ai" | "manual">("ai");
const modeOptions = [
  { label: "AI Generate", value: "ai" },
  { label: "Manual", value: "manual" },
];

const raceGoalOptions = [
  { label: "5K", value: "5k" },
  { label: "10K", value: "10k" },
  { label: "Half Marathon", value: "halfmarathon" },
  { label: "Marathon", value: "marathon" },
];

const form = reactive({
  name: "",
  endDate: new Date(),
  weeks: 10,
  importJson: "",
  baseKmPerWeek: 30,
  runsPerWeek: 4,
  raceGoal: "marathon",
});

const manualPayload = ref<Record<string, any>>({});

const { exec: submitManual, loading: manualLoading } = useApi({
  exec: () => api.post("/plans", manualPayload.value),
  successToast: "Training plan created",
  onSuccess: async ({ data }) => {
    const planId = data.plan.id;

    if (form.importJson.trim()) {
      try {
        const parsed = JSON.parse(form.importJson);
        await api.post(`/plans/${planId}/workouts/bulk`, parsed);
      } catch {
        // Plan was created, navigate anyway — user can fix workouts later
      }
    }

    router.push({ name: "plan", params: { id: planId } });
  },
});

const generatePayload = ref<Record<string, any>>({});

const { exec: submitGenerate, loading: generateLoading } = useApi({
  exec: () => api.post("/plans/generate", generatePayload.value),
  successToast: "Training plan generated",
  onSuccess: async ({ data }) => {
    const planId = data.plan.id;
    router.push({ name: "plan", params: { id: planId } });
  },
});

const loading = computed(() => manualLoading.value || generateLoading.value);

const startDate = computed(() => {
  const { endDate, weeks } = form;
  if (!endDate || !weeks) return "";
  return formatDateToYYYYMMDD(calcStartDate(endDate, weeks));
});

function onSubmit() {
  if (mode.value === "ai") {
    generatePayload.value = {
      name: form.name,
      endDate: formatDateToYYYYMMDD(form.endDate),
      weeks: form.weeks,
      baseKmPerWeek: form.baseKmPerWeek,
      runsPerWeek: form.runsPerWeek,
      raceGoal: form.raceGoal,
    };
    submitGenerate();
  } else {
    manualPayload.value = {
      name: form.name,
      endDate: formatDateToYYYYMMDD(form.endDate),
      weeks: form.weeks,
      raceGoal: form.raceGoal,
    };
    submitManual();
  }
}
</script>
