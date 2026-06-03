import type React from "react";
import { Mail } from "lucide-react";
import { TableCell, TableRow } from "@/components/ui/table";
import { TeamChip } from "./team-chip";
import type { PendingRow as PendingRowModel } from "../hooks/use-users";
import { formatRelative } from "../lib/format-relative";

interface PendingRowProps {
  row: PendingRowModel;
  actions?: React.ReactNode;
  /** Name of the default team in the org (to render star on chip) */
  defaultTeamName?: string;
}

export function PendingRow({ row, actions, defaultTeamName }: PendingRowProps) {
  // Use invite.created_at for "invited X ago" label
  const invitedAgo = formatRelative(row.invite.created_at);

  // Build a synthetic TeamMembership shape for TeamChip
  const teamMembership = row.team_name
    ? {
      team_name: row.team_name,
      role: row.role,
      default_team: defaultTeamName ? row.team_name === defaultTeamName : false,
    }
    : null;

  return (
    <TableRow className="border-b border-border hover:bg-muted/50">
      {/* User */}
      <TableCell className="py-3.5">
        <div className="flex items-center gap-3">
          <div className="h-7 w-7 rounded shrink-0 flex items-center justify-center bg-muted text-muted-foreground">
            <Mail className="h-3.5 w-3.5" />
          </div>
          <div className="min-w-0">
            <div className="flex items-center gap-2 flex-wrap">
              <span className="font-mono text-xs text-foreground truncate">{row.email}</span>
              <span className="inline-flex items-center px-2 py-px rounded border border-warn-border bg-warn-bg text-warn text-[11px] font-mono shrink-0">
                invited
              </span>
            </div>
            <div className="font-mono text-[11px] text-muted-foreground truncate mt-0.5">
              Invited by {row.invited_by ?? "—"} · {invitedAgo}
            </div>
          </div>
        </div>
      </TableCell>

      {/* Org role — en dash for pending */}
      <TableCell className="py-3.5">
        <span className="font-mono text-[11px] text-muted-foreground">–</span>
      </TableCell>

      {/* Teams */}
      <TableCell className="py-3.5">
        {teamMembership ? (
          <TeamChip
            membership={teamMembership}
            isDefault={defaultTeamName ? row.team_name === defaultTeamName : undefined}
          />
        ) : (
          <span className="font-mono text-[11px] text-muted-foreground">–</span>
        )}
      </TableCell>

      {/* Last active */}
      <TableCell className="py-3.5">
        <span className="font-mono text-[11px] text-muted-foreground">–</span>
      </TableCell>

      <TableCell className="py-3.5 text-right">{actions}</TableCell>
    </TableRow>
  );
}
