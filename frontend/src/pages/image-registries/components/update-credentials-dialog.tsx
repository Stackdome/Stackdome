import { useEffect, useState } from "react";
import {
  Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { FieldShell } from "@/components/branded";
import { useToast } from "@/components/ui/use-toast";
import { updateRegistryCredential, type RegistryCredential } from "@/api/registry-credentials";
import { getErrorMessage } from "@/api/client";
import { getCurrentOrganizationId } from "@/helpers/common";
import { rotateRegistrySchema } from "../lib/form-schemas";
import { providerIdForHost, REGISTRY_PROVIDERS } from "../lib/providers";
import { ProviderLogo } from "./provider-logo";

interface UpdateCredentialsDialogProps {
  credential: RegistryCredential | null;
  onOpenChange: (open: boolean) => void;
  /** Fired after a successful rotation so the page can refresh the list. */
  onUpdated: () => void;
}

export function UpdateCredentialsDialog({ credential, onOpenChange, onUpdated }: UpdateCredentialsDialogProps) {
  const { toast } = useToast();
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [saving, setSaving] = useState(false);
  const [fieldErrors, setFieldErrors] = useState<{ username?: string; password?: string }>({});

  const credentialId = credential?.id;
  useEffect(() => {
    if (credentialId == null) return;
    // Username is readable and prefilled; password is write-only, never prefilled.
    setUsername(credential?.username ?? "");
    setPassword("");
    setFieldErrors({});
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [credentialId]);

  const submit = async () => {
    const parsed = rotateRegistrySchema.safeParse({ username, password });
    if (!parsed.success) {
      const flat = parsed.error.flatten().fieldErrors;
      setFieldErrors({ username: flat.username?.[0], password: flat.password?.[0] });
      return;
    }
    setFieldErrors({});
    const orgId = getCurrentOrganizationId();
    if (!orgId || !credential?.id) {
      toast({
        title: "Couldn't update credentials",
        description: !orgId ? "No organization selected." : "Registry is no longer available.",
        variant: "destructive",
      });
      return;
    }
    setSaving(true);
    try {
      await updateRegistryCredential(orgId, credential.id, {
        host: credential.host,
        purpose: credential.purpose,
        username: parsed.data.username,
        password: parsed.data.password,
      });
      toast({ title: "Credentials updated" });
      onOpenChange(false);
      onUpdated();
    } catch (e) {
      toast({ title: "Couldn't update credentials", description: getErrorMessage(e), variant: "destructive" });
    } finally {
      setSaving(false);
    }
  };

  const providerId = providerIdForHost(credential?.host);
  const providerLabel = REGISTRY_PROVIDERS.find((p) => p.id === providerId)?.label ?? "Registry";

  return (
    <Dialog open={credential != null} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[440px]">
        <DialogHeader>
          <DialogTitle>Update credentials</DialogTitle>
          <DialogDescription>
            Replaces the stored login for this registry. Builds pick up the new credentials on their next run.
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-4">
          {credential && (
            <div className="flex items-center gap-3 rounded-lg border border-border bg-muted/30 px-3 py-2">
              <ProviderLogo providerId={providerId} className="h-5 w-5 shrink-0" />
              <div className="min-w-0">
                <p className="text-sm font-medium text-foreground">{providerLabel}</p>
                <p className="truncate font-mono text-[11.5px] text-fg-muted">{credential.host}</p>
              </div>
            </div>
          )}
          <FieldShell label="Username" htmlFor="rotate-username" required error={fieldErrors.username}>
            <Input
              id="rotate-username"
              autoComplete="off"
              value={username}
              onChange={(e) => {
                setUsername(e.target.value);
                setFieldErrors((prev) => ({ ...prev, username: undefined }));
              }}
              aria-invalid={!!fieldErrors.username}
            />
          </FieldShell>
          <FieldShell label="Password" htmlFor="rotate-password" required error={fieldErrors.password}>
            <Input
              id="rotate-password"
              type="password"
              autoComplete="new-password"
              value={password}
              onChange={(e) => {
                setPassword(e.target.value);
                setFieldErrors((prev) => ({ ...prev, password: undefined }));
              }}
              aria-invalid={!!fieldErrors.password}
            />
          </FieldShell>
        </div>
        <DialogFooter>
          <Button variant="ghost" onClick={() => onOpenChange(false)} disabled={saving}>
            Cancel
          </Button>
          <Button onClick={() => void submit()} disabled={saving}>
            Update credentials
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
