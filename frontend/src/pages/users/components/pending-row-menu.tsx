import { useState } from "react";
import { MoreHorizontal } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { useToast } from "@/components/ui/use-toast";
import { useConfirm } from "@/components/branded/confirm";
import { useInvites } from "../hooks/use-invites";
import type { PendingRow as PendingRowModel } from "../hooks/use-users";

interface PendingRowMenuProps {
  row: PendingRowModel;
  onChanged: () => void;
}

export function PendingRowMenu({ row, onChanged }: PendingRowMenuProps) {
  const { resend, revoke } = useInvites();
  const { toast } = useToast();
  const confirm = useConfirm();
  const [busy, setBusy] = useState(false);

  async function handleResend() {
    if (busy) return;
    setBusy(true);
    try {
      await resend(row.id);
      toast({ title: "Invite resent", description: `Resent invite to ${row.email}.`, variant: "success" });
      onChanged();
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : "Failed to resend invite.";
      toast({ title: "Failed to resend invite", description: msg, variant: "destructive" });
    } finally {
      setBusy(false);
    }
  }

  async function handleRevoke() {
    if (busy) return;
    const ok = await confirm({
      title: `Revoke invite for ${row.email}?`,
      confirmLabel: "Revoke",
      variant: "destructive",
    });
    if (!ok) return;
    setBusy(true);
    try {
      await revoke(row.id);
      toast({ title: "Invite revoked", description: `Invite for ${row.email} has been revoked.`, variant: "success" });
      onChanged();
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : "Failed to revoke invite.";
      toast({ title: "Failed to revoke invite", description: msg, variant: "destructive" });
    } finally {
      setBusy(false);
    }
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" size="icon" aria-label="Invite actions" disabled={busy}>
          <MoreHorizontal />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        <DropdownMenuItem onSelect={handleResend}>Resend</DropdownMenuItem>
        <DropdownMenuItem variant="destructive" onSelect={handleRevoke}>Revoke</DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
