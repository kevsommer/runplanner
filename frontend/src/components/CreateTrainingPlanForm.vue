<template>
  <form
    class="flex flex-column gap-3"
    @submit.prevent="onSubmit">
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
      <label for="weeks">Number of Weeks</label>
      <InputNumber
        id="weeks"
        v-model="form.weeks"
        show-buttons
        :min="1"
        :max="30" />
    </div>

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

    <Button
      type="submit"
      :loading="loading"
      label="Submit" />
  </form>
</template>

<script setup lang="ts">
import { reactive, ref } from "vue";
import InputText from "primevue/inputtext";
import DatePicker from "primevue/datepicker";
import Button from "primevue/button";
import InputNumber from "primevue/inputnumber";
import Textarea from "primevue/textarea";
import { useRouter } from "vue-router";
import { api } from "@/api";
import { useApi } from "@/composables/useApi";
import { formatDateToYYYYMMDD } from "@/utils";

const router = useRouter();

const form = reactive({
  name: "",
  endDate: new Date(),
  weeks: 10,
  importJson: "",
});

const payload = ref<Record<string, any>>({});

const { exec: submitPlan, loading } = useApi({
  exec: () => api.post("/plans/", payload.value),
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

function onSubmit() {
  payload.value = {
    ...form,
    endDate: formatDateToYYYYMMDD(form.endDate),
  };
  submitPlan();
}
</script>
