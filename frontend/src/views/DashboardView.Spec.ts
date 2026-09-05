import { describe, it, expect, beforeEach, vi } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import PrimeVue from "primevue/config";
import ToastService from "primevue/toastservice";
import { api } from "@/tests/mocks";
import DashboardView from "./DashboardView.vue";
import type { Plan } from "@/components/TrainingPlanCard.vue";

function makePlan(overrides: Partial<Plan> = {}): Plan {
  return {
    id: "plan-1",
    name: "Marathon Training",
    startDate: "2026-02-09",
    endDate: "2026-03-08",
    weeks: 4,
    totalPlannedKm: 300,
    totalDoneKm: 100,
    archivedAt: null,
    ...overrides,
  };
}

function mountDashboard() {
  return mount(DashboardView, {
    global: {
      plugins: [PrimeVue, ToastService],
      stubs: { TodayWorkoutSection: true },
    },
  });
}

async function mountWithPlans(plans: Plan[]) {
  api.get.mockResolvedValue({ data: { plans } });
  const wrapper = mountDashboard();
  await flushPromises();
  return wrapper;
}

beforeEach(() => {
  api.get.mockReset();
  api.post.mockReset();
});

describe("DashboardView — archived plans", () => {
  it("shows unarchived plans in the main grid", async () => {
    const wrapper = await mountWithPlans([makePlan({ id: "p1", name: "Current Plan" })]);

    expect(wrapper.text()).toContain("Current Plan");
    expect(wrapper.find('[data-test="archived-panel"]').exists()).toBe(false);
  });

  it("moves archived plans into the archived panel", async () => {
    const wrapper = await mountWithPlans([
      makePlan({ id: "p1", name: "Current Plan" }),
      makePlan({ id: "p2", name: "Old Plan", archivedAt: "2026-01-01T00:00:00Z" }),
    ]);

    const panel = wrapper.find('[data-test="archived-panel"]');
    expect(panel.exists()).toBe(true);
    expect(panel.text()).toContain("Archived (1)");
    expect(panel.text()).toContain("Old Plan");
    expect(panel.text()).not.toContain("Current Plan");
  });

  it("counts every archived plan in the panel header", async () => {
    const wrapper = await mountWithPlans([
      makePlan({ id: "p1", name: "Old One", archivedAt: "2026-01-01T00:00:00Z" }),
      makePlan({ id: "p2", name: "Old Two", archivedAt: "2026-01-02T00:00:00Z" }),
    ]);

    expect(wrapper.find('[data-test="archived-panel"]').text()).toContain("Archived (2)");
  });

  it("explains the empty grid when every plan is archived", async () => {
    const wrapper = await mountWithPlans([
      makePlan({ id: "p1", name: "Old Plan", archivedAt: "2026-01-01T00:00:00Z" }),
    ]);

    expect(wrapper.text()).toContain("No current plans. Your archived plans are below.");
  });

  it("prompts to create a plan when there are none at all", async () => {
    const wrapper = await mountWithPlans([]);

    expect(wrapper.text()).toContain("No training plans yet. Create one to get started!");
    expect(wrapper.find('[data-test="archived-panel"]').exists()).toBe(false);
  });

  it("refetches plans after a card is archived", async () => {
    const wrapper = await mountWithPlans([makePlan({ id: "p1", name: "Current Plan" })]);
    expect(api.get).toHaveBeenCalledTimes(1);

    api.post.mockResolvedValue({ data: { plan: {}, activePlanId: null } });
    await wrapper.find(".archive-btn").trigger("click");
    await vi.dynamicImportSettled();
    await flushPromises();

    expect(api.post).toHaveBeenCalledWith("/plans/p1/archive", { archived: true });
    expect(api.get).toHaveBeenCalledTimes(2);
  });
});
