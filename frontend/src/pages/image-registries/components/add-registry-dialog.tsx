import { useEffect, useState } from "react";
import { isAxiosError } from "axios";
import {
  Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from "@/components/ui/select";
import { FieldShell } from "@/components/branded";
import { useToast } from "@/components/ui/use-toast";
import { createRegistryCredential, type RegistryCredentialPurpose } from "@/api/registry-credentials";
import { getErrorMessage } from "@/api/client";
import { getCurrentOrganizationId } from "@/helpers/common";
import { createRegistrySchema } from "../lib/form-schemas";
import {
  REGISTRY_PROVIDERS, PURPOSE_BOTH, PURPOSE_PULL, PURPOSE_PUSH, PURPOSE_LABELS,
  type RegistryProvider,
} from "../lib/providers";
import { ProviderLogo } from "./provider-logo";

interface AddRegistryDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onCreated: () => void;
}

interface FieldErrors {
  host?: string;
  username?: string;
  password?: string;
}

export function AddRegistryDialog({ open, onOpenChange, onCreated }: AddRegistryDialogProps) {
  const { toast } = useToast();
  const [provider, setProvider] = useState<RegistryProvider | null>(null);
  const [host, setHost] = useState("");
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [purpose, setPurpose] = useState<RegistryCredentialPurpose>(PURPOSE_BOTH);
  const [saving, setSaving] = useState(false);
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({});

  useEffect(() => {
    if (!open) return;
    setProvider(null);
    setHost("");
    setUsername("");
    setPassword("");
    setPurpose(PURPOSE_BOTH);
    setFieldErrors({});
  }, [open]);

  const pickProvider = (p: RegistryProvider) => {
    setProvider(p);
    setHost(p.hostPrefill);
    setFieldErrors({});
  };

  const submit = async () => {
    const parsed = createRegistrySchema.safeParse({ host, username, password });
    if (!parsed.success) {
      const flat = parsed.error.flatten().fieldErrors;
      setFieldErrors({ host: flat.host?.[0], username: flat.username?.[0], password: flat.password?.[0] });
      return;
    }
    setFieldErrors({});
    const orgId = getCurrentOrganizationId();
    if (!orgId) {
      toast({ title: "Couldn't add registry", description: "No organization selected.", variant: "destructive" });
      return;
    }
    setSaving(true);
    try {
      await createRegistryCredential(orgId, { ...parsed.data, purpose });
      toast({ title: "Registry added" });
      onOpenChange(false);
      onCreated();
    } catch (e) {
      // 409 = host+purpose duplicate; anchor it to the host field instead of a toast.
      if (isAxiosError(e) && e.response?.status === 409) {
        setFieldErrors({ host: "Credentials for this registry and purpose already exist." });
      } else {
        toast({ title: "Couldn't add registry", description: getErrorMessage(e), variant: "destructive" });
      }
    } finally {
      setSaving(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[480px]">
        <DialogHeader>
          <DialogTitle>Add image registry</DialogTitle>
          <DialogDescription>
            Store pull/push credentials so builds can use private registries.
          </DialogDescription>
        </DialogHeader>

        {provider == null ? (
          <div className="grid grid-cols-2 gap-2">
            {REGISTRY_PROVIDERS.map((p) => (
              <button
                key={p.id}
                type="button"
                onClick={() => pickProvider(p)}
                className="flex items-center gap-3 rounded-lg border border-border bg-card px-3 py-3 text-left hover:border-brand-border hover:bg-brand-bg-hover"
              >
                <ProviderLogo providerId={p.id} className="h-5 w-5 shrink-0" />
                <span className="min-w-0">
                  <span className="block text-sm font-medium text-foreground">{p.label}</span>
                  <span className="block truncate font-mono text-[11px] text-fg-muted">
                    {p.hostPrefill || "custom host"}
                  </span>
                </span>
              </button>
            ))}
          </div>
        ) : (
          <div className="space-y-4">
            <p className="text-xs text-muted-foreground">{provider.hint}</p>
            <FieldShell label="Host" htmlFor="registry-host" required error={fieldErrors.host}>
              <Input
                id="registry-host"
                placeholder={provider.hostPlaceholder}
                value={host}
                onChange={(e) => {
                  setHost(e.target.value);
                  setFieldErrors((prev) => ({ ...prev, host: undefined }));
                }}
                aria-invalid={!!fieldErrors.host}
              />
            </FieldShell>
            <FieldShell label="Username" htmlFor="registry-username" required error={fieldErrors.username}>
              <Input
                id="registry-username"
                autoComplete="off"
                value={username}
                onChange={(e) => {
                  setUsername(e.target.value);
                  setFieldErrors((prev) => ({ ...prev, username: undefined }));
                }}
                aria-invalid={!!fieldErrors.username}
              />
            </FieldShell>
            <FieldShell label="Password" htmlFor="registry-password" required error={fieldErrors.password}>
              <Input
                id="registry-password"
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
            <FieldShell label="Purpose" htmlFor="registry-purpose">
              <Select value={purpose} onValueChange={(v) => setPurpose(v as RegistryCredentialPurpose)}>
                <SelectTrigger id="registry-purpose">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value={PURPOSE_BOTH}>{PURPOSE_LABELS[PURPOSE_BOTH]}</SelectItem>
                  <SelectItem value={PURPOSE_PULL}>{PURPOSE_LABELS[PURPOSE_PULL]}</SelectItem>
                  <SelectItem value={PURPOSE_PUSH}>{PURPOSE_LABELS[PURPOSE_PUSH]}</SelectItem>
                </SelectContent>
              </Select>
            </FieldShell>
          </div>
        )}

        <DialogFooter>
          {provider != null && (
            <Button variant="ghost" onClick={() => setProvider(null)} disabled={saving}>
              Back
            </Button>
          )}
          <Button variant="ghost" onClick={() => onOpenChange(false)} disabled={saving}>
            Cancel
          </Button>
          {provider != null && (
            <Button onClick={() => void submit()} disabled={saving}>
              Add registry
            </Button>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
