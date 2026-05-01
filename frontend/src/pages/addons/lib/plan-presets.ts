export type PlanId = "basic" | "starter" | "launch" | "scale" | "performance" | "custom";

export interface PlanPreset {
  id: PlanId;
  label: string;
  cpu: string;
  memory: string;
}

export const PLAN_PRESETS: PlanPreset[] = [
  { id: "basic", label: "Basic", cpu: "0.5 CPU", memory: "1 GiB" },
  { id: "starter", label: "Starter", cpu: "1 CPU", memory: "2 GiB" },
  { id: "launch", label: "Launch", cpu: "2 CPU", memory: "8 GiB" },
  { id: "scale", label: "Scale", cpu: "4 CPU", memory: "32 GiB" },
  { id: "performance", label: "Performance", cpu: "8 CPU", memory: "64 GiB" },
  { id: "custom", label: "Custom", cpu: "Configure manually", memory: "—" },
];

export const DEFAULT_PLAN: PlanId = "basic";

interface ResourceMapping {
  cpu: { request: string; limit: string };
  memory: { request: string; limit: string };
}

const PLAN_RESOURCES: Record<Exclude<PlanId, "custom">, ResourceMapping> = {
  basic: {
    cpu: { request: "250m", limit: "500m" },
    memory: { request: "1Gi", limit: "1Gi" },
  },
  starter: {
    cpu: { request: "500m", limit: "1" },
    memory: { request: "2Gi", limit: "2Gi" },
  },
  launch: {
    cpu: { request: "1", limit: "2" },
    memory: { request: "8Gi", limit: "8Gi" },
  },
  scale: {
    cpu: { request: "2", limit: "4" },
    memory: { request: "32Gi", limit: "32Gi" },
  },
  performance: {
    cpu: { request: "4", limit: "8" },
    memory: { request: "64Gi", limit: "64Gi" },
  },
};

export function resourcesForPlan(plan: PlanId): ResourceMapping | undefined {
  if (plan === "custom") return undefined;
  return PLAN_RESOURCES[plan];
}
