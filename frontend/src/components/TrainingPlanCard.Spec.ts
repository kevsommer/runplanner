import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { mount } from "@vue/test-utils";
import PrimeVue from "primevue/config";
import { api, router } from "@/tests/mocks";
import TrainingPlanCard, { type Plan } from "./TrainingPlanCard.vue";

function mountCard(plan: Plan) {
  return mount(TrainingPlanCard, {
    props: { plan },
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
};

beforeEach(() => {
  router.push.mockReset();
  api.delete.mockReset();
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

  it("calls api.delete and emits deleted when confirmed", async () => {
    vi.spyOn(window, "confirm").mockReturnValue(true);
    api.delete.mockResolvedValue({ data: { deleted: true } });

    const wrapper = mountCard(basePlan);
    await wrapper.find(".delete-btn").trigger("click");
    await vi.dynamicImportSettled();

    expect(window.confirm).toHaveBeenCalledWith("Are you sure you want to delete this training plan?");
    expect(api.delete).toHaveBeenCalledWith("/plans/plan-1");
    expect(wrapper.emitted("deleted")).toBeTruthy();
  });

  it("does not call api.delete when confirm is cancelled", async () => {
    vi.spyOn(window, "confirm").mockReturnValue(false);

    const wrapper = mountCard(basePlan);
    await wrapper.find(".delete-btn").trigger("click");

    expect(window.confirm).toHaveBeenCalled();
    expect(api.delete).not.toHaveBeenCalled();
    expect(wrapper.emitted("deleted")).toBeFalsy();
  });

  it("does not navigate when delete button is clicked", async () => {
    vi.spyOn(window, "confirm").mockReturnValue(false);

    const wrapper = mountCard(basePlan);
    await wrapper.find(".delete-btn").trigger("click");
    expect(router.push).not.toHaveBeenCalled();
  });
});
