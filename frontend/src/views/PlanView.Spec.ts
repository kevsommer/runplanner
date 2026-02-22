import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import PrimeVue from "primevue/config";
import ToastService from "primevue/toastservice";
import { api, router } from "@/tests/mocks";
import PlanView from "./PlanView.vue";
import type { Workout } from "@/components/WorkoutCard.vue";

function makeWorkout(overrides: Partial<Workout> = {}): Workout {
  return {
    id: "w-1",
    planId: "plan-1",
    runType: "easy_run",
    day: "2026-02-16",
    description: "Morning jog",
    notes: "",
    status: "pending",
    distance: 8,
    ...overrides,
  };
}

function makePlan(overrides: Record<string, unknown> = {}) {
  return {
    id: "plan-1",
    name: "Marathon Training",
    startDate: "2026-02-09",
    endDate: "2026-03-08",
    weeks: 4,
    weeksSummary: [
      {
        number: 1,
        days: [
          { date: "2026-02-09", dayName: "Monday", workouts: [makeWorkout({ day: "2026-02-09" })] },
          { date: "2026-02-10", dayName: "Tuesday", workouts: [] },
          { date: "2026-02-11", dayName: "Wednesday", workouts: [] },
          { date: "2026-02-12", dayName: "Thursday", workouts: [] },
          { date: "2026-02-13", dayName: "Friday", workouts: [] },
          { date: "2026-02-14", dayName: "Saturday", workouts: [] },
          { date: "2026-02-15", dayName: "Sunday", workouts: [] },
        ],
        plannedKm: 30,
        doneKm: 0,
        allDone: false,
      },
      {
        number: 2,
        days: [
          { date: "2026-02-16", dayName: "Monday", workouts: [makeWorkout({ id: "w-2", day: "2026-02-16" })] },
          { date: "2026-02-17", dayName: "Tuesday", workouts: [] },
          { date: "2026-02-18", dayName: "Wednesday", workouts: [] },
          { date: "2026-02-19", dayName: "Thursday", workouts: [] },
          { date: "2026-02-20", dayName: "Friday", workouts: [] },
          { date: "2026-02-21", dayName: "Saturday", workouts: [] },
          { date: "2026-02-22", dayName: "Sunday", workouts: [] },
        ],
        plannedKm: 35,
        doneKm: 10,
        allDone: false,
      },
      {
        number: 3,
        days: [
          { date: "2026-02-23", dayName: "Monday", workouts: [] },
          { date: "2026-02-24", dayName: "Tuesday", workouts: [] },
          { date: "2026-02-25", dayName: "Wednesday", workouts: [] },
          { date: "2026-02-26", dayName: "Thursday", workouts: [] },
          { date: "2026-02-27", dayName: "Friday", workouts: [] },
          { date: "2026-02-28", dayName: "Saturday", workouts: [] },
          { date: "2026-03-01", dayName: "Sunday", workouts: [] },
        ],
        plannedKm: 40,
        doneKm: 0,
        allDone: false,
      },
      {
        number: 4,
        days: [
          { date: "2026-03-02", dayName: "Monday", workouts: [] },
          { date: "2026-03-03", dayName: "Tuesday", workouts: [] },
          { date: "2026-03-04", dayName: "Wednesday", workouts: [] },
          { date: "2026-03-05", dayName: "Thursday", workouts: [] },
          { date: "2026-03-06", dayName: "Friday", workouts: [] },
          { date: "2026-03-07", dayName: "Saturday", workouts: [] },
          { date: "2026-03-08", dayName: "Sunday", workouts: [makeWorkout({ id: "w-3", day: "2026-03-08", status: "completed" })] },
        ],
        plannedKm: 20,
        doneKm: 20,
        allDone: true,
      },
    ],
    ...overrides,
  };
}

function mountView() {
  return mount(PlanView, {
    global: {
      plugins: [PrimeVue, ToastService],
      stubs: {
        DayCard: true,
      },
    },
  });
}

async function mountWithPlan(planOverrides: Record<string, unknown> = {}) {
  const plan = makePlan(planOverrides);
  api.get.mockResolvedValue({ data: { plan } });
  const wrapper = mountView();
  await flushPromises();
  return { wrapper, plan };
}

beforeEach(() => {
  vi.useFakeTimers();
  vi.setSystemTime(new Date("2026-02-17"));
  api.get.mockReset();
  router.push.mockReset();
  router.currentRoute.value = { params: { id: "plan-1" } };
});

afterEach(() => {
  vi.useRealTimers();
});

