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
import { useToast } from "@/components/ui/use-toast";

type ActionResult = { ok: true } | { ok: false; error: string };
type Role = "Developer" | "Viewer";

interface MemberRowMenuProps {
  membershipId: string;
  currentRole: Role | undefined;
  memberName?: string;
  onChangeRole: (membershipId: string, nextRole: Role) => Promise<ActionResult>;
  onRemove: (membershipId: string) => Promise<ActionResult>;
}

export function MemberRowMenu({
  membershipId,
  currentRole,
  memberName,
  onChangeRole,
  onRemove,
}: MemberRowMenuProps) {
  const { toast } = useToast();
  const [busy, setBusy] = useState(false);

  const nextRole: Role | null =
    currentRole === "Developer"
      ? "Viewer"
      : currentRole === "Viewer"
        ? "Developer"
        : null;

  const changeLabel =
    currentRole === "Developer" ? "Change to Viewer" : "Promote to Developer";

  async function handleChangeRole() {
    if (busy || !nextRole) return;
    setBusy(true);
    try {
      const result = await onChangeRole(membershipId, nextRole);
      if (result.ok) {
        toast({
          title: "Role updated",
          description: memberName
            ? `${memberName} is now a ${nextRole}.`
            : `Role changed to ${nextRole}.`,
          variant: "success",
        });
      } else {
        toast({
          title: "Failed to update role",
          description: result.error,
          variant: "destructive",
        });
      }
    } finally {
      setBusy(false);
    }
  }

  async function handleRemove() {
    if (busy) return;
    setBusy(true);
    try {
      const result = await onRemove(membershipId);
      if (result.ok) {
        toast({
          title: "Member removed",
          description: memberName
            ? `${memberName} has been removed from the team.`
            : "Member removed from the team.",
          variant: "success",
        });
      } else {
        toast({
          title: "Failed to remove member",
          description: result.error,
          variant: "destructive",
        });
      }
    } finally {
      setBusy(false);
    }
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" size="icon" aria-label="Member actions" disabled={busy}>
          <MoreHorizontal />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="min-w-[180px]">
        <DropdownMenuItem
          disabled={!nextRole || busy}
          onSelect={() => void handleChangeRole()}
        >
          {changeLabel}
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem
          disabled={busy}
          className="text-destructive focus:text-destructive"
          onSelect={() => void handleRemove()}
        >
          Remove from team
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
