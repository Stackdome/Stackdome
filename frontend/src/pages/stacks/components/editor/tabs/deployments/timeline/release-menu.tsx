import { MoreVertical } from "lucide-react";
import { DropdownMenu, DropdownMenuTrigger, DropdownMenuContent, DropdownMenuItem } from "@/components/ui/dropdown-menu";
import { Button } from "@/components/ui/button";
import type { StackRelease } from "@/api/releases";
import { ReleaseState } from "../release-states";

export interface ReleaseMenuProps {
  release: StackRelease;
  onCancel: (id: string) => void;
}

export function ReleaseMenu({ release, onCancel }: ReleaseMenuProps) {
  // Only a Pending release can be cancelled — once InProgress the backend
  // rejects it (the rollout is already applied to the cluster). Cancel is the
  // menu's only item, so there is nothing to open otherwise.
  if (release.state !== ReleaseState.Pending || !release.id) return null;

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" size="icon" aria-label="Release actions" onClick={(e) => e.stopPropagation()}>
          <MoreVertical className="h-4 w-4" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="min-w-[180px]">
        <DropdownMenuItem variant="destructive" onClick={() => onCancel(release.id!)}>Cancel release</DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
