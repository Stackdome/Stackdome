import type { components } from "@/api/types/openapi";

type ProjectMembership = components["schemas"]["UserProjectMembership"];

interface ProjectChipProps {
  membership: ProjectMembership;
  /** Pass true when this project is the organisation's default project */
  isDefault?: boolean;
}

export function ProjectChip({ membership, isDefault }: ProjectChipProps) {
  const showDefault = isDefault ?? membership.default_project;
  // Collapse to badge-only when the project is literally named "default"
  // (matches the invite-dialog precedent)
  const collapseName = showDefault && membership.project_name === "default";

  return (
    <span className="inline-flex items-center gap-1 px-2 py-px rounded border border-border bg-card text-[11px] font-mono">
      {showDefault && (
        <span className="px-1 py-px rounded text-[9px] uppercase tracking-wider text-fg-2 bg-foreground/5 border border-border">
          default
        </span>
      )}
      {!collapseName && (
        <span className="text-foreground">{membership.project_name}</span>
      )}
      {membership.role && (
        <span className="text-muted-foreground">{membership.role}</span>
      )}
    </span>
  );
}
