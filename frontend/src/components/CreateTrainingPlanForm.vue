<template>
  <Card class="w-full md:w-6 lg:w-4">
    <template #title>Create a new training plan</template>

    <template #content>
      <form @submit.prevent="onSubmit" class="flex flex-column gap-3">
        <Message v-if="error" severity="error" :closable="false">{{ error }}</Message>

        <div class="flex flex-column gap-2">
          <label for="name">Name</label>
          <InputText id="name" v-model="form.name" type="text" placeholder="Training Plan Name" />
        </div>

        <div class="flex flex-column gap-2">
          <label for="goal">Goal</label>
          <SelectButton
            id="goal"
            v-model="form.goal"
            :options="['MARATHON', 'HALF', '10K', '5K']"
          />
        </div>

        <div class="flex flex-column gap-2">
          <label for="start_date">Start Date</label>
          <DatePicker id="start_date" v-model="form.start_date" dateFormat="yy-mm-dd" showIcon />
        </div>

        <div class="flex flex-column gap-2">
          <label for="number_of_weeks">Number of Weeks</label>
          <InputNumber
            id="number_of_weeks"
            v-model="form.number_of_weeks"
            show-buttons
            :min="1"
            :max="30"
          />
        </div>

        <div class="flex flex-column gap-2">
          <label for="activities_per_week">Activities Per Week</label>
          <InputNumber
            id="activities_per_week"
            v-model="form.activities_per_week"
            show-buttons
            :min="1"
            :max="7"
          />
        </div>

        <Button type="submit" :loading="loading" label="Submit" />
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

const form = reactive({
  name: "",
  goal: "MARATHON",
  start_date: new Date(),
  number_of_weeks: 10,
  activities_per_week: 4,
});

const loading = ref(false);
const error = ref<string | null>(null);

function onSubmit() {
  api
    .post("/plans/", form)
    .catch(() => {
      error.value = "Failed to create training plan. Please try again.";
    })
    .finally(() => {
      loading.value = false;
    });
  return;
}
</script>
