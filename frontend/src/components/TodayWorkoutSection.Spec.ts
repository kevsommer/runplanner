import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import PrimeVue from "primevue/config";
import { api, router } from "@/tests/mocks";
import TodayWorkoutSection from "./TodayWorkoutSection.vue";
import type { Plan } from "./TrainingPlanCard.vue";
import type { Workout } from "./WorkoutCard.vue";

const TODAY = "2026-03-10";

function mountSection(plan: Plan | null) {
  return mount(TodayWorkoutSection, {
    props: { plan },
    global: {
      plugins: [PrimeVue],
    },
  });
}

function makeWorkout(overrides: Partial<Workout> = {}): Workout {
  return {
    id: "w-1",
    planId: "plan-1",
    runType: "easy_run",
    day: `${TODAY}T00:00:00Z`,
    description: "Morning jog",
    notes: "",
    status: "pending",
    distance: 8,
    ...overrides,
  };
}

const basePlan: Plan = {
  id: "plan-1",
  name: "Marathon Training",
  startDate: "2026-01-01",
  endDate: "2026-04-01",
  weeks: 13,
};

beforeEach(() => {
  vi.useFakeTimers();
  vi.setSystemTime(new Date(TODAY));
  api.get.mockReset();
  api.put.mockReset();
  router.push.mockReset();
});

afterEach(() => {
  vi.useRealTimers();
});

describe("TodayWorkoutSection — no plan", () => {
  it("renders nothing when plan is null", () => {
    api.get.mockResolvedValue({ data: { workouts: [] } });
    const wrapper = mountSection(null);
    expect(wrapper.html()).toBe("<!--v-if-->");
  });

  it("does not call the API when plan is null", () => {
    const wrapper = mountSection(null);
    expect(api.get).not.toHaveBeenCalled();
  });
});

describe("TodayWorkoutSection — with plan", () => {
  it("renders the section heading", async () => {
    api.get.mockResolvedValue({ data: { workouts: [] } });
    const wrapper = mountSection(basePlan);
    await flushPromises();
    expect(wrapper.text()).toContain("Today's Workout");
  });

  it("renders the plan name", async () => {
    api.get.mockResolvedValue({ data: { workouts: [] } });
    const wrapper = mountSection(basePlan);
    await flushPromises();
    expect(wrapper.text()).toContain("Marathon Training");
  });

  it("calls GET /plans/:id/workouts on mount", async () => {
    api.get.mockResolvedValue({ data: { workouts: [] } });
    mountSection(basePlan);
    await flushPromises();
    expect(api.get).toHaveBeenCalledWith("/plans/plan-1/workouts");
  });

  it("shows rest day message when no workouts today", async () => {
    api.get.mockResolvedValue({ data: { workouts: [] } });
    const wrapper = mountSection(basePlan);
    await flushPromises();
    expect(wrapper.text()).toContain("Rest day");
  });

  it("shows today's workouts", async () => {
    const workout = makeWorkout();
    api.get.mockResolvedValue({ data: { workouts: [workout] } });
    const wrapper = mountSection(basePlan);
    await flushPromises();
    expect(wrapper.text()).toContain("Easy Run");
    expect(wrapper.text()).toContain("8 km");
  });

  it("does not show workouts from other days", async () => {
    const other = makeWorkout({ id: "w-2", day: "2026-03-09T00:00:00Z", runType: "long_run", distance: 20 });
    api.get.mockResolvedValue({ data: { workouts: [other] } });
    const wrapper = mountSection(basePlan);
    await flushPromises();
    expect(wrapper.text()).toContain("Rest day");
    expect(wrapper.text()).not.toContain("Long Run");
  });

  it("shows multiple workouts scheduled for today", async () => {
    const w1 = makeWorkout({ id: "w-1", runType: "easy_run" });
    const w2 = makeWorkout({ id: "w-2", runType: "strength_training", distance: 0 });
    api.get.mockResolvedValue({ data: { workouts: [w1, w2] } });
    const wrapper = mountSection(basePlan);
    await flushPromises();
    expect(wrapper.text()).toContain("Easy Run");
    expect(wrapper.text()).toContain("Strength Training");
  });

  it("navigates to plan on 'View plan' click", async () => {
    api.get.mockResolvedValue({ data: { workouts: [] } });
    const wrapper = mountSection(basePlan);
    await flushPromises();
    await wrapper.find("button").trigger("click");
    expect(router.push).toHaveBeenCalledWith("/plans/plan-1");
  });

  it("re-fetches workouts after a workout is updated", async () => {
    const workout = makeWorkout();
    api.get.mockResolvedValue({ data: { workouts: [workout] } });
    api.put.mockResolvedValue({});
    const wrapper = mountSection(basePlan);
    await flushPromises();

    // Trigger complete on the WorkoutCard
    await wrapper.find("[data-test='complete-button']").trigger("click");
    await flushPromises();

    expect(api.get).toHaveBeenCalledTimes(2);
    expect(api.get).toHaveBeenNthCalledWith(2, "/plans/plan-1/workouts");
  });

  it("navigates to plan when edit is triggered on a workout", async () => {
    const workout = makeWorkout();
    api.get.mockResolvedValue({ data: { workouts: [workout] } });
    const wrapper = mountSection(basePlan);
    await flushPromises();

    await wrapper.find(".pi-pencil").trigger("click");
    expect(router.push).toHaveBeenCalledWith("/plans/plan-1");
  });
});
