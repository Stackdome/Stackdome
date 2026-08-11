import type React from "react";
import { TableCell, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { ProjectChip } from "./project-chip";
import type { ActiveRow } from "../hooks/use-users";
import { formatRelative } from "../lib/format-relative";

interface UserRowProps {
  row: ActiveRow;
  actions?: React.ReactNode;
  /** Name of the default project in the org (to render star on chip) */
  defaultProjectName?: string;
}

function monogram(name: string): string {
  const parts = name.trim().split(/\s+/);
  if (parts.length >= 2) {
    return ((parts[0][0] ?? "") + (parts[parts.length - 1][0] ?? "")).toUpperCase();
  }
  return (name.slice(0, 2) || "?").toUpperCase();
}

export function UserRow({ row, actions, defaultProjectName }: UserRowProps) {
  const isAdmin = row.role === "OrgAdmin";
  // last_active_at is not yet on the User model — show — until it is
  const lastActive = formatRelative((row.user as Record<string, unknown>)["last_active_at"] as string | undefined);

  return (
    <TableRow className="border-b border-border hover:bg-muted/50">
      {/* User */}
      <TableCell className="py-3.5">
        <div className="flex items-center gap-3">
          <div className="h-7 w-7 rounded shrink-0 flex items-center justify-center bg-muted text-foreground text-[10px] font-mono uppercase select-none">
            {monogram(row.name)}
          </div>
          <div className="min-w-0">
            <div className="text-sm font-medium text-foreground truncate">{row.name}</div>
            <div className="font-mono text-xs text-muted-foreground truncate">{row.email}</div>
          </div>
        </div>
      </TableCell>

      {/* Org role */}
      <TableCell className="py-3.5">
        {isAdmin ? (
          <Badge variant="secondary">{row.role}</Badge>
        ) : (
          <Badge variant="outline">{row.role ?? "—"}</Badge>
        )}
      </TableCell>

      {/* Projects */}
      <TableCell className="py-3.5">
        <div className="flex flex-wrap gap-1.5">
          {row.projects.length > 0
            ? row.projects.map((t, i) => (
              <ProjectChip
                key={t.project_id ?? i}
                membership={t}
                isDefault={defaultProjectName ? t.project_name === defaultProjectName : undefined}
              />
            ))
            : <span className="text-muted-foreground text-xs">—</span>
          }
        </div>
      </TableCell>

      {/* Last active */}
      <TableCell className="py-3.5">
        <span className="font-mono text-[11px] text-muted-foreground">{lastActive}</span>
      </TableCell>

      <TableCell className="py-3.5 text-right">{actions}</TableCell>
    </TableRow>
  );
}
