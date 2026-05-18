import type React from "react";
import { MailIcon } from "lucide-react";
import { TableCell, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import type { PendingRow as PendingRowModel } from "../hooks/use-users";

interface PendingRowProps {
  row: PendingRowModel;
  actions?: React.ReactNode;
}

export function PendingRow({ row, actions }: PendingRowProps) {
  return (
    <TableRow className="hover:bg-muted/50">
      <TableCell className="py-3">
        <div className="flex items-center gap-3">
          <div className="size-7 shrink-0 flex items-center justify-center rounded-full border border-dashed border-border bg-muted">
            <MailIcon className="size-3.5 text-muted-foreground" />
          </div>
          <div className="min-w-0">
            <div className="font-mono text-[11px] text-foreground truncate">{row.email}</div>
            {row.invited_by && (
              <div className="font-mono text-[11px] text-muted-foreground truncate">
                invited by {row.invited_by}
              </div>
            )}
          </div>
        </div>
      </TableCell>
      <TableCell className="py-3">
        <Badge
          variant="secondary"
          className="font-mono text-[11px] border border-warn-border bg-warn-bg text-warn"
        >
          invited
        </Badge>
      </TableCell>
      <TableCell className="py-3">
        {row.team_name ? (
          <span className="inline-flex h-[22px] items-center rounded-[2px] border border-border bg-muted px-2 font-mono text-[11px] text-foreground">
            {row.team_name}
            {row.role && (
              <span className="ml-1.5 text-muted-foreground">{row.role}</span>
            )}
          </span>
        ) : (
          <span className="font-mono text-[11px] text-muted-foreground">—</span>
        )}
      </TableCell>
      <TableCell className="py-3">
        {row.expires_at ? (
          <span className="font-mono text-[11px] text-warn">
            expires {new Date(row.expires_at).toLocaleDateString()}
          </span>
        ) : (
          <span className="font-mono text-[11px] text-muted-foreground">—</span>
        )}
      </TableCell>
      <TableCell className="py-3 text-right">{actions}</TableCell>
    </TableRow>
  );
}
