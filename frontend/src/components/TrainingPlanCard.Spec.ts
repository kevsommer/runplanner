import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import PrimeVue from "primevue/config";
import { api, confirm, router } from "@/tests/mocks";
import TrainingPlanCard, { type Plan } from "./TrainingPlanCard.vue";

function mountCard(plan: Plan, activePlanId?: string | null) {
  return mount(TrainingPlanCard, {
    props: { plan, activePlanId },
    global: {
      plugins: [PrimeVue],
    },
  });
}

const basePlan: Plan = {
  id: "plan-1",
  name: "Marathon Training",
  startDate: "2026-01-01",
  endDate: "2026-04-01",
  weeks: 13,
  totalPlannedKm: 500,
  totalDoneKm: 120,
};

beforeEach(() => {
  router.push.mockReset();
  api.delete.mockReset();
  api.post.mockReset();
  confirm.require.mockReset();
});

afterEach(() => {
  vi.useRealTimers();
});

describe("TrainingPlanCard", () => {
  it("renders plan name", () => {
    const wrapper = mountCard(basePlan);
    expect(wrapper.text()).toContain("Marathon Training");
  });

  it("renders date range", () => {
    const wrapper = mountCard(basePlan);
    expect(wrapper.text()).toContain("Jan 1");
    expect(wrapper.text()).toContain("Apr 1");
  });

  it("renders weeks", () => {
    const wrapper = mountCard(basePlan);
    expect(wrapper.text()).toContain("13 weeks");
  });

  it("renders total done and planned km", () => {
    const wrapper = mountCard(basePlan);
    expect(wrapper.text()).toContain("120 / 500 km");
  });

  it("renders 0 done km when no workouts completed", () => {
    const wrapper = mountCard({ ...basePlan, totalDoneKm: 0, totalPlannedKm: 300 });
    expect(wrapper.text()).toContain("0 / 300 km");
  });

  it("navigates to plan on click", async () => {
    const wrapper = mountCard(basePlan);
    await wrapper.find(".training-plan-card").trigger("click");
    expect(router.push).toHaveBeenCalledWith("/plans/plan-1");
  });

  it("shows current week badge when plan is active", () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-01-15"));

    const wrapper = mountCard(basePlan);
    expect(wrapper.text()).toContain("Week 3");
  });

  it("does not show current week badge for future plan", () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2025-12-01"));

    const wrapper = mountCard(basePlan);
    expect(wrapper.text()).not.toMatch(/Week \d/);
  });

  it("shows days remaining for upcoming race", () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-02-01"));

    const wrapper = mountCard(basePlan);
    expect(wrapper.text()).toContain("days until race day");
  });

  it("shows 'Plan completed' when plan is past", () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-05-01"));

    const wrapper = mountCard(basePlan);
    expect(wrapper.text()).toContain("Plan completed");
  });

  it("shows 'Race day is today!' on end date", () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-04-01"));

    const wrapper = mountCard(basePlan);
    expect(wrapper.text()).toContain("Race day is today!");
  });

  it("renders progress bar when plan has started", () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-02-01"));

    const wrapper = mountCard(basePlan);
    expect(wrapper.findComponent({ name: "ProgressBar" }).exists()).toBe(true);
  });

  it("renders progress bar at 0 for future plan", () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2025-12-01"));

    const wrapper = mountCard(basePlan);
    const progressBar = wrapper.findComponent({ name: "ProgressBar" });
    expect(progressBar.exists()).toBe(true);
    expect(progressBar.props("value")).toBe(0);
  });

  it("opens a confirm dialog when delete button is clicked", async () => {
    const wrapper = mountCard(basePlan);
    await wrapper.find(".delete-btn").trigger("click");

    expect(confirm.require).toHaveBeenCalledOnce();
    expect(confirm.require).toHaveBeenCalledWith(
      expect.objectContaining({
        message: "Are you sure you want to delete this training plan?",
        header: "Delete Training Plan",
      })
    );
  });

  it("calls api.delete and emits deleted when confirmed", async () => {
    api.delete.mockResolvedValue({ data: { deleted: true } });

    const wrapper = mountCard(basePlan);
    await wrapper.find(".delete-btn").trigger("click");
    const { accept } = confirm.require.mock.calls[0][0];
    accept();
    await flushPromises();

    expect(api.delete).toHaveBeenCalledWith("/plans/plan-1");
    expect(wrapper.emitted("deleted")).toBeTruthy();
  });

  it("does not call api.delete when confirm is cancelled", async () => {
    const wrapper = mountCard(basePlan);
    await wrapper.find(".delete-btn").trigger("click");
    const { reject } = confirm.require.mock.calls[0][0];
    if (reject) reject();
    await flushPromises();

    expect(api.delete).not.toHaveBeenCalled();
    expect(wrapper.emitted("deleted")).toBeFalsy();
  });

  it("does not navigate when delete button is clicked", async () => {
    const wrapper = mountCard(basePlan);
    await wrapper.find(".delete-btn").trigger("click");
    expect(router.push).not.toHaveBeenCalled();
  });
});

describe("TrainingPlanCard — active plan selection", () => {
  it("shows 'Active' badge when activePlanId matches plan id", () => {
    const wrapper = mountCard(basePlan, "plan-1");
    expect(wrapper.text()).toContain("Active");
  });

  it("does not show 'Active' badge when activePlanId does not match", () => {
    const wrapper = mountCard(basePlan, "other-plan");
    expect(wrapper.text()).not.toContain("Active");
  });

  it("does not show 'Active' badge when activePlanId is null", () => {
    const wrapper = mountCard(basePlan, null);
    expect(wrapper.text()).not.toContain("Active");
  });

  it("activate button has 'Set as active plan' label when not active", () => {
    const wrapper = mountCard(basePlan, null);
    const btn = wrapper.find(".activate-btn");
    expect(btn.attributes("aria-label")).toBe("Set as active plan");
  });

  it("activate button has 'Remove active plan' label when active", () => {
    const wrapper = mountCard(basePlan, "plan-1");
    const btn = wrapper.find(".activate-btn");
    expect(btn.attributes("aria-label")).toBe("Remove active plan");
  });

  it("calls api.post and emits activated with plan id on activation", async () => {
    api.post.mockResolvedValue({ data: { activePlanId: "plan-1" } });

    const wrapper = mountCard(basePlan, null);
    await wrapper.find(".activate-btn").trigger("click");
    await vi.dynamicImportSettled();

    expect(api.post).toHaveBeenCalledWith("/plans/plan-1/activate");
    expect(wrapper.emitted("activated")).toBeTruthy();
    expect(wrapper.emitted("activated")![0]).toEqual(["plan-1"]);
  });

  it("calls api.post and emits activated with null on deactivation", async () => {
    api.post.mockResolvedValue({ data: { activePlanId: null } });

    const wrapper = mountCard(basePlan, "plan-1");
    await wrapper.find(".activate-btn").trigger("click");
    await vi.dynamicImportSettled();

    expect(api.post).toHaveBeenCalledWith("/plans/plan-1/activate");
    expect(wrapper.emitted("activated")![0]).toEqual([null]);
  });

  it("does not navigate when activate button is clicked", async () => {
    api.post.mockResolvedValue({ data: { activePlanId: "plan-1" } });

    const wrapper = mountCard(basePlan, null);
    await wrapper.find(".activate-btn").trigger("click");
    expect(router.push).not.toHaveBeenCalled();
  });
});
