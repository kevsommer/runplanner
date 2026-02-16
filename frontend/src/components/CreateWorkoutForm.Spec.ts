import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import PrimeVue from 'primevue/config';
import CreateWorkoutForm from "./CreateWorkoutForm.vue";

const postMock = vi.fn();
vi.mock("@/api", () => ({
  api: { post: (...args: unknown[]) => postMock(...args) },
}));

function mountForm(props: { planId: string; initialDate?: string }) {
  return mount(CreateWorkoutForm, {
    props,
    global: {
      plugins: [PrimeVue],
      stubs: {
        DatePicker: true,
      },
    },
  });
}

beforeEach(() => {
  postMock.mockReset();
});

describe("CreateWorkoutForm", () => {
  it("submits correct payload with defaults", async () => {
    postMock.mockResolvedValue({});
    const wrapper = mountForm({ planId: "plan-1", initialDate: "2026-03-15" });

    await wrapper.find("form").trigger("submit");
    await flushPromises();

    expect(postMock).toHaveBeenCalledWith("/workouts/", {
      planId: "plan-1",
      runType: "easy_run",
      day: "2026-03-15",
      description: "",
      distance: 5,
    });
  });

  it("sets distance to 0 when runType is strength_training", async () => {
    postMock.mockResolvedValue({});
    const wrapper = mountForm({ planId: "plan-1", initialDate: "2026-03-15" });

    const fields = wrapper.findComponent({ name: "WorkoutFormFields" });
    (fields.props("form") as Record<string, unknown>).runType =
      "strength_training";

    await wrapper.find("form").trigger("submit");
    await flushPromises();

    expect(postMock).toHaveBeenCalledWith(
      "/workouts/",
      expect.objectContaining({ distance: 0 }),
    );
  });

  it("emits 'created' on successful submission", async () => {
    postMock.mockResolvedValue({});
    const wrapper = mountForm({ planId: "plan-1", initialDate: "2026-03-15" });

    await wrapper.find("form").trigger("submit");
    await flushPromises();

    expect(wrapper.emitted("created")).toHaveLength(1);
  });

  it("shows error message on API failure", async () => {
    postMock.mockRejectedValue(new Error("Server error"));
    const wrapper = mountForm({ planId: "plan-1", initialDate: "2026-03-15" });

    await wrapper.find("form").trigger("submit");
    await flushPromises();

    const errorMsg = wrapper.find("[data-test='error-message']");
    expect(errorMsg.exists()).toBe(true);
    expect(errorMsg.text()).toContain("Failed to create workout");
    expect(wrapper.emitted("created")).toBeUndefined();
  });

  it("emits 'cancel' when cancel button is clicked", async () => {
    const wrapper = mountForm({ planId: "plan-1" });

    await wrapper.find("[data-test='cancel-button']").trigger("click");

    expect(wrapper.emitted("cancel")).toHaveLength(1);
  });

  it("hides day field when initialDate is provided", () => {
    const wrapper = mountForm({ planId: "plan-1", initialDate: "2026-03-15" });

    expect(wrapper.find("[data-test='day-field']").exists()).toBe(false);
  });

  it("shows day field when no initialDate is provided", () => {
    const wrapper = mountForm({ planId: "plan-1" });

    expect(wrapper.find("[data-test='day-field']").exists()).toBe(true);
  });
});
