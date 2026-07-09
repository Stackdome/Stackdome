import { StatusPill as BrandedStatusPill } from "@/components/branded";
import { statusVariant } from "@/components/branded/status-variant";
import type { PostgresAddonState } from "@/api/addons";

interface StatusPillProps {
  state?: PostgresAddonState;
}

export function StatusPill({ state }: StatusPillProps) {
  const label = state ?? "Unknown";
  return <BrandedStatusPill variant={statusVariant("addon", state)}>{label}</BrandedStatusPill>;
}
