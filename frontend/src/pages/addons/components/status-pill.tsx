import { Badge } from "@/components/ui/badge";
import type { PostgresAddonState } from "@/api/addons";

interface StatusPillProps {
  state?: PostgresAddonState;
}

function variantFor(state?: PostgresAddonState): "default" | "secondary" | "destructive" | "outline" {
  switch (state) {
    case "Ready":
      return "default";
    case "Error":
      return "destructive";
    case "Hibernated":
    case "Fenced":
      return "outline";
    default:
      return "secondary";
  }
}

export function StatusPill({ state }: StatusPillProps) {
  const label = state ?? "Unknown";
  return (
    <Badge variant={variantFor(state)} className="text-xs font-medium">
      {label}
    </Badge>
  );
}
