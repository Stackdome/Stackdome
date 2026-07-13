import { useState } from "react";
import { MoreHorizontal } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useToast } from "@/components/ui/use-toast";
import { useUserActions } from "../hooks/use-user-actions";
import { useProjectOptions } from "../hooks/use-project-options";
import type { ActiveRow } from "../hooks/use-users";

interface UserRowMenuProps {
  row: ActiveRow;
  onChanged: () => void;
}

export function UserRowMenu({ row, onChanged }: UserRowMenuProps) {
  const { promote, demote } = useUserActions();
  const { projects } = useProjectOptions();
  const { toast } = useToast();
  const [busy, setBusy] = useState(false);
  const [demoteOpen, setDemoteOpen] = useState(false);
  const [demoteProject, setDemoteProject] = useState("");
  const [demoteRole, setDemoteRole] = useState<"Developer" | "Viewer">("Developer");

  const isOrgMember = row.role === "OrgMember";
  const isOrgAdmin = row.role === "OrgAdmin";

  async function handlePromote() {
    if (busy) return;
    setBusy(true);
    try {
      const result = await promote(row.id);
      if (result.ok) {
        toast({ title: "User promoted", description: `${row.name} is now an OrgAdmin.`, variant: "success" });
        onChanged();
      } else {
        toast({ title: "Failed to promote user", description: result.error, variant: "destructive" });
      }
    } finally {
      setBusy(false);
    }
  }

  async function handleDemoteConfirm() {
    if (busy || !demoteProject) return;
    setBusy(true);
    try {
      const result = await demote(row.id, demoteProject, demoteRole);
      if (result.ok) {
        toast({ title: "User demoted", description: `${row.name} has been demoted to OrgMember.`, variant: "success" });
        setDemoteOpen(false);
        setDemoteProject("");
        onChanged();
      } else {
        toast({ title: "Failed to demote user", description: result.error, variant: "destructive" });
      }
    } finally {
      setBusy(false);
    }
  }

  async function handleCopyId() {
    try {
      await navigator.clipboard.writeText(row.id);
      toast({ title: "Copied", description: "User ID copied to clipboard." });
    } catch (err) {
      toast({
        title: "Couldn't copy ID",
        description: err instanceof Error ? err.message : "Clipboard access was denied.",
        variant: "destructive",
      });
    }
  }

  return (
    <DropdownMenu onOpenChange={(open) => { if (!open) setDemoteOpen(false); }}>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" size="icon" aria-label="User actions" disabled={busy}>
          <MoreHorizontal />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="min-w-[200px]">
        {isOrgMember && (
          <DropdownMenuItem onSelect={handlePromote}>
            Promote to OrgAdmin
          </DropdownMenuItem>
        )}
        {isOrgAdmin && !demoteOpen && (
          <DropdownMenuItem onSelect={(e) => { e.preventDefault(); setDemoteOpen(true); }}>
            Demote
          </DropdownMenuItem>
        )}
        {isOrgAdmin && demoteOpen && (
          <div className="px-2 py-1.5 space-y-2">
            <p className="text-xs text-muted-foreground font-medium">Demote to project member</p>
            <Select value={demoteProject} onValueChange={setDemoteProject}>
              <SelectTrigger size="sm" className="w-full">
                <SelectValue placeholder="Select project" />
              </SelectTrigger>
              <SelectContent>
                {projects.map((t) => (
                  <SelectItem key={t.id ?? t.name} value={t.name ?? ""}>
                    {t.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <Select value={demoteRole} onValueChange={(v) => setDemoteRole(v as "Developer" | "Viewer")}>
              <SelectTrigger size="sm" className="w-full">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="Developer">Developer</SelectItem>
                <SelectItem value="Viewer">Viewer</SelectItem>
              </SelectContent>
            </Select>
            <div className="flex gap-1.5 pt-0.5">
              <Button
                size="sm"
                className="flex-1 h-7 text-xs"
                onClick={handleDemoteConfirm}
                disabled={!demoteProject || busy}
              >
                Confirm
              </Button>
              <Button
                size="sm"
                variant="ghost"
                className="flex-1 h-7 text-xs"
                onClick={() => setDemoteOpen(false)}
              >
                Cancel
              </Button>
            </div>
          </div>
        )}
        <DropdownMenuSeparator />
        <DropdownMenuItem onSelect={handleCopyId}>Copy ID</DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
