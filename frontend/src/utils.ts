export function calcStartDate(endDate: Date, weeks: number): Date {
  const weekday = endDate.getDay(); // 0=Sun, 1=Mon, ..., 6=Sat
  const daysSinceMonday = weekday === 0 ? 6 : weekday - 1;
  const mondayOfRaceWeek = new Date(endDate);
  mondayOfRaceWeek.setDate(endDate.getDate() - daysSinceMonday);
  const mondayOfWeek1 = new Date(mondayOfRaceWeek);
  mondayOfWeek1.setDate(mondayOfRaceWeek.getDate() - (weeks - 1) * 7);
  return mondayOfWeek1;
}

export function formatDateToYYYYMMDD(date: Date): string {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}

export function formatDateISO(date: Date): string {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}

export function formatDate(dateStr: string): string {
  const date = new Date(dateStr);
  return date.toLocaleDateString("en-US", { month: "short", day: "numeric" });
}
