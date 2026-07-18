import { MoreVertical } from "lucide-react";
import { DropdownMenu, DropdownMenuTrigger, DropdownMenuContent, DropdownMenuItem } from "@/components/ui/dropdown-menu";
import { Button } from "@/components/ui/button";
import type { StackRelease } from "@/api/releases";
import { ReleaseState } from "../release-states";

export interface ReleaseMenuProps {
  release: StackRelease;
  onRollback: (id: string) => void;
  onCancel: (id: string) => void;
  onCopyId: (id: string) => void;
}

export function ReleaseMenu({ release, onRollback, onCancel, onCopyId }: ReleaseMenuProps) {
  const id = release.id ?? "";
  const state = release.state ?? "";
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" size="icon" aria-label="Release actions" className="h-7 w-7" onClick={(e) => e.stopPropagation()}>
          <MoreVertical className="h-4 w-4" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="min-w-[180px]">
        {state === ReleaseState.Released && release.id && <DropdownMenuItem onClick={() => onRollback(release.id!)}>Rollback to this</DropdownMenuItem>}
        {/* Only a Pending release can be cancelled — once InProgress the backend
            rejects it (the rollout is already applied to the cluster). */}
        {state === ReleaseState.Pending && release.id && (
          <DropdownMenuItem variant="destructive" onClick={() => onCancel(release.id!)}>Cancel release</DropdownMenuItem>
        )}
        <DropdownMenuItem onClick={() => onCopyId(id)}>Copy release ID</DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
