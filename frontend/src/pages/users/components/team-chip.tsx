import { Star } from "lucide-react";
import type { components } from "@/api/types/openapi";

type TeamMembership = components["schemas"]["UserTeamMembership"];

export function TeamChip({ t }: { t: TeamMembership }) {
  return (
    <span className="inline-flex h-[22px] items-center gap-1.5 rounded-[2px] border border-border bg-muted px-2">
      {t.default_team && <Star className="size-2.5 text-brand" />}
      <span className="font-mono text-[11px] text-foreground">{t.team_name}</span>
      <span className="font-mono text-[11px] text-muted-foreground">{t.role}</span>
    </span>
  );
}
