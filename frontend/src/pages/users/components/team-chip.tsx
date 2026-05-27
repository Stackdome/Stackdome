import type { components } from "@/api/types/openapi";

type TeamMembership = components["schemas"]["UserTeamMembership"];

interface TeamChipProps {
  membership: TeamMembership;
  /** Pass true when this team is the organisation's default team */
  isDefault?: boolean;
}

export function TeamChip({ membership, isDefault }: TeamChipProps) {
  const showDefault = isDefault ?? membership.default_team;
  // Collapse to badge-only when the team is literally named "default"
  // (matches the invite-dialog precedent)
  const collapseName = showDefault && membership.team_name === "default";

  return (
    <span className="inline-flex items-center gap-1 px-2 py-px rounded border border-border bg-card text-[11px] font-mono">
      {showDefault && (
        <span className="px-1 py-px rounded text-[9px] uppercase tracking-wider text-brand bg-brand-bg border border-brand-border">
          default
        </span>
      )}
      {!collapseName && (
        <span className="text-foreground">{membership.team_name}</span>
      )}
      {membership.role && (
        <span className="text-muted-foreground">{membership.role}</span>
      )}
    </span>
  );
}
