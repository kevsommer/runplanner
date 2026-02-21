import { describe, it, expect, beforeEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import PrimeVue from 'primevue/config';
import { ToastService } from "primevue";
import { api, toast } from "@/tests/mocks";
import CreateWorkoutForm from "./CreateWorkoutForm.vue";

function mountForm(props: { planId: string; initialDate?: string }) {
  return mount(CreateWorkoutForm, {
    props,
    global: {
      plugins: [PrimeVue, ToastService],
      stubs: {
        DatePicker: true,
      },
    },
  });
}

beforeEach(() => {
  api.post.mockReset();
  toast.add.mockReset();
});

describe("CreateWorkoutForm", () => {
  it("submits correct payload with defaults", async () => {
    api.post.mockResolvedValue({});
    const wrapper = mountForm({ planId: "plan-1", initialDate: "2026-03-15" });

    await wrapper.find("form").trigger("submit");
    await flushPromises();

    expect(api.post).toHaveBeenCalledWith("/workouts", {
      planId: "plan-1",
      runType: "easy_run",
      day: "2026-03-15",
      description: "",
      distance: 5,
    });
  });

  it("sets distance to 0 when runType is strength_training", async () => {
    api.post.mockResolvedValue({});
    const wrapper = mountForm({ planId: "plan-1", initialDate: "2026-03-15" });

    const fields = wrapper.findComponent({ name: "WorkoutFormFields" });
    (fields.props("form") as Record<string, unknown>).runType =
      "strength_training";

    await wrapper.find("form").trigger("submit");
    await flushPromises();

    expect(api.post).toHaveBeenCalledWith(
      "/workouts",
      expect.objectContaining({ distance: 0 }),
    );
  });

  it("emits 'created' on successful submission", async () => {
    api.post.mockResolvedValue({});
    const wrapper = mountForm({ planId: "plan-1", initialDate: "2026-03-15" });

    await wrapper.find("form").trigger("submit");
    await flushPromises();

    expect(wrapper.emitted("created")).toHaveLength(1);
  });

  it("shows error toast on API failure", async () => {
    api.post.mockRejectedValue(new Error("Server error"));
    const wrapper = mountForm({ planId: "plan-1", initialDate: "2026-03-15" });

    await wrapper.find("form").trigger("submit");
    await flushPromises();

    expect(toast.add).toHaveBeenCalledWith(
      expect.objectContaining({
        severity: "error",
        detail: "Server error",
      }),
    );
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
