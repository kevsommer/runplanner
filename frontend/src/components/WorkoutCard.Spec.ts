import { describe, it, expect, beforeEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import PrimeVue from "primevue/config";
import { api } from "@/tests/mocks";
import WorkoutCard from "./WorkoutCard.vue";
import type { Workout } from "./WorkoutCard.vue";

function makeWorkout(overrides: Partial<Workout> = {}): Workout {
  return {
    id: "w-1",
    planId: "plan-1",
    runType: "easy_run",
    day: "2026-03-15",
    description: "Morning jog",
    notes: "",
    status: "pending",
    distance: 8,
    ...overrides,
  };
}

function mountCard(workout: Workout) {
  return mount(WorkoutCard, {
    props: { workout },
    global: {
      plugins: [PrimeVue],
    },
  });
}

beforeEach(() => {
  api.put.mockReset();
});

describe("WorkoutCard skip button", () => {
  it("renders skip button", () => {
    const wrapper = mountCard(makeWorkout());
    expect(wrapper.find("[data-test='skip-button']").exists()).toBe(true);
  });

  it("does not render skip button for completed workout", () => {
    const wrapper = mountCard(makeWorkout({ status: "completed" }));
    expect(wrapper.find("[data-test='skip-button']").exists()).toBe(false);
  });

  it("does not render complete button for skipped workout", () => {
    const wrapper = mountCard(makeWorkout({ status: "skipped" }));
    expect(wrapper.find("[data-test='complete-button']").exists()).toBe(false);
  });

  it("clicking skip calls API with skipped status", async () => {
    api.put.mockResolvedValue({});
    const wrapper = mountCard(makeWorkout());

    await wrapper.find("[data-test='skip-button']").trigger("click");
    await flushPromises();

    expect(api.put).toHaveBeenCalledWith("/workouts/w-1", { status: "skipped" });
  });

  it("clicking skip on skipped workout reverts to pending", async () => {
    api.put.mockResolvedValue({});
    const wrapper = mountCard(makeWorkout({ status: "skipped" }));

    await wrapper.find("[data-test='skip-button']").trigger("click");
    await flushPromises();

    expect(api.put).toHaveBeenCalledWith("/workouts/w-1", { status: "pending" });
  });

  it("skipped workout has skipped styling", () => {
    const wrapper = mountCard(makeWorkout({ status: "skipped" }));
    expect(wrapper.find("div").classes()).toContain("workout-skipped");
  });

  it("emits updated after skip", async () => {
    api.put.mockResolvedValue({});
    const wrapper = mountCard(makeWorkout());

    await wrapper.find("[data-test='skip-button']").trigger("click");
    await flushPromises();

    expect(wrapper.emitted("updated")).toHaveLength(1);
  });
});
