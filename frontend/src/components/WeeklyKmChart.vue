<template>
  <Chart
    type="line"
    :data="chartData"
    :options="chartOptions"
    class="weekly-km-chart"
    @select="onChartClick"
  />
</template>

<script setup lang="ts">
import Chart from "primevue/chart";
import { computed } from "vue";

type WeekSummary = {
  number: number;
  plannedKm: number;
  doneKm: number;
};

const props = defineProps<{
  weeksSummary: WeekSummary[];
  currentWeekIndex: number | null;
}>();

const emit = defineEmits<{
  weekSelected: [index: number];
}>();

const chartData = computed(() => {
  const labels = props.weeksSummary.map((w) => `W${w.number}`);
  const planned = props.weeksSummary.map((w) => Math.round(w.plannedKm));
  const done = props.weeksSummary.map((w) => Math.round(w.doneKm));

  return {
    labels,
    datasets: [
      {
        label: "Planned km",
        data: planned,
        borderColor: "rgba(148, 163, 184, 0.8)",
        backgroundColor: "rgba(148, 163, 184, 0.15)",
        borderWidth: 2,
        pointRadius: 3,
        pointHoverRadius: 5,
        tension: 0,
        fill: true,
        order: 2,
      },
      {
        label: "Done km",
        data: done,
        borderColor: "rgb(34, 197, 94)",
        backgroundColor: "rgba(34, 197, 94, 0.15)",
        borderWidth: 2,
        pointRadius: 3,
        pointHoverRadius: 5,
        tension: 0,
        fill: true,
        order: 1,
      },
    ],
  };
});

const chartOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  interaction: {
    mode: "index" as const,
    intersect: false,
  },
  plugins: {
    legend: {
      display: true,
      position: "bottom" as const,
      labels: {
        boxWidth: 12,
        padding: 12,
        usePointStyle: true,
      },
    },
    tooltip: {
      callbacks: {
        label: (ctx: { dataset: { label: string }; parsed: { y: number } }) =>
          `${ctx.dataset.label}: ${ctx.parsed.y} km`,
      },
    },
  },
  scales: {
    x: {
      grid: { display: false },
    },
    y: {
      beginAtZero: true,
      ticks: {
        callback: (value: number) => `${value}`,
      },
      grid: {
        color: "rgba(148, 163, 184, 0.15)",
      },
    },
  },
  onClick: (_event: unknown, elements: { index: number }[]) => {
    if (elements.length > 0) {
      emit("weekSelected", elements[0].index);
    }
  },
}));

function onChartClick(event: { element: { index: number } }) {
  emit("weekSelected", event.element.index);
}
</script>

<style scoped>
.weekly-km-chart {
  height: 200px;
}
</style>
