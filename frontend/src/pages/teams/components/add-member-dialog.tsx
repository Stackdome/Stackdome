import { useEffect, useState } from "react";
import { Check } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/ui/command";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { getCurrentOrganizationId } from "@/helpers/common";
import { listOrganizationUsers } from "@/api/organizations";
import type { User } from "@/api/organizations";

type ActionResult = { ok: true } | { ok: false; error: string };
type Role = "Developer" | "Viewer";

interface AddMemberDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onAdd: (userId: string, role: Role) => Promise<ActionResult>;
  existingMemberUserIds: string[];
}

export function AddMemberDialog({
  open,
  onOpenChange,
  onAdd,
  existingMemberUserIds,
}: AddMemberDialogProps) {
  const [users, setUsers] = useState<User[]>([]);
  const [loadingUsers, setLoadingUsers] = useState(false);
  const [selectedUserId, setSelectedUserId] = useState<string | null>(null);
  const [role, setRole] = useState<Role>("Developer");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!open) return;
    const orgId = getCurrentOrganizationId();
    if (!orgId) return;
    setLoadingUsers(true);
    listOrganizationUsers(orgId, 1, 200)
      .then((res) => setUsers(res.items ?? []))
      .catch(() => setUsers([]))
      .finally(() => setLoadingUsers(false));
  }, [open]);

  const availableUsers = users.filter(
    (u) => !existingMemberUserIds.includes(u.id ?? "")
  );

  const selectedUser = availableUsers.find((u) => u.id === selectedUserId);

  function handleClose() {
    if (busy) return;
    setSelectedUserId(null);
    setRole("Developer");
    setError(null);
    onOpenChange(false);
  }

  async function handleConfirm() {
    if (!selectedUserId || busy) return;
    setBusy(true);
    setError(null);
    try {
      const result = await onAdd(selectedUserId, role);
      if (result.ok) {
        setSelectedUserId(null);
        setRole("Developer");
        onOpenChange(false);
      } else {
        setError(result.error);
      }
    } finally {
      setBusy(false);
    }
  }

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>Add member</DialogTitle>
        </DialogHeader>

        <div className="space-y-4">
          {/* User combobox */}
          <div className="space-y-1.5">
            <p className="text-sm font-medium">User</p>
            <div className="rounded-md border border-border overflow-hidden">
              <Command>
                <CommandInput placeholder="Search users…" />
                <CommandList>
                  {loadingUsers ? (
                    <CommandEmpty>Loading…</CommandEmpty>
                  ) : availableUsers.length === 0 ? (
                    <CommandEmpty>No users to add.</CommandEmpty>
                  ) : (
                    <>
                      <CommandEmpty>No users found.</CommandEmpty>
                      <CommandGroup>
                        {availableUsers.map((u) => (
                          <CommandItem
                            key={u.id}
                            value={`${u.name ?? ""} ${u.email ?? ""}`}
                            onSelect={() => setSelectedUserId(u.id ?? null)}
                          >
                            <Check
                              className={cn(
                                "mr-2 size-4 shrink-0",
                                selectedUserId === u.id
                                  ? "opacity-100"
                                  : "opacity-0"
                              )}
                            />
                            <span className="truncate">
                              {u.name ?? u.email ?? u.id}
                            </span>
                            {u.name && u.email && (
                              <span className="ml-2 text-muted-foreground text-xs truncate">
                                {u.email}
                              </span>
                            )}
                          </CommandItem>
                        ))}
                      </CommandGroup>
                    </>
                  )}
                </CommandList>
              </Command>
            </div>
            {selectedUser && (
              <p className="text-xs text-muted-foreground">
                Selected: {selectedUser.name ?? selectedUser.email}
              </p>
            )}
          </div>

          {/* Role radio group */}
          <div className="space-y-1.5">
            <p className="text-sm font-medium">Team role</p>
            <RadioGroup
              value={role}
              onValueChange={(v) => setRole(v as Role)}
              className="flex gap-6"
            >
              <label className="flex items-center gap-2 cursor-pointer">
                <RadioGroupItem value="Developer" id="add-role-dev" />
                <span className="text-sm">Developer</span>
              </label>
              <label className="flex items-center gap-2 cursor-pointer">
                <RadioGroupItem value="Viewer" id="add-role-viewer" />
                <span className="text-sm">Viewer</span>
              </label>
            </RadioGroup>
          </div>

          {error && (
            <p className="text-sm text-destructive">{error}</p>
          )}
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={handleClose} disabled={busy}>
            Cancel
          </Button>
          <Button
            onClick={() => void handleConfirm()}
            disabled={!selectedUserId || busy}
          >
            Add member
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
