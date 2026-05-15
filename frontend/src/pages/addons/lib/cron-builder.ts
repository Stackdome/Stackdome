// Bridges a friendly schedule builder and the 6-field cron string CloudNativePG
// (robfig/cron) expects: `seconds minutes hours day-of-month month day-of-week`.

export type Frequency = "hourly" | "daily" | "weekly" | "monthly" | "custom";

export interface ScheduleParts {
  frequency: Frequency;
  /** 0-59 */
  minute: number;
  /** 0-23 */
  hour: number;
  /** 0-6, 0 = Sunday (robfig convention) */
  dayOfWeek: number;
  /** 1-31 */
  dayOfMonth: number;
  /** Raw expression, used only when frequency === "custom". */
  custom: string;
}

export const DEFAULT_SCHEDULE_PARTS: ScheduleParts = {
  frequency: "daily",
  minute: 0,
  hour: 3,
  dayOfWeek: 0,
  dayOfMonth: 1,
  custom: "0 0 3 * * *",
};

/** Trim, collapse internal whitespace, and upgrade 5-field crontab to 6-field. */
export function normalizeCron(raw: string): string {
  const expr = raw.trim().replace(/\s+/g, " ");
  if (!expr) return expr;
  if (expr.startsWith("@")) return expr;
  const fields = expr.split(" ");
  if (fields.length === 5) return `0 ${expr}`;
  return expr;
}

/**
 * True when `raw` is a usable cron arity: an `@macro`, or exactly 6
 * whitespace-separated fields once a 5-field crontab is upgraded. Catches the
 * common "missed a space" mistake before it reaches cronstrue/the backend.
 */
// One comma-list entry: *, ?, number, range, 3-letter name — optional /step.
const CRON_ATOM = String.raw`(\*|\?|\d+(-\d+)?|[A-Za-z]{3}(-[A-Za-z]{3})?)(\/\d+)?`;
const CRON_FIELD = new RegExp(`^${CRON_ATOM}(,${CRON_ATOM})*$`);

export function isValidCronArity(raw: string): boolean {
  const expr = normalizeCron(raw);
  if (!expr) return false;
  if (expr.startsWith("@")) return true;
  const fields = expr.split(" ");
  if (fields.length !== 6) return false;
  return fields.every((f) => CRON_FIELD.test(f));
}

export function buildCron(p: ScheduleParts): string {
  const { minute, hour, dayOfWeek, dayOfMonth } = p;
  switch (p.frequency) {
    case "hourly":
      return `0 ${minute} * * * *`;
    case "daily":
      return `0 ${minute} ${hour} * * *`;
    case "weekly":
      return `0 ${minute} ${hour} * * ${dayOfWeek}`;
    case "monthly":
      return `0 ${minute} ${hour} ${dayOfMonth} * *`;
    case "custom":
      return normalizeCron(p.custom);
  }
}

function intIn(token: string, min: number, max: number): number | null {
  if (!/^\d+$/.test(token)) return null;
  const n = Number(token);
  return n >= min && n <= max ? n : null;
}

/**
 * Infer builder parts from a cron string. Anything that doesn't match one of
 * the four builder shapes is returned as `custom` with the normalized raw.
 */
export function parseCron(expr: string): ScheduleParts {
  const norm = normalizeCron(expr);
  const fallback: ScheduleParts = {
    ...DEFAULT_SCHEDULE_PARTS,
    frequency: "custom",
    custom: norm || DEFAULT_SCHEDULE_PARTS.custom,
  };

  if (norm.startsWith("@")) return fallback;
  const f = norm.split(" ");
  if (f.length !== 6) return fallback;
  const [sec, mi, hr, dom, mon, dow] = f;
  if (sec !== "0" || mon !== "*") return fallback;

  const minute = intIn(mi, 0, 59);
  if (minute === null) return fallback;

  // hourly: every hour at :minute
  if (hr === "*" && dom === "*" && dow === "*") {
    return { ...DEFAULT_SCHEDULE_PARTS, frequency: "hourly", minute, custom: norm };
  }

  const hour = intIn(hr, 0, 23);
  if (hour === null) return fallback;

  // daily
  if (dom === "*" && dow === "*") {
    return { ...DEFAULT_SCHEDULE_PARTS, frequency: "daily", minute, hour, custom: norm };
  }
  // weekly
  if (dom === "*") {
    const dw = intIn(dow, 0, 6);
    if (dw === null) return fallback;
    return {
      ...DEFAULT_SCHEDULE_PARTS,
      frequency: "weekly",
      minute,
      hour,
      dayOfWeek: dw,
      custom: norm,
    };
  }
  // monthly
  if (dow === "*") {
    const dm = intIn(dom, 1, 31);
    if (dm === null) return fallback;
    return {
      ...DEFAULT_SCHEDULE_PARTS,
      frequency: "monthly",
      minute,
      hour,
      dayOfMonth: dm,
      custom: norm,
    };
  }
  return fallback;
}
