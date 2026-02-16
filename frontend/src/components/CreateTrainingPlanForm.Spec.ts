import { describe, it, expect, beforeEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import PrimeVue from "primevue/config";
import { ToastService } from "primevue";
import { api, router, toast } from "@/tests/mocks";
import CreateTrainingPlanForm from "./CreateTrainingPlanForm.vue";

function mountForm() {
  return mount(CreateTrainingPlanForm, {
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
  router.push.mockReset();
  toast.add.mockReset();
});

describe("CreateTrainingPlanForm", () => {
  it("redirects to plan page on successful creation", async () => {
    api.post.mockResolvedValue({ data: { plan: { id: "plan-42" } } });
    const wrapper = mountForm();

    await wrapper.find("form").trigger("submit");
    await flushPromises();

    expect(api.post).toHaveBeenCalledWith(
      "/plans/",
      expect.objectContaining({ name: "", weeks: 10 }),
    );
    expect(router.push).toHaveBeenCalledWith({
      name: "plan",
      params: { id: "plan-42" },
    });
  });

  it("shows error toast and does not redirect on API failure", async () => {
    api.post.mockRejectedValue(new Error("Server error"));
    const wrapper = mountForm();

    await wrapper.find("form").trigger("submit");
    await flushPromises();

    expect(router.push).not.toHaveBeenCalled();
    expect(toast.add).toHaveBeenCalledWith(
      expect.objectContaining({
        severity: "error",
        detail: "Server error",
      }),
    );
  });
});
