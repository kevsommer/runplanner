import { describe, it, expect } from "vitest";
import { calcStartDate, formatDateToYYYYMMDD } from "./utils";

// Helper: parse "YYYY-MM-DD" as a local date (no timezone shift)
function d(str: string): Date {
  const [year, month, day] = str.split("-").map(Number);
  return new Date(year, month - 1, day);
}

describe("calcStartDate", () => {
  it("returns a Monday", () => {
    const result = calcStartDate(d("2025-10-05"), 10); // Sunday end date
    expect(result.getDay()).toBe(1); // 1 = Monday
  });

  it("1-week plan: start is the Monday of the race week", () => {
    // 2025-10-05 is a Sunday → Monday of that week is 2025-09-29
    const result = calcStartDate(d("2025-10-05"), 1);
    expect(formatDateToYYYYMMDD(result)).toBe("2025-09-29");
  });

  it("end date is a Monday: start is the Monday of the race week itself for 1 week", () => {
    // 2025-09-29 is a Monday
    const result = calcStartDate(d("2025-09-29"), 1);
    expect(formatDateToYYYYMMDD(result)).toBe("2025-09-29");
  });

  it("end date is a Monday: 10-week plan goes back 9 weeks", () => {
    // Monday 2025-09-29 minus 9 weeks = 2025-07-28
    const result = calcStartDate(d("2025-09-29"), 10);
    expect(formatDateToYYYYMMDD(result)).toBe("2025-07-28");
  });

  it("end date is a Wednesday: finds the Monday of that week", () => {
    // 2025-10-01 is a Wednesday → Monday is 2025-09-29
    const result = calcStartDate(d("2025-10-01"), 1);
    expect(formatDateToYYYYMMDD(result)).toBe("2025-09-29");
  });

  it("end date is a Saturday: finds the Monday of that week", () => {
    // 2025-10-04 is a Saturday → Monday is 2025-09-29
    const result = calcStartDate(d("2025-10-04"), 1);
    expect(formatDateToYYYYMMDD(result)).toBe("2025-09-29");
  });

  it("12-week plan with Friday end date", () => {
    // 2025-11-28 is a Friday → Monday of race week is 2025-11-24
    // minus 11 weeks = 2025-09-08
    const result = calcStartDate(d("2025-11-28"), 12);
    expect(formatDateToYYYYMMDD(result)).toBe("2025-09-08");
  });

  it("matches the backend behaviour for the same inputs", () => {
    // Verified against the Go StartDateFor function:
    // endDate=2026-04-26 (Sunday), weeks=18 → start=2025-12-22
    const result = calcStartDate(d("2026-04-26"), 18);
    expect(formatDateToYYYYMMDD(result)).toBe("2025-12-22");
  });
});
