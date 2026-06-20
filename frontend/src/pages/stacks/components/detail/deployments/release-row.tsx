import { MoreVertical } from "lucide-react";
import { StatusPill, variantFromState } from "@/components/branded";
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
} from "@/components/ui/dropdown-menu";
import { Button } from "@/components/ui/button";
import type { StackRelease } from "@/api/releases";
import { causeLabel, formatDuration, releaseGitSha } from "./derive";

export interface ReleaseRowProps {
  release: StackRelease;
  onViewDetails: (id: string) => void;
  onRollback: (id: string) => void;
  onCancel: (id: string) => void;
}

export function ReleaseRow({ release, onViewDetails, onRollback, onCancel }: ReleaseRowProps) {
  const id = release.id ?? "";
  const state = release.state ?? "";
  const sha = releaseGitSha(release);
  const subline =
    state === "Failed"
      ? release.message
      : state === "Released"
        ? [sha, formatDuration(release.rendered_at, release.completed_at)].filter(Boolean).join(" · ")
        : undefined;

  return (
    <div className="flex items-center justify-between gap-4 border-b border-border px-4 py-3 last:border-0">
      <div className="flex min-w-0 items-center gap-3">
        <StatusPill variant={variantFromState(state)}>{state}</StatusPill>
        <span className="font-mono text-[13px] text-foreground">#{release.sequence}</span>
        <span className="truncate text-[13px] text-muted-foreground">
          {causeLabel(release.cause)}
          {subline ? (
            <span className={state === "Failed" ? "text-danger" : ""}>
              {" "}· {subline}
            </span>
          ) : null}
        </span>
      </div>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="ghost" size="icon" aria-label="Release actions" className="h-7 w-7">
            <MoreVertical className="h-4 w-4" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end" className="min-w-[180px]">
          <DropdownMenuItem onClick={() => onViewDetails(id)}>View details</DropdownMenuItem>
          {state === "Released" && (
            <DropdownMenuItem onClick={() => onRollback(id)}>Rollback to this</DropdownMenuItem>
          )}
          {state === "Pending" && (
            <DropdownMenuItem variant="destructive" onClick={() => onCancel(id)}>
              Cancel
            </DropdownMenuItem>
          )}
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  );
}
