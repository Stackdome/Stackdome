import { MoreHorizontal, Edit, Trash2, KeyRound } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import type { Secret } from "../types";

interface SecretListProps {
  secrets: Secret[];
  onEdit: (secret: Secret) => void;
  onDelete: (secret: Secret) => void;
  canWrite?: (projectId?: string) => boolean;
}

export function formatSecretType(type: string): string {
  switch (type) {
    case "Generic":
      return "Generic";
    case "DockerRegistry":
      return "Docker Registry";
    case "GitCredentials":
      return "Git Credentials";
    case "UsernamePassword":
      return "Username/Password";
    case "Token":
      return "Token";
    case "SSHKey":
      return "SSH Key";
    default:
      return type;
  }
}

export function SecretList({ secrets, onEdit, onDelete, canWrite }: SecretListProps) {
  if (!secrets.length) {
    return <div className="text-muted-foreground p-4">No secrets found.</div>;
  }

  return (
    <div className="divide-y divide-border">
      {secrets.map((secret) => {
        const rowCanWrite = canWrite ? canWrite(secret.project_id) : true;
        return (
          <div key={secret.id} className="flex w-full items-center gap-4 px-4 py-3">
            <div className="flex min-w-0 flex-1 items-center gap-3">
              <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg border border-border bg-card">
                <KeyRound className="h-5 w-5 shrink-0 text-muted-foreground" />
              </div>
              <div className="min-w-0">
                <p className="truncate text-[15px] font-medium text-foreground">{secret.name}</p>
                <p className="truncate font-mono text-[11.5px] text-fg-muted">
                  {secret.description || "No description"}
                </p>
              </div>
            </div>
            <span className="inline-flex items-center rounded-full border border-border px-2 py-0.5 text-xs text-muted-foreground">
              {formatSecretType(secret.type)}
            </span>
            <span className="w-[90px] text-right font-mono text-[11px] text-muted-foreground tabular-nums">
              {secret.created_at ? new Date(secret.created_at).toLocaleDateString() : "—"}
            </span>
            {rowCanWrite ? (
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button variant="ghost" size="icon" className="h-8 w-8" aria-label="Secret actions">
                    <MoreHorizontal className="h-4 w-4" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end" className="w-[160px]">
                  {/* Edit opens a dialog; deferred so the dialog mounts only after
                      the menu has closed and released its body pointer-events lock.
                      See https://github.com/radix-ui/primitives/issues/1836 */}
                  <DropdownMenuItem onSelect={() => setTimeout(() => onEdit(secret), 0)}>
                    <Edit className="h-4 w-4" />
                    Edit
                  </DropdownMenuItem>
                  <DropdownMenuItem
                    className="text-danger focus:text-danger"
                    onSelect={() => onDelete(secret)}
                  >
                    <Trash2 className="h-4 w-4" />
                    Delete
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            ) : (
              <span className="h-8 w-8" />
            )}
          </div>
        );
      })}
    </div>
  );
}