describe("PlanView", () => {
  describe("initial load", () => {
    it("fetches the training plan on mount", async () => {
      await mountWithPlan();
      expect(api.get).toHaveBeenCalledWith("/plans/plan-1");
    });

    it("renders the plan name and date range", async () => {
      const { wrapper } = await mountWithPlan();
      expect(wrapper.text()).toContain("Marathon Training");
      expect(wrapper.text()).toContain("4 weeks");
    });

    it("selects the current week on initial load", async () => {
      // 2026-02-17 is in week 2 (index 1) since plan starts 2026-02-09
      const { wrapper } = await mountWithPlan();
      expect(wrapper.text()).toContain("Week 2");
    });

    it("selects first week when today is before plan start", async () => {
      vi.setSystemTime(new Date("2026-01-01"));
      const { wrapper } = await mountWithPlan();
      expect(wrapper.text()).toContain("Week 1");
    });

    it("selects last week when today is after plan end", async () => {
      vi.setSystemTime(new Date("2026-04-01"));
      const { wrapper } = await mountWithPlan();
      expect(wrapper.text()).toContain("Week 4");
    });
  });

  describe("week navigation", () => {
    it("navigates to next week on right arrow click", async () => {
      vi.setSystemTime(new Date("2026-02-10")); // week 1
      const { wrapper } = await mountWithPlan();
      expect(wrapper.text()).toContain("Week 1");

      await wrapper.findAll("button").find(b => b.find(".pi-chevron-right").exists())!.trigger("click");
      await flushPromises();

      expect(wrapper.text()).toContain("Week 2");
    });

    it("navigates to previous week on left arrow click", async () => {
      // Starts on week 2 (current week)
      const { wrapper } = await mountWithPlan();
      expect(wrapper.text()).toContain("Week 2");

      await wrapper.findAll("button").find(b => b.find(".pi-chevron-left").exists())!.trigger("click");
      await flushPromises();

      expect(wrapper.text()).toContain("Week 1");
    });

    it("does not go before first week", async () => {
      vi.setSystemTime(new Date("2026-02-10")); // week 1
      const { wrapper } = await mountWithPlan();
      expect(wrapper.text()).toContain("Week 1");

      await wrapper.findAll("button").find(b => b.find(".pi-chevron-left").exists())!.trigger("click");
      await flushPromises();

      expect(wrapper.text()).toContain("Week 1");
    });

    it("does not go past last week", async () => {
      vi.setSystemTime(new Date("2026-04-01")); // after plan end, so last week
      const { wrapper } = await mountWithPlan();
      expect(wrapper.text()).toContain("Week 4");

      await wrapper.findAll("button").find(b => b.find(".pi-chevron-right").exists())!.trigger("click");
      await flushPromises();

      expect(wrapper.text()).toContain("Week 4");
    });
  });

  describe("week summary display", () => {
    it("shows planned and done kilometers", async () => {
      const { wrapper } = await mountWithPlan();
      // Week 2: 10 / 35 km
      expect(wrapper.text()).toContain("10 / 35 km");
    });

    it("shows total done and planned km across all weeks in the header", async () => {
      const { wrapper } = await mountWithPlan();
      // totalDoneKm = 0 + 10 + 0 + 20 = 30, totalPlannedKm = 30 + 35 + 40 + 20 = 125
      expect(wrapper.text()).toContain("30 / 125 km");
    });

    it("shows total km of 0 when plan has no workouts", async () => {
      const { wrapper } = await mountWithPlan({
        weeksSummary: [
          { number: 1, plannedKm: 0, doneKm: 0, allDone: false, days: [] },
        ],
        weeks: 1,
      });
      expect(wrapper.text()).toContain("0 / 0 km");
    });

    it("shows Current badge on current week", async () => {
      const { wrapper } = await mountWithPlan();
      expect(wrapper.text()).toContain("Current");
    });

    it("does not show Current badge on non-current week", async () => {
      vi.setSystemTime(new Date("2026-02-10")); // week 1
      const { wrapper } = await mountWithPlan();
      // Navigate to week 2
      await wrapper.findAll("button").find(b => b.find(".pi-chevron-right").exists())!.trigger("click");
      await flushPromises();

      // We're viewing week 2 but current week is 1
      expect(wrapper.text()).toContain("Week 2");
      expect(wrapper.text()).not.toContain("Current");
    });

    it("shows check icon when all workouts are done", async () => {
      vi.setSystemTime(new Date("2026-04-01")); // last week (allDone: true)
      const { wrapper } = await mountWithPlan();
      expect(wrapper.find(".pi-check-circle").exists()).toBe(true);
    });
  });

  describe("data refetch", () => {
    it("refetches plan when DayCard emits workoutCreated", async () => {
      const { wrapper } = await mountWithPlan();
      api.get.mockResolvedValue({ data: { plan: makePlan() } });

      const dayCard = wrapper.findComponent({ name: "DayCard" });
      await dayCard.vm.$emit("workoutCreated");
      await flushPromises();

      expect(api.get).toHaveBeenCalledTimes(2);
    });

    it("refetches plan when DayCard emits workoutUpdated", async () => {
      const { wrapper } = await mountWithPlan();
      api.get.mockResolvedValue({ data: { plan: makePlan() } });

      const dayCard = wrapper.findComponent({ name: "DayCard" });
      await dayCard.vm.$emit("workoutUpdated");
      await flushPromises();

      expect(api.get).toHaveBeenCalledTimes(2);
    });
  });
});
