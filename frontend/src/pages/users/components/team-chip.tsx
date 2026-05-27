import { Star } from "lucide-react";
import type { components } from "@/api/types/openapi";

type TeamMembership = components["schemas"]["UserTeamMembership"];

interface TeamChipProps {
  membership: TeamMembership;
  /** Pass true when this team is the organisation's default team */
  isDefault?: boolean;
}

export function TeamChip({ membership, isDefault }: TeamChipProps) {
  const showStar = isDefault ?? membership.default_team;
  return (
    <span className="inline-flex items-center gap-1 px-2 py-px rounded border border-border bg-card text-[11px] font-mono">
      {showStar && (
        <Star className="h-3 w-3 text-brand fill-current shrink-0" />
      )}
      <span className="text-foreground">{membership.team_name}</span>
      {membership.role && (
        <span className="text-muted-foreground">{membership.role}</span>
      )}
    </span>
  );
}
