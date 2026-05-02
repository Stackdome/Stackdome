import { StatusPill as BrandedStatusPill, variantFromState } from "@/components/branded";
import type { PostgresAddonState } from "@/api/addons";

interface StatusPillProps {
  state?: PostgresAddonState;
}

export function StatusPill({ state }: StatusPillProps) {
  const label = state ?? "Unknown";
  return <BrandedStatusPill variant={variantFromState(state)}>{label}</BrandedStatusPill>;
}
