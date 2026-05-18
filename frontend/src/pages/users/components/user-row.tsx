import type React from "react";
import { TableCell, TableRow } from "@/components/ui/table";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { TeamChip } from "./team-chip";
import type { ActiveRow } from "../hooks/use-users";

interface UserRowProps {
  row: ActiveRow;
  actions?: React.ReactNode;
}

function initials(name: string): string {
  return name
    .split(" ")
    .map((n) => n[0])
    .join("")
    .slice(0, 2)
    .toUpperCase();
}

export function UserRow({ row, actions }: UserRowProps) {
  return (
    <TableRow className="hover:bg-muted/50">
      <TableCell className="py-3">
        <div className="flex items-center gap-3">
          <Avatar className="size-7 shrink-0">
            <AvatarImage src={undefined} alt={row.name} />
            <AvatarFallback className="text-[11px] font-mono">
              {initials(row.name)}
            </AvatarFallback>
          </Avatar>
          <div className="min-w-0">
            <div className="text-sm font-medium text-foreground truncate">{row.name}</div>
            <div className="font-mono text-[11px] text-muted-foreground truncate">{row.email}</div>
          </div>
        </div>
      </TableCell>
      <TableCell className="py-3">
        <Badge variant="outline" className="font-mono text-[11px]">
          {row.role ?? "—"}
        </Badge>
      </TableCell>
      <TableCell className="py-3">
        <div className="flex flex-wrap gap-1">
          {row.teams.length > 0
            ? row.teams.map((t, i) => <TeamChip key={t.team_id ?? i} membership={t} />)
            : <span className="text-muted-foreground text-xs">—</span>
          }
        </div>
      </TableCell>
      <TableCell className="py-3">
        <span className="font-mono text-[11px] text-muted-foreground">—</span>
      </TableCell>
      <TableCell className="py-3 text-right">{actions}</TableCell>
    </TableRow>
  );
}
