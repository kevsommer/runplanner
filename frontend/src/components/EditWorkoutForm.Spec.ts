import { describe, it, expect, beforeEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import PrimeVue from "primevue/config";
import { api } from "@/tests/mocks";
import EditWorkoutForm from "./EditWorkoutForm.vue";
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

function mountForm(workout: Workout) {
  return mount(EditWorkoutForm, {
    props: { workout },
    global: {
      plugins: [PrimeVue],
      stubs: {
        DatePicker: true,
      },
    },
  });
}

beforeEach(() => {
  api.put.mockReset();
});

describe("EditWorkoutForm", () => {
  it("submits correct payload from workout props", async () => {
    api.put.mockResolvedValue({});
    const workout = makeWorkout();
    const wrapper = mountForm(workout);

    await wrapper.find("form").trigger("submit");
    await flushPromises();

    expect(api.put).toHaveBeenCalledWith("/workouts/w-1", {
      runType: "easy_run",
      description: "Morning jog",
      distance: 8,
    });
  });

  it("sets distance to 0 when runType is strength_training", async () => {
    api.put.mockResolvedValue({});
    const workout = makeWorkout();
    const wrapper = mountForm(workout);

    const fields = wrapper.findComponent({ name: "WorkoutFormFields" });
    (fields.props("form") as Record<string, unknown>).runType =
      "strength_training";

    await wrapper.find("form").trigger("submit");
    await flushPromises();

    expect(api.put).toHaveBeenCalledWith(
      "/workouts/w-1",
      expect.objectContaining({ distance: 0 }),
    );
  });

  it("emits 'updated' on successful submission", async () => {
    api.put.mockResolvedValue({});
    const wrapper = mountForm(makeWorkout());

    await wrapper.find("form").trigger("submit");
    await flushPromises();

    expect(wrapper.emitted("updated")).toHaveLength(1);
  });

  it("shows error message on API failure", async () => {
    api.put.mockRejectedValue(new Error("Server error"));
    const wrapper = mountForm(makeWorkout());

    await wrapper.find("form").trigger("submit");
    await flushPromises();

    const errorMsg = wrapper.find("[data-test='error-message']");
    expect(errorMsg.exists()).toBe(true);
    expect(errorMsg.text()).toContain("Failed to update workout");
    expect(wrapper.emitted("updated")).toBeUndefined();
  });

  it("emits 'cancel' when cancel button is clicked", async () => {
    const wrapper = mountForm(makeWorkout());

    await wrapper.find("[data-test='cancel-button']").trigger("click");

    expect(wrapper.emitted("cancel")).toHaveLength(1);
  });
});
