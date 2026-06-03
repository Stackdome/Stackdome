import { MoreHorizontal } from "lucide-react";
import { useNavigate } from "react-router-dom";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import type { Team } from "@/api/teams";

interface TeamRowMenuProps {
  team: Team;
  onRename?: (team: Team) => void;
  onDelete?: (team: Team) => void;
}

export function TeamRowMenu({ team, onRename, onDelete }: TeamRowMenuProps) {
  const navigate = useNavigate();

  function handleManageMembers() {
    void navigate("/settings/teams/" + encodeURIComponent(team.name));
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" size="icon" aria-label="Team actions">
          <MoreHorizontal />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="min-w-[180px]">
        <DropdownMenuItem onSelect={handleManageMembers}>
          Manage members
        </DropdownMenuItem>
        <DropdownMenuItem
          disabled={team.default_team}
          onSelect={() => {
            if (!team.default_team && onRename) onRename(team);
          }}
        >
          Rename
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem
          disabled={team.default_team}
          className="text-destructive focus:text-destructive"
          onSelect={() => {
            if (!team.default_team && onDelete) onDelete(team);
          }}
        >
          Delete
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
