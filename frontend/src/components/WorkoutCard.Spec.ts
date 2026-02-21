import { describe, it, expect, beforeEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import PrimeVue from "primevue/config";
import { api, confirm } from "@/tests/mocks";
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
  api.delete.mockReset();
  confirm.require.mockReset();
});

describe("WorkoutCard delete button", () => {
  it("opens a confirm dialog when the delete button is clicked", async () => {
    const wrapper = mountCard(makeWorkout());
    await wrapper.find("[data-test='delete-button']").trigger("click");

    expect(confirm.require).toHaveBeenCalledOnce();
    expect(confirm.require).toHaveBeenCalledWith(
      expect.objectContaining({
        message: "Are you sure you want to delete this workout?",
        header: "Delete Workout",
      })
    );
  });

  it("calls api.delete when confirm is accepted", async () => {
    api.delete.mockResolvedValue({ data: {} });
    const wrapper = mountCard(makeWorkout());

    await wrapper.find("[data-test='delete-button']").trigger("click");
    const { accept } = confirm.require.mock.calls[0][0];
    accept();
    await flushPromises();

    expect(api.delete).toHaveBeenCalledWith("/workouts/w-1");
  });

  it("emits updated after successful deletion", async () => {
    api.delete.mockResolvedValue({ data: {} });
    const wrapper = mountCard(makeWorkout());

    await wrapper.find("[data-test='delete-button']").trigger("click");
    const { accept } = confirm.require.mock.calls[0][0];
    accept();
    await flushPromises();

    expect(wrapper.emitted("updated")).toHaveLength(1);
  });

  it("does not call api.delete when confirm is rejected", async () => {
    const wrapper = mountCard(makeWorkout());

    await wrapper.find("[data-test='delete-button']").trigger("click");
    const { reject } = confirm.require.mock.calls[0][0];
    if (reject) reject();
    await flushPromises();

    expect(api.delete).not.toHaveBeenCalled();
  });
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
