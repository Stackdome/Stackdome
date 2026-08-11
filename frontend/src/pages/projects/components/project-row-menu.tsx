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
import type { Project } from "@/api/projects";

interface ProjectRowMenuProps {
  project: Project;
  onRename?: (project: Project) => void;
  onDelete?: (project: Project) => void;
}

export function ProjectRowMenu({ project, onRename, onDelete }: ProjectRowMenuProps) {
  const navigate = useNavigate();

  function handleManageMembers() {
    void navigate("/settings/projects/" + encodeURIComponent(project.name));
  }

  return (
    // Non-modal: menu items open the Rename/Delete dialogs in the same tick;
    // a modal menu's body pointer-events lock would race those opens
    // (radix-ui/primitives#1836).
    <DropdownMenu modal={false}>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" size="icon" aria-label="Project actions">
          <MoreHorizontal />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="min-w-[180px]">
        <DropdownMenuItem onSelect={handleManageMembers}>
          Manage members
        </DropdownMenuItem>
        <DropdownMenuItem
          disabled={project.default_project}
          onSelect={() => {
            if (!project.default_project && onRename) onRename(project);
          }}
        >
          Rename
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem
          variant="destructive"
          disabled={project.default_project}
          onSelect={() => {
            if (!project.default_project && onDelete) onDelete(project);
          }}
        >
          Delete
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
