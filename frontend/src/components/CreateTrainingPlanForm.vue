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
          <label for="endDate">End Date</label>
          <DatePicker id="endDate" v-model="form.endDate" dateFormat="yy-mm-dd" showIcon />
        </div>

        <div class="flex flex-column gap-2">
          <label for="weeks">Number of Weeks</label>
          <InputNumber id="weeks" v-model="form.weeks" show-buttons :min="1" :max="30" />
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
import { useRouter } from "vue-router";
import { api } from "@/api";
import { formatDateToYYYYMMDD } from "@/utils";

const router = useRouter();

const form = reactive({
  name: "",
  endDate: new Date(),
  weeks: 10,
});

const loading = ref(false);
const error = ref<string | null>(null);

async function onSubmit() {
  loading.value = true;
  const payload = {
    ...form,
    endDate: formatDateToYYYYMMDD(form.endDate),
  };
  try {
    const { data } = await api.post("/plans/", payload);
    router.push({ name: "plan", params: { id: data.plan.id } });
  } catch {
    error.value = "Failed to create training plan. Please try again.";
  } finally {
    loading.value = false;
  }
}
</script>
